package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"temp_backend/config"
	"temp_backend/internal/api"
	"temp_backend/internal/handler"
	"temp_backend/internal/repository"
	"temp_backend/internal/service"
	"temp_backend/pkg/mongodb"
	pks3 "temp_backend/pkg/s3"
	pksendgrid "temp_backend/pkg/sendgrid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load failed: %v\n", err)
		os.Exit(1)
	}

	logger := newLogger(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mongoClient, err := mongodb.InitMongoDB(ctx, cfg.Mongo.URI, cfg.Mongo.ConnectTimeout)
	if err != nil {
		logger.Error("mongo connection failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(disconnectCtx)
	}()

	s3Client, err := pks3.InitS3Config(
		ctx,
		cfg.S3.EndpointURL,
		cfg.S3.Region,
		cfg.S3.AccessKey,
		cfg.S3.SecretKey,
	)
	if err != nil {
		logger.Error("s3 client creation failed", slog.Any("error", err))
		os.Exit(1)
	}

	if err := ensureBucket(ctx, s3Client, cfg); err != nil {
		logger.Error("s3 bucket check failed", slog.Any("error", err))
		os.Exit(1)
	}

	// Item repositories and service
	itemRepo := repository.NewMongoItemRepository(mongoClient.Database(cfg.Mongo.Database))
	storageRepo := repository.NewS3Repository(s3Client, cfg.S3.Bucket, cfg.S3.PublicURL)
	itemService := service.NewItemService(itemRepo, storageRepo)
	itemHandler := handler.NewItemHandler(itemService)

	// User repository and service
	userRepo, err := repository.NewMongoUserRepository(mongoClient.Database(cfg.Mongo.Database))
	if err != nil {
		logger.Error("user repository initialization failed", slog.Any("error", err))
		os.Exit(1)
	}

	// Auth repositories
	refreshTokenRepo, err := repository.NewMongoRefreshTokenRepository(mongoClient.Database(cfg.Mongo.Database))
	if err != nil {
		logger.Error("refresh token repository initialization failed", slog.Any("error", err))
		os.Exit(1)
	}

	// Email repositories and service (initialized early for use in user handler)
	emailRepo, err := repository.NewMongoEmailRepository(mongoClient.Database(cfg.Mongo.Database))
	if err != nil {
		logger.Error("email repository initialization failed", slog.Any("error", err))
		os.Exit(1)
	}

	sendgridClient := pksendgrid.NewClient(cfg.SendGrid.APIKey)
	emailService := service.NewEmailService(
		emailRepo,
		sendgridClient,
		cfg.SendGrid.SenderEmail,
		cfg.SendGrid.Enabled,
		cfg.SendGrid.VerificationTemplateID,
		cfg.SendGrid.PasswordResetTemplateID,
		logger,
	)

	userService := service.NewUserService(userRepo, refreshTokenRepo)
	userHandler := handler.NewUserHandler(userService, emailService)

	// Auth services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	// Room repositories and service
	roomRepo := repository.NewMongoRoomRepository(mongoClient.Database(cfg.Mongo.Database))
	roomService := service.NewRoomService(roomRepo, userRepo)
	roomHandler := handler.NewRoomHandler(roomService)

	// Post repositories and service
	postRepo := repository.NewMongoPostRepository(mongoClient.Database(cfg.Mongo.Database))
	postService := service.NewPostService(postRepo, roomRepo)
	postHandler := handler.NewPostHandler(postService, storageRepo, roomRepo, userRepo)

	// Comment repositories and service
	commentRepo := repository.NewMongoCommentRepository(mongoClient.Database(cfg.Mongo.Database).Collection("comments"))
	commentService := service.NewCommentService(commentRepo, postRepo)
	commentHandler := handler.NewCommentHandler(commentService, userRepo)

	server := api.NewServer(cfg, logger, itemHandler, userHandler, authHandler, roomHandler, postHandler, commentHandler, authService)

	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", slog.Any("error", err))
			stop()
		}
	}()

	logger.Info("server started", slog.String("port", cfg.App.Port))

	<-ctx.Done()

	logger.Info("shutting down gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("server stopped")
}

func ensureBucket(ctx context.Context, client *s3.Client, cfg config.Config) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.S3.Bucket),
	})
	if err == nil {
		return nil
	}

	input := &s3.CreateBucketInput{
		Bucket: aws.String(cfg.S3.Bucket),
	}

	if cfg.S3.Region != "us-east-1" && cfg.S3.Region != "" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(cfg.S3.Region),
		}
	}

	_, err = client.CreateBucket(ctx, input)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "BucketAlreadyOwnedByYou") || strings.Contains(errStr, "BucketAlreadyExists") {
			return nil
		}
		return err
	}
	return nil
}

func newLogger(cfg config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.App.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}
