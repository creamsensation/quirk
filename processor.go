package quirk

import (
	"cmp"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	pg "github.com/lib/pq"
)

const (
	ParamPrefix = "@"
	Placeholder = "?"

	tableColCharRegexPattern = `[a-zA-Z0-9_]`
)

var (
	tableColCharRegex = regexp.MustCompile(tableColCharRegexPattern)
)

type paramSorter struct {
	index int
	name  string
	value any
}

func processQueryParts(q *Quirk) (string, []any, error) {
	parts := make([]string, 0)
	args := make([]any, 0)
	for _, p := range q.parts {
		partArgs := make([]any, 0)
		containsNamedParams := containsQueryNamedParam(p.query)
		if containsNamedParams && len(p.args) > 0 {
			argRef := reflect.ValueOf(p.args[0])
			switch argRef.Kind() {
			case reflect.Map:
				processedQuery, processedArgs := processNamedParamsMap(p.query, argRef)
				processedQuery, processedArgs, err := processPart(processedQuery, processedArgs...)
				if err != nil {
					return "", args, err
				}
				p.query = processedQuery
				partArgs = append(partArgs, processedArgs...)
			case reflect.Struct:
				processedQuery, processedArgs := processNamedParamsStruct(p.query, argRef)
				processedQuery, processedArgs, err := processPart(processedQuery, processedArgs...)
				if err != nil {
					return "", args, err
				}
				p.query = processedQuery
				partArgs = append(partArgs, processedArgs...)
			}
			parts = append(parts, p.query)
		}
		if !containsNamedParams && len(p.args) > 0 {
			processedQuery, processedArgs, err := processPart(p.query, p.args...)
			if err != nil {
				return "", args, err
			}
			partArgs = append(partArgs, processedArgs...)
			parts = append(parts, processedQuery)
		}
		if len(p.args) == 0 {
			parts = append(parts, p.query)
		}
		args = append(args, partArgs...)
	}
	mergedQuery := strings.Join(parts, " ")
	placeholdersCounts := strings.Count(mergedQuery, Placeholder)
	for i := 1; i <= placeholdersCounts; i++ {
		switch q.driverName {
		case Postgres:
			mergedQuery = strings.Replace(mergedQuery, Placeholder, fmt.Sprintf("$%d", i), 1)
		}
	}
	return mergedQuery, args, nil
}

func processPart(q string, args ...any) (string, []any, error) {
	resultArgs := make([]any, 0)
	placeholdersIndexes := getSubstringIndexes(q, Placeholder)
	if len(placeholdersIndexes) != len(args) {
		return q, resultArgs, ErrorMismatchArgs
	}
	for i, arg := range args {
		switch a := arg.(type) {
		case Safe:
			q = replaceStringAtIndex(
				q, Placeholder, fmt.Sprintf("%v", a.Value), placeholdersIndexes[i],
			)
			continue
		}
		argRef := reflect.ValueOf(arg)
		switch argRef.Kind() {
		case reflect.Slice:
			if strings.Contains(strings.ToLower(q), " in ") {
				q = strings.Replace(strings.ToLower(q), " in ", " = ", 1)
			}
			if strings.Contains(strings.ToLower(q), "(?)") {
				q = strings.Replace(strings.ToLower(q), "(?)", "ANY(?)", 1)
			}
			resultArgs = append(resultArgs, pg.Array(arg))
		default:
			resultArgs = append(resultArgs, arg)
		}
	}
	return q, resultArgs, nil
}

func processNamedParamsMap(query string, mapArgRef reflect.Value) (string, []any) {
	args := make([]any, 0)
	params := make([]paramSorter, 0)
	originQuery := query
	for _, key := range mapArgRef.MapKeys() {
		name := strcase.ToSnake(key.String())
		param := ParamPrefix + name
		value := mapArgRef.MapIndex(key)
		if !strings.Contains(query, param) {
			continue
		}
		index := strings.Index(originQuery, param)
		suitableIndex := findMostSuitableParamIndex(originQuery, param)
		params = append(params, paramSorter{index: suitableIndex, name: name, value: value.Interface()})
		if index != suitableIndex {
			query = replaceStringAtIndex(query, param, Placeholder, findMostSuitableParamIndex(query, param))
		}
		if index == suitableIndex {
			query = strings.Replace(query, param, Placeholder, 1)
		}
	}
	slices.SortFunc(
		params, func(a, b paramSorter) int {
			return cmp.Compare(a.index, b.index)
		},
	)
	for _, item := range params {
		args = append(args, item.value)
	}
	return query, args
}

func processNamedParamsStruct(query string, structArgRef reflect.Value) (string, []any) {
	args := make([]any, 0)
	params := make([]paramSorter, 0)
	originQuery := query
	for i := 0; i < structArgRef.NumField(); i++ {
		field := structArgRef.Field(i)
		fieldName := structArgRef.Type().Field(i).Name
		param := ParamPrefix + fieldName
		if !strings.Contains(query, param) {
			continue
		}
		index := strings.Index(originQuery, param)
		suitableIndex := findMostSuitableParamIndex(originQuery, param)
		params = append(params, paramSorter{index: suitableIndex, name: fieldName, value: field.Interface()})
		if index != suitableIndex {
			query = replaceStringAtIndex(query, param, Placeholder, findMostSuitableParamIndex(query, param))
		}
		if index == suitableIndex {
			query = strings.Replace(query, param, Placeholder, 1)
		}
	}
	slices.SortFunc(
		params, func(a, b paramSorter) int {
			return cmp.Compare(a.index, b.index)
		},
	)
	for _, item := range params {
		args = append(args, item.value)
	}
	return query, args
}

func findMostSuitableParamIndex(query, param string) int {
	qn := len(query)
	for _, i := range getSubstringIndexes(query, param) {
		n := len(param)
		if i+n < qn {
			nextChar := query[i+n]
			if query[i:i+n] == param && !tableColCharRegex.MatchString(string(nextChar)) {
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
