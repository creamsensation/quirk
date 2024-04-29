package quirk

import (
	"database/sql"
	"time"
)

type DB struct {
	*sql.DB
	driverName  string
	transaction bool
	rollback    bool
	log         bool
}

const (
	Postgres = "postgres"
	Mysql    = "mysql"
)

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	return wrapConnection(db, driverName), err
}

func wrapConnection(db *sql.DB, driverName string) *DB {
	return &DB{
		DB:          db,
		driverName:  driverName,
		transaction: false,
		rollback:    false,
		log:         false,
	}
}

func (d *DB) Q(query string, arg ...Map) *Quirk {
	return New(d).Q(query, arg...)
}

func (d *DB) DriverName() string {
	return d.driverName
}

func (d *DB) Log(use ...bool) {
	l := true
	if len(use) > 0 {
		l = use[0]
	}
	d.log = l
}

func (d *DB) Begin() (*DB, error) {
	db := &DB{
		DB:          d.DB,
		driverName:  d.driverName,
		transaction: true,
		rollback:    false,
		log:         d.log,
	}
	q := "BEGIN;"
	t := time.Now()
	_, err := d.DB.Query(q)
	log(db.log, q, time.Now().Sub(t))
	return db, err
}

func (d *DB) Rollback() error {
	if !d.transaction {
		return nil
	}
	d.rollback = true
	q := "ROLLBACK;"
	t := time.Now()
	_, err := d.DB.Query(q)
	log(d.log, q, time.Now().Sub(t))
	return err
}

func (d *DB) Commit() error {
	if !d.transaction {
		return nil
	}
	q := "COMMIT;"
	t := time.Now()
	_, err := d.DB.Query(q)
	log(d.log, q, time.Now().Sub(t))
	return err
}

func (d *DB) MustBegin() *DB {
	db, err := d.Begin()
	if err != nil {
		panic(err)
	}
	d.transaction = true
	return db
}

func (d *DB) MustRollback() {
	if !d.transaction {
		return
	}
	err := d.Rollback()
	if err != nil {
		panic(err)
	}
}

func (d *DB) MustCommit() {
	if !d.transaction {
		return
	}
	if err := d.Commit(); err != nil {
		panic(err)
	}
}
