package model

import (
	"encoding/json"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	t "github.com/veganbase/backend/services/social-service/model/types"
	"strings"
	"time"
)

//Thread is a representation of a conversation between entities (users and orgs)
type Thread struct {
	ID           string         `db:"id" json:"id"`
	Subject      string         `db:"subject" json:"subject"`
	Author       string         `db:"author" json:"author"`
	Content      string         `db:"content" json:"content"`
	Attachments  Attachments    `db:"attachments" json:"attachments"`
	LockReply    bool           `db:"lock_reply" json:"lock_reply"`
	Participants pq.StringArray `db:"participants" json:"participants"`
	Status       t.ThreadStatus `db:"status" json:"status"`
	IsEdited     bool           `db:"is_edited" json:"is_edited"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

type Message struct {
	ID          string      `db:"id" json:"id"`
	ParentID    string      `db:"parent_id" json:"parent_id"`
	Author      string      `db:"author" json:"author"`
	Content     string      `db:"content" json:"content"`
	Attachments Attachments `db:"attachments" json:"attachments"`
	IsEdited    bool        `db:"is_edited" json:"is_edited"`
	IsDeleted   bool        `db:"is_deleted" json:"is_deleted"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
}

func (t *Thread) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in thread data")
	}

	// Part of Step 3: check that read-only fields aren't included.
	roBad := []string{}
	chassis.ReadOnlyField(fields, "id", &roBad)
	chassis.ReadOnlyField(fields, "owner", &roBad)
	chassis.ReadOnlyField(fields, "created_at", &roBad)
	if len(roBad) > 0 {
		return errors.New("attempt to set read-only fields: " + strings.Join(roBad, ","))
	}
	*t = Thread{}

	chassis.StringField(&t.Subject, fields, "subject")
	chassis.StringListField(&t.Participants, fields, "participants")
	chassis.BoolField(&t.LockReply, fields, "lock_reply")

	chassis.StringField(&t.Content, fields, "content")

	if v, ok := fields["attachments"]; ok {
		jsonAttach, err := json.Marshal(v)
		if err != nil {
			return errors.Wrap(err, "marshalling attribute attachments to JSON")
		}
		if err = json.Unmarshal(jsonAttach, &t.Attachments); err != nil {
			return errors.Wrap(err, "unmarshaling attribute attachments")
		}
	}

	return nil
}

func (m *Message) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in thread data")
	}

	// Part of Step 3: check that read-only fields aren't included.
	roBad := []string{}
	chassis.ReadOnlyField(fields, "id", &roBad)
	chassis.ReadOnlyField(fields, "author", &roBad)
	chassis.ReadOnlyField(fields, "parent_id", &roBad)
	chassis.ReadOnlyField(fields, "created_at", &roBad)
	if len(roBad) > 0 {
		return errors.New("attempt to set read-only fields: " + strings.Join(roBad, ","))
	}

	chassis.StringField(&m.Content, fields, "content")

	if v, ok := fields["attachments"]; ok {
		jsonAttach, err := json.Marshal(v)
		if err != nil {
			return errors.Wrap(err, "marshalling attribute attachments to JSON")
		}
		if err = json.Unmarshal(jsonAttach, &m.Attachments); err != nil {
			return errors.Wrap(err, "unmarshaling attribute attachments")
		}
	}

	chassis.BoolField(&m.IsEdited, fields, "is_edited")

	return nil
}

func (m *Message) Patch(patch []byte) error {

	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	roFields := map[string]string{
		"id":         "ID",
		"parent_id":  "parent_id",
		"author":     "author",
		"is_edited":  "is_edited",
		"is_deleted": "is_deleted",
		"created_at": "creation date",
	}
	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch message's " + label)
		}
	}

	if err = chassis.StringField(&m.Content, updates, "content"); err != nil {
		return err
	}

	if v, ok := updates["attachments"]; ok {
		jsonAttach, err := json.Marshal(v)
		if err != nil {
			return errors.Wrap(err, "marshalling attribute attachments to JSON")
		}
		if err = json.Unmarshal(jsonAttach, &m.Attachments); err != nil {
			return errors.Wrap(err, "unmarshaling attribute attachments")
		}
	}

	return nil
}

func (t *Thread) Patch(patch []byte) error {

	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	roFields := map[string]string{
		"id":           "ID",
		"subject":      "subject",
		"author":       "author",
		"status":       "status",
		"is_edited":    "is_edited",
		"participants": "participants",
		"created_at":   "creation date",
	}
	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch thread's " + label)
		}
	}

	status := ""
	if err = chassis.StringField(&status, updates, "status"); err != nil {
		return err
	}
	if status != "" {
		if err = t.Status.FromString(status); err != nil {
			return err
		}
	}

	if err = chassis.StringField(&t.Content, updates, "content"); err != nil {
		return err
	}

	if err = chassis.BoolField(&t.LockReply, updates, "lock_reply"); err != nil {
		return err
	}

	if v, ok := updates["attachments"]; ok {
		jsonAttach, err := json.Marshal(v)
		if err != nil {
			return errors.Wrap(err, "marshalling attribute attachments to JSON")
		}
		if err = json.Unmarshal(jsonAttach, &t.Attachments); err != nil {
			return errors.Wrap(err, "unmarshaling attribute attachments")
		}
	}

	return nil
}
