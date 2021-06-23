package model

import (
	"encoding/json"
	"github.com/veganbase/backend/chassis"
	"strings"

	sqlx_types "github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
	"github.com/veganbase/backend/services/user-service/model/types"
)

// Patch applies a patch represented as a JSON object to a user value.
// This takes a JSON patch value, the user ID of the user making the
// change and the admin flag of the user making the change (the two
// latter are needed to manage updates to the IsAdmin flag).
func (u *User) Patch(patch []byte, actingID string, adminAction bool) error {
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}
	roFields := map[string]string{
		"id":         "ID",
		"email":      "email",
		"last_login": "last login time",
		"api_key":    "API key",
	}
	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch user " + label)
		}
	}
	if _, ok := updates["is_admin"]; ok && !adminAction {
		return errors.New("can't patch user admin flag")
	}
	if v, ok := updates["is_admin"]; ok {
		val, ok := v.(bool)
		if !ok {
			return errors.New("invalid value for is_admin flag")
		}
		// An update to the is_admin flag can only be performed by an
		// administrator (already checked above), and an administrator is
		// not allowed to change their own is_admin flag.
		if u.ID == actingID {
			return errors.New("administrators are not allowed to change their own admin status")
		}
		u.IsAdmin = val
	}
	if err := stringUpdate(updates, "name", u.Name); err != nil {
		return err
	}
	if err := stringUpdate(updates, "display_name", u.DisplayName); err != nil {
		return err
	}
	if err := stringUpdate(updates, "avatar", u.Avatar); err != nil {
		return err
	}
	if err := stringUpdate(updates, "country", u.Country); err != nil {
		return err
	}
	return nil
}

// Patch applies a patch represented as a JSON object to an
// organisation value.
func (org *Organisation) Patch(patch []byte) error {
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}
	roFields := map[string]string{
		"id":         "ID",
		"slug":       "slug",
		"created_at": "creation time",
	}
	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch organisation " + label)
		}
	}
	if err := stringUpdate(updates, "name", &org.Name); err != nil {
		return err
	}
	if err := stringUpdate(updates, "description", &org.Description); err != nil {
		return err
	}
	if err := stringUpdate(updates, "logo", &org.Logo); err != nil {
		return err
	}
	if v, ok := updates["address"]; ok {
		jsonRaw, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		org.Address = sqlx_types.JSONText(jsonRaw)
	}
	if err := stringUpdate(updates, "phone", &org.Phone); err != nil {
		return err
	}
	if err := stringUpdate(updates, "email", &org.Email); err != nil {
		return err
	}
	if v, ok := updates["urls"]; ok {
		tmp, err := json.Marshal(v)
		if err != nil {
			return errors.New("invalid JSON for urls")
		}
		urls := types.URLMap{}
		err = json.Unmarshal(tmp, &urls)
		if err != nil {
			return errors.New("invalid value for urls")
		}
		org.URLs = urls
	}
	if v, ok := updates["industry"]; ok {
		ival, ok := v.([]interface{})
		if !ok {
			return errors.New("can't convert industry to array")
		}
		industry := []string{}
		for _, i := range ival {
			istr, ok := i.(string)
			if !ok {
				return errors.New("can't convert industry entry to string")
			}
			industry = append(industry, istr)
		}
		org.Industry = industry
	}
	if v, ok := updates["year_founded"]; ok {
		yval, ok := v.(float64)
		if !ok {
			return errors.New("can't convert year_founded to number")
		}
		org.YearFounded = uint(yval)
	}
	if v, ok := updates["employees"]; ok {
		eval, ok := v.(float64)
		if !ok {
			return errors.New("can't convert employees to number")
		}
		org.Employees = uint(eval)
	}

	// TODO: VALIDATE THE PATCHED ORGANISATION AGAINST THE JSON SCHEMA.
	return nil
}

// Patch applies a patch represented as a JSON object to an payout account value.
func (u *PayoutAccount) Patch(patch []byte) error {
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}
	roFields := map[string]string{
		"id":         "ID",
		"created_at": "created_at",
		"owner":      "owner",
	}

	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch user's payout" + label)
		}
	}
	if err := stringUpdate(updates, "account", &u.AccountNumber); err != nil {
		return err
	}
	return nil
}

// Patch applies a patch represented as a JSON object to an payout account value.
func (u *PaymentMethod) Patch(patch []byte) error {
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}
	roFields := map[string]string{
		"id":         "ID",
		"pm_id":      "pm_id",
		"user_id":    "user_id",
		"other_info": "other_info",
		"type":       "type",
		"created_at": "created_at",
	}

	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch user's payout" + label)
		}
	}
	if err := stringUpdate(updates, "description", &u.Description); err != nil {
		return err
	}

	//getting boolean value for is_default field
	if v, ok := updates["is_default"]; ok {
		s, ok := v.(bool)
		if !ok {
			return errors.New("non-boolean value for '" + "is_default" + "'")
		}
		u.IsDefault = s
	}

	return nil
}

// Patch applies a patch represented as a JSON object to an address value.
func (a *Address) Patch(patch []byte) error {
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}
	roFields := map[string]string{
		"id":         "ID",
		"owner":      "owner",
		"created_at": "created_at",
	}

	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch user's address" + label)
		}
	}

	if err := stringUpdate(updates, "description", &a.Description); err != nil {
		return err
	}
	if err := stringUpdate(updates, "street_address", &a.StreetAddress); err != nil {
		return err
	}
	if err := stringUpdate(updates, "city", &a.City); err != nil {
		return err
	}
	if err := stringUpdate(updates, "postcode", &a.Postcode); err != nil {
		return err
	}
	if err := stringUpdate(updates, "country", &a.Country); err != nil {
		return err
	}
	if err := stringUpdate(updates, "house_number", &a.HouseNumber); err != nil {
		return err
	}
	if err := stringUpdate(updates, "region_postal", &a.RegionPostalCode); err != nil {
		return err
	}

	//getting boolean value for is_default field
	if v, ok := updates["is_default"]; ok {
		s, ok := v.(bool)
		if !ok {
			return errors.New("non-boolean value for '" + "is_default" + "'")
		}
		a.IsDefault = s
	}
	if v, ok := updates["recipient"]; ok {
		c, err := json.Marshal(v)
		if err != nil {
			return errors.New("error parsing JSON for '" + "coordinates" + "'")
		}
		if err = json.Unmarshal(c, &a.Recipient); err != nil {
			return err
		}
	}

	//Validating address with new fields
	addressValidateRaw, err := json.Marshal(a.AddressValidate)
	if err != nil {
		return errors.Wrap(err, "marshalling user-address fields for validation")
	}
	res, err := Validate("user-address", addressValidateRaw)
	if err != nil {
		return errors.Wrap(err, "unmarshalling user-address fields")
	}
	if !res.Valid() {
		msgs := []string{}
		for _, err := range res.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors for user-address fields: " + strings.Join(msgs, "; "))
	}

	return nil
}

// Patch applies a patch represented as a JSON object to an delivery fees configuration value.
func (df *DeliveryFees) Patch(patch []byte) error {
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}
	roFields := map[string]string{
		"id":         "ID",
		"owner":      "owner",
		"created_at": "created_at",
	}

	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch user's delivery fees " + label)
		}
	}

	if err := chassis.StringField(&df.Currency, updates, "currency", ); err != nil {
		return err
	}
	if err := chassis.IntField(&df.FreeDeliveryAbove, updates, "free_delivery_above"); err != nil {
		return err
	}
	if err := chassis.IntField(&df.NormalOrderPrice, updates, "normal_order_price"); err != nil {
		return err
	}
	if err := chassis.IntField(&df.ChilledOrderPrice, updates, "chilled_order_price"); err != nil {
		return err
	}
	return nil
}

func stringUpdate(updates map[string]interface{}, k string, dst *string) error {
	v, ok := updates[k]
	if !ok {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return errors.New("non-string value for '" + k + "'")
	}
	*dst = s
	return nil
}

func optStringUpdate(updates map[string]interface{}, k string, dst **string) error {
	v, ok := updates[k]
	if !ok {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return errors.New("non-string value for '" + k + "'")
	}
	*dst = &s
	return nil
}
