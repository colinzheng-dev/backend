package model

import (
	"encoding/json"
	"fmt"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/model"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/veganbase/backend/services/item-service/model/types"
)

// UnmarshalJSON is a validating unmarshaller for full item
// definitions for all item types. (This is used to unmarshal new item
// values for item creation, which is the only place we get send full
// JSON representations of items.) It does the following:
//
// 1. Unmarshal JSON data to a generic map[string]interface{}
//    representation.
//
// 2. Validate the "fixed" fields for an item (i.e. those fields that
//    do not depend on the item type), using a special JSON schema.
//
// 3. Process the "fixed" fields for an Item (i.e. those fields that
//    do not depend on the item type) one by one, removing them from
//    the generic map as they're added to a new Item value.
//
// 4. Marshal the remaining fields, which should contain the item
//    type-specific attributes fields, back to a byte buffer.
//
// 5. Validate the attributes byte buffer against the item
//    type-specific JSON schema validator managed by the schema
//    package.
//
// 6. Do semantic validation of any complex compound fields (e.g.
//    opening hours).
//
// 7. If the attribute data validation is successful, then the
//    attributes field of the new Item value can be filled in from the
//    JSON data in the byte buffer.
func (item *Item) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in item data")
	}

	// Part of Step 3: check that read-only fields aren't included.
	roBad := []string{}
	chassis.ReadOnlyField(fields, "id", &roBad)
	chassis.ReadOnlyField(fields, "slug", &roBad)
	chassis.ReadOnlyField(fields, "approval", &roBad)
	chassis.ReadOnlyField(fields, "creator", &roBad)
	chassis.ReadOnlyField(fields, "ownership", &roBad)
	chassis.ReadOnlyField(fields, "created_at", &roBad)
	if len(roBad) > 0 {
		return errors.New("attempt to set read-only fields: " + strings.Join(roBad, ","))
	}

	// Step 2.
	res, err := Validate("item-fixed", data)
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

	// Step 3. (We can ignore errors from stringField here because the
	// types have already been checked during schema validation. We keep
	// the checks in stringField too so that we can use it for checking
	// patches.)
	*item = Item{}
	chassis.StringField(&item.Owner, fields, "owner")
	chassis.StringField(&item.Lang, fields, "lang")
	chassis.StringField(&item.Name, fields, "name")
	chassis.StringField(&item.Description, fields, "description")
	chassis.StringField(&item.FeaturedPicture, fields, "featured_picture")
	tmp := ""
	chassis.StringField(&tmp, fields, "item_type")
	err = item.ItemType.FromString(tmp)
	if err != nil {
		return errors.New("invalid item type '" + tmp + "'")
	}
	if !item.ItemType.Creatable() {
		return errors.New("can't create items of abstract item type '" + tmp + "'")
	}
	if err = chassis.StringListField(&item.Tags, fields, "tags"); err != nil {
		return err
	}
	if err = chassis.StringListField(&item.Pictures, fields, "pictures"); err != nil {
		return err
	}
	// TODO: REMOVE THIS DEFAULT WHEN DASHBOARD IS UPDATED, AND ADD
	// "featured_picture" TO REQUIRED FIELDS IN item-fixed JSON SCHEMA
	if item.FeaturedPicture == "" {
		item.FeaturedPicture = item.Pictures[0]
	}
	if !item.checkFeaturedPicture() {
		return errors.New("featured_picture must be a member of pictures list")
	}
	if err = urlMapField(&item.URLs, fields); err != nil {
		return err
	}

	// Step 4.
	attrFields, err := json.Marshal(fields)
	if err != nil {
		return errors.Wrap(err, "marshalling attribute fields to JSON")
	}

	// Step 5.
	it := item.ItemType.String()
	attrRes, err := Validate(it, attrFields)
	if err != nil {
		return errors.Wrap(err, "unmarshalling item attributes for '"+it+"'")
	}
	if !attrRes.Valid() {
		msgs := []string{}
		for _, err := range attrRes.Errors() {
			msgs = append(msgs, err.String())
		}
		return errors.New("validation errors in item attributes for '" + it +
			"': " + strings.Join(msgs, "; "))
	}

	// Step 6.
	for k, v := range fields {
		val, ok := semanticValidators[k]
		if !ok {
			continue
		}
		if err := val(v); err != nil {
			return err
		}
	}

	// Step 7.
	item.Attrs = fields

	return nil
}



// UnrestrictedUnmarshalJSON unmarshal the response for purchase service.
func (item *Item) UnrestrictedUnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in item data")
	}

	*item = Item{}
	owner, _ := fields["owner"].(map[string]interface{})

	chassis.StringField(&item.ID, fields, "id")
	chassis.StringField(&item.Lang, fields, "lang")
	chassis.StringField(&item.Owner, owner, "id")
	chassis.StringField(&item.Name, fields, "name")
	chassis.StringField(&item.Description, fields, "description")
	chassis.StringField(&item.FeaturedPicture, fields, "featured_picture")
	tmp := ""
	chassis.StringField(&tmp, fields, "item_type")
	err = item.ItemType.FromString(tmp)
	if err != nil {
		return errors.New("invalid item type '" + tmp + "'")
	}
	if !item.ItemType.Creatable() {
		return errors.New("can't create items of abstract item type '" + tmp + "'")
	}
	if err = chassis.StringListField(&item.Tags, fields, "tags"); err != nil {
		return err
	}
	if err = chassis.StringListField(&item.Pictures, fields, "pictures"); err != nil {
		return err
	}
	// TODO: REMOVE THIS DEFAULT WHEN DASHBOARD IS UPDATED, AND ADD
	// "featured_picture" TO REQUIRED FIELDS IN item-fixed JSON SCHEMA
	if item.FeaturedPicture == "" {
		item.FeaturedPicture = item.Pictures[0]
	}
	if !item.checkFeaturedPicture() {
		return errors.New("featured_picture must be a member of pictures list")
	}
	if err = urlMapField(&item.URLs, fields); err != nil {
		return err
	}

	// Step 7.
	item.Attrs = fields

	return nil
}

// Process URL list field.
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

// Semantic validators for complex fields: all field values passed in
// to these validation functions have passed JSON schema validation,
// so we can rely on the structure required by the schemas and
// concentrate on checking only additional semantic constraints.
var semanticValidators = map[string]func(interface{}) error{
	"opening_hours": validateOpeningHours,
	"special_hours": validateSpecialHours,
}


func validateOpeningHours(oh interface{}) error {
	for _, ohi := range oh.([]interface{}) {
		oh := ohi.(map[string]interface{})

		periodByDay := make(map[int]Intervals)
		daysOpened := [7]bool{}

		for i := 0; i< 7; i++ {
			periodByDay[i] = Intervals{}
		}

		for _, periodi := range oh["periods"].([]interface{}) {
			period := periodi.(map[string]interface{})
			var p Period
			start := period["start"].(string)
			end, hasEnd := period["end"].(string)
			p.Day = int(period["day"].(float64))
			p.IsOvernight = period["is_overnight"].(bool)

			//STEP 1 - Check if time passed is valid
			if !checkDayTime(start) {
				return errors.New("invalid time 'start' in opening hours: " + start)
			}
			if hasEnd {
				if !checkDayTime(end) {
					return errors.New("invalid time 'end' in opening hours: " + end)
				}
			} else {
				end = "0000"
			}
			p.Start, _ = strconv.Atoi(start)
			p.End, _ = strconv.Atoi(end)
			//STEP 2 - check if flag is_overnight is correct
			if p.IsOvernight != checkOvernight(p.Start, p.End) {
				return errors.New("invalid 'is_overnight' flag in opening hours: start=" + start + ", end=" + end)
			}

			//STEP 3 - saving period by its day
			periodByDay[p.Day] = append(periodByDay[p.Day], p)
			daysOpened[p.Day] = true
		}
		//STEP 4 - check for overlapping opening hours
		for k, v := range periodByDay {
			var periods Intervals
			periods = v
			// checking the existence of overnight period on previous day.
			// For mondays (idx = 0), previous day will have idx = 6
			idxDayBefore := 6
			if k != 0 {
				idxDayBefore = k-1
			}
			//checking if there were opening hours on the day before
			if daysOpened[idxDayBefore] {
				dayBefore := periodByDay[idxDayBefore]
				for _, p := range dayBefore {
					//if any opening hour is overnight, we include period from 00:00 to the real closing time.
					if p.IsOvernight {
						newP := p
						newP.Start = 0
						periods = append(periods, newP)
					}
				}
			}
			//2. check for overlapping periods. assuming that each day may have at most 3 opening hours windows
			// we call checkOverlap that is O(nlogn)
			if err := periods.checkOverlap(periods); err != nil {
				return err
			}
		}

	}
	return nil
}

func checkDayTime(t string) bool {
	return checkHour(t[0:2]) && checkMinutes(t[2:4])
}

// Check hour values in times.
func checkHour(h string) bool {
	if h[0] == '0' {
		h = h[1:]
	}
	i, _ := strconv.Atoi(h)
	return i <= 23
}

// Check minute values in times.
func checkMinutes(m string) bool {
	if m[0] == '0' {
		m = m[1:]
	}
	i, _ := strconv.Atoi(m)
	return i <= 59
}

// Check hour values in times.
func checkOvernight(start, end int) bool {
	if end == 0 {
		return false
	}
	return start > end
}

// Period is a helper struct to holds opening_hours information with more suited typing
type Period struct {
	Day         int
	IsOvernight bool
	Start       int
	End         int
}

// Intervals implements sort.Interface based on the Start field.
type Intervals []Period

func (a Intervals) Len() int           { return len(a) }
func (a Intervals) Less(i, j int) bool { return a[i].Start < a[j].Start }
func (a Intervals) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (a Intervals) checkOverlap(interval []Period) error {

	// Sort intervals in increasing order of start time
	sort.Sort(a)

	// In the sorted slice, if start time of an interval is less than end of
	// previous interval,  then there is an overlap
	for i := 1; i < a.Len(); i++ {
		if a[i-1].End > a[i].Start {
			return errors.New("overlap between periods: "+ padZeros(a[i].Start)+ "-" + padZeros(a[i].End) +  " and " +
				padZeros(a[i-1].Start)+ "-" + padZeros(a[i-1].End))
		}
	}
	// If we reach here, then no overlap
	return nil
}

func padZeros(num int) string {
	return fmt.Sprintf("%04d", num)
}

func validateSpecialHours(sh interface{}) error {
	sphs := sh.([]interface{})
	periodByDay := make(map[int]Intervals)
	daysOpened := [366]bool{}

	for i := 0; i < 365; i++ {
		periodByDay[i] = Intervals{}
	}

	for _, special := range sphs {
		period := special.(map[string]interface{})
		var p Period
		start := period["start"].(string)
		end, hasEnd := period["end"].(string)
		month, day, err := parseDate(period["day"].(string))
		if err != nil {
			return errors.New("date validation: " + err.Error())
		}
		p.Day = monthDaysAcc[month - 1] + day - 1
		p.IsOvernight = period["is_overnight"].(bool)

		//STEP 1 - Check if time passed is valid
		if !checkDayTime(start) {
			return errors.New("invalid time 'start' in opening hours: " + start)
		}
		if hasEnd {
			if !checkDayTime(end) {
				return errors.New("invalid time 'end' in opening hours: " + end)
			}
		} else {
			end = "0000"
		}
		p.Start, _ = strconv.Atoi(start)
		p.End, _ = strconv.Atoi(end)
		//STEP 2 - check if flag is_overnight is correct
		if p.IsOvernight != checkOvernight(p.Start, p.End) {
			return errors.New("invalid 'is_overnight' flag in opening hours: start=" + start + ", end=" + end)
		}

		//STEP 3 - saving period by its day
		periodByDay[p.Day] = append(periodByDay[p.Day], p)
		daysOpened[p.Day] = true
	}

	//STEP 4 - check for overlapping opening hours
	for k, v := range periodByDay {
		var periods Intervals
		periods = v
		// checking the existence of overnight period on previous day.
		idxDayBefore := 365
		if k != 0 {
			idxDayBefore = k-1
		}
		//checking if there were opening hours on the day before
		if daysOpened[idxDayBefore] && k != 0 {
			dayBefore := periodByDay[idxDayBefore]
			for _, p := range dayBefore {
				//if any opening hour is overnight, we include period from 00:00 to the real closing time.
				if p.IsOvernight {
					newP := p
					newP.Start = 0
					periods = append(periods, newP)
				}
			}
		}
		//2. check for overlapping periods. assuming that each day may have at most 3 opening hours windows
		// we call checkOverlap that is O(nlogn)
		if err := periods.checkOverlap(periods); err != nil {
			return err
		}
	}

	return nil
}


// Monthly day counts and cumulative monthly day counts for opening
// hours checking (based on a leap year, since we only really care
// about overlaps and checking that dates are OK).
var monthDays = []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
var monthDaysAcc = []int{0, 31, 60, 91, 121, 152, 182, 213, 244, 274, 305, 335}


// Parse a MM-DD date string to give month and day indexes.
func parseDate(d string) (int, int, error) {
	ds := strings.Split(d, "-")
	month := ds[1]
	if month[0] == '0' {
		month = month[1:]
	}
	imonth, _ := strconv.Atoi(month)
	if imonth < 1 || imonth > 12 {
		return 0, 0, errors.New("invalid month in date string, got:" + month)
	}
	day := ds[2]
	if day[0] == '0' {
		day = day[1:]
	}
	iday, _ := strconv.Atoi(day)
	if iday < 1 || iday > monthDays[imonth-1] {
		return 0, 0, errors.New("invalid day in date string, got:" + day)
	}

	return imonth, iday, nil
}

// UnmarshalJSON this method will be used when we need to retrieve attrs by calling the endpoint internally
// currently this method doesn't unmarshall the object completely, only the required fields
func (i *ItemFull) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in item data")
	}

	*i = ItemFull{}
	// Step 2 - Format owner as UserModel.Info struct
	rawOwner, ok := fields["owner"]
	if ok {
		jsonOwner, err := json.Marshal(rawOwner)
		if err != nil {
			return err
		}
		owner := model.Info{}
		if err = json.Unmarshal(jsonOwner, &owner); err != nil {
			return err
		}
		i.Owner = &owner

	}

	chassis.StringField(&i.ID, fields, "id")
	chassis.StringField(&i.Slug, fields, "slug")
	tmp := ""
	chassis.StringField(&tmp, fields, "item_type")
	if err = i.ItemType.FromString(tmp); err != nil {
		return errors.New("invalid item type '" + tmp + "'")
	}

	chassis.StringField(&i.Lang, fields, "lang")
	chassis.StringField(&i.Name, fields, "name")
	chassis.StringField(&i.Description, fields, "description")
	chassis.StringField(&i.FeaturedPicture, fields, "featured_picture")

	// we are deleting this fields because they are not needed when calling internally
	delete(fields, "upvotes")
	delete(fields, "user_upvoted")
	delete(fields, "rank")
	delete(fields, "collections")
	delete(fields, "links")

	// Step 7.
	i.Attrs = fields

	return nil
}



// UnmarshalJSON this method will be used when we need to retrieve attrs by calling the endpoint internally
// currently this method doesn't unmarshall the object completely, only the required fields
func (i *ItemFullWithLink) UnmarshalJSON(data []byte) error {
	// Step 1.
	fields := map[string]interface{}{}
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return errors.New("invalid JSON in item data")
	}

	*i = ItemFullWithLink{}
	// Step 2 - Format owner as UserModel.Info struct
	rawOwner, ok := fields["owner"]
	if ok {
		jsonOwner, err := json.Marshal(rawOwner)
		if err != nil {
			return err
		}
		owner := model.Info{}
		if err = json.Unmarshal(jsonOwner, &owner); err != nil {
			return err
		}
		i.Owner = &owner

	}

	chassis.StringField(&i.ID, fields, "id")
	chassis.StringField(&i.Slug, fields, "slug")
	tmp := ""
	chassis.StringField(&tmp, fields, "item_type")
	if err = i.ItemType.FromString(tmp); err != nil {
		return errors.New("invalid item type '" + tmp + "'")
	}

	chassis.StringField(&i.Lang, fields, "lang")
	chassis.StringField(&i.Name, fields, "name")
	chassis.StringField(&i.Description, fields, "description")
	chassis.StringField(&i.FeaturedPicture, fields, "featured_picture")

	// we are deleting this fields because they are not needed when calling internally
	delete(fields, "upvotes")
	delete(fields, "user_upvoted")
	delete(fields, "rank")
	delete(fields, "collections")

	link := fields["links"]
	delete(fields, "links")

	var links []LinkLinkWithFull
	rawLink, err := json.Marshal(link)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(rawLink, &links); err != nil {
		return err
	}

	if len(links) > 0 {
		i.Link = &links[0]
	}

	// Step 7.
	i.Attrs = fields

	return nil
}