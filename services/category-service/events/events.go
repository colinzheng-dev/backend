package events

// Event names used in category service.
const (
	CategoryUpdate = "category-update"
)

// Category is a representation of a single category.
type Category map[string]interface{}

// CategoryMap is a map from category names to category information.
type CategoryMap map[string]*Category

// CategoryUpdateInfo is a message published when a category's entries
// change.
type CategoryUpdateInfo struct {
	Name    string   `json:"name"`
	Entries Category `json:"entries"`
}
