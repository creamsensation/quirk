package quirk

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Jsonb[T any] map[string]T

func (j Jsonb[T]) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *Jsonb[T]) Scan(value any) error {
	switch v := value.(type) {
	case []uint8:
		return json.Unmarshal(v, &j)
	default:
		return errors.New("incompatible scan type for jsonb")
	}
}
