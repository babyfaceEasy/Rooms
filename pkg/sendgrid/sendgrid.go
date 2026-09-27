package sendgrid

import "github.com/sendgrid/sendgrid-go"

// NewClient creates and returns a new SendGrid client.
func NewClient(apiKey string) *sendgrid.Client {
	return sendgrid.NewSendClient(apiKey)
}
