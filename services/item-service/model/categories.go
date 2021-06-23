package model

import (
	"github.com/rs/zerolog/log"
	category_client "github.com/veganbase/backend/services/category-service/client"
)

// CategoryFormatChecker is a JSON schema format checker for strings
// that should come from the labels of a category.
type CategoryFormatChecker struct {
	client category_client.Client
	name   string
}

// NewCategoryChecker creates a new format checker for the labels of a
// category.
func NewCategoryChecker(client category_client.Client, name string) *CategoryFormatChecker {
	return &CategoryFormatChecker{client, name}
}

// IsFormat implements the gojsonschema.FormatChecker interface.
func (f *CategoryFormatChecker) IsFormat(input interface{}) bool {
	s, ok := input.(string)
	if !ok {
		return false
	}

	return f.client.IsValidLabel(f.name, s)
}

// LoadCategoryCheckers creates format checkers for categories.
func LoadCategoryCheckers(c category_client.Client) map[string]*CategoryFormatChecker {
	checkers := map[string]*CategoryFormatChecker{}
	for name := range c.Categories() {
		log.Info().Str("name", "category:"+name).Msg("adding schema format checker")
		checkers["category:"+name] = NewCategoryChecker(c, name)
	}
	return checkers
}
