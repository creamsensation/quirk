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
	return Safe{fmt.Sprintf("to_tsvector('simple', '%s')", strings.Join(result, " "))}
}

func CreateTsQuery(value string) Safe {
	result := make([]string, 0)
	for _, v := range strings.Split(simplify(value), " ") {
		n := len(v)
		if n > 1 {
			v = v[:n-1]
		}
		result = append(result, v+":*")
	}
	return Safe{fmt.Sprintf("to_tsquery('simple', '%s')", strings.Join(result, " & "))}
}
