package client

import "github.com/veganbase/backend/services/category-service/events"

// Client is the service client API for the site service.
type Client interface {
	// Sites returns the current category list (which is populated and
	// updated asynchronously based on events from the category
	// service).
	Categories() events.CategoryMap

	// IsValidLabel determines whether a label is valid for a category.
	IsValidLabel(catName string, label string) bool
}
