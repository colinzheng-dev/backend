package chassis

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"github.com/jmoiron/sqlx/types"
)

// THIS IS A TEMPORARY SOLUTION FOR MESSAGING. WE'LL SWITCH TO USING
// PROTOCOL BUFFERS AT SOME POINT.

//GenericMap will be used inside messages that'll have an array of maps in its attributes
type GenericMap map[string]interface{}

// Scan implements the sql.Scanner interface.
func (m *GenericMap) Scan(src interface{}) error {
	j := types.JSONText{}
	err := j.Scan(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, m)
}

// Value implements the driver.Value interface.
func (m GenericMap) Value() (driver.Value, error) {
	v, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return types.JSONText(v).Value()
}


type FixedFields struct {
	Email    string `json:"email"`
	Language string `json:"language"`
	Site     string `json:"site"`
}

// LoginEmailRequestMsg is the message published when a login email is
// requested by a user.
type LoginEmailRequestMsg struct {
	FixedFields
	LoginToken string `json:"login_token"`
}


type GenericEmailMsg struct {
	FixedFields
	Data GenericMap `json:"data"`
}

func (msg *GenericEmailMsg) MarshalJSON() ([]byte, error) {
	// Marshal fixed fields.
	jsonFixed, err := json.Marshal(msg.FixedFields)
	if err != nil {
		return nil, err
	}

	// Return right away if there are no attributes or links.
	if len(msg.Data) == 0 {
		return jsonFixed, nil
	}
	// Compose into final JSON.
	var b bytes.Buffer
	b.Write(jsonFixed[:len(jsonFixed)-1])
	if len(msg.Data) > 0 {
		// Marshall type-specific attributes.
		jsonData, err := json.Marshal(msg.Data)
		if err != nil {
			return nil, err
		}
		b.WriteByte(',')
		b.Write(jsonData[1 : len(jsonData)-1])
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}
