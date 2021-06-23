package model

import (
	"encoding/json"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"strings"
	"time"
)

type Webhook struct {
	ID        string         `json:"id" db:"id"`
	Owner     string         `json:"owner" db:"owner"`
	URL       string         `json:"url" db:"url"`
	Enabled   bool           `json:"enabled" db:"enabled"`
	LiveMode  bool           `json:"livemode" db:"livemode"`
	Events    pq.StringArray `json:"events" db:"events"`
	Secret    string         `json:"secret" db:"secret"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

func (w *Webhook) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in webhook data")
	}

	//Part of Step 3: check that read-only fields aren't included.
	roBad := []string{}
	chassis.ReadOnlyField(fields, "id", &roBad)
	chassis.ReadOnlyField(fields, "secret", &roBad)
	chassis.ReadOnlyField(fields, "created_at", &roBad)
	if len(roBad) > 0 {
		return errors.New("attempt to set read-only fields: " + strings.Join(roBad, ","))
	}

	*w = Webhook{}
	chassis.StringField(&w.Owner, fields, "owner")
	chassis.StringField(&w.URL, fields, "url")
	chassis.BoolField(&w.Enabled, fields, "enabled")
	chassis.BoolField(&w.LiveMode, fields, "livemode")
	chassis.StringListField(&w.Events, fields, "events")

	return nil
}

// Patch applies a patch represented as a JSON object to an address value.
func (w *Webhook) Patch(patch []byte) error {
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}
	roFields := map[string]string{
		"id":         "ID",
		"secret":     "secret",
		"created_at": "created_at",
	}

	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch webhook's field" + label)
		}
	}

	if err := chassis.StringField(&w.URL, updates, "url"); err != nil {
		return err
	}
	if err := chassis.StringField(&w.Owner, updates, "owner"); err != nil {
		return err
	}
	if err := chassis.BoolField(&w.LiveMode, updates, "livemode"); err != nil {
		return err
	}
	if err := chassis.BoolField(&w.Enabled, updates, "enabled"); err != nil {
		return err
	}
	if err := chassis.StringListField(&w.Events, updates, "events"); err != nil {
		return err
	}

	return nil
}
