package storage

import (
	"github.com/veganbase/backend/services/blob-service/model"
)

type stored struct {
	format string
	size   int64
}

type MockStorage struct {
	B map[string]stored
}

func NewMockStorage(blobs *map[string]*model.Blob) *MockStorage {
	store := map[string]stored{}
	for id, blob := range *blobs {
		store[id] = stored{blob.Format, int64(blob.Size)}
	}
	return &MockStorage{store}
}

func (mock *MockStorage) Write(id, format string, data []byte) (string, int64, error) {
	mock.B[id] = stored{format, int64(len(data))}
	return "http://dummy", int64(len(data)), nil
}

func (mock *MockStorage) Delete(id, format string) error {
	delete(mock.B, id)
	return nil
}
