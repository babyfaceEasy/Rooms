package repository

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ObjectStorage defines a thin abstraction over an S3-compatible object store
// (MinIO for local development, S3 in production).
type ObjectStorage interface {
	PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error)
	RemoveObject(ctx context.Context, key string) error
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
}

type s3Repository struct {
	client *s3.Client
	bucket string
}

// NewS3Repository returns an ObjectStorage implementation backed by the supplied S3 client.
func NewS3Repository(client *s3.Client, bucket string) ObjectStorage {
	return &s3Repository{
		client: client,
		bucket: bucket,
	}
}

func (s *s3Repository) PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	return key, nil
}

func (s *s3Repository) RemoveObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *s3Repository) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}
