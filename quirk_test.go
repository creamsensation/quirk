package quirk

import (
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
	       		created_at timestamp not null default current_timestamp
	    		)`,
				).Exec(),
			)
		},
	)
	t.Run(
		"insert with struct and names params", func(t *testing.T) {
			id := 0
			data := test{
				Name:          "Dominik",
				Lastname:      "Linduska",
				Active:        true,
				Amount:        999.99,
				AmountSpecial: 999.99,
				Quantity:      55,
				Roles:         []string{"owner", "admin"},
				Note:          NullString("go go go"),
			}
			assert.Nil(
				t, New(db).
					Q(`INSERT INTO tests`).
					Q(`(id, name, lastname, active, amount, amount_special, quantity, roles, note, created_at)`).
					Q(
						`VALUES (DEFAULT, @Name, @Lastname, @Active, @Amount, @AmountSpecial, @Quantity, @Roles, @Note, DEFAULT)`,
						data,
					).
					Q(`RETURNING id`).
					Exec(&id),
			)
			assert.Equal(t, 1, id, "should create row and return new id")
		},
	)
	t.Run(
		"update from struct with named params", func(t *testing.T) {
			data := test{Id: 1}
			assert.Nil(
				t, New(db).
					Q(`UPDATE tests`).
					Q(`SET active = @Active`, data).
					Q(`WHERE id = @Id`, data).
					Exec(),
			)
			active := true
			assert.Nil(
				t, New(db).
					Q(`SELECT active`).
					Q(`FROM tests`).
					Q(`WHERE id = ?`, data.Id).
					Exec(&active),
			)
			assert.Equal(t, false, active, "should be deactivated")
		},
	)
	t.Run(
		"update roles", func(t *testing.T) {
			data := test{Id: 1, Roles: []string{"admin", "test"}}
			assert.Nil(
				t, New(db).
					Q(`UPDATE tests`).
					Q(`SET roles = @Roles`, data).
					Q(`WHERE id = @Id`, data).
					Exec(),
			)
			data.Roles = []string{}
			assert.Nil(
				t, New(db).
					Q(`SELECT roles`).
					Q(`FROM tests`).
					Q(`WHERE id = ?`, data.Id).
					Exec(&data),
			)
			assert.Equal(t, []string{"admin", "test"}, data.Roles)
		},
	)
	t.Run(
		"select", func(t *testing.T) {
			var r test
			assert.Nil(
				t, New(db).
					Q(`SELECT *`).
					Q(`FROM tests`).
					Q(`WHERE id = ?`, 1).
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
					Q(`WHERE id IN (?)`, []int{1, 2}).
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
					Q(`WHERE id = ?`, 1).
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
