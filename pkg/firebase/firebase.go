package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// InitializeApp initializes a Firebase app from a service account key file
func InitializeApp(ctx context.Context, credentialsPath string) (*firebase.App, *messaging.Client, error) {
	if credentialsPath == "" {
		return nil, nil, fmt.Errorf("firebase credentials path is empty")
	}

	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("firebase app initialization failed: %w", err)
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("firebase messaging client initialization failed: %w", err)
	}

	return app, messagingClient, nil
}
