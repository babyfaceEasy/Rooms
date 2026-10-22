package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// AppSyncRealtimePublisher publishes events to AWS AppSync Event API (not GraphQL).
// It uses the Event routing mechanism for real-time alerts on mobile devices.
// Publishing is non-blocking: if AppSync is unavailable, the API continues without failing.
//
// Event Schema (in AppSync):
//   type RealtimeEvent @event(level: PUBLIC) {
//     eventType: String! @eventVersion(type: "new_post" | "new_comment")
//     resourceId: String!
//     objectId: String!
//     payload: AWSJSON!
//     publishedAt: AWSDateTime!
//   }
type AppSyncRealtimePublisher struct {
	endpoint string
	apiKey   string
	enabled  bool
	client   *http.Client
	logger   *slog.Logger
}

func NewAppSyncRealtimePublisher(endpoint, apiKey string, enabled bool, logger *slog.Logger) *AppSyncRealtimePublisher {
	if endpoint == "" || apiKey == "" {
		enabled = false
	}

	return &AppSyncRealtimePublisher{
		endpoint: endpoint,
		apiKey:   apiKey,
		enabled:  enabled,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

// PublishNewPost publishes a new_post event to AppSync Event API.
// roomID is the room/channel identifier for routing.
// data should contain post details (title, content, author, etc.)
func (p *AppSyncRealtimePublisher) PublishNewPost(roomID, postID string, data interface{}) {
	p.publish("new_post", roomID, postID, data)
}

// PublishCommentCreated publishes a new_comment event to AppSync Event API.
// postID is the post identifier for routing.
// data should contain comment details (text, author, etc.)
func (p *AppSyncRealtimePublisher) PublishCommentCreated(postID, commentID string, data interface{}) {
	p.publish("new_comment", postID, commentID, data)
}

// publish sends an event to the AppSync Event API endpoint.
// The event is routed to all subscribers listening for that channel.
// AppSync Event API requires events to be stringified JSON in an array with a channel field.
func (p *AppSyncRealtimePublisher) publish(eventType, resourceID, objectID string, data interface{}) {
	if p == nil || !p.enabled || p.endpoint == "" || p.apiKey == "" {
		return
	}

	// Create the event object
	event := map[string]interface{}{
		"eventType":   eventType,
		"resourceId":  resourceID,
		"objectId":    objectID,
		"payload":     data,
		"publishedAt": time.Now().UTC().Format(time.RFC3339Nano),
	}

	// Stringify the event as JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("appsync event marshal failed", slog.Any("error", err), slog.String("event_type", eventType))
		}
		return
	}

	// AppSync Event API requires events as stringified JSON in an array with a channel
	payload := map[string]interface{}{
		"channel": "default/channel", // Channel for routing
		"events":  []string{string(eventJSON)},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("appsync event payload marshal failed", slog.Any("error", err), slog.String("event_type", eventType))
		}
		return
	}

	// Debug: Log the payload being sent
	if p.logger != nil && p.logger.Enabled(context.Background(), slog.LevelDebug) {
		p.logger.Debug("appsync event payload details",
			slog.String("event_type", eventType),
			slog.String("payload_json", string(jsonBody)),
			slog.Int("events_count", len(payload["events"].([]string))),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// POST to AppSync Event API endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("appsync event request creation failed", slog.Any("error", err), slog.String("event_type", eventType))
		}
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("X-Api-User-Agent", "temp_backend/1.0")

	// Debug: Log request details before sending
	if p.logger != nil && p.logger.Enabled(context.Background(), slog.LevelDebug) {
		p.logger.Debug("appsync event request sending",
			slog.String("endpoint", p.endpoint),
			slog.String("method", http.MethodPost),
			slog.String("body", string(jsonBody)),
		)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("appsync event request failed", slog.Any("error", err), slog.String("event_type", eventType))
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var responseBody bytes.Buffer
		if _, readErr := responseBody.ReadFrom(resp.Body); readErr != nil {
			if p.logger != nil {
				p.logger.Warn("appsync event response read failed", slog.Any("error", readErr), slog.String("event_type", eventType), slog.Int("status_code", resp.StatusCode))
			}
			return
		}
		if p.logger != nil {
			p.logger.Warn("appsync event publish failed",
				slog.String("event_type", eventType),
				slog.Int("status_code", resp.StatusCode),
				slog.String("response", strings.TrimSpace(responseBody.String())),
			)
		}
		return
	}

	if p.logger != nil && p.logger.Enabled(context.Background(), slog.LevelDebug) {
		p.logger.Debug("appsync event published",
			slog.String("channel", "default/channel"),
			slog.String("event_type", eventType),
			slog.String("resource_id", resourceID),
			slog.String("object_id", objectID),
		)
	}
}

func (p *AppSyncRealtimePublisher) IsEnabled() bool {
	return p != nil && p.enabled
}

func (p *AppSyncRealtimePublisher) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("AppSyncRealtimePublisher{endpoint=%s, enabled=%t}", p.endpoint, p.enabled)
}
