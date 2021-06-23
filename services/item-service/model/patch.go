package model

import (
	"encoding/json"
	"github.com/veganbase/backend/chassis"
	"strings"

	"github.com/pkg/errors"
)

// Patch applies a patch represented as a JSON object to an item
// value. It does the following (which has a lot in common with the
// UnmarshalJSON method for Item):
//
// 1. Unmarshal JSON patch data to a generic map[string]interface{}
//    representation.
//
// 2. Check that no read-only fields are included in the patch.
//
// 3. Process the "fixed" fields for an Item (i.e. those fields that
//    do not depend on the item type) one by one, removing them from
//    the generic map as they're assigned to the Item value.
//
// 4. Extract the Attrs value from the item as a generic
//    map[string]interface{} value.
//
// 5. Merge the remaining fields from the patch into the type-specific
//    item attributes.
//
// 6. Marshal the updated attributes, back to a byte buffer.
//
// 7. Validate the patched attributes byte buffer against the item
//    type-specific JSON schema validator managed by the schema
//    package.
//
// 8. Do semantic validation of any complex compound fields (e.g.
//    opening hours).
//
// 9. If the attribute data validation is successful, then the
//    attributes field of the new Item value can be filled in from the
//    JSON data in the byte buffer.
func (it *Item) Patch(patch []byte) error {
	// Step 1.
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2.
	roFields := map[string]string{
		"id":         "ID",
		"item_type":  "type",
		"slug":       "slug",
		"approval":   "approval state",
		"creator":    "creator",
		"owner":      "owner",
		"ownership":  "ownership status",
		"created_at": "creation date",
	}
	for fld, label := range roFields {
		if _, ok := updates[fld]; ok {
			return errors.New("can't patch item " + label)
		}
	}

	// Step 3.
	if err = chassis.StringField(&it.Lang, updates, "lang"); err != nil {
		return err
	}
	if err = chassis.StringField(&it.Name, updates, "name"); err != nil {
		return err
	}
	if err = chassis.StringField(&it.Description, updates, "description"); err != nil {
		return err
	}
	if err = chassis.StringField(&it.FeaturedPicture, updates, "featured_picture"); err != nil {
		return err
	}
	if err = chassis.StringListField(&it.Tags, updates, "tags"); err != nil {
		return err
	}
	if err = chassis.StringListField(&it.Pictures, updates, "pictures"); err != nil {
		return err
	}
	if !it.checkFeaturedPicture() {
		return errors.New("featured_picture must be a member of pictures list")
	}
	if err = urlMapField(&it.URLs, updates); err != nil {
		return err
	}
	fixed := FixedValidate(it)
	fixedData, err := json.Marshal(fixed)
	if err != nil {
		return errors.Wrap(err, "marshalling fixed fields for validation")
	}
	res, err := Validate("item-fixed", fixedData)
	if err != nil {
		return errors.Wrap(err, "unmarshalling fixed item fields")
	}
	if !res.Valid() {
		msgs := []string{}
		for _, err := range res.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors for fixed item fields: " + strings.Join(msgs, "; "))
	}

	// Step 4.
	attrs := map[string]interface{}(it.Attrs)

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
	itt := it.ItemType.String()
	attrRes, err := Validate(itt, attrFields)
	if err != nil {
		return errors.Wrap(err, "processing item attributes for '"+itt+"'")
	}
	if !attrRes.Valid() {
		msgs := []string{}
		for _, err := range attrRes.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors in item attributes for '" + itt +
			"': " + strings.Join(msgs, "; "))
	}

	// Step 8.
	for k, v := range updates {
		val, ok := semanticValidators[k]
		if !ok {
			continue
		}
		if err := val(v); err != nil {
			return err
		}
	}

	// Step 9.
	it.Attrs = attrs

	return nil
}

// UpdateAvailability patches available_inventory_quantity field on attrs attribute.
func (it *Item) UpdateAvailability(patch []byte) error {
	// Step 1 - unmarshal fields to be updated
	updates := map[string]interface{}{}
	err := json.Unmarshal(patch, &updates)
	if err != nil {
		return errors.Wrap(err, "unmarshaling patch")
	}

	// Step 2 - turns attrs into an map of interfaces
	attrs := map[string]interface{}(it.Attrs)

	//Step 3 - getting current item availability
	var currentAvailability int
	if err = chassis.IntField(&currentAvailability, attrs, "available_quantity"); err != nil {
		return err
	}

	// Step 4 - getting value to be increased or decreased on the availability
	var value int
	if err = chassis.IntField(&value, updates, "quantity"); err != nil {
		return err
	}

	//Step 5 -summing them up (if value will decrease availability, it will be negative. positive otherwise)
	attrs["available_quantity"] = currentAvailability + value

	// Step 6 - updating item attrs attribute.
	it.Attrs = attrs

	return nil
}