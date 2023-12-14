package quirk

import (
	"database/sql"
	"reflect"
	"regexp"
	"strings"
	"time"
	
	"github.com/iancoleman/strcase"
	pg "github.com/lib/pq"
)

type Quirk struct {
	*DB
	driverName    string
	dbname        string
	parts         []queryPart
	rows          *sql.Rows
	subscriptions []subscription
}

type Safe struct {
	Value any
}

type queryPart struct {
	query string
	args  []any
}

const (
	querySuffix = ";"
)

func New(db *DB) *Quirk {
	q := &Quirk{
		DB:            db,
		driverName:    db.driverName,
		subscriptions: make([]subscription, 0),
	}
	return q
}

func (q *Quirk) Q(query string, args ...any) *Quirk {
	q.parts = append(q.parts, queryPart{query, args})
	return q
}

func (q *Quirk) If(condition bool, query string, args ...any) *Quirk {
	if !condition {
		return q
	}
	q.Q(query, args...)
	return q
}

func (q *Quirk) Subscribe(s subscription) {
	q.subscriptions = append(q.subscriptions, s)
}

func (q *Quirk) CreateSql() string {
	mergedQueryParts, _, err := processQueryParts(q)
	if !strings.HasSuffix(mergedQueryParts, querySuffix) {
		mergedQueryParts += querySuffix
	}
	if err != nil {
		return ""
	}
	return mergedQueryParts
}

func (q *Quirk) CreateMatcher() string {
	return regexp.QuoteMeta(q.CreateSql())
}

func (q *Quirk) Exec(r ...any) error {
	return q.exec(r...)
}

func (q *Quirk) MustExec(r ...any) {
	if err := q.exec(r...); err != nil {
		panic(err)
	}
}

func (q *Quirk) exec(result ...any) error {
	t := time.Now()
	mergedQueryParts, args, err := processQueryParts(q)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(mergedQueryParts, querySuffix) {
		mergedQueryParts += querySuffix
	}
	rows, err := q.DB.Query(mergedQueryParts, args...)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()
	if len(result) == 0 {
		q.afterQuery(t, mergedQueryParts, args)
		return nil
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		q.afterQuery(t, mergedQueryParts, args)
		return nil
	}
	if len(result) > 1 {
		q.scanMultiple(rows, result...)
	}
	if len(result) == 1 {
		q.scanSingle(rows, columns, result[0])
	}
	q.afterQuery(t, mergedQueryParts, args)
	return nil
}

func (q *Quirk) afterQuery(t time.Time, query string, args []any) {
	queryLog := createQueryLog(q.driverName, query, args...)
	duration := time.Now().Sub(t)
	for _, sub := range q.subscriptions {
		sub(queryLog, duration)
	}
	log(q.log, queryLog, duration)
}

func (q *Quirk) scanSingle(rows *sql.Rows, columns []string, result any) {
	res := reflect.ValueOf(result)
	rv := reflect.ValueOf(result)
	resKind := rv.Elem().Type().Kind()
	rvKind := rv.Elem().Type().Kind()
	if rvKind == reflect.Slice {
		rv = reflect.New(rv.Elem().Type().Elem())
		rvKind = rv.Elem().Type().Kind()
	}
	for rows.Next() {
		rowData := make([]any, len(columns))
		switch rvKind {
		case reflect.Map:
			rv.Elem().Set(reflect.MakeMap(rv.Elem().Type()))
			for i := range columns {
				field := reflect.New(rv.Elem().Type().Elem())
				rowData[i] = field.Interface()
			}
		case reflect.Struct:
			visibleFields := reflect.VisibleFields(rv.Elem().Type())
			model := make(map[string]any)
			for i := 0; i < rv.Elem().NumField(); i++ {
				field := rv.Elem().Field(i)
				fieldName := rv.Elem().Type().Field(i).Name
				exported := false
				for _, vf := range visibleFields {
					if vf.Name == fieldName && vf.IsExported() {
						exported = true
					}
				}
				if !exported {
					continue
				}
				switch field.Type().Kind() {
				case reflect.Slice:
					model[strcase.ToSnake(fieldName)] = pg.Array(field.Addr().Interface())
				default:
					model[strcase.ToSnake(fieldName)] = field.Addr().Interface()
				}
			}
			for i, c := range columns {
				modelField, ok := model[c]
				if !ok {
					rowData[i] = new(any)
					continue
				}
				rowData[i] = modelField
			}
		default:
			switch rv.Type().Kind() {
			case reflect.Slice:
				rowData[0] = pg.Array(rv.Interface())
			default:
				rowData[0] = rv.Interface()
			}
		}
		if len(rowData) > 0 {
			if scanErr := rows.Scan(rowData...); scanErr != nil {
				panic(scanErr)
			}
		}
		switch resKind {
		case reflect.Slice:
			res.Elem().Set(reflect.Append(res.Elem(), rv.Elem()))
		case reflect.Map:
			for i, c := range columns {
				rv.Elem().SetMapIndex(reflect.ValueOf(c), reflect.ValueOf(rowData[i]).Elem())
			}
		}
	}
}

func (q *Quirk) scanMultiple(rows *sql.Rows, result ...any) {
	resultValues := make([]reflect.Value, len(result))
	for i, r := range result {
		resultValues[i] = reflect.ValueOf(r)
	}
	for rows.Next() {
		rowData := make([]any, 0)
		for _, rv := range resultValues {
			rowData = append(rowData, rv.Interface())
		}
		if len(rowData) > 0 {
			if scanErr := rows.Scan(rowData...); scanErr != nil {
				panic(scanErr)
			}
		}
	}
}
