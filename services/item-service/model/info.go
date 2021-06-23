package model


// Info is a minimal view of item information used in
// various places as a helper.
type Info struct {
	ID    string  `json:"id"`
	Slug  *string `json:"slug,omitempty"`
	Name  *string `json:"name"`
}