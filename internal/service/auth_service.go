package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/config"
	"temp_backend/internal/domain"
	"temp_backend/internal/repository"
)

// AuthService defines authentication operations.
type AuthService interface {
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (newAccessToken string, err error)
	Logout(ctx context.Context, userID string) error
	ValidateToken(token string) (*domain.JWTClaims, error)
}

type authService struct {
	userRepo          repository.UserRepository
	refreshTokenRepo  repository.RefreshTokenRepository
	cfg               config.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, cfg config.Config) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		cfg:              cfg,
	}
}

// Login authenticates a user and returns access and refresh tokens.
func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("login failed: %w", domain.ErrInvalidCredentials)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", fmt.Errorf("login failed: %w", domain.ErrInvalidCredentials)
	}

	// Generate access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate and store refresh token
	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token.
func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// Hash the refresh token
	tokenHash := hashToken(refreshToken)

	// Get refresh token from database
	storedToken, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

// Logout invalidates all refresh tokens for a user.
func (s *authService) Logout(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", domain.ErrInvalidInput)
	}

	return s.refreshTokenRepo.DeleteByUserID(ctx, oid)
}

// ValidateToken validates a JWT token and returns its claims.
func (s *authService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	claims := &domain.JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid signing method: %w", domain.ErrTokenInvalid)
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", domain.ErrTokenInvalid)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid: %w", domain.ErrTokenInvalid)
	}

	return claims, nil
}

// generateAccessToken creates a JWT access token for a user.
func (s *authService) generateAccessToken(user *domain.User) (string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.cfg.JWT.AccessTokenTTL)

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
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// generateRefreshToken creates a JWT refresh token and stores its hash in the database.
func (s *authService) generateRefreshToken(ctx context.Context, user *domain.User) (string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.cfg.JWT.RefreshTokenTTL)

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
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", err
	}

	// Store hashed token in database
	tokenHash := hashToken(tokenString)
	refreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return "", err
	}

	return tokenString, nil
}

// hashToken creates a SHA256 hash of a token.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
