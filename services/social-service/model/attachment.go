package model

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
)

// Attachment is used to represent blobs, products or orders/bookings attached to a message or thread
type AttachmentFixed struct {
	AttachType string             `json:"type"`
	Ref        string             `json:"ref"`
}
type Attachment struct {
	AttachmentFixed
	Attrs      chassis.GenericMap `json:"-,omitempty"`
}

func (a *Attachment) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in attachment data")
	}
	if err = chassis.StringField(&a.AttachType, fields, "type"); err != nil {
		return errors.Wrap(err, "invalid field type")
	}
	if err = chassis.StringField(&a.Ref, fields, "ref"); err != nil {
		return errors.Wrap(err, "invalid field ref")
	}

	a.Attrs = fields

	return nil
}


// MarshalJSON marshals an attachment to a full JSON view, including both
// fixed and type-specific fields. We do this by marshalling a view of
// the fixed fields to JSON, replacing the final close bracket of the
// JSON object with a comma and appending the raw JSON for the
// type-specific fields minus their opening bracket.
func (a *Attachment) MarshalJSON() ([]byte, error) {
	// Marshal fixed fields.
	jsonFixed, err := json.Marshal(a.AttachmentFixed)
	if err != nil {
		return nil, err
	}

	// Return right away if there are no attributes.
	if len(a.Attrs) == 0 {
		return jsonFixed, nil
	}
	// Compose into final JSON.
	var b bytes.Buffer
	b.Write(jsonFixed[:len(jsonFixed)-1])
	if len(a.Attrs) > 0 {
		// Marshall type-specific attributes.
		jsonAttrs, err := json.Marshal(a.Attrs)
		if err != nil {
			return nil, err
		}
		b.WriteByte(',')
		b.Write(jsonAttrs[1:len(jsonAttrs)-1])
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

type Attachments []Attachment
// Make the Attachments struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (a Attachments) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Make the Attachments struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (a *Attachments) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}
