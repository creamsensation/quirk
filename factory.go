package quirk

import (
	"fmt"
	"strings"
)

func CreateInsert(data map[string]any) (string, string) {
	n := len(data)
	columns := make([]string, n)
	placeholders := make([]string, n)
	i := 0
	for k := range data {
		columns[i] = k
		placeholders[i] = "@" + k
		i++
	}
	return strings.Join(columns, ", "), strings.Join(placeholders, ", ")
}

func CreateUpdate(data map[string]any) string {
	n := len(data)
	result := make([]string, n)
	i := 0
	for k := range data {
		result[i] = fmt.Sprintf("%[1]s = @%[1]s", k)
		i++
	}
	return strings.Join(result, ", ")
}
