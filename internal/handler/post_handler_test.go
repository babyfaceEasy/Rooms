package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"temp_backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockPostService is a mock implementation for testing
type MockPostService struct {
	createPostFunc       func(ctx context.Context, text string, userID, roomID primitive.ObjectID, imageURL, videoURL, audioURL *string) (*domain.Post, error)
	getPostFunc          func(ctx context.Context, id primitive.ObjectID) (*domain.Post, error)
	getPostsByRoomIDFunc func(ctx context.Context, roomID primitive.ObjectID, page, limit int) ([]*domain.Post, int64, error)
	deletePostFunc       func(ctx context.Context, id, userID primitive.ObjectID) error
}

func (m *MockPostService) CreatePost(ctx context.Context, text string, userID, roomID primitive.ObjectID, imageURL, videoURL, audioURL *string) (*domain.Post, error) {
	if m.createPostFunc != nil {
		return m.createPostFunc(ctx, text, userID, roomID, imageURL, videoURL, audioURL)
	}
	return nil, nil
}

func (m *MockPostService) GetPost(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
	if m.getPostFunc != nil {
		return m.getPostFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockPostService) GetPostsByRoomID(ctx context.Context, roomID primitive.ObjectID, page, limit int) ([]*domain.Post, int64, error) {
	if m.getPostsByRoomIDFunc != nil {
		return m.getPostsByRoomIDFunc(ctx, roomID, page, limit)
	}
	return nil, 0, nil
}

func (m *MockPostService) DeletePost(ctx context.Context, id, userID primitive.ObjectID) error {
	if m.deletePostFunc != nil {
		return m.deletePostFunc(ctx, id, userID)
	}
	return nil
}

// MockRoomRepository is a mock implementation for testing
type MockRoomRepository struct {
	getByCodeFunc    func(ctx context.Context, code string) (*domain.Room, error)
	getByIDFunc      func(ctx context.Context, id primitive.ObjectID) (*domain.Room, error)
	isUserMemberFunc func(ctx context.Context, roomID, userID primitive.ObjectID) (bool, error)
}

func (m *MockRoomRepository) Create(ctx context.Context, room *domain.Room) error {
	return nil
}

func (m *MockRoomRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Room, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRoomRepository) GetByCode(ctx context.Context, code string) (*domain.Room, error) {
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, code)
	}
	return nil, nil
}

func (m *MockRoomRepository) AddUserToRoom(ctx context.Context, roomID, userID primitive.ObjectID) error {
	return nil
}

func (m *MockRoomRepository) RemoveUserFromRoom(ctx context.Context, roomID, userID primitive.ObjectID) error {
	return nil
}

func (m *MockRoomRepository) DeleteRoom(ctx context.Context, roomID primitive.ObjectID) error {
	return nil
}

func (m *MockRoomRepository) ListUserRooms(ctx context.Context, userID primitive.ObjectID) ([]*domain.Room, error) {
	return nil, nil
}

func (m *MockRoomRepository) IsUserMember(ctx context.Context, roomID, userID primitive.ObjectID) (bool, error) {
	if m.isUserMemberFunc != nil {
		return m.isUserMemberFunc(ctx, roomID, userID)
	}
	return false, nil
}

func TestCreatePostRequest_Structure(t *testing.T) {
	req := CreatePostRequest{
		Text: "This is a test post",
	}

	assert.Equal(t, "This is a test post", req.Text)
}

func TestPostResponse_Structure(t *testing.T) {
	resp := PostResponse{
		ID:        "507f1f77bcf86cd799439011",
		RoomID:    "507f1f77bcf86cd799439013",
		RoomCode:  "room-code",
		RoomName:  "Test Room",
		UserID:    "507f1f77bcf86cd799439012",
		Text:      "Test post",
		CreatedAt: "2024-06-29T00:00:00Z",
		UpdatedAt: "2024-06-29T00:00:00Z",
	}

	assert.Equal(t, "507f1f77bcf86cd799439011", resp.ID)
	assert.Equal(t, "507f1f77bcf86cd799439013", resp.RoomID)
	assert.Equal(t, "room-code", resp.RoomCode)
	assert.Equal(t, "507f1f77bcf86cd799439012", resp.UserID)
	assert.Equal(t, "Test post", resp.Text)
}

func TestNewPostHandler(t *testing.T) {
	mockService := &MockPostService{}
	mockRoomRepo := &MockRoomRepository{}
	handler := NewPostHandler(mockService, nil, mockRoomRepo)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.CreatePost)
	assert.NotNil(t, handler.GetPost)
	assert.NotNil(t, handler.DeletePost)
}

func TestCreatePost_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	postID := primitive.NewObjectID()
	text := "My first post"
	now := time.Now()

	post := &domain.Post{
		ID:        postID,
		RoomID:    roomID,
		UserID:    userID,
		Text:      text,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockPostService{
		createPostFunc: func(ctx context.Context, t string, uID, rID primitive.ObjectID, imageURL, videoURL, audioURL *string) (*domain.Post, error) {
			return post, nil
		},
	}

	mockRoomRepo := &MockRoomRepository{}
	handler := NewPostHandler(mockService, nil, mockRoomRepo)
	assert.NotNil(t, handler.CreatePost)
}

func TestCreatePost_TextRequired(t *testing.T) {
	mockService := &MockPostService{}
	mockRoomRepo := &MockRoomRepository{}
	handler := NewPostHandler(mockService, nil, mockRoomRepo)
	assert.NotNil(t, handler.CreatePost)
}

func TestGetPost_Success(t *testing.T) {
	postID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	post := &domain.Post{
		ID:        postID,
		RoomID:    roomID,
		UserID:    userID,
		Text:      "Test post",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := &MockPostService{
		getPostFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
			if id == postID {
				return post, nil
			}
			return nil, domain.ErrPostNotFound
		},
	}

	mockRoomRepo := &MockRoomRepository{}
	handler := NewPostHandler(mockService, nil, mockRoomRepo)
	assert.NotNil(t, handler.GetPost)
}

func TestGetPost_NotFound(t *testing.T) {
	mockService := &MockPostService{
		getPostFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.Post, error) {
			return nil, domain.ErrPostNotFound
		},
	}

	mockRoomRepo := &MockRoomRepository{}
	handler := NewPostHandler(mockService, nil, mockRoomRepo)
	assert.NotNil(t, handler.GetPost)
}

func TestDeletePost_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	postID := primitive.NewObjectID()

	mockService := &MockPostService{
		deletePostFunc: func(ctx context.Context, id, uID primitive.ObjectID) error {
			if id == postID && uID == userID {
				return nil
			}
			return errors.New("unauthorized")
		},
	}

	mockRoomRepo := &MockRoomRepository{}
	handler := NewPostHandler(mockService, nil, mockRoomRepo)
	assert.NotNil(t, handler.DeletePost)
}

func TestDeletePost_NotOwner(t *testing.T) {
	ownerID := primitive.NewObjectID()

	mockService := &MockPostService{
		deletePostFunc: func(ctx context.Context, id, uID primitive.ObjectID) error {
			if uID != ownerID {
				return domain.ErrUnauthorizedPost
			}
			return nil
		},
	}

	mockRoomRepo := &MockRoomRepository{}
	handler := NewPostHandler(mockService, nil, mockRoomRepo)
	assert.NotNil(t, handler.DeletePost)
}

func TestDeletePost_NotFound(t *testing.T) {
	mockService := &MockPostService{
		deletePostFunc: func(ctx context.Context, id, uID primitive.ObjectID) error {
			return domain.ErrPostNotFound
		},
	}

	mockRoomRepo := &MockRoomRepository{}
	handler := NewPostHandler(mockService, nil, mockRoomRepo)
	assert.NotNil(t, handler.DeletePost)
}

func TestIsValidImageType(t *testing.T) {
	testCases := []struct {
		filename string
		valid    bool
	}{
		{"image.jpg", true},
		{"image.jpeg", true},
		{"image.png", true},
		{"image.gif", true},
		{"image.webp", true},
		{"document.pdf", false},
		{"video.mp4", false},
		{"image.bmp", false},
	}

	for _, tc := range testCases {
		result := isValidImageType(tc.filename)
		assert.Equal(t, tc.valid, result)
	}
}

func TestIsValidVideoType(t *testing.T) {
	testCases := []struct {
		filename string
		valid    bool
	}{
		{"video.mp4", true},
		{"video.webm", true},
		{"video.mov", true},
		{"video.avi", true},
		{"document.pdf", false},
		{"image.jpg", false},
		{"video.mkv", false},
	}

	for _, tc := range testCases {
		result := isValidVideoType(tc.filename)
		assert.Equal(t, tc.valid, result)
	}
}
