package service

import (
	"context"
	"fmt"
	"time"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentService defines the interface for comment business logic
type CommentService interface {
	CreateComment(ctx context.Context, postID, userID primitive.ObjectID, text string) (*domain.Comment, error)
	GetComment(ctx context.Context, id primitive.ObjectID) (*domain.Comment, error)
	GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page, limit int, sortOrder string) ([]*domain.Comment, int64, error)
	DeleteComment(ctx context.Context, id, userID primitive.ObjectID) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	roomRepo    repository.RoomRepository
}

// NewCommentService creates a new CommentService
func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository, roomRepo repository.RoomRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		roomRepo:    roomRepo,
	}
}

// CreateComment creates a new comment on a post
func (s *commentService) CreateComment(ctx context.Context, postID, userID primitive.ObjectID, text string) (*domain.Comment, error) {
	// Validate post exists
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", domain.ErrPostNotFound)
	}

	if post == nil {
		return nil, domain.ErrPostNotFound
	}

	// Create comment
	comment := &domain.Comment{
		PostID:    postID,
		UserID:    userID,
		Text:      text,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Validate comment
	if err := comment.ValidateComment(); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

// GetComment retrieves a comment by ID
func (s *commentService) GetComment(ctx context.Context, id primitive.ObjectID) (*domain.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetCommentsByPostID retrieves paginated comments for a post
func (s *commentService) GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page, limit int, sortOrder string) ([]*domain.Comment, int64, error) {
	comments, total, err := s.commentRepo.GetByPostID(ctx, postID, page, limit, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// DeleteComment deletes a comment (owner or room owner)
func (s *commentService) DeleteComment(ctx context.Context, id, userID primitive.ObjectID) error {
	// Verify user is the comment owner or room owner
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user is the comment owner
	if comment.UserID == userID {
		return s.commentRepo.DeleteComment(ctx, id)
	}

	// Check if user is the room owner
	post, err := s.postRepo.GetByID(ctx, comment.PostID)
	if err != nil {
		return err
	}

	room, err := s.roomRepo.GetByID(ctx, post.RoomID)
	if err != nil {
		return err
	}

	if room.CreatedBy != userID {
		return domain.ErrUnauthorizedComment
	}

	return s.commentRepo.DeleteComment(ctx, id)
}
