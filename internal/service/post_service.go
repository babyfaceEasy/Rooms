package service

import (
	"context"
	"time"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostService defines the interface for post business logic
type PostService interface {
	CreatePost(ctx context.Context, text string, userID, roomID primitive.ObjectID, imageURL, videoURL, audioURL *string) (*domain.Post, error)
	GetPost(ctx context.Context, id primitive.ObjectID) (*domain.Post, error)
	GetPostsByRoomID(ctx context.Context, roomID primitive.ObjectID) ([]*domain.Post, error)
	DeletePost(ctx context.Context, id, userID primitive.ObjectID) error
}

type postService struct {
	repo repository.PostRepository
}

// NewPostService creates a new post service
func NewPostService(repo repository.PostRepository) PostService {
	return &postService{repo: repo}
}

// CreatePost creates a new post with validation
func (s *postService) CreatePost(ctx context.Context, text string, userID, roomID primitive.ObjectID, imageURL, videoURL, audioURL *string) (*domain.Post, error) {
	// Create post object
	now := time.Now()
	post := &domain.Post{
		RoomID:    roomID,
		UserID:    userID,
		Text:      text,
		Image:     imageURL,
		Video:     videoURL,
		Audio:     audioURL,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate post
	if err := post.Validate(); err != nil {
		return nil, err
	}

	// Create post in repository
	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// GetPost retrieves a post by ID
func (s *postService) GetPost(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return post, nil
}

// GetPostsByRoomID retrieves all posts for a specific room
func (s *postService) GetPostsByRoomID(ctx context.Context, roomID primitive.ObjectID) ([]*domain.Post, error) {
	posts, err := s.repo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// DeletePost soft-deletes a post (only creator can delete)
func (s *postService) DeletePost(ctx context.Context, id, userID primitive.ObjectID) error {
	// Get post to verify ownership
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user is the creator
	if post.UserID != userID {
		return domain.ErrUnauthorizedPost
	}

	// Soft delete the post
	if err := s.repo.DeletePost(ctx, id); err != nil {
		return err
	}

	return nil
}
