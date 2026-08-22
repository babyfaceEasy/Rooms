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
	GetPostsByRoomID(ctx context.Context, roomID primitive.ObjectID, page, limit int) ([]*domain.Post, int64, error)
	DeletePost(ctx context.Context, id, userID primitive.ObjectID) error
	ValidatePost(ctx context.Context, postID, userID primitive.ObjectID) (*domain.Post, error)
	RemoveValidation(ctx context.Context, postID, userID primitive.ObjectID) error
	RespectPost(ctx context.Context, postID, userID primitive.ObjectID) (*domain.Post, error)
	RemoveRespect(ctx context.Context, postID, userID primitive.ObjectID) error
}

type postService struct {
	repo     repository.PostRepository
	roomRepo repository.RoomRepository
}

// NewPostService creates a new post service
func NewPostService(repo repository.PostRepository, roomRepo repository.RoomRepository) PostService {
	return &postService{repo: repo, roomRepo: roomRepo}
}

// CreatePost creates a new post with validation
func (s *postService) CreatePost(ctx context.Context, text string, userID, roomID primitive.ObjectID, imageURL, videoURL, audioURL *string) (*domain.Post, error) {
	// Create post object
	now := time.Now()
	post := &domain.Post{
		RoomID:      roomID,
		UserID:      userID,
		Text:        text,
		Image:       imageURL,
		Video:       videoURL,
		Audio:       audioURL,
		Validations: []primitive.ObjectID{},
		CreatedAt:   now,
		UpdatedAt:   now,
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

// GetPostsByRoomID retrieves paginated posts for a specific room
func (s *postService) GetPostsByRoomID(ctx context.Context, roomID primitive.ObjectID, page, limit int) ([]*domain.Post, int64, error) {
	posts, total, err := s.repo.GetByRoomID(ctx, roomID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
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

// ValidatePost adds the user's validation to a post.
func (s *postService) ValidatePost(ctx context.Context, postID, userID primitive.ObjectID) (*domain.Post, error) {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Check if user already validated this post
	for _, v := range post.Validations {
		if v == userID {
			return nil, domain.ErrAlreadyValidated
		}
	}

	if err := s.repo.AddValidation(ctx, postID, userID); err != nil {
		return nil, err
	}

	// Return updated post
	return s.repo.GetByID(ctx, postID)
}

// RemoveValidation removes the user's validation from a post.
func (s *postService) RemoveValidation(ctx context.Context, postID, userID primitive.ObjectID) error {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Check if user has validated this post
	validated := false
	for _, v := range post.Validations {
		if v == userID {
			validated = true
			break
		}
	}
	if !validated {
		return domain.ErrNotValidated
	}

	return s.repo.RemoveValidation(ctx, postID, userID)
}

// RespectPost adds the user's respect to a post.
func (s *postService) RespectPost(ctx context.Context, postID, userID primitive.ObjectID) (*domain.Post, error) {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Check if user already respected this post
	for _, r := range post.Respects {
		if r == userID {
			return nil, domain.ErrAlreadyRespected
		}
	}

	if err := s.repo.AddRespect(ctx, postID, userID); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, postID)
}

// RemoveRespect removes the user's respect from a post.
func (s *postService) RemoveRespect(ctx context.Context, postID, userID primitive.ObjectID) error {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Check if user has respected this post
	respected := false
	for _, r := range post.Respects {
		if r == userID {
			respected = true
			break
		}
	}
	if !respected {
		return domain.ErrNotRespected
	}

	return s.repo.RemoveRespect(ctx, postID, userID)
}
