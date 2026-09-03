package service

import (
	"crypto/rand"
	"fmt"
	"sync"
)

// SSEEvent represents an event sent to connected SSE clients
type SSEEvent struct {
	Type   string      `json:"type"`
	PostID string      `json:"post_id"`
	Data   interface{} `json:"data,omitempty"`
}

// subscription wraps a channel with metadata for tracking
type subscription struct {
	ch chan SSEEvent
}

// SSEManager manages SSE connections and broadcasts events
type SSEManager struct {
	// clients maps room_id -> map of subscription IDs to channels
	clients map[string]map[string]subscription
	mu      sync.RWMutex
}

// NewSSEManager creates a new SSE manager
func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[string]map[string]subscription),
	}
}

// Subscribe subscribes a client to events for a specific room
// Returns a channel that will receive SSEEvent messages and a subscription ID for cleanup
func (m *SSEManager) Subscribe(roomID string) (<-chan SSEEvent, string) {
	ch := make(chan SSEEvent, 10) // buffered channel to prevent blocking

	// Generate a unique ID for this subscription
	subID := generateSubID()

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[roomID]; !exists {
		m.clients[roomID] = make(map[string]subscription)
	}

	m.clients[roomID][subID] = subscription{ch: ch}

	return ch, subID
}

// Unsubscribe removes a client channel from a room
func (m *SSEManager) Unsubscribe(roomID string, subID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, exists := m.clients[roomID]; exists {
		if sub, ok := room[subID]; ok {
			close(sub.ch)
			delete(room, subID)
		}
		// Clean up empty room entries
		if len(m.clients[roomID]) == 0 {
			delete(m.clients, roomID)
		}
	}
}

// PublishNewPost publishes a new post event to all subscribers in a room
func (m *SSEManager) PublishNewPost(roomID string, postID string, data interface{}) {
	event := SSEEvent{
		Type:   "new_post",
		PostID: postID,
		Data:   data,
	}
	m.publish(roomID, event)
}

// SubscribeToPost subscribes a client to events for a specific post
// Returns a channel that will receive SSEEvent messages and a subscription ID for cleanup
func (m *SSEManager) SubscribeToPost(postID string) (<-chan SSEEvent, string) {
	return m.Subscribe(postID) // Reuse room subscription logic with post ID
}

// UnsubscribeFromPost removes a client channel from a post
func (m *SSEManager) UnsubscribeFromPost(postID string, subID string) {
	m.Unsubscribe(postID, subID) // Reuse room unsubscription logic with post ID
}

// PublishCommentCreated publishes a new comment event to all subscribers of a post
func (m *SSEManager) PublishCommentCreated(postID string, commentID string, data interface{}) {
	event := SSEEvent{
		Type:   "new_comment",
		PostID: commentID,
		Data:   data,
	}
	m.publish(postID, event)
}

// publish sends an event to all clients subscribed to a resource (room or post)
func (m *SSEManager) publish(resourceID string, event SSEEvent) {
	m.mu.RLock()
	subscribers := m.clients[resourceID]
	m.mu.RUnlock()

	for _, sub := range subscribers {
		select {
		case sub.ch <- event:
		default:
			// Channel is full, skip to prevent blocking
		}
	}
}

// generateSubID generates a unique subscription ID
func generateSubID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
