package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"temp_backend/internal/domain"
)

var mongoContainer testcontainers.Container

func setupMongoContainer(t *testing.T) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	mongoContainer = container

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "27017")
	require.NoError(t, err)

	mongoURL := "mongodb://" + host + ":" + port.Port()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	require.NoError(t, err)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	return client
}

func teardownMongoContainer(t *testing.T) {
	if mongoContainer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := mongoContainer.Terminate(ctx)
		require.NoError(t, err)
	}
}

func TestCreateRefreshToken_Success(t *testing.T) {
	client := setupMongoContainer(t)
	defer teardownMongoContainer(t)

	db := client.Database("test_db")
	repo, err := NewMongoRefreshTokenRepository(db)
	require.NoError(t, err)

	userID := primitive.NewObjectID()
	token := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: "test_hash_123",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = repo.Create(ctx, token)
	require.NoError(t, err)
	assert.NotNil(t, token.ID)
	assert.False(t, token.ID.IsZero())
	assert.NotZero(t, token.CreatedAt)
}

func TestGetByTokenHash_Success(t *testing.T) {
	client := setupMongoContainer(t)
	defer teardownMongoContainer(t)

	db := client.Database("test_db")
	repo, err := NewMongoRefreshTokenRepository(db)
	require.NoError(t, err)

	userID := primitive.NewObjectID()
	tokenHash := "test_hash_456"
	expiresAt := time.Now().Add(7 * 24 * time.Hour).UTC()

	token := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = repo.Create(ctx, token)
	require.NoError(t, err)

	retrievedToken, err := repo.GetByTokenHash(ctx, tokenHash)
	require.NoError(t, err)
	assert.NotNil(t, retrievedToken)
	assert.Equal(t, userID, retrievedToken.UserID)
	assert.Equal(t, tokenHash, retrievedToken.TokenHash)
	assert.Equal(t, expiresAt.Unix(), retrievedToken.ExpiresAt.Unix())
}

func TestGetByTokenHash_NotFound(t *testing.T) {
	client := setupMongoContainer(t)
	defer teardownMongoContainer(t)

	db := client.Database("test_db")
	repo, err := NewMongoRefreshTokenRepository(db)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	retrievedToken, err := repo.GetByTokenHash(ctx, "nonexistent_hash")
	assert.Error(t, err)
	assert.Nil(t, retrievedToken)
}

func TestDelete_Success(t *testing.T) {
	client := setupMongoContainer(t)
	defer teardownMongoContainer(t)

	db := client.Database("test_db")
	repo, err := NewMongoRefreshTokenRepository(db)
	require.NoError(t, err)

	userID := primitive.NewObjectID()
	token := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: "test_hash_delete",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = repo.Create(ctx, token)
	require.NoError(t, err)

	err = repo.Delete(ctx, token.ID)
	assert.NoError(t, err)

	retrievedToken, err := repo.GetByTokenHash(ctx, token.TokenHash)
	assert.Error(t, err)
	assert.Nil(t, retrievedToken)
}

func TestDeleteByUserID_Success(t *testing.T) {
	client := setupMongoContainer(t)
	defer teardownMongoContainer(t)

	db := client.Database("test_db")
	repo, err := NewMongoRefreshTokenRepository(db)
	require.NoError(t, err)

	userID := primitive.NewObjectID()

	token1 := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: "hash_1",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
	}

	token2 := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: "hash_2",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = repo.Create(ctx, token1)
	require.NoError(t, err)

	err = repo.Create(ctx, token2)
	require.NoError(t, err)

	err = repo.DeleteByUserID(ctx, userID)
	assert.NoError(t, err)

	retrievedToken1, _ := repo.GetByTokenHash(ctx, "hash_1")
	retrievedToken2, _ := repo.GetByTokenHash(ctx, "hash_2")

	assert.Nil(t, retrievedToken1)
	assert.Nil(t, retrievedToken2)
}

func TestRefreshTokenExpiration(t *testing.T) {
	client := setupMongoContainer(t)
	defer teardownMongoContainer(t)

	db := client.Database("test_db")
	repo, err := NewMongoRefreshTokenRepository(db)
	require.NoError(t, err)

	userID := primitive.NewObjectID()

	token := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: "expired_token_hash",
		ExpiresAt: time.Now().Add(-1 * time.Second).UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = repo.Create(ctx, token)
	require.NoError(t, err)

	retrievedToken, err := repo.GetByTokenHash(ctx, "expired_token_hash")

	assert.Error(t, err)
	assert.Nil(t, retrievedToken)
}

func TestCreateMultipleTokensForUser(t *testing.T) {
	client := setupMongoContainer(t)
	defer teardownMongoContainer(t)

	db := client.Database("test_db")
	repo, err := NewMongoRefreshTokenRepository(db)
	require.NoError(t, err)

	userID := primitive.NewObjectID()

	token1 := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: "multi_hash_1",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
	}

	token2 := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: "multi_hash_2",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = repo.Create(ctx, token1)
	require.NoError(t, err)

	err = repo.Create(ctx, token2)
	require.NoError(t, err)

	retrievedToken1, err := repo.GetByTokenHash(ctx, "multi_hash_1")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedToken1)

	retrievedToken2, err := repo.GetByTokenHash(ctx, "multi_hash_2")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedToken2)
}
