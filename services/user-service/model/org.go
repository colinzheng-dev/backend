package model

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"time"

	sqlx_types "github.com/jmoiron/sqlx/types"

	"github.com/lib/pq"
	"github.com/veganbase/backend/services/user-service/model/types"
)

// Organisation represents an organisation, which may be composed of
// multiple users, and may own items.
type Organisation struct {
	// Unique ID of the organisation.
	ID string `json:"id" db:"id"`
	// Slug for use in URLs derived from organisation name.
	Slug string `json:"slug" db:"slug"`
	// Human-readable name for organisation.
	Name string `json:"name" db:"name"`
	// Full description of organisation.
	Description string `json:"description" db:"description"`
	// Logo for organisation.
	Logo string `json:"logo,omitempty" db:"logo"`
	// Address for the organisation.
	Address sqlx_types.JSONText `json:"address,omitempty" db:"address"`
	// Contact phone number for the organisation.
	Phone string `json:"contact_phone,omitempty" db:"phone"`
	// Contact email address for the organisation.
	Email string `json:"contact_email,omitempty" db:"email"`
	// URLs associated with organisation, along with their link types (e.g. website, Facebook, etc.)
	URLs types.URLMap `json:"urls,omitempty" db:"urls"`
	// Industry labels: these are taken from the "org-industry" category.
	Industry pq.StringArray `json:"industry,omitempty" db:"industry"`
	// (Optional) year the organisation was founded.
	YearFounded uint `json:"year_founded,omitempty" db:"year_founded"`
	// (Optional) count of employees in the the organisation.
	Employees uint `json:"employees,omitempty" db:"employees"`
	// Creation time of organisation.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}


func (o *Organisation) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in organisation data")
	}

	// Part of Step 2: check that read-only fields aren't included.
	//roBad := []string{}
	//TODO: OrgWithUserInfo uses Organisation as a sub struct,
	// making read-only validations break in many other places
	//chassis.ReadOnlyField(fields, "id", &roBad)
	//chassis.ReadOnlyField(fields, "owner", &roBad)
	//chassis.ReadOnlyField(fields, "created_at", &roBad)
	//if len(roBad) > 0 {
	//	return errors.New("attempt to set read-only fields: " + strings.Join(roBad, ","))
	//}

	// Step 3 - validate against schema.
	//res, err := Validate("organisation", data)
	//if err != nil {
	//	return errors.Wrap(err, "unmarshalling fixed organisation fields")
	//}
	//if !res.Valid() {
	//	msgs := []string{}
	//	for _, err := range res.Errors() {
	//		msgs = append(msgs, err.String())
	//	}
	//	return errors.New("validation errors for organisation fields: " + strings.Join(msgs, "; "))
	//}

	//Step 4 - Create object (we are ignoring errors because they were already checked by schema)
	*o = Organisation{}
	//TODO: REMOVE WHEN FIX VALIDATIONS
	chassis.StringField(&o.ID, fields, "id")
	chassis.StringField(&o.Slug, fields, "slug")

	chassis.StringField(&o.Name, fields, "name")
	chassis.StringField(&o.Description, fields, "description")
	chassis.StringField(&o.Logo, fields, "logo")
	chassis.StringField(&o.Phone, fields, "contact_phone")
	chassis.StringField(&o.Email, fields, "contact_email")
	chassis.StringListField(&o.Industry, fields, "industry")

	chassis.UintField(&o.YearFounded, fields, "year_founded")
	chassis.UintField(&o.Employees, fields, "employees")

	if err = urlMapField(&o.URLs, fields); err != nil {
		return err
	}

	address, ok := fields["address"]
	if !ok {
		return nil
	}

	o.Address, _ = json.Marshal(address)

	return nil
}
//TODO: URLMap may be moved to chassis, because it is reused on other services
func urlMapField(dst *types.URLMap, fields map[string]interface{}) error {
	val, ok := fields["urls"]
	if !ok {
		if *dst == nil {
			empty := types.URLMap{}
			*dst = empty
		}
		return nil
	}
	urls, ok := val.(map[string]interface{})
	if !ok {
		return errors.New("invalid JSON for 'urls' field")
	}

	res := map[types.URLType]string{}
	for k, v := range urls {
		t := types.UnknownURL
		if err := t.FromString(k); err != nil {
			return err
		}
		s, ok := v.(string)
		if !ok {
			return errors.New("non-string value for URL '" + k + "'")
		}
		res[t] = s
	}

	*dst = res
	delete(fields, "urls")

	return nil
}

// OrgUser represents the membership of a user in an organisation.
type OrgUser struct {
	// ID is a unique integer ID for the user/organisation association.
	ID int `db:"id"`

	// OrgID is the organisation ID.
	OrgID string `db:"org_id"`

	// UserID is the user ID.
	UserID string `db:"user_id"`

	// IsOrgAdmin is used to indicate users that are organisation
	// administrators.
	IsOrgAdmin bool `db:"is_org_admin"`

	// CreatedAt is a creation timestamp.
	CreatedAt time.Time `db:"created_at"`
}

