package repository

import (
	"context"
	"errors"
	"io"
)

// NoOpStorage is a no-operation storage implementation used when S3 is disabled.
// It allows the application to run without file storage capabilities.
type NoOpStorage struct{}

// NewNoOpStorage creates a new no-operation storage repository.
func NewNoOpStorage() ObjectStorage {
	return &NoOpStorage{}
}

// PutObject is a no-op implementation that returns an error.
func (n *NoOpStorage) PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	// Consume the reader to avoid issues
	if reader != nil {
		_, _ = io.Copy(io.Discard, reader)
	}
	return "", errors.New("file storage is disabled")
}

// RemoveObject is a no-op implementation that returns an error.
func (n *NoOpStorage) RemoveObject(ctx context.Context, key string) error {
	return errors.New("file storage is disabled")
}

// GetObject is a no-op implementation that returns an error.
func (n *NoOpStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, errors.New("file storage is disabled")
}
