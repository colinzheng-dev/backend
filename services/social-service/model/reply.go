package model

import (
	"bytes"
	"encoding/json"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"strings"
	"time"
)

type ReplyFixed struct {
	Id        string             `db:"id" json:"id"`
	ParentId  string             `db:"parent_id" json:"parent_id"`
	Owner     string             `db:"owner" json:"owner"`
	IsEdited  bool               `db:"is_edited" json:"is_edited"`
	Pictures  pq.StringArray     `db:"pictures" json:"pictures"`
	IsDeleted bool               `db:"is_deleted" json:"is_deleted"`
	CreatedAt time.Time          `db:"created_at" json:"created_at"`
}
type Reply struct {
	ReplyFixed
	Attrs     chassis.GenericMap `db:"attrs" json:"attrs"`
}

func (rpl *Reply) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in item data")
	}

	// Part of Step 3: check that read-only fields aren't included.
	roBad := []string{}
	readOnlyField(fields, "id", &roBad)
	readOnlyField(fields, "owner", &roBad)
	readOnlyField(fields, "created_at", &roBad)
	if len(roBad) > 0 {
		return errors.New("attempt to set read-only fields: " + strings.Join(roBad, ","))
	}

	// Step 2 - validating fixed fields with reply-fixed schema
	res, err := Validate("reply-fixed", data)
	if err != nil {
		return errors.Wrap(err, "unmarshalling fixed post fields")
	}
	if !res.Valid() {
		msgs := []string{}
		for _, err := range res.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors for fixed post fields: " + strings.Join(msgs, "; "))
	}

	// Step 3. (We can ignore errors from stringField here because the
	// types have already been checked during schema validation. We keep
	// the checks in stringField too so that we can use it for checking
	// patches.)
	*rpl = Reply{}
	stringField(&rpl.ParentId, fields, "parent_id")
	boolField(&rpl.IsEdited, fields, "is_edited")
	boolField(&rpl.IsDeleted, fields, "is_deleted")


	if err = stringListField(&rpl.Pictures, fields, "pictures"); err != nil {
		return err
	}

	// Step 4.
	attrFields, err := json.Marshal(fields)
	if err != nil {
		return errors.Wrap(err, "marshalling attribute fields to JSON")
	}

	// Step 5.

	attrRes, err := Validate("reply", attrFields)
	if err != nil {
		return errors.Wrap(err, "unmarshalling reply attributes")
	}
	if !attrRes.Valid() {
		msgs := []string{}
		for _, err := range attrRes.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors in reply attributes: " + strings.Join(msgs, "; "))
	}

	// Step 6
	rpl.Attrs = fields

	return nil
}

func (rpl *Reply) Patch(patch []byte) error {
	// Step 1.
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2.
	roFields := map[string]string{
		"id":         "ID",
		"owner":      "owner",
		"subject":    "subject",
		"post_type":  "post_type",
		"created_at": "creation date",
	}
	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch post " + label)
		}
	}

	// Step 3.
	if err = boolField(&rpl.IsEdited, updates, "is_edited"); err != nil {
		return err
	}
	if err = boolField(&rpl.IsDeleted, updates, "is_deleted"); err != nil {
		return err
	}

	if err = stringListField(&rpl.Pictures, updates, "pictures"); err != nil {
		return err
	}
	//validating fixed post fields
	fixed := GetReplyFixedValidate(rpl)
	fixedData, err := json.Marshal(fixed)
	if err != nil {
		return errors.Wrap(err, "marshalling fixed fields for validation")
	}
	res, err := Validate("reply-fixed", fixedData)
	if err != nil {
		return errors.Wrap(err, "unmarshalling fixed reply fields")
	}
	if !res.Valid() {
		msgs := []string{}
		for _, err := range res.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors for fixed reply fields: " + strings.Join(msgs, "; "))
	}

	// Step 4.
	attrs := map[string]interface{}(rpl.Attrs)

	// Step 5.
	for k, v := range updates {
		attrs[k] = v
	}

	// Step 6.
	attrFields, err := json.Marshal(attrs)
	if err != nil {
		return errors.New("couldn't marshal patched attributes back to JSON")
	}

	// Step 7.
	attrRes, err := Validate("reply", attrFields)
	if err != nil {
		return errors.Wrap(err, "processing reply attributes")
	}
	if !attrRes.Valid() {
		msgs := []string{}
		for _, err := range attrRes.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors in reply attributes " + strings.Join(msgs, "; "))
	}

	// Step 8.
	rpl.Attrs = attrs

	return nil
}


func (r *Reply) MarshalJSON() ([]byte, error) {
	// Marshal fixed fields.
	jsonFixed, err := json.Marshal(r.ReplyFixed)
	if err != nil {
		return nil, err
	}

	// Compose into final JSON.
	var b bytes.Buffer
	b.Write(jsonFixed[:len(jsonFixed)-1])
	if len(r.Attrs) > 0 {
		// Marshall type-specific attributes.
		jsonAttrs, err := json.Marshal(r.Attrs)
		if err != nil {
			return nil, err
		}
		b.WriteByte(',')
		b.Write(jsonAttrs[1 : len(jsonAttrs)-1])
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}