package e2e

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"
	"temp_backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupMongoContainer starts a MongoDB testcontainer and returns the URI and cleanup function.
func setupMongoContainer(t *testing.T) (string, func()) {
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

	uri := "mongodb://" + host + ":" + port.Port() + "/test_db"

	cleanup := func() {
		_ = container.Terminate(context.Background())
	}

	return uri, cleanup
}

func TestRegistration_FullFlow(t *testing.T) {
	mongoURI, cleanup := setupMongoContainer(t)
	defer cleanup()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	assert.NoError(t, err)
	defer client.Disconnect(context.Background())

	db := client.Database("test_db")

	// Create repositories and services
	userRepo, err := repository.NewMongoUserRepository(db)
	assert.NoError(t, err)
	refreshTokenRepo, err := repository.NewMongoRefreshTokenRepository(db)
	assert.NoError(t, err)
	userService := service.NewUserService(userRepo, refreshTokenRepo)

	// Test successful registration
	user, err := userService.Register(context.Background(), "John Doe", "john@example.com", "SecurePass123!", true)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.True(t, user.IsAgeVerified)

	// Verify user can be retrieved
	retrieved, err := userService.GetUserByID(context.Background(), user.ID.Hex())
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, "John Doe", retrieved.Name)
}

func TestRegistration_ConcurrentRegistrations(t *testing.T) {
	mongoURI, cleanup := setupMongoContainer(t)
	defer cleanup()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	assert.NoError(t, err)
	defer client.Disconnect(context.Background())

	db := client.Database("test_db")

	userRepo, err := repository.NewMongoUserRepository(db)
	assert.NoError(t, err)
	refreshTokenRepo, err := repository.NewMongoRefreshTokenRepository(db)
	assert.NoError(t, err)
	userService := service.NewUserService(userRepo, refreshTokenRepo)

	// Register multiple users concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []*domain.User
	var errs []error

	numGoroutines := 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			email := fmt.Sprintf("user%d@example.com", index)
			user, err := userService.Register(
				context.Background(),
				fmt.Sprintf("User %d", index),
				email,
				"SecurePass123!",
				true,
			)

			mu.Lock()
			results = append(results, user)
			if err != nil {
				errs = append(errs, err)
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 0, len(errs), "Should have no registration errors")
	assert.Equal(t, numGoroutines, len(results), "Should have registered all users")

	// Verify all users can be retrieved
	for _, user := range results {
		retrieved, err := userService.GetUserByID(context.Background(), user.ID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	}
}

func TestRegistration_EmailUniquenessConstraint(t *testing.T) {
	mongoURI, cleanup := setupMongoContainer(t)
	defer cleanup()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	assert.NoError(t, err)
	defer client.Disconnect(context.Background())

	db := client.Database("test_db")

	userRepo, err := repository.NewMongoUserRepository(db)
	assert.NoError(t, err)
	refreshTokenRepo, err := repository.NewMongoRefreshTokenRepository(db)
	assert.NoError(t, err)
	userService := service.NewUserService(userRepo, refreshTokenRepo)

	// Register first user
	user1, err := userService.Register(
		context.Background(),
		"John Doe",
		"john@example.com",
		"SecurePass123!",
		true,
	)

	assert.NoError(t, err)
	assert.NotNil(t, user1)

	// Try to register with same email
	user2, err := userService.Register(
		context.Background(),
		"Jane Doe",
		"john@example.com",
		"SecurePass456!",
		true,
	)

	assert.Error(t, err)
	assert.Nil(t, user2)
	// Should fail due to duplicate email
}

func TestRegistration_DataPersistence(t *testing.T) {
	mongoURI, cleanup := setupMongoContainer(t)
	defer cleanup()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	assert.NoError(t, err)
	defer client.Disconnect(context.Background())

	db := client.Database("test_db")

	// First connection: register user
	{
		userRepo, err := repository.NewMongoUserRepository(db)
		assert.NoError(t, err)
		refreshTokenRepo, err := repository.NewMongoRefreshTokenRepository(db)
		assert.NoError(t, err)
		userService := service.NewUserService(userRepo, refreshTokenRepo)

		user, err := userService.Register(
			context.Background(),
			"John Doe",
			"john@example.com",
			"SecurePass123!",
			true,
		)

		assert.NoError(t, err)
		assert.NotNil(t, user)
	}

	// Second connection: verify user persists
	{
		userRepo, err := repository.NewMongoUserRepository(db)
		assert.NoError(t, err)

		user, err := userRepo.GetByEmail(context.Background(), "john@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
	}
}

func TestRegistration_ValidationErrors(t *testing.T) {
	mongoURI, cleanup := setupMongoContainer(t)
	defer cleanup()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	assert.NoError(t, err)
	defer client.Disconnect(context.Background())

	db := client.Database("test_db")

	userRepo, err := repository.NewMongoUserRepository(db)
	assert.NoError(t, err)
	refreshTokenRepo, err := repository.NewMongoRefreshTokenRepository(db)
	assert.NoError(t, err)
	userService := service.NewUserService(userRepo, refreshTokenRepo)

	testCases := []struct {
		name        string
		input       RegisterRequest
		expectError bool
	}{
		{
			name: "valid registration",
			input: RegisterRequest{
				Name:        "John Doe",
				Email:       "john@example.com",
				Password:    "SecurePass123!",
				AgeVerified: true,
			},
			expectError: false,
		},
		{
			name: "missing age verification",
			input: RegisterRequest{
				Name:        "John Doe",
				Email:       "john@example.com",
				Password:    "SecurePass123!",
				AgeVerified: false,
			},
			expectError: true,
		},
		{
			name: "invalid email",
			input: RegisterRequest{
				Name:        "John Doe",
				Email:       "notanemail",
				Password:    "SecurePass123!",
				AgeVerified: true,
			},
			expectError: true,
		},
		{
			name: "weak password",
			input: RegisterRequest{
				Name:        "John Doe",
				Email:       "john2@example.com",
				Password:    "weak",
				AgeVerified: true,
			},
			expectError: true,
		},
		{
			name: "empty name",
			input: RegisterRequest{
				Name:        "",
				Email:       "john3@example.com",
				Password:    "SecurePass123!",
				AgeVerified: true,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := userService.Register(
				context.Background(),
				tc.input.Name,
				tc.input.Email,
				tc.input.Password,
				tc.input.AgeVerified,
			)

			if tc.expectError {
				assert.Error(t, err, "expected error for case: %s", tc.name)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err, "expected no error for case: %s", tc.name)
				assert.NotNil(t, user)
			}
		})
	}
}

func TestRegistration_UserDeletion(t *testing.T) {
	mongoURI, cleanup := setupMongoContainer(t)
	defer cleanup()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	assert.NoError(t, err)
	defer client.Disconnect(context.Background())

	db := client.Database("test_db")

	userRepo, err := repository.NewMongoUserRepository(db)
	assert.NoError(t, err)
	refreshTokenRepo, err := repository.NewMongoRefreshTokenRepository(db)
	assert.NoError(t, err)
	userService := service.NewUserService(userRepo, refreshTokenRepo)

	// Register user
	user, err := userService.Register(
		context.Background(),
		"John Doe",
		"john@example.com",
		"SecurePass123!",
		true,
	)

	assert.NoError(t, err)

	// Delete user
	err = userService.DeleteUser(context.Background(), user.ID.Hex())
	assert.NoError(t, err)

	// Verify user is deleted
	retrieved, err := userService.GetUserByID(context.Background(), user.ID.Hex())
	assert.Error(t, err)
	assert.Nil(t, retrieved)
}
