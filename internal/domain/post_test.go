package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func TestValidatePost_Success(t *testing.T) {
	post := &Post{
		ID:        primitive.NewObjectID(),
		UserID:    primitive.NewObjectID(),
		Text:      "This is a valid post",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.NoError(t, post.Validate())
}

func TestValidatePost_EmptyText(t *testing.T) {
	post := &Post{
		ID:        primitive.NewObjectID(),
		UserID:    primitive.NewObjectID(),
		Text:      "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Error(t, post.Validate())
	assert.Equal(t, ErrPostTextRequired, post.Validate())
}

func TestValidatePost_TextTooLong(t *testing.T) {
	longText := ""
	for i := 0; i < 5001; i++ {
		longText += "a"
	}

	post := &Post{
		ID:        primitive.NewObjectID(),
		UserID:    primitive.NewObjectID(),
		Text:      longText,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Error(t, post.Validate())
	assert.Equal(t, ErrPostTextTooLong, post.Validate())
}

func TestValidatePost_WithImage(t *testing.T) {
	imageURL := "https://s3.amazonaws.com/bucket/image.jpg"
	post := &Post{
		ID:        primitive.NewObjectID(),
		UserID:    primitive.NewObjectID(),
		Text:      "Post with image",
		Image:     &imageURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.NoError(t, post.Validate())
}

func TestValidatePost_WithVideo(t *testing.T) {
	videoURL := "https://s3.amazonaws.com/bucket/video.mp4"
	post := &Post{
		ID:        primitive.NewObjectID(),
		UserID:    primitive.NewObjectID(),
		Text:      "Post with video",
		Video:     &videoURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.NoError(t, post.Validate())
}

func TestPost_Structure(t *testing.T) {
	post := &Post{
		ID:        primitive.NewObjectID(),
		UserID:    primitive.NewObjectID(),
		Text:      "Test post",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.NotNil(t, post.ID)
	assert.NotNil(t, post.UserID)
	assert.Equal(t, "Test post", post.Text)
	assert.Nil(t, post.Image)
	assert.Nil(t, post.Video)
	assert.Nil(t, post.DeletedAt)
}
