package quirk

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestUtil(t *testing.T) {
	t.Run(
		"get substring indexes", func(t *testing.T) {
			assert.Equal(
				t, []int{13, 29, 37, 42}, getSubstringIndexes(`WHERE neco = ? AND neco2 IN (?) AND (?) = ?`, Placeholder),
			)
		},
	)
	t.Run(
		"replace placeholder with slice placeholders", func(t *testing.T) {
			q := `WHERE neco = (?)`
			arg := []int{1, 2, 3}
			slicePlaceholder := createSlicePlaceholder(len(arg))
			for _, i := range getSubstringIndexes(q, Placeholder) {
				q = replaceStringAtIndex(q, Placeholder, slicePlaceholder, i)
			}
			assert.Equal(t, "?,?,?", slicePlaceholder)
			assert.Equal(
				t, `WHERE neco = (?,?,?)`, q,
			)
		},
	)
	t.Run(
		"contains query named param", func(t *testing.T) {
			containsQueryNamedParam(`:Neco::tsquery`)
		},
	)
}
