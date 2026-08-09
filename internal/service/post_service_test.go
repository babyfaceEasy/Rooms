package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"temp_backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockPostRepository is a mock implementation for testing
type MockPostRepository struct {
	createFunc      func(ctx context.Context, post *domain.Post) error
	getByIDFunc     func(ctx context.Context, id primitive.ObjectID) (*domain.Post, error)
	deletePostFunc  func(ctx context.Context, id primitive.ObjectID) error
	getByRoomIDFunc func(ctx context.Context, roomID primitive.ObjectID) ([]*domain.Post, error)
}

func (m *MockPostRepository) Create(ctx context.Context, post *domain.Post) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, post)
	}
	return nil
}

func (m *MockPostRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrPostNotFound
}

func (m *MockPostRepository) DeletePost(ctx context.Context, id primitive.ObjectID) error {
	if m.deletePostFunc != nil {
		return m.deletePostFunc(ctx, id)
	}
	return nil
}

func (m *MockPostRepository) GetByRoomID(ctx context.Context, roomID primitive.ObjectID) ([]*domain.Post, error) {
	if m.getByRoomIDFunc != nil {
		return m.getByRoomIDFunc(ctx, roomID)
	}
	return []*domain.Post{}, nil
}

func TestCreatePost_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	text := "This is my first post"

	repoMock := &MockPostRepository{
		createFunc: func(ctx context.Context, post *domain.Post) error {
			post.ID = primitive.NewObjectID()
			return nil
		},
	}

	svc := NewPostService(repoMock)
	post, err := svc.CreatePost(context.Background(), text, userID, roomID, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, text, post.Text)
	assert.Equal(t, userID, post.UserID)
	assert.Equal(t, roomID, post.RoomID)
	assert.Nil(t, post.Image)
	assert.Nil(t, post.Video)
	assert.Nil(t, post.Audio)
}

func TestCreatePost_EmptyText(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()

	repoMock := &MockPostRepository{}

	svc := NewPostService(repoMock)
	post, err := svc.CreatePost(context.Background(), "", userID, roomID, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, domain.ErrPostTextRequired, err)
}

func TestCreatePost_TextTooLong(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	longText := ""
	for i := 0; i < 5001; i++ {
		longText += "a"
	}

	repoMock := &MockPostRepository{}

	svc := NewPostService(repoMock)
	post, err := svc.CreatePost(context.Background(), longText, userID, roomID, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, domain.ErrPostTextTooLong, err)
}

func TestCreatePost_WithImage(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	text := "Check out this image!"
	imageURL := "https://s3.amazonaws.com/bucket/image.jpg"

	repoMock := &MockPostRepository{
		createFunc: func(ctx context.Context, post *domain.Post) error {
			post.ID = primitive.NewObjectID()
			return nil
		},
	}

	svc := NewPostService(repoMock)
	post, err := svc.CreatePost(context.Background(), text, userID, roomID, &imageURL, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.NotNil(t, post.Image)
	assert.Equal(t, imageURL, *post.Image)
}

func TestCreatePost_WithVideo(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	text := "Check out this video!"
	videoURL := "https://s3.amazonaws.com/bucket/video.mp4"

	repoMock := &MockPostRepository{
		createFunc: func(ctx context.Context, post *domain.Post) error {
			post.ID = primitive.NewObjectID()
			return nil
		},
	}

	svc := NewPostService(repoMock)
	post, err := svc.CreatePost(context.Background(), text, userID, roomID, nil, &videoURL, nil)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.NotNil(t, post.Video)
	assert.Equal(t, videoURL, *post.Video)
}

func TestCreatePost_RepositoryError(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	text := "This post will fail"

	repoMock := &MockPostRepository{
		createFunc: func(ctx context.Context, post *domain.Post) error {
			return errors.New("database error")
		},
	}

	svc := NewPostService(repoMock)
	post, err := svc.CreatePost(context.Background(), text, userID, roomID, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, post)
}

func TestGetPost_Success(t *testing.T) {
	postID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	post := &domain.Post{
		ID:        postID,
		UserID:    userID,
		Text:      "Test post",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repoMock := &MockPostRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
			if id == postID {
				return post, nil
			}
			return nil, domain.ErrPostNotFound
		},
	}

	svc := NewPostService(repoMock)
	retrievedPost, err := svc.GetPost(context.Background(), postID)

	assert.NoError(t, err)
	assert.NotNil(t, retrievedPost)
	assert.Equal(t, postID, retrievedPost.ID)
}

func TestGetPost_NotFound(t *testing.T) {
	repoMock := &MockPostRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
			return nil, domain.ErrPostNotFound
		},
	}

	svc := NewPostService(repoMock)
	post, err := svc.GetPost(context.Background(), primitive.NewObjectID())

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, domain.ErrPostNotFound, err)
}

func TestDeletePost_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	postID := primitive.NewObjectID()
	post := &domain.Post{
		ID:        postID,
		UserID:    userID,
		Text:      "Test post",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repoMock := &MockPostRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
			if id == postID {
				return post, nil
			}
			return nil, domain.ErrPostNotFound
		},
		deletePostFunc: func(ctx context.Context, id primitive.ObjectID) error {
			return nil
		},
	}

	svc := NewPostService(repoMock)
	err := svc.DeletePost(context.Background(), postID, userID)

	assert.NoError(t, err)
}

func TestDeletePost_NotOwner(t *testing.T) {
	ownerID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	postID := primitive.NewObjectID()
	post := &domain.Post{
		ID:        postID,
		UserID:    ownerID,
		Text:      "Test post",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repoMock := &MockPostRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
			if id == postID {
				return post, nil
			}
			return nil, domain.ErrPostNotFound
		},
	}

	svc := NewPostService(repoMock)
	err := svc.DeletePost(context.Background(), postID, otherUserID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUnauthorizedPost, err)
}

func TestDeletePost_NotFound(t *testing.T) {
	repoMock := &MockPostRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
			return nil, domain.ErrPostNotFound
		},
	}

	svc := NewPostService(repoMock)
	err := svc.DeletePost(context.Background(), primitive.NewObjectID(), primitive.NewObjectID())

	assert.Error(t, err)
	assert.Equal(t, domain.ErrPostNotFound, err)
}
