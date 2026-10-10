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

// AppSyncRealtimePublisher sends realtime events to an AWS AppSync GraphQL endpoint using API key auth.
// It is intentionally non-blocking for request handling: if the broker is disabled or unreachable,
// the API continues without failing the original mutation.
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

func (p *AppSyncRealtimePublisher) PublishNewPost(roomID, postID string, data interface{}) {
	p.publish("new_post", roomID, postID, data)
}

func (p *AppSyncRealtimePublisher) PublishCommentCreated(postID, commentID string, data interface{}) {
	p.publish("new_comment", postID, commentID, data)
}

func (p *AppSyncRealtimePublisher) publish(eventType, resourceID, objectID string, data interface{}) {
	if p == nil || !p.enabled || p.endpoint == "" || p.apiKey == "" {
		return
	}

	payload := map[string]interface{}{
		"event_type": eventType,
		"resource_id": resourceID,
		"object_id": objectID,
		"payload": data,
		"published_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	body := map[string]interface{}{
		"query": `mutation PublishRealtimeEvent($input: PublishRealtimeEventInput!) { publishRealtimeEvent(input: $input) { id } }`,
		"variables": map[string]interface{}{
			"input": payload,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("appsync realtime publish marshal failed", slog.Any("error", err))
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("appsync realtime publish request creation failed", slog.Any("error", err))
		}
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("appsync realtime publish request failed", slog.Any("error", err), slog.String("event_type", eventType))
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var responseBody bytes.Buffer
		if _, readErr := responseBody.ReadFrom(resp.Body); readErr != nil {
			if p.logger != nil {
				p.logger.Warn("appsync realtime publish response read failed", slog.Any("error", readErr), slog.String("event_type", eventType), slog.Int("status_code", resp.StatusCode))
			}
			return
		}
		if p.logger != nil {
			p.logger.Warn("appsync realtime publish failed",
				slog.String("event_type", eventType),
				slog.Int("status_code", resp.StatusCode),
				slog.String("response", strings.TrimSpace(responseBody.String())),
			)
		}
		return
	}

	if p.logger != nil && p.logger.Enabled(context.Background(), slog.LevelDebug) {
		p.logger.Debug("appsync realtime publish succeeded",
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
