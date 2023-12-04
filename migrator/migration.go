package migrator

type Migration struct {
	*migrator
	name string
	up   func(c Control)
	down func(c Control)
}

func (m *Migration) Up(fn func(c Control)) *Migration {
	m.up = fn
	return m
}

func (m *Migration) Down(fn func(c Control)) *Migration {
	m.down = fn
	return m
}
