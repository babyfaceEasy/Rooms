package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"temp_backend/config"
	"temp_backend/internal/domain"
)

// MockRefreshTokenRepository is a mock implementation of RefreshTokenRepository for testing.
type MockRefreshTokenRepository struct {
	createFunc         func(ctx context.Context, token *domain.RefreshToken) error
	getByTokenHashFunc func(ctx context.Context, hash string) (*domain.RefreshToken, error)
	deleteFunc         func(ctx context.Context, id primitive.ObjectID) error
	deleteByUserIDFunc func(ctx context.Context, userID primitive.ObjectID) error
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, token)
	}
	token.ID = primitive.NewObjectID()
	token.CreatedAt = time.Now().UTC()
	return nil
}

func (m *MockRefreshTokenRepository) GetByTokenHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	if m.getByTokenHashFunc != nil {
		return m.getByTokenHashFunc(ctx, hash)
	}
	return nil, domain.ErrTokenNotFound
}

func (m *MockRefreshTokenRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	if m.deleteByUserIDFunc != nil {
		return m.deleteByUserIDFunc(ctx, userID)
	}
	return nil
}

func newTestConfig() config.Config {
	cfg := config.Config{}
	cfg.JWT.Secret = "this-is-a-test-secret-key-that-is-at-least-32-chars"
	cfg.JWT.AccessTokenTTL = 1 * time.Hour
	cfg.JWT.RefreshTokenTTL = 7 * 24 * time.Hour
	return cfg
}

func TestLogin_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &domain.User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  string(hashedPassword),
		IsAgeVerified: true,
	}

	mockUserRepo := &MockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			if email == "john@example.com" {
				return user, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		createFunc: func(ctx context.Context, token *domain.RefreshToken) error {
			token.ID = primitive.NewObjectID()
			token.CreatedAt = time.Now().UTC()
			return nil
		},
	}

	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, newTestConfig())
	ctx := context.Background()

	accessToken, refreshToken, err := svc.Login(ctx, "john@example.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestLogin_InvalidPassword(t *testing.T) {
	userID := primitive.NewObjectID()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

	user := &domain.User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  string(hashedPassword),
		IsAgeVerified: true,
	}

	mockUserRepo := &MockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			if email == "john@example.com" {
				return user, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, newTestConfig())
	ctx := context.Background()

	accessToken, refreshToken, err := svc.Login(ctx, "john@example.com", "wrong_password")

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.True(t, errors.Is(err, domain.ErrInvalidCredentials))
}

func TestLogin_UserNotFound(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, newTestConfig())
	ctx := context.Background()

	accessToken, refreshToken, err := svc.Login(ctx, "nonexistent@example.com", "password")

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.True(t, errors.Is(err, domain.ErrInvalidCredentials))
}

func TestLogin_GeneratesValidAccessToken(t *testing.T) {
	userID := primitive.NewObjectID()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &domain.User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  string(hashedPassword),
		IsAgeVerified: true,
	}

	mockUserRepo := &MockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return user, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		createFunc: func(ctx context.Context, token *domain.RefreshToken) error {
			token.ID = primitive.NewObjectID()
			token.CreatedAt = time.Now().UTC()
			return nil
		},
	}

	cfg := newTestConfig()
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)
	ctx := context.Background()

	accessToken, _, err := svc.Login(ctx, "john@example.com", "password123")
	require.NoError(t, err)

	claims, err := svc.ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "john@example.com", claims.Email)
}

func TestRefreshAccessToken_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	user := &domain.User{
		ID:    userID,
		Email: "john@example.com",
		Name:  "John Doe",
	}

	refreshTokenEntity := &domain.RefreshToken{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TokenHash: "test_hash",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
		CreatedAt: time.Now().UTC(),
	}

	mockUserRepo := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			if id == userID {
				return user, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		getByTokenHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return refreshTokenEntity, nil
		},
	}

	cfg := newTestConfig()
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)
	ctx := context.Background()

	// Create a valid refresh token to hash
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, domain.JWTClaims{
		UserID: userID,
		Email:  "john@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	})
	refreshTokenString, _ := refreshToken.SignedString([]byte(cfg.JWT.Secret))

	newAccessToken, err := svc.RefreshAccessToken(ctx, refreshTokenString)

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
}

func TestRefreshAccessToken_InvalidToken(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		getByTokenHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return nil, domain.ErrTokenNotFound
		},
	}

	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, newTestConfig())
	ctx := context.Background()

	newAccessToken, err := svc.RefreshAccessToken(ctx, "invalid_token")

	assert.Error(t, err)
	assert.Empty(t, newAccessToken)
}

func TestLogout_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	userIDHex := userID.Hex()

	mockUserRepo := &MockUserRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		deleteByUserIDFunc: func(ctx context.Context, id primitive.ObjectID) error {
			if id == userID {
				return nil
			}
			return domain.ErrUserNotFound
		},
	}

	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, newTestConfig())
	ctx := context.Background()

	err := svc.Logout(ctx, userIDHex)

	assert.NoError(t, err)
}

func TestLogout_InvalidUserID(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, newTestConfig())
	ctx := context.Background()

	err := svc.Logout(ctx, "invalid_id")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestValidateToken_Success(t *testing.T) {
	cfg := newTestConfig()
	userID := primitive.NewObjectID()

	claims := domain.JWTClaims{
		UserID: userID,
		Email:  "john@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWT.Secret))

	mockUserRepo := &MockUserRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	validatedClaims, err := svc.ValidateToken(tokenString)

	assert.NoError(t, err)
	assert.Equal(t, userID, validatedClaims.UserID)
	assert.Equal(t, "john@example.com", validatedClaims.Email)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	cfg := newTestConfig()
	userID := primitive.NewObjectID()

	claims := domain.JWTClaims{
		UserID: userID,
		Email:  "john@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWT.Secret))

	mockUserRepo := &MockUserRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	validatedClaims, err := svc.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	cfg := newTestConfig()
	userID := primitive.NewObjectID()

	claims := domain.JWTClaims{
		UserID: userID,
		Email:  "john@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret-key-at-least-32-chars-long-for-testing"))

	mockUserRepo := &MockUserRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	validatedClaims, err := svc.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestValidateToken_MalformedToken(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, newTestConfig())

	validatedClaims, err := svc.ValidateToken("malformed.token.string")

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, newTestConfig())

	validatedClaims, err := svc.ValidateToken("")

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestLogin_RefreshTokenIsStoredHashed(t *testing.T) {
	userID := primitive.NewObjectID()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := &domain.User{
		ID:            userID,
		Name:          "John Doe",
		Email:         "john@example.com",
		PasswordHash:  string(hashedPassword),
		IsAgeVerified: true,
	}

	var storedToken *domain.RefreshToken

	mockUserRepo := &MockUserRepository{
		getByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return user, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		createFunc: func(ctx context.Context, token *domain.RefreshToken) error {
			storedToken = token
			token.ID = primitive.NewObjectID()
			token.CreatedAt = time.Now().UTC()
			return nil
		},
	}

	cfg := newTestConfig()
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)
	ctx := context.Background()

	_, _, err := svc.Login(ctx, "john@example.com", "password123")
	require.NoError(t, err)

	assert.NotNil(t, storedToken)
	assert.NotEmpty(t, storedToken.TokenHash)
	// Verify it's a hash (hex string of 64 chars = SHA256)
	assert.Len(t, storedToken.TokenHash, 64)
}

func TestAccessTokenIncludesCorrectClaims(t *testing.T) {
	cfg := newTestConfig()
	userID := primitive.NewObjectID()
	email := "john@example.com"

	mockUserRepo := &MockUserRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	user := &domain.User{
		ID:    userID,
		Email: email,
		Name:  "John Doe",
	}

	// Get access token through generateAccessToken by accessing service methods
	// We'll create it manually to verify
	now := time.Now().UTC()
	expiresAt := now.Add(cfg.JWT.AccessTokenTTL)

	claims := domain.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWT.Secret))

	validatedClaims, err := svc.ValidateToken(tokenString)
	require.NoError(t, err)

	assert.Equal(t, userID, validatedClaims.UserID)
	assert.Equal(t, email, validatedClaims.Email)
	assert.NotNil(t, validatedClaims.ExpiresAt)
	assert.NotNil(t, validatedClaims.IssuedAt)
}
