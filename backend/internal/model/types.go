package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// StringArray is a slice of strings that serializes to/from JSON in SQL databases
type StringArray []string

// Value implements the driver.Valuer interface for database writes
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for database reads
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal StringArray value:", value))
	}

	if len(bytes) == 0 {
		*a = []string{}
		return nil
	}

	return json.Unmarshal(bytes, a)
}
