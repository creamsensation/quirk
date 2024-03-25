package quirk

import (
	"fmt"
	"strings"
)

func CreateTsVectors(values ...any) Safe {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = Normalize(fmt.Sprintf("%v", v))
	}
	return Safe{fmt.Sprintf("to_tsvector('simple', '%s')", strings.Join(result, " "))}
}

func CreateTsQuery(value string) Safe {
	result := make([]string, 0)
	for _, v := range strings.Split(Normalize(value), " ") {
		n := len(v)
		if n > 1 {
			v = v[:n-1]
		}
		result = append(result, v+":*")
	}
	return Safe{fmt.Sprintf("to_tsquery('simple', '%s')", strings.Join(result, " & "))}
}

func MapToJsonb[T any](mv map[string]T) Jsonb[T] {
	result := make(Jsonb[T])
	for k, v := range mv {
		result[k] = v
	}
	return result
}

func MapToTsVectorsValue[T any](m map[string]T) string {
	n := len(m)
	result := make([]string, n)
	if n == 0 {
		return ""
	}
	i := 0
	for _, v := range m {
		result[i] = fmt.Sprintf("%v", v)
		i++
	}
	return strings.Join(result, " ")
}
