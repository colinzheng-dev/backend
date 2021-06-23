package model

// ItemCollectionInfo is a sort of raw JSON view of the JSON request
// used to create an item collection (and used to return information
// about an item collection). There are rules that determine what
// combinations of fields are allowed, depending on the collection
// type.
type ItemCollectionInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// ItemCollectionView is a view of a collection that includes IDs for
// manual collections.
type ItemCollectionView struct {
	ItemCollectionInfo
	IDs []string `json:"ids,omitempty"`
}
