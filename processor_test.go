package quirk

import (
	"reflect"
	"strings"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {
	t.Run(
		"process named params struct", func(t *testing.T) {
			q, _ := processNamedParamsStruct(
				`INSERT INTO posts (name, lastname, vectors) VALUES (@Name, @Lastname, to_tsvector('Daar Walker'))`,
				reflect.ValueOf(
					test{
						Name:     "Daar",
						Lastname: "Walker",
					},
				),
			)
			assert.Equal(t, 2, strings.Count(q, Placeholder), "all named params should be placeholders")
		},
	)
}
