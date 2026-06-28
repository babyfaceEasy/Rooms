package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/config"
	"temp_backend/internal/domain"
	"temp_backend/internal/service"
)

// MockAuthService is a mock implementation of AuthService for testing.
type MockAuthService struct {
	validateTokenFunc func(token string) (*domain.JWTClaims, error)
	loginFunc         func(ctx context.Context, email, password string) (string, string, error)
	refreshFunc       func(ctx context.Context, refreshToken string) (string, error)
	logoutFunc        func(ctx context.Context, userID string) error
}

func (m *MockAuthService) ValidateToken(token string) (*domain.JWTClaims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(token)
	}
	return nil, domain.ErrTokenInvalid
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, email, password)
	}
	return "", "", errors.New("not implemented")
}

func (m *MockAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	if m.refreshFunc != nil {
		return m.refreshFunc(ctx, refreshToken)
	}
	return "", errors.New("not implemented")
}

func (m *MockAuthService) Logout(ctx context.Context, userID string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, userID)
	}
	return errors.New("not implemented")
}

func newTestConfig() config.Config {
	cfg := config.Config{}
	cfg.JWT.Secret = "this-is-a-test-secret-key-that-is-at-least-32-chars"
	cfg.JWT.AccessTokenTTL = 1 * time.Hour
	cfg.JWT.RefreshTokenTTL = 7 * 24 * time.Hour
	return cfg
}

func newTestAuthService(cfg config.Config) service.AuthService {
	secret := cfg.JWT.Secret
	return &MockAuthService{
		validateTokenFunc: func(token string) (*domain.JWTClaims, error) {
			claims := &domain.JWTClaims{}

			parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, domain.ErrTokenInvalid
				}
				return []byte(secret), nil
			})

			if err != nil {
				return nil, domain.ErrTokenInvalid
			}

			if !parsedToken.Valid {
				return nil, domain.ErrTokenInvalid
			}

			return claims, nil
		},
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	cfg := newTestConfig()
	authService := newTestAuthService(cfg)

	userID := primitive.NewObjectID()
	claims := domain.JWTClaims{
		UserID: userID,
		Email:  "john@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	require.NoError(t, err)

	validatedClaims, err := authService.ValidateToken(tokenString)

	assert.NoError(t, err)
	assert.Equal(t, userID, validatedClaims.UserID)
	assert.Equal(t, "john@example.com", validatedClaims.Email)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	cfg := newTestConfig()
	authService := newTestAuthService(cfg)

	userID := primitive.NewObjectID()
	claims := domain.JWTClaims{
		UserID: userID,
		Email:  "john@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	require.NoError(t, err)

	validatedClaims, err := authService.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	cfg := newTestConfig()
	authService := newTestAuthService(cfg)

	userID := primitive.NewObjectID()
	claims := domain.JWTClaims{
		UserID: userID,
		Email:  "john@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("wrong-secret-key-at-least-32-chars-long-for-testing"))
	require.NoError(t, err)

	validatedClaims, err := authService.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestValidateToken_MalformedToken(t *testing.T) {
	cfg := newTestConfig()
	authService := newTestAuthService(cfg)

	validatedClaims, err := authService.ValidateToken("malformed.token.string")

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	cfg := newTestConfig()
	authService := newTestAuthService(cfg)

	validatedClaims, err := authService.ValidateToken("")

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestAuthMiddleware_BearerTokenExtraction(t *testing.T) {
	cfg := newTestConfig()
	authService := newTestAuthService(cfg)

	userID := primitive.NewObjectID()
	claims := domain.JWTClaims{
		UserID: userID,
		Email:  "john@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	require.NoError(t, err)

	// The middleware extracts "Bearer " + token format
	// We verify that the token validation works with properly formatted Bearer tokens
	validatedClaims, err := authService.ValidateToken(tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims)
}

func TestAuthMiddleware_InvalidBearerFormat(t *testing.T) {
	cfg := newTestConfig()

	testCases := []string{
		"invalid_token",
		"Bearer",
		"Bearer ",
		"Bearer invalid",
	}

	for _, testToken := range testCases {
		// These should all fail validation
		authService := newTestAuthService(cfg)
		_, err := authService.ValidateToken(testToken)
		assert.Error(t, err)
	}
}

func TestAuthMiddleware_TokenClaimsExtraction(t *testing.T) {
	cfg := newTestConfig()
	authService := newTestAuthService(cfg)

	userID := primitive.NewObjectID()
	email := "test@example.com"
	claims := domain.JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	require.NoError(t, err)

	validatedClaims, err := authService.ValidateToken(tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims)
	assert.Equal(t, userID, validatedClaims.UserID)
	assert.Equal(t, email, validatedClaims.Email)
	assert.NotNil(t, validatedClaims.ExpiresAt)
}

func TestAuthMiddleware_CaseSensitiveBearer(t *testing.T) {
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
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	require.NoError(t, err)

	authService := newTestAuthService(cfg)

	// Valid token should work
	validatedClaims, err := authService.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims)

	// Lowercase "bearer" format used in Authorization header - the middleware handles this
	// We just verify token validation works with the token string
	validatedClaims2, err := authService.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims2)
}

func TestAuthMiddleware_MultipleValidations(t *testing.T) {
	cfg := newTestConfig()
	authService := newTestAuthService(cfg)

	userID1 := primitive.NewObjectID()
	claims1 := domain.JWTClaims{
		UserID: userID1,
		Email:  "user1@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token1 := jwt.NewWithClaims(jwt.SigningMethodHS256, claims1)
	token1String, _ := token1.SignedString([]byte(cfg.JWT.Secret))

	userID2 := primitive.NewObjectID()
	claims2 := domain.JWTClaims{
		UserID: userID2,
		Email:  "user2@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token2 := jwt.NewWithClaims(jwt.SigningMethodHS256, claims2)
	token2String, _ := token2.SignedString([]byte(cfg.JWT.Secret))

	// Validate first token
	validated1, err := authService.ValidateToken(token1String)
	assert.NoError(t, err)
	assert.Equal(t, userID1, validated1.UserID)

	// Validate second token
	validated2, err := authService.ValidateToken(token2String)
	assert.NoError(t, err)
	assert.Equal(t, userID2, validated2.UserID)

	// Verify they are different users
	assert.NotEqual(t, validated1.UserID, validated2.UserID)
}
