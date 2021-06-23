package chassis

import (
	"errors"
	"github.com/lib/pq"
	"time"
)

// Process a bool field.
func BoolField(dst *bool, fields map[string]interface{}, key string) error {
	val, ok := fields[key]
	if !ok {
		return nil
	}
	*dst, ok = val.(bool)
	if !ok {
		return errors.New("invalid value for '" + key + "': not a bool")
	}
	delete(fields, key)
	return nil
}
// Process a string field.
func StringField(dst *string, fields map[string]interface{}, key string) error {
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

// Process an int field.
func IntField(dst *int, fields map[string]interface{}, key string) error {
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

// Process an uint field.
func UintField(dst *uint, fields map[string]interface{}, key string) error {
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
	*dst = uint(dstFloat)
	return nil
}

func TimeField(dst *time.Time, fields map[string]interface{}, key string) error {
	val, ok := fields[key]
	if !ok {
		return nil
	}
	rawTime, ok := val.(string)
	if !ok {
		return errors.New("invalid value for '" + key + "': not a string")
	}

	datetime, err := time.Parse(time.RFC3339, rawTime)
	if err != nil {
		return errors.New("invalid value for '" + key + "': not a valid datetime")
	}
	*dst = datetime
	delete(fields, key)

	return nil
}

// Process a field that has a list of strings.
func StringListField(dst *pq.StringArray, fields map[string]interface{}, key string) error {
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

// Process a field that has a list of strings.
func StringSliceField(dst *[]string, fields map[string]interface{}, key string) error {
	val, ok := fields[key]
	if !ok {
		if *dst == nil {
			empty := []string{}
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

// Check a read-only field.
func ReadOnlyField(fields map[string]interface{}, key string, ro *[]string) {
	_, ok := fields[key]
	if ok {
		*ro = append(*ro, key)
	}
}