package model

import (
	"errors"
	"github.com/lib/pq"
)
//TODO: THIS CAN BE PLACED ON CHASSIS AND REMOVED FROM HERE AND ITEM-SERVICE
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

func boolField(dst *bool, fields map[string]interface{}, key string) error {
	val, ok := fields[key]
	if !ok {
		return nil
	}
	*dst, ok = val.(bool)
	if !ok {
		return errors.New("invalid value for '" + key + "': not a boolean")
	}
	delete(fields, key)
	return nil
}

// Process an int field.
func intField(dst *int, fields map[string]interface{}, key string) error {
	val, ok := fields[key]
	if !ok {
		return nil
	}
	var dstFloat float64
	dstFloat, ok = val.(float64)
	if !ok {
		return errors.New("invalid value for '" + key + "': not numeric")
	}
	delete(fields, key)
	*dst = int(dstFloat)
	return nil
}

// Process a field that has a list of strings.
func stringListField(dst *pq.StringArray, fields map[string]interface{}, key string) error {
	val, ok := fields[key]
	if !ok {
		if *dst == nil {
			empty := pq.StringArray{}
			*dst = empty
		}
		return nil
	}
	vals, ok := val.([]interface{})
	if !ok {
		return errors.New("invalid value for '" + key + "': not an array")
	}
	ss := []string{}
	for _, si := range vals {
		s, ok := si.(string)
		if !ok {
			return errors.New("non-string value in '" + key + "' array")
		}
		ss = append(ss, s)
	}
	*dst = ss
	delete(fields, key)
	return nil
}

// Process URL list field.


// Check a read-only field.
func readOnlyField(fields map[string]interface{}, key string, ro *[]string) {
	_, ok := fields[key]
	if ok {
		*ro = append(*ro, key)
	}
}