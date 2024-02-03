package migrator

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"time"
	
	"github.com/creamsensation/quirk"
)

type Migrator interface {
	Run()
}

type migrator struct {
	dir        string
	databases  map[string]*quirk.DB
	migrations []*Migration
	init       *bool
	new        *bool
	up         *bool
	down       *bool
}

const (
	migrationsTable string = "quirk_migrations"
)

const (
	migrationFileContent = `package main

import "github.com/creamsensation/quirk/migrator"

func init() {
	manager.Add().
		Up(
			func(c migrator.Control) {

			},
		).
		Down(
			func(c migrator.Control) {
			
			},
		)
}
`
)

func New(dir string, databases map[string]*quirk.DB, migrations []*Migration) Migrator {
	m := &migrator{dir: dir, databases: databases, migrations: migrations}
	return m
}

func (m *migrator) Run() {
	m.init = flag.Bool("init", false, "Init migrations")
	m.new = flag.Bool("new", false, "New migration")
	m.up = flag.Bool("up", false, "Up migrations")
	m.down = flag.Bool("down", false, "Down migration")
	flag.Parse()
	flag.Parse()
	if *m.init {
		m.Init()
		return
	}
	if *m.new {
		m.New()
		return
	}
	if *m.up {
		m.Up()
		return
	}
	if *m.down {
		m.Down()
		return
	}
}

func (m *migrator) Init() {
	for _, db := range m.databases {
		quirk.New(db).Q(
			fmt.Sprintf(
				`CREATE TABLE IF NOT EXISTS %s (
    id serial primary key,
    name varchar(255) not null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp
    )`, migrationsTable,
			),
		).MustExec()
	}
}

func (m *migrator) New() {
	if len(m.dir) == 0 {
		return
	}
	m.check(os.MkdirAll(m.dir, os.ModePerm))
	filepath := fmt.Sprintf("%s/%d.go", m.dir, time.Now().UnixNano())
	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		return
	}
	file, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(migrationFileContent)
	if err != nil {
		panic(err)
	}
}

func (m *migrator) Up() {
	existingMigrationsNames := m.getExistingMigrationsNames()
	for _, db := range m.databases {
		db.MustBegin()
	}
	for _, item := range m.migrations {
		if slices.Contains(existingMigrationsNames, item.name) {
			continue
		}
		fmt.Printf("Up [%s]...\n", item.name)
		item.up(&control{m})
		m.insertMigration(item.name)
	}
	for _, db := range m.databases {
		db.MustCommit()
	}
	for _, item := range m.migrations {
		if slices.Contains(existingMigrationsNames, item.name) {
			continue
		}
		fmt.Printf("Up successful [%s]!\n", item.name)
	}
}

func (m *migrator) Down() {
	lastMigrationName := m.getLastMigrationName()
	for _, db := range m.databases {
		db.MustBegin()
	}
	for _, item := range m.migrations {
		if !slices.Contains(lastMigrationName, item.name) {
			continue
		}
		fmt.Printf("Down [%s]...\n", item.name)
		item.down(&control{m})
		m.deleteMigration(item.name)
	}
	for _, db := range m.databases {
		db.MustCommit()
	}
	for _, item := range m.migrations {
		if !slices.Contains(lastMigrationName, item.name) {
			continue
		}
		fmt.Printf("Down successful [%s]!\n", item.name)
	}
}

func (m *migrator) check(err error) {
	if err == nil {
		return
	}
	panic(err)
}

func (m *migrator) getLastMigrationName() []string {
	result := make([]string, 0)
	for _, db := range m.databases {
		var r string
		quirk.New(db).Q(fmt.Sprintf(`SELECT name FROM %s ORDER BY created_at DESC LIMIT 1`, migrationsTable)).MustExec(&r)
		if !slices.Contains(result, r) {
			result = append(result, r)
		}
	}
	return result
}

func (m *migrator) getExistingMigrationsNames() []string {
	result := make([]string, 0)
	for _, db := range m.databases {
		r := make([]string, 0)
		quirk.New(db).Q(fmt.Sprintf(`SELECT name FROM %s ORDER BY created_at ASC`, migrationsTable)).MustExec(&r)
		for _, name := range r {
			if !slices.Contains(result, name) {
				result = append(result, name)
			}
		}
	}
	return result
}

func (m *migrator) insertMigration(name string) {
	for _, db := range m.databases {
		quirk.New(db).
			Q(
				fmt.Sprintf(
					`INSERT INTO %s (id, name, created_at, updated_at) VALUES (DEFAULT, @name, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
					migrationsTable,
				),
				quirk.Map{"name": name},
			).
			MustExec()
	}
}

func (m *migrator) deleteMigration(name string) {
	for _, db := range m.databases {
		quirk.New(db).Q(
			fmt.Sprintf(`DELETE FROM %s WHERE name = @name`, migrationsTable), quirk.Map{"name": name},
		).MustExec()
	}
}
