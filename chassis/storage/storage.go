package storage

// Storage describes the blob storage operations used by the blob
// service.
type Storage interface {
	// Write uploads data for a given blob ID to the blob storage.
	Write(id, format string, data []byte) (string, int64, error)

	// Delete deletes the data stored for a given blob ID.
	Delete(id, format string) error
}
