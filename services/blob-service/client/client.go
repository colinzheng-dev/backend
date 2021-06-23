package client

// Client is the service client API for the blob service.
type Client interface {
	AddItemBlobs(itemID string, blobIDs []string) error
	RemoveItemBlobs(itemID string, blobIDs []string) error
}
