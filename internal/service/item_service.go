package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ItemService orchestrates item metadata and file storage.
type ItemService interface {
	Create(ctx context.Context, name, description string, file *multipart.FileHeader) (*domain.Item, error)
	GetByID(ctx context.Context, id string) (*domain.Item, error)
	List(ctx context.Context, page, limit int) ([]domain.Item, error)
	Update(ctx context.Context, id, name, description string, file *multipart.FileHeader) (*domain.Item, error)
	Delete(ctx context.Context, id string) error
	Download(ctx context.Context, id string) (io.ReadCloser, *domain.Item, error)
}

type itemService struct {
	items   repository.ItemRepository
	storage repository.ObjectStorage
}

// NewItemService creates a new ItemService.
func NewItemService(items repository.ItemRepository, storage repository.ObjectStorage) ItemService {
	return &itemService{
		items:   items,
		storage: storage,
	}
}

func (s *itemService) Create(ctx context.Context, name, description string, file *multipart.FileHeader) (*domain.Item, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required: %w", domain.ErrInvalidInput)
	}

	item := &domain.Item{
		Name:        name,
		Description: strings.TrimSpace(description),
	}

	if err := s.items.Create(ctx, item); err != nil {
		return nil, err
	}

	if file != nil {
		if err := s.attachFile(ctx, item, file); err != nil {
			// Best-effort rollback: if the upload failed, remove the empty record.
			_ = s.items.Delete(ctx, item.ID)
			return nil, err
		}
	}

	return item, nil
}

func (s *itemService) GetByID(ctx context.Context, id string) (*domain.Item, error) {
	oid, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}
	return s.items.GetByID(ctx, oid)
}

func (s *itemService) List(ctx context.Context, page, limit int) ([]domain.Item, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return s.items.List(ctx, int64(page), int64(limit))
}

func (s *itemService) Update(ctx context.Context, id, name, description string, file *multipart.FileHeader) (*domain.Item, error) {
	oid, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}

	item, err := s.items.GetByID(ctx, oid)
	if err != nil {
		return nil, err
	}

	if name = strings.TrimSpace(name); name != "" {
		item.Name = name
	}
	if description = strings.TrimSpace(description); description != "" {
		item.Description = description
	}

	if file != nil {
		// Replace existing file if present.
		if item.FileKey != "" {
			_ = s.storage.RemoveObject(ctx, item.FileKey)
		}
		if err := s.attachFile(ctx, item, file); err != nil {
			return nil, err
		}
	}

	item.UpdatedAt = time.Now().UTC()
	if err := s.items.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *itemService) Delete(ctx context.Context, id string) error {
	oid, err := parseObjectID(id)
	if err != nil {
		return err
	}

	item, err := s.items.GetByID(ctx, oid)
	if err != nil {
		return err
	}

	if item.FileKey != "" {
		if err := s.storage.RemoveObject(ctx, item.FileKey); err != nil {
			return fmt.Errorf("failed to remove object %q: %w", item.FileKey, err)
		}
	}

	return s.items.Delete(ctx, oid)
}

func (s *itemService) Download(ctx context.Context, id string) (io.ReadCloser, *domain.Item, error) {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if item.FileKey == "" {
		return nil, nil, fmt.Errorf("item has no file: %w", domain.ErrInvalidInput)
	}

	obj, err := s.storage.GetObject(ctx, item.FileKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve object %q: %w", item.FileKey, err)
	}
	return obj, item, nil
}

func (s *itemService) attachFile(ctx context.Context, item *domain.Item, file *multipart.FileHeader) error {
	key := buildObjectKey(item.ID, file.Filename)

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(path.Ext(file.Filename))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	uploadedKey, err := s.storage.PutObject(ctx, key, src, file.Size, contentType)
	if err != nil {
		return fmt.Errorf("failed to store object %q: %w", key, err)
	}

	item.FileKey = uploadedKey
	item.FileName = file.Filename
	item.ContentType = contentType
	item.Size = file.Size
	item.UpdatedAt = time.Now().UTC()
	return nil
}

func parseObjectID(id string) (primitive.ObjectID, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.ObjectID{}, fmt.Errorf("invalid id %q: %w", id, domain.ErrInvalidInput)
	}
	return oid, nil
}

func buildObjectKey(itemID primitive.ObjectID, filename string) string {
	ext := path.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}
	base := strings.TrimSuffix(filename, ext)
	safe := sanitizeFilename(base)
	uid := primitive.NewObjectID().Hex()
	return fmt.Sprintf("items/%s/%s-%s%s", itemID.Hex(), uid[:8], safe, ext)
}

func sanitizeFilename(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case (r >= 'a' && r <= 'z'),
			(r >= 'A' && r <= 'Z'),
			(r >= '0' && r <= '9'),
			r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		return "file"
	}
	return out
}
