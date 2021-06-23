package model

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"strings"
	"time"
)
//Address holds all information needed to send parcels to an user
type Address struct {
	ID    string `json:"id" db:"id"`
	Owner string `json:"owner" db:"owner"`
	AddressValidate
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
//AddressValidate holds all Address fields that need to be validated against a schema
type AddressValidate struct {
	Description      string        `json:"description" db:"description"`
	StreetAddress    string        `json:"street_address" db:"street_address"`
	City             string        `json:"city" db:"city"`
	Postcode         string        `json:"postcode" db:"postcode"`
	Country          string        `json:"country" db:"country"`
	RegionPostalCode string        `json:"region_postal" db:"region_postal"`
	HouseNumber      string        `json:"house_number" db:"house_number"`
	IsDefault        bool          `json:"is_default" db:"is_default"`
	Recipient        RecipientInfo `json:"recipient,omitempty" db:"recipient"`
	Coordinates      Coordinates   `json:"coordinates,omitempty" db:"coordinates"`
}

//RecipientInfo holds all contact information about who the parcel is address to
type RecipientInfo struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Company   string `json:"company"`
	Email     string `json:"contact_email"`
	Telephone string `json:"contact_phone"`
}

//Coordinates holds the geographical coordinates of an address
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (a *Address) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in address data")
	}

	//Part of Step 3: check that read-only fields aren't included.
	roBad := []string{}
	chassis.ReadOnlyField(fields, "id", &roBad)
	chassis.ReadOnlyField(fields, "owner", &roBad)
	chassis.ReadOnlyField(fields, "created_at", &roBad)
	if len(roBad) > 0 {
		return errors.New("attempt to set read-only fields: " + strings.Join(roBad, ","))
	}

	// Step 2.
	res, err := Validate("user-address", data)
	if err != nil {
		return errors.Wrap(err, "unmarshalling fixed user address fields")
	}
	if !res.Valid() {
		msgs := []string{}
		for _, err := range res.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors for user address fields: " + strings.Join(msgs, "; "))
	}

	*a = Address{}
	chassis.StringField(&a.Description, fields, "description")
	chassis.StringField(&a.StreetAddress, fields, "street_address")
	chassis.StringField(&a.City, fields, "city")
	chassis.StringField(&a.Postcode, fields, "postcode")
	chassis.StringField(&a.Country, fields, "country")
	chassis.StringField(&a.RegionPostalCode, fields, "region_postal")
	chassis.StringField(&a.HouseNumber, fields, "house_number")
	chassis.BoolField(&a.IsDefault, fields, "is_default")

	rawRecipient := fields["recipient"]
	recipient, err := json.Marshal(rawRecipient)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(recipient, &a.Recipient); err != nil {
		return err
	}

	rawCoordinates := fields["coordinates"]
	c, err := json.Marshal(rawCoordinates)
	if err != nil {

	}
	if err = json.Unmarshal(c, &a.Coordinates); err != nil {
		return err
	}

	return nil
}

//UnrestrictedUnmarshalJSON will unmarshal a JSON to Address object without validate any field.
//This method is used when another service is calling user-service to acquire this object.
func (a *Address) UnrestrictedUnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in address data")
	}

	//Step 2. Assign every field directly to struct
	*a = Address{}
	chassis.StringField(&a.ID, fields, "id")
	chassis.StringField(&a.Owner, fields, "owner")
	chassis.StringField(&a.Description, fields, "description")
	chassis.StringField(&a.StreetAddress, fields, "street_address")
	chassis.StringField(&a.City, fields, "city")
	chassis.StringField(&a.Postcode, fields, "postcode")
	chassis.StringField(&a.RegionPostalCode, fields, "region_postal")
	chassis.StringField(&a.Country, fields, "country")
	chassis.StringField(&a.HouseNumber, fields, "house_number")
	chassis.BoolField(&a.IsDefault, fields, "is_default")
	chassis.TimeField(&a.CreatedAt, fields, "created_at")

	rawRecipient := fields["recipient"]
	recipient, err := json.Marshal(rawRecipient)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(recipient, &a.Recipient); err != nil {
		return err
	}

	rawCoordinates := fields["coordinates"]
	c, err := json.Marshal(rawCoordinates)
	if err != nil {

	}
	if err = json.Unmarshal(c, &a.Coordinates); err != nil {
		return err
	}

	return nil
}

// Make the RecipientInfo struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (r RecipientInfo) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Make the RecipientInfo struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (r *RecipientInfo) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &r)
}

// Make the Coordinates struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (c Coordinates) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Make the Coordinates struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (c *Coordinates) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &c)
}