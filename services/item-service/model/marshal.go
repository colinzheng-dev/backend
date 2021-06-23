package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// MarshalJSON marshals an item to a full JSON view, including both
// fixed and type-specific fields. We do this by marshalling a view of
// the fixed fields to JSON, replacing the final close bracket of the
// JSON object with a comma and appending the raw JSON for the
// type-specific fields minus their opening bracket.
func (view *ItemFull) MarshalJSON() ([]byte, error) {
	// Marshal fixed fields.
	jsonFixed, err := json.Marshal(view.ItemFullFixed)
	if err != nil {
		return nil, err
	}

	// Return right away if there are no attributes or links.
	//if len(view.Attrs) == 0 && view.Links == nil{
	//	return jsonFixed, nil
	//}
	// Compose into final JSON.
	var b bytes.Buffer
	b.Write(jsonFixed[:len(jsonFixed)-1])
	if len(view.Attrs) > 0 {
		// Marshall type-specific attributes.
		jsonAttrs, err := json.Marshal(view.Attrs)
		if err != nil {
			return nil, err
		}
		b.WriteByte(',')
		b.Write(jsonAttrs[1:len(jsonAttrs)-1])
	}

	if view.Links != nil {
		b.Write([]byte(`,"links":`))
		jsonLinks, err := json.Marshal(view.Links)
		if err != nil {
			return nil, err
		}
		b.Write(jsonLinks)

	}
	b.Write([]byte(`, "upvotes": ` + strconv.Itoa(view.Upvotes)))
	b.Write([]byte(`, "user_upvoted": ` + strconv.FormatBool(view.UserUpvoted)))
	jsonCollections, err := json.Marshal(view.Collections)
	b.Write([]byte(`, "rank": ` + fmt.Sprintf("%g", view.Rank)))
	b.Write([]byte(`, "collections": `))
	b.Write(jsonCollections)
	b.WriteByte('}')
	return b.Bytes(), nil
}
