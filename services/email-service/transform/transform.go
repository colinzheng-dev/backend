package transform

import (
	"encoding/json"
	"errors"
)

// Unmarshal does topic-specific unmarshalling for messages that
// generate emails.
func Unmarshal(topic string, data []byte, fields map[string]interface{}) error {
	switch topic {
	case "user-created":
		return userCreated(data, fields)
	default:
		return genericValidations(data, fields)
	}
}

func userCreated(data []byte, fields map[string]interface{}) error {
	return defaultFields(data, fields)
}

func genericValidations(data []byte, fields map[string]interface{}) error {
	var err error
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil
	}

	if err = checkStringField("email", true, fields); err != nil {
		return err
	}
	if err = checkStringField("site", false, fields);  err != nil {
		return err
	}
	if err = checkStringField("language", true, fields); err != nil {
		return err
	}
	//not required because we can set "customer" as default value on templates
	if err = checkStringField("customer_name", false, fields); err != nil {
		return err
	}

	return nil
}

func defaultFields(data []byte, fields map[string]interface{}) error {
	flds := map[string]interface{}{}
	err := json.Unmarshal(data, &flds)
	if err != nil {
		return nil
	}
	err = addStringField("email", true, flds, fields)
	if err != nil {
		return err
	}
	err = addStringField("site", false, flds, fields)
	if err != nil {
		return err
	}
	return addStringField("language", false, flds, fields)
}

func addStringField(name string, required bool,
	data map[string]interface{}, fields map[string]interface{}) error {
	f, ok := data[name]
	if !ok {
		if required {
			return errMissingField(name)
		}
		return nil
	}

	if s, chk := f.(string); chk {
		fields[name] = s
	} else {
		return errInvalidFieldType(name)
	}
	return nil
}

func checkStringField(name string, required bool, data map[string]interface{}) error {
	f, ok := data[name]
	if !ok {
		if required {
			return errMissingField(name)
		}
		return nil
	}
	if _ , chk := f.(string); !chk {
		return errInvalidFieldType(name)
	}
	return nil
}


// errInvalidFieldType is the error returned when a field in an
// event causing an email has an incorrect type.
func errInvalidFieldType(name string) error {
	return errors.New("invalid field type '" + name + "' in email event")
}

// errMissingField is the error returned when a required field in an
// event causing an email is missing.
func errMissingField(name string) error {
	return errors.New("required field '" + name + "' is missing")
}
