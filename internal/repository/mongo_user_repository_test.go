package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"temp_backend/internal/domain"
)

// setupMongoDB starts a MongoDB testcontainer and returns the client and cleanup function.
func setupMongoDB(t *testing.T) (*mongo.Client, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:7",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	host, err := container.Host(ctx)
	assert.NoError(t, err)

	port, err := container.MappedPort(ctx, "27017")
	assert.NoError(t, err)

	uri := "mongodb://" + host + ":" + port.Port()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	assert.NoError(t, err)

	err = client.Ping(ctx, nil)
	assert.NoError(t, err)

	cleanup := func() {
		_ = client.Disconnect(context.Background())
		_ = container.Terminate(context.Background())
	}

	return client, cleanup
}

func TestMongoUserRepository_Create_Success(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	user := &domain.User{
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  "hashedpassword",
		IsAgeVerified: true,
	}

	err = repo.Create(context.Background(), user)

	assert.NoError(t, err)
	assert.NotEqual(t, primitive.NilObjectID, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
}

func TestMongoUserRepository_Create_DuplicateEmail(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	user1 := &domain.User{
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  "hashedpassword",
		IsAgeVerified: true,
	}

	err = repo.Create(context.Background(), user1)
	assert.NoError(t, err)

	user2 := &domain.User{
		Name:          "Jane Doe",
		Email:         "john@example.com",
		PasswordHash:  "hashedpassword2",
		IsAgeVerified: true,
	}

	err = repo.Create(context.Background(), user2)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrEmailAlreadyExists))
}

func TestMongoUserRepository_GetByEmail_Found(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	user := &domain.User{
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  "hashedpassword",
		IsAgeVerified: true,
	}

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)

	retrieved, err := repo.GetByEmail(context.Background(), "john@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, user.Name, retrieved.Name)
}

func TestMongoUserRepository_GetByEmail_NotFound(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	retrieved, err := repo.GetByEmail(context.Background(), "nonexistent@example.com")

	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestMongoUserRepository_GetByEmail_CaseInsensitive(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	user := &domain.User{
		Name:          "John Doe",
		Email:         "John@EXAMPLE.COM",
		PasswordHash:  "hashedpassword",
		IsAgeVerified: true,
	}

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)

	retrieved, err := repo.GetByEmail(context.Background(), "john@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestMongoUserRepository_GetByID_Found(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	user := &domain.User{
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  "hashedpassword",
		IsAgeVerified: true,
	}

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), user.ID)

	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, user.ID, retrieved.ID)
}

func TestMongoUserRepository_GetByID_NotFound(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), primitive.NewObjectID())

	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestMongoUserRepository_Update_Success(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	user := &domain.User{
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  "hashedpassword",
		IsAgeVerified: true,
	}

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)

	user.Name = "Jane Doe"
	user.UpdatedAt = time.Now().UTC()

	err = repo.Update(context.Background(), user)

	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", retrieved.Name)
}

func TestMongoUserRepository_Update_NotFound(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	user := &domain.User{
		ID:            primitive.NewObjectID(),
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  "hashedpassword",
		IsAgeVerified: true,
		UpdatedAt:     time.Now().UTC(),
	}

	err = repo.Update(context.Background(), user)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestMongoUserRepository_Delete_Success(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	user := &domain.User{
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  "hashedpassword",
		IsAgeVerified: true,
	}

	err = repo.Create(context.Background(), user)
	assert.NoError(t, err)

	err = repo.Delete(context.Background(), user.ID)

	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), user.ID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestMongoUserRepository_Delete_NotFound(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	err = repo.Delete(context.Background(), primitive.NewObjectID())

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestMongoUserRepository_EmailIndexCreated(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	db := client.Database("test_db")
	repo, err := NewMongoUserRepository(db)
	assert.NoError(t, err)

	// Verify unique index on email exists
	indexView := repo.collection.Indexes()
	cursor, err := indexView.List(context.Background())
	assert.NoError(t, err)
	defer cursor.Close(context.Background())

	foundEmailIndex := false
	for cursor.Next(context.Background()) {
		var index bson.M
		err := cursor.Decode(&index)
		assert.NoError(t, err)

		if key, ok := index["key"].(bson.M); ok {
			if _, hasEmail := key["email"]; hasEmail {
				if unique, ok := index["unique"].(bool); ok && unique {
					foundEmailIndex = true
				}
			}
		}
	}

	assert.True(t, foundEmailIndex, "unique email index should exist")
}
