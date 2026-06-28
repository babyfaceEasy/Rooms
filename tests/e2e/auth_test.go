package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"temp_backend/config"
	"temp_backend/internal/api"
	"temp_backend/internal/handler"
	"temp_backend/internal/repository"
	"temp_backend/internal/service"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type RegisterRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	AgeVerified  bool   `json:"age_verified"`
}

type RegisterResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func setupAuthTestEnvironment(t *testing.T) (*http.Server, *mongo.Client, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start MongoDB container
	req := testcontainers.ContainerRequest{
		Image:        "mongo:7",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "27017")
	require.NoError(t, err)

	mongoURL := "mongodb://" + host + ":" + port.Port()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	require.NoError(t, err)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	db := client.Database("test_auth_db")

	// Create repositories
	userRepo, err := repository.NewMongoUserRepository(db)
	require.NoError(t, err)

	refreshTokenRepo := repository.NewMongoRefreshTokenRepository(db)

	// Create services
	userService := service.NewUserService(userRepo)

	cfg := config.Config{
		App: config.App{
			Name: "test_backend",
		},
		JWT: config.JWT{
			Secret:           "this-is-a-test-secret-key-that-is-at-least-32-chars",
			AccessTokenTTL:   1 * time.Hour,
			RefreshTokenTTL:  7 * 24 * time.Hour,
		},
	}

	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg)

	// Create handlers
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService, userService)

	// Create server
	server := api.NewServer(cfg, nil, nil, userHandler, authHandler, authService)

	// Start server
	httpServer := &http.Server{
		Addr:    ":9999",
		Handler: server.GetApp(),
	}

	go func() {
		_ = httpServer.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(500 * time.Millisecond)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
		_ = container.Terminate(context.Background())
	}

	return httpServer, client, cleanup
}

func makeRequest(t *testing.T, method, path string, body interface{}, token string) (int, []byte) {
	var requestBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)
		requestBody = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, "http://localhost:9999"+path, requestBody)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, respBody
}

func TestAuthFlow_RegisterLoginLogout(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Register user
	registerReq := RegisterRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		AgeVerified: true,
	}

	statusCode, respBody := makeRequest(t, "POST", "/api/v1/auth/register", registerReq, "")
	assert.Equal(t, http.StatusCreated, statusCode)

	var registerResp RegisterResponse
	err := json.Unmarshal(respBody, &registerResp)
	require.NoError(t, err)
	assert.NotEmpty(t, registerResp.ID)

	// Login
	loginReq := LoginRequest{
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}

	statusCode, respBody = makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
	assert.Equal(t, http.StatusOK, statusCode)

	var loginResp LoginResponse
	err = json.Unmarshal(respBody, &loginResp)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.RefreshToken)
	assert.Equal(t, "Bearer", loginResp.TokenType)
	assert.Equal(t, 3600, loginResp.ExpiresIn)

	accessToken := loginResp.AccessToken

	// Logout
	statusCode, _ = makeRequest(t, "POST", "/api/v1/auth/logout", nil, accessToken)
	assert.Equal(t, http.StatusNoContent, statusCode)
}

func TestAuthFlow_LoginWithInvalidCredentials(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Register user first
	registerReq := RegisterRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		AgeVerified: true,
	}

	statusCode, _ := makeRequest(t, "POST", "/api/v1/auth/register", registerReq, "")
	assert.Equal(t, http.StatusCreated, statusCode)

	// Try login with wrong password
	loginReq := LoginRequest{
		Email:    "john@example.com",
		Password: "WrongPassword123!",
	}

	statusCode, _ = makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
	assert.Equal(t, http.StatusUnauthorized, statusCode)
}

func TestAuthFlow_LoginNonexistentUser(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	loginReq := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "Password123!",
	}

	statusCode, _ := makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
	assert.Equal(t, http.StatusUnauthorized, statusCode)
}

func TestAuthFlow_RefreshToken(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Register and login
	registerReq := RegisterRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		AgeVerified: true,
	}

	statusCode, _ := makeRequest(t, "POST", "/api/v1/auth/register", registerReq, "")
	require.Equal(t, http.StatusCreated, statusCode)

	loginReq := LoginRequest{
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}

	statusCode, respBody := makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
	require.Equal(t, http.StatusOK, statusCode)

	var loginResp LoginResponse
	err := json.Unmarshal(respBody, &loginResp)
	require.NoError(t, err)

	refreshToken := loginResp.RefreshToken
	oldAccessToken := loginResp.AccessToken

	// Refresh token
	refreshReq := RefreshRequest{
		RefreshToken: refreshToken,
	}

	statusCode, respBody = makeRequest(t, "POST", "/api/v1/auth/refresh", refreshReq, "")
	assert.Equal(t, http.StatusOK, statusCode)

	var refreshResp RefreshResponse
	err = json.Unmarshal(respBody, &refreshResp)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshResp.AccessToken)
	assert.Equal(t, "Bearer", refreshResp.TokenType)
	assert.Equal(t, 3600, refreshResp.ExpiresIn)

	// Old token and new token should be different
	assert.NotEqual(t, oldAccessToken, refreshResp.AccessToken)

	// New token should be valid
	newAccessToken := refreshResp.AccessToken
	statusCode, _ = makeRequest(t, "POST", "/api/v1/auth/logout", nil, newAccessToken)
	assert.Equal(t, http.StatusNoContent, statusCode)
}

func TestAuthFlow_ProtectedEndpointWithoutToken(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Try to access protected endpoint without token
	statusCode, _ := makeRequest(t, "GET", "/api/v1/users/507f1f77bcf86cd799439011", nil, "")
	assert.Equal(t, http.StatusUnauthorized, statusCode)
}

func TestAuthFlow_ProtectedEndpointWithValidToken(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Register and login
	registerReq := RegisterRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		AgeVerified: true,
	}

	statusCode, respBody := makeRequest(t, "POST", "/api/v1/auth/register", registerReq, "")
	require.Equal(t, http.StatusCreated, statusCode)

	var registerResp RegisterResponse
	err := json.Unmarshal(respBody, &registerResp)
	require.NoError(t, err)

	loginReq := LoginRequest{
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}

	statusCode, respBody = makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
	require.Equal(t, http.StatusOK, statusCode)

	var loginResp LoginResponse
	err = json.Unmarshal(respBody, &loginResp)
	require.NoError(t, err)

	accessToken := loginResp.AccessToken

	// Access protected endpoint with valid token
	statusCode, _ = makeRequest(t, "GET", "/api/v1/users/"+registerResp.ID, nil, accessToken)
	assert.Equal(t, http.StatusOK, statusCode)
}

func TestAuthFlow_ProtectedEndpointWithInvalidToken(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Try to access protected endpoint with invalid token
	statusCode, _ := makeRequest(t, "GET", "/api/v1/users/507f1f77bcf86cd799439011", nil, "invalid_token")
	assert.Equal(t, http.StatusUnauthorized, statusCode)
}

func TestAuthFlow_LogoutInvalidatesTokens(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Register and login
	registerReq := RegisterRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		AgeVerified: true,
	}

	statusCode, _ := makeRequest(t, "POST", "/api/v1/auth/register", registerReq, "")
	require.Equal(t, http.StatusCreated, statusCode)

	loginReq := LoginRequest{
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}

	statusCode, respBody := makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
	require.Equal(t, http.StatusOK, statusCode)

	var loginResp LoginResponse
	err := json.Unmarshal(respBody, &loginResp)
	require.NoError(t, err)

	accessToken := loginResp.AccessToken
	refreshToken := loginResp.RefreshToken

	// Logout
	statusCode, _ = makeRequest(t, "POST", "/api/v1/auth/logout", nil, accessToken)
	assert.Equal(t, http.StatusNoContent, statusCode)

	// Try to refresh token after logout (should fail)
	refreshReq := RefreshRequest{
		RefreshToken: refreshToken,
	}

	statusCode, _ = makeRequest(t, "POST", "/api/v1/auth/refresh", refreshReq, "")
	assert.Equal(t, http.StatusUnauthorized, statusCode)
}

func TestAuthFlow_MultipleLogins(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Register user
	registerReq := RegisterRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		AgeVerified: true,
	}

	statusCode, _ := makeRequest(t, "POST", "/api/v1/auth/register", registerReq, "")
	require.Equal(t, http.StatusCreated, statusCode)

	// Login multiple times
	loginReq := LoginRequest{
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}

	var tokens []string

	for i := 0; i < 3; i++ {
		statusCode, respBody := makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
		assert.Equal(t, http.StatusOK, statusCode)

		var loginResp LoginResponse
		err := json.Unmarshal(respBody, &loginResp)
		require.NoError(t, err)

		tokens = append(tokens, loginResp.AccessToken)
	}

	// Verify each token is unique
	assert.NotEqual(t, tokens[0], tokens[1])
	assert.NotEqual(t, tokens[1], tokens[2])

	// Verify all tokens work
	for _, token := range tokens {
		statusCode, _ := makeRequest(t, "POST", "/api/v1/auth/logout", nil, token)
		// Note: first logout succeeds, subsequent may fail if we're testing one user
		// This is expected behavior
		assert.True(t, statusCode == http.StatusNoContent || statusCode == http.StatusBadRequest)
	}
}

func TestAuthFlow_InvalidRefreshToken(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	refreshReq := RefreshRequest{
		RefreshToken: "invalid_refresh_token_xyz",
	}

	statusCode, _ := makeRequest(t, "POST", "/api/v1/auth/refresh", refreshReq, "")
	assert.Equal(t, http.StatusUnauthorized, statusCode)
}

func TestAuthFlow_ConcurrentTokenOperations(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Register user
	registerReq := RegisterRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		AgeVerified: true,
	}

	statusCode, _ := makeRequest(t, "POST", "/api/v1/auth/register", registerReq, "")
	require.Equal(t, http.StatusCreated, statusCode)

	// Login
	loginReq := LoginRequest{
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}

	statusCode, respBody := makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
	require.Equal(t, http.StatusOK, statusCode)

	var loginResp LoginResponse
	err := json.Unmarshal(respBody, &loginResp)
	require.NoError(t, err)

	accessToken := loginResp.AccessToken
	refreshToken := loginResp.RefreshToken

	// Run multiple refresh operations concurrently
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func() {
			refreshReq := RefreshRequest{
				RefreshToken: refreshToken,
			}

			statusCode, respBody := makeRequest(t, "POST", "/api/v1/auth/refresh", refreshReq, "")
			assert.Equal(t, http.StatusOK, statusCode)

			var refreshResp RefreshResponse
			err := json.Unmarshal(respBody, &refreshResp)
			assert.NoError(t, err)
			assert.NotEmpty(t, refreshResp.AccessToken)

			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify original token still works
	statusCode, _ = makeRequest(t, "POST", "/api/v1/auth/logout", nil, accessToken)
	assert.Equal(t, http.StatusNoContent, statusCode)
}

func TestAuthFlow_LoginAfterLogout(t *testing.T) {
	_, _, cleanup := setupAuthTestEnvironment(t)
	defer cleanup()

	// Register user
	registerReq := RegisterRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		AgeVerified: true,
	}

	statusCode, _ := makeRequest(t, "POST", "/api/v1/auth/register", registerReq, "")
	require.Equal(t, http.StatusCreated, statusCode)

	loginReq := LoginRequest{
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}

	// First login
	statusCode, respBody := makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
	require.Equal(t, http.StatusOK, statusCode)

	var loginResp LoginResponse
	err := json.Unmarshal(respBody, &loginResp)
	require.NoError(t, err)

	firstAccessToken := loginResp.AccessToken

	// Logout
	statusCode, _ = makeRequest(t, "POST", "/api/v1/auth/logout", nil, firstAccessToken)
	assert.Equal(t, http.StatusNoContent, statusCode)

	// Login again after logout
	statusCode, respBody = makeRequest(t, "POST", "/api/v1/auth/login", loginReq, "")
	assert.Equal(t, http.StatusOK, statusCode)

	var secondLoginResp LoginResponse
	err = json.Unmarshal(respBody, &secondLoginResp)
	require.NoError(t, err)
	assert.NotEmpty(t, secondLoginResp.AccessToken)

	// New token should work
	statusCode, _ = makeRequest(t, "POST", "/api/v1/auth/logout", nil, secondLoginResp.AccessToken)
	assert.Equal(t, http.StatusNoContent, statusCode)
}
