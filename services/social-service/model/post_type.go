
package model

import (
"database/sql/driver"
"encoding/json"
"strings"

"github.com/pkg/errors"
)

// PostType is an enumeration that represents all the known types of
// posts.
type PostType int

// Constants for all item types.
const (
	UnknownPost PostType = iota
	QuestionPost
	ReviewPost
	CommentPost
)

// UnmarshalJSON unmarshals a post type from a JSON string.
func (post *PostType) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal post type")
	}
	return post.FromString(s)
}

// FromString converts a string to an post type.
func (post *PostType) FromString(s string) error {
	switch strings.ToLower(s) {
	default:
		return errors.New("unknown item type '" + s + "'")
	case "question":
		*post = QuestionPost
	case "review":
		*post = ReviewPost
	case "comment":
		*post = CommentPost
	}
	return nil
}

// String converts a post type from its internal representation to a string.
func (post PostType) String() string {
	switch post {
	default:
		return "<unknown post type>"
	case QuestionPost:
		return "question"
	case ReviewPost:
		return "review"
	case CommentPost:
		return "comment"
	}
}

// MarshalJSON converts an internal post type to JSON.
func (post PostType) MarshalJSON() ([]byte, error) {
	s := post.String()
	if s == "<unknown post type>" {
		return nil, errors.New("unknown post type")
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface.
func (post *PostType) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for PostType")
	}
	return post.FromString(s)
}

// Value implements the driver.Value interface.
func (post PostType) Value() (driver.Value, error) {
	return post.String(), nil
}

// ID prefixes by post type.
var PostTypeIDPrefixes = map[PostType]string{
	QuestionPost: "qst",
	ReviewPost: "rvw",
	CommentPost: "cmt",
}
