package migrator

import "quirk"

type Control interface {
	DB(name ...string) *quirk.Quirk
}

type control struct {
	*migrator
}

func (c *control) DB(name ...string) *quirk.Quirk {
	n := mainDbname
	if len(name) > 0 {
		n = name[0]
	}
	d, ok := c.databases[n]
	if !ok {
		panic(ErrorInvalidDatabase)
	}
	return quirk.New(d)
}
