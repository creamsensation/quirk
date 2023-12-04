package quirk

import (
	"database/sql"
	"time"
)

type test struct {
	Id       int
	Name     string
	Lastname string
	Active   bool
	Amount   float64
	Quantity int
	// Roles     Array[string]
	Roles     []string
	Note      sql.NullString
	CreatedAt time.Time
}

func createConnection() (*DB, error) {
	return Connect(
		WithPostgres(),
		WithHost("localhost"),
		WithPort(5432),
		WithDbname("test"),
		WithUser("cream"),
		WithPassword("cream"),
		WithSslDisable(),
	)
}
