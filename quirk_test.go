package quirk

import (
	"database/sql"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestQuirk(t *testing.T) {
	db, err := createConnection()
	assert.Nil(t, err)
	t.Cleanup(
		func() {
			assert.Nil(
				t,
				New(db).Q(`DROP TABLE tests`).Exec(),
			)
		},
	)
	t.Run(
		"create table", func(t *testing.T) {
			assert.Nil(
				t,
				New(db).Q(
					`CREATE TABLE IF NOT EXISTS tests (
	    			id serial,
	       		name varchar(255) not null,
	       		lastname varchar(255) not null,
	       		active bool not null default false,
	       		amount float not null default 0,
	       		amount_special float not null default 0,
	       		quantity int not null default 0,
	       		roles varchar[] not null default array[]::varchar[],
	       		note text,
	       		vectors tsvector not null default '',
	       		created_at timestamp not null default current_timestamp
	    		)`,
				).Exec(),
			)
		},
	)
	t.Run(
		"insert", func(t *testing.T) {
			id := 0
			note := "go go go"
			data := Map{
				"name":           "Dominik",
				"lastname":       "Linduska",
				"active":         true,
				"amount":         999.99,
				"amount-special": 999.99,
				"quantity":       55,
				"roles":          []string{"owner", "admin"},
				"note":           sql.Null[string]{V: note, Valid: true},
				"vectors":        CreateTsVector("Dominik", "Linduska"),
			}
			assert.Nil(
				t, New(db).
					Q(`INSERT INTO tests`).
					Q(`(id, name, lastname, active, amount, amount_special, quantity, roles, note, vectors, created_at)`).
					Q(
						`VALUES (DEFAULT, @name, @lastname, @active, @amount, @amount-special, @quantity, @roles, @note, to_tsvector(@vectors), DEFAULT)`,
						data,
					).
					Q(`RETURNING id`).
					Exec(&id),
			)
			testResult := make(Map)
			New(db).Q(`SELECT * FROM tests WHERE id = @id`, Map{"id": id}).MustExec(&testResult)
			assert.Equal(t, note, testResult["note"])
			assert.Equal(t, 1, id, "should create row and return new id")
			
			fulltextResult := make(Map)
			New(db).Q(
				`SELECT * FROM tests WHERE vectors @@ to_tsquery(@query)`, Map{"query": CreateTsQuery("linduska")},
			).MustExec(&fulltextResult)
			assert.Equal(t, int64(1), fulltextResult["id"])
		},
	)
	t.Run(
		"scan with struct", func(t *testing.T) {
			testResult := new(test)
			New(db).Q(`SELECT * FROM tests WHERE id = @id`, Map{"id": 1}).MustExec(testResult)
			assert.Equal(t, 1, testResult.Id)
			assert.Equal(t, "Dominik", testResult.Name)
			assert.Equal(t, "Linduska", testResult.Lastname)
		},
	)
	t.Run(
		"update from struct with named params", func(t *testing.T) {
			assert.Nil(
				t, New(db).
					Q(`UPDATE tests`).
					Q(`SET active = @active`, Map{"active": false}).
					Q(`WHERE id = @id`, Map{"id": 1}).
					Exec(),
			)
			active := true
			assert.Nil(
				t, New(db).
					Q(`SELECT active`).
					Q(`FROM tests`).
					Q(`WHERE id = @id`, Map{"id": 1}).
					Exec(&active),
			)
			assert.Equal(t, false, active, "should be deactivated")
		},
	)
	t.Run(
		"update roles", func(t *testing.T) {
			roles := []string{"admin", "test"}
			args := Map{"id": 1, "roles": roles}
			assert.Nil(
				t, New(db).
					Q(`UPDATE tests`).
					Q(`SET roles = @roles`, args).
					Q(`WHERE id = @id`, args).
					Exec(),
			)
			data := test{}
			assert.Nil(
				t, New(db).
					Q(`SELECT roles`).
					Q(`FROM tests`).
					Q(`WHERE id = @id`, Map{"id": 1}).
					Exec(&data),
			)
			assert.Equal(t, roles, data.Roles)
		},
	)
	t.Run(
		"select", func(t *testing.T) {
			var r test
			assert.Nil(
				t, New(db).
					Q(`SELECT *`).
					Q(`FROM tests`).
					Q(`WHERE id = @id`, Map{"id": 1}).
					Q(`LIMIT 1`).
					Exec(&r),
			)
			assert.Equal(t, 1, r.Id)
		},
	)
	t.Run(
		"select where in", func(t *testing.T) {
			var r test
			assert.Nil(
				t, New(db).
					Q(`SELECT *`).
					Q(`FROM tests`).
					Q(`WHERE id IN (@id)`, Map{"id": []int{1, 2}}).
					Exec(&r),
			)
			assert.Equal(t, 1, r.Id)
		},
	)
	t.Run(
		"delete", func(t *testing.T) {
			var count int
			assert.Nil(
				t, New(db).
					Q(`SELECT count(id)`).
					Q(`FROM tests`).
					Exec(&count),
			)
			assert.Equal(t, 1, count)
			assert.Nil(
				t, New(db).
					Q(`DELETE FROM tests`).
					Q(`WHERE id = @id`, Map{"id": 1}).
					Exec(),
			)
			assert.Nil(
				t, New(db).
					Q(`SELECT count(id)`).
					Q(`FROM tests`).
					Exec(&count),
			)
			assert.Equal(t, 0, count)
		},
	)
}
