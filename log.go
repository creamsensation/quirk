package quirk

import (
	"fmt"
	"strings"
	"time"
)

func log(use bool, q string, d time.Duration) {
	if !use {
		return
	}
	fmt.Printf("[%s] %s\n", d.String(), q)
}

func createQueryLog(driverName string, q string, args ...any) string {
	q = formatSql(q)
	for i, a := range args {
		switch driverName {
		case Postgres:
			q = strings.Replace(q, fmt.Sprintf("$%d", i+1), fmt.Sprintf("%v", a), 1)
		case Mysql:
			q = strings.Replace(q, "?", fmt.Sprintf("%v", a), 1)
		}
	}
	return q
}
