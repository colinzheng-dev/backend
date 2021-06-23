package model

import "github.com/jmoiron/sqlx/types"

// CategorySummary is a view used for the category summary route.
type CategorySummary struct {
	Label      string         `json:"label"`
	Extensible bool           `json:"extensible"`
	Schema     types.JSONText `json:"schema"`
}

// Summary generates a summary view of a category.
func Summary(cat *Category) *CategorySummary {
	return &CategorySummary{
		Label:      cat.Label,
		Extensible: cat.Extensible,
		Schema:     cat.Schema,
	}
}
