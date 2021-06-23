package model

import (
	"errors"
)

// Process a string field.
func stringField(dst *string, fields map[string]interface{}, key string) error {
	val, ok := fields[key]
	if !ok {
		return nil
	}
	*dst, ok = val.(string)
	if !ok {
		return errors.New("invalid value for '" + key + "': not a string")
	}
	delete(fields, key)
	return nil
}