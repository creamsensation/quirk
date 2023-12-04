package quirk

import (
	"fmt"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	t.Run(
		"basic connection", func(t *testing.T) {
			host, port, user, password, dbname := "localhost", 5432, "test", "12345", "test"
			driver, dataSource, _, err := createConnectionDataSource(
				WithPostgres(),
				WithHost(host),
				WithPort(port),
				WithUser(user),
				WithPassword(password),
				WithDbname(dbname),
				WithSslDisable(),
			)
			assert.Nil(t, err)
			assert.Equal(t, Postgres, driver)
			assert.Equal(
				t,
				fmt.Sprintf(
					"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname,
				),
				dataSource,
			)
		},
	)
}
