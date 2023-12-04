package migrator

import (
	"runtime"
	"strings"
)

type Manager struct {
	migrations []*Migration
}

func (m *Manager) Add() *Migration {
	_, file, _, _ := runtime.Caller(1)
	parts := strings.Split(file, "/")
	mig := new(Migration)
	mig.name = strings.TrimSuffix(parts[len(parts)-1], ".go")
	m.migrations = append(m.migrations, mig)
	return mig
}

func (m *Manager) GetAll() []*Migration {
	return m.migrations
}
