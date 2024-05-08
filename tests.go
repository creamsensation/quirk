package quirk

import (
	"database/sql"
	"time"
)

type test struct {
	Id            int            `db:"id"`
	Name          string         `db:"name"`
	Lastname      string         `db:"lastname"`
	Active        bool           `db:"active"`
	Amount        float64        `db:"amount"`
	AmountSpecial float64        `db:"amount_special"`
	Quantity      int            `db:"quantity"`
	Roles         []string       `db:"roles"`
	Note          sql.NullString `db:"note"`
	CreatedAt     time.Time      `db:"created_at"`
}

func createConnection() (*DB, error) {
	return Connect(
		WithPostgres(),
		WithHost("localhost"),
		WithPort(5432),
		WithDbname("test"),
		WithUser("swtp"),
		WithPassword("swtp"),
		WithSslDisable(),
	)
}
