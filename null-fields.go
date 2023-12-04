package quirk

import (
	"database/sql"
	"time"
)

func NullBool(v bool) sql.NullBool {
	return sql.NullBool{
		Bool:  v,
		Valid: !!v,
	}
}

func NullString(v string) sql.NullString {
	return sql.NullString{
		String: v,
		Valid:  len(v) != 0,
	}
}

func NullInt(v int) sql.NullInt64 {
	return sql.NullInt64{
		Int64: int64(v),
		Valid: v != 0,
	}
}

func NullInt16(v int16) sql.NullInt16 {
	return sql.NullInt16{
		Int16: v,
		Valid: v != 0,
	}
}

func NullInt32(v int32) sql.NullInt32 {
	return sql.NullInt32{
		Int32: v,
		Valid: v != 0,
	}
}

func NullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: v,
		Valid: v != 0,
	}
}

func NullFloat64(v float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: v,
		Valid:   v != 0,
	}
}

func NullTime(v time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  v,
		Valid: !v.IsZero(),
	}
}

func NullByte(v byte) sql.NullByte {
	return sql.NullByte{
		Byte:  v,
		Valid: v != 0,
	}
}
