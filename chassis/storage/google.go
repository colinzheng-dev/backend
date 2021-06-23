package storage

import (
	"context"
	"fmt"

	google "cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// GoogleStorage represents blob storage in a Google Cloud Storage
// bucket.
type GoogleStorage struct {
	ctx       context.Context
	blobstore *google.Client
	bucket    *google.BucketHandle
}

// NewGoogleClient creates a new connection to Google Cloud Storage
// for blob storage.
func NewGoogleClient(ctx context.Context,
	credsPath string, bucket string) (*GoogleStorage, error) {
	client := &GoogleStorage{ctx: ctx}

	// Connect to blob storage.
	c, err := google.NewClient(ctx, option.WithCredentialsFile(credsPath))
	if err != nil {
		return nil, err
	}

	client.blobstore = c
	client.bucket = c.Bucket(bucket)
	return client, nil
}

// Write writes data to the Google Cloud Storage bucket using a name
// and file extension to generate the object name.
func (s *GoogleStorage) Write(id, format string, data []byte) (string, int64, error) {
	blob := s.bucket.Object(fmt.Sprintf("%v.%v", id, format))
	w := blob.NewWriter(s.ctx)
	if _, err := w.Write(data); err != nil {
		return "", 0, err
	}
	if err := w.Close(); err != nil {
		return "", 0, err
	}

	attrs, err := blob.Attrs(s.ctx)
	if err != nil {
		return "", 0, err
	}

	return attrs.MediaLink, attrs.Size, nil
}

// Delete deletes a blob with a given ID and file type from the Google
// Cloud Storage bucket.
func (s *GoogleStorage) Delete(id, format string) error {
	blob := s.bucket.Object(fmt.Sprintf("%v.%v", id, format))
	return blob.Delete(s.ctx)
}
