package quirk

import (
	"cmp"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"
	
	pg "github.com/lib/pq"
)

const (
	ParamPrefix = "@"
	Placeholder = "?"
)

var (
	tableColCharMatcher   = regexp.MustCompile(`[a-zA-Z0-9_]`)
	argPlaceholderMatcher = regexp.MustCompile(`\(\$[0-9]+\)`)
)

var (
	safeType = reflect.TypeOf(Safe(""))
)

type indexedArg struct {
	index int
	name  string
	value any
}

func processQueryParts(q *Quirk) (string, []any, error) {
	parts := make([]string, 0)
	args := make([]any, 0)
	pgi := 1
	for _, p := range q.parts {
		existingNames := make([]string, 0)
		for argKey := range p.arg {
			if !existsNameInQuery(argKey, p.query) {
				continue
			}
			existingNames = append(existingNames, argKey)
		}
		argLen := len(existingNames)
		partArgs := make([]indexedArg, argLen)
		i := 0
		for argKey, argValue := range p.arg {
			if !slices.Contains(existingNames, argKey) {
				continue
			}
			partArgs[i] = indexedArg{
				index: strings.Index(p.query, ParamPrefix+argKey),
				name:  argKey,
				value: argValue,
			}
			i++
		}
		slices.SortFunc(
			partArgs, func(a, b indexedArg) int {
				return cmp.Compare(a.index, b.index)
			},
		)
		for _, partArg := range partArgs {
			argValueType := reflect.TypeOf(partArg.value)
			isSafe := argValueType == safeType
			isSlice := argValueType.Kind() == reflect.Slice
			isMap := argValueType.Kind() == reflect.Map
			name := ParamPrefix + partArg.name
			if !isSafe {
				p.query = replaceStringAtIndex(p.query, name, fmt.Sprintf("$%d", pgi), findParamIndex(p.query, name))
				pgi++
			}
			if isSafe {
				p.query = replaceStringAtIndex(
					p.query, name, fmt.Sprintf("%s", string(partArg.value.(Safe))), findParamIndex(p.query, name),
				)
			}
			if isSlice {
				partArg.value = pg.Array(partArg.value)
			}
			if isMap {
				partArg.value = transformMapToJsonb(partArg.value)
			}
			if !isSafe {
				args = append(args, partArg.value)
			}
		}
		sliceExists := false
		for _, partArg := range partArgs {
			if reflect.TypeOf(partArg.value).Kind() == reflect.Slice {
				sliceExists = true
				break
			}
		}
		if sliceExists {
			p.query = processQueryInOperator(p.query)
		}
		parts = append(parts, p.query)
	}
	return strings.Join(parts, " "), args, nil
}

func transformMapToJsonb(value any) any {
	switch m := value.(type) {
	case map[string]string:
		return MapToJsonb[string](m)
	case map[string]int:
		return MapToJsonb[int](m)
	case map[string]float64:
		return MapToJsonb[float64](m)
	case map[string]bool:
		return MapToJsonb[bool](m)
	case map[string]time.Time:
		return MapToJsonb[time.Time](m)
	default:
		return nil
	}
}

func processQueryInOperator(q string) string {
	if strings.Contains(strings.ToLower(q), " in (") {
		q = strings.ToUpper(strings.Replace(strings.ToLower(q), " in (", " = any(", 1))
	}
	return q
}

func existsNameInQuery(name, query string) bool {
	return regexp.MustCompile(regexp.QuoteMeta(ParamPrefix+name)+`([ ,)'])`).MatchString(query) || strings.HasSuffix(
		query, ParamPrefix+name,
	)
}

func findParamIndex(query, param string) int {
	qn := len(query)
	for _, i := range getSubstringIndexes(query, param) {
		n := len(param)
		if i+n < qn {
			nextChar := query[i+n]
			if query[i:i+n] == param && !tableColCharMatcher.MatchString(string(nextChar)) {
				return i
			}
		}
		if i+n >= qn {
			if query[i:i+n] == param {
				return i
			}
		}
	}
	return -1
}
