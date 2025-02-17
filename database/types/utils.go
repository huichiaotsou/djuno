package types

import (
	"database/sql"
	"strings"
	"time"
)

// ToNullString converts the given value to a nullable string
func ToNullString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	return sql.NullString{
		Valid:  value != "",
		String: value,
	}
}

// ToNullTime converts the given value to a nullable time
func ToNullTime(value *time.Time) sql.NullTime {
	if value == nil || value.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Valid: true, Time: *value}
}
