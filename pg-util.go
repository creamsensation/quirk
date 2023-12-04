package quirk

import (
	"fmt"
	"strings"
)

func CreateTsVectors(values ...any) Safe {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = simplify(fmt.Sprintf("%v", v))
	}
	return Safe{fmt.Sprintf("to_tsvector('%s')", strings.Join(result, " "))}
}

func CreateTsQuery(values ...any) Safe {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = simplify(fmt.Sprintf("%v", v)) + ":*"
	}
	return Safe{fmt.Sprintf("to_tsquery('%s')", strings.Join(result, " & "))}
}
