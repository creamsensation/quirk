package quirk

import (
	"fmt"
	"slices"
	"strings"
)

func CreateTsVector(values ...any) string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = Normalize(fmt.Sprintf("%v", v))
		result[i] = Escape(result[i])
	}
	return strings.Join(result, " ")
}

func CreateTsQuery(values ...any) string {
	result := make([]string, 0)
	for _, item := range values {
		value := fmt.Sprintf("%v", item)
		if len(value) == 0 {
			continue
		}
		for _, v := range strings.Split(Normalize(value), " ") {
			n := len(v)
			if n > 1 {
				v = v[:n-1]
			}
			escaped := Escape(v) + ":*"
			if slices.Contains(result, escaped) {
				continue
			}
			result = append(result, escaped)
		}
	}
	return strings.Join(result, " & ")
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
