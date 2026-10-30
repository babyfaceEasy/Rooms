# AWS AppSync Event API Integration Guide for Mobile

This document explains how to integrate AWS AppSync Event API into your mobile client for real-time alerts on new posts and comments.

---

## Table of Contents

1. [Overview](#overview)
2. [Backend Publishing](#backend-publishing)
3. [Mobile Subscription](#mobile-subscription)
4. [Event Schema](#event-schema)
5. [Implementation Examples](#implementation-examples)
6. [Troubleshooting](#troubleshooting)

---

## Overview

### What is AppSync Event API?

AWS AppSync Event API is a real-time event routing system optimized for mobile clients. It's designed for simple, efficient event publishing and subscription:

- ✅ Simple event publishing (send stringified JSON to a channel)
- ✅ Low latency (~100-200ms)
- ✅ Minimal battery impact on mobile devices
- ✅ Scales automatically for millions of concurrent connections
- ✅ API Key authentication for development (IAM for production)
- ✅ WebSocket-based subscriptions for real-time delivery

### Architecture

```
Backend (Go)
├─ REST API (CRUD: /posts, /comments, etc.)
└─ Event Publisher ──→ AppSync Event API
                           ├─ Channel routing
                           └─ WebSocket delivery
                                    ↓
Mobile Client (Subscribed via WebSocket)
├─ Receives events in real-time
└─ Shows notifications to user
```

### Event Flow

```
1. User creates a post/comment via REST API
   POST /api/v1/posts → Backend processes request
                       ↓
2. Backend publishes event to AppSync Event API
   POST https://<appsync>.appsync-api.us-east-1.amazonaws.com/event
   
   Payload:
   {
     "channel": "default/channel",
     "events": ["{stringified JSON}"]
   }
                       ↓
3. AppSync Event Router receives event
   Publishes to WebSocket channel
                       ↓
4. Mobile client receives event via WebSocket
   Shows notification → User refreshes to see new post/comment
```

---

## Backend Publishing

### Implementation

The backend publishes events using the `AppSyncRealtimePublisher` service. When a post or comment is created, an event is automatically published to the Event API:

**New Post Event:**
```go
appSyncPublisher.PublishNewPost(roomID, postID, postData)
```

**New Comment Event:**
```go
appSyncPublisher.PublishCommentCreated(postID, commentID, commentData)
```

### Configuration

The backend requires these environment variables:

```env
# Enable AppSync Event publishing
APPSYNC_ENABLED=true

# Event API endpoint (note: `/event` suffix)
APPSYNC_ENDPOINT=https://<your-appsync-id>.appsync-api.us-east-1.amazonaws.com/event

# API Key for authentication
APPSYNC_API_KEY=da2-xxxxxxxxxxxxxxxxxxxxxxxxxx

# AWS region (must match AppSync endpoint region)
AWS_REGION=us-east-1
```

### Event Publishing Format

The backend sends events to AppSync Event API with this exact format:

```json
{
  "channel": "default/channel",
  "events": [
    "{\"eventType\":\"new_post\",\"resourceId\":\"room_id\",\"objectId\":\"post_id\",\"payload\":{...},\"publishedAt\":\"ISO timestamp\"}"
  ]
}
```

**Key points:**
- `channel` field is required for event routing
- `events` is an array of **stringified JSON objects** (not plain objects)
- Each event inside must be serialized as a JSON string
- AppSync returns: `{"successful": [...], "failed": [...]}`

### Event Payload Structure

Each event contains:

```json
{
  "eventType": "new_post" | "new_comment",
  "resourceId": "room_id" | "post_id",
  "objectId": "post_id" | "comment_id",
  "payload": {
    "id": "...",
    "title": "...",
    "content": "...",
    "author": {...},
    "createdAt": "..."
  },
  "publishedAt": "2026-10-18T19:10:00Z"
}
```
```

---

## Mobile Subscription

### AWS AppSync Events Setup

AppSync Events uses WebSocket connections to deliver real-time updates. Mobile clients subscribe to channels and receive events as they're published.

#### iOS (Swift)

```swift
import AWSAppSync
import URLSession

class RealtimeManager: NSObject {
    private var webSocketTask: URLSessionWebSocketTask?
    private let endpoint = "wss://<your-appsync-id>.appsync-api.us-east-1.amazonaws.com/event"
    private let apiKey = "da2-xxxxxxxxxxxxxxxxxxxxxxxxxx"
    private let logger = Logger()
    
    func subscribeToEvents() {
        var request = URLRequest(url: URL(string: endpoint)!)
        request.setValue(apiKey, forHTTPHeaderField: "x-api-key")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        let urlSession = URLSession(configuration: .default)
        webSocketTask = urlSession.webSocketTask(with: request)
        webSocketTask?.resume()
        
        // Listen for incoming events
        receiveMessage()
    }
    
    private func receiveMessage() {
        webSocketTask?.receive { [weak self] result in
            switch result {
            case .success(let message):
                switch message {
                case .string(let text):
                    if let data = text.data(using: .utf8),
                       let event = try? JSONDecoder().decode(RealtimeEvent.self, from: data) {
                        self?.handleEvent(event)
                    }
                case .data(let data):
                    if let event = try? JSONDecoder().decode(RealtimeEvent.self, from: data) {
                        self?.handleEvent(event)
                    }
                @unknown default:
                    break
                }
                self?.receiveMessage() // Continue listening
            case .failure(let error):
                self?.logger.error("WebSocket error: \(error)")
            }
        }
    }
    
    private func handleEvent(_ event: RealtimeEvent) {
        print("Received event: \(event.eventType)")
        
        switch event.eventType {
        case "new_post":
            handleNewPostNotification(event)
        case "new_comment":
            handleNewCommentNotification(event)
        default:
            break
        }
    }
    
    private func handleNewPostNotification(_ event: RealtimeEvent) {
        let content = UNMutableNotificationContent()
        content.title = "New Post"
        content.body = event.payload?["title"] as? String ?? "A new post has been added"
        content.sound = .default
        content.badge = NSNumber(value: UIApplication.shared.applicationIconBadgeNumber + 1)
        content.userInfo = [
            "postId": event.objectId,
            "roomId": event.resourceId
        ]
        
        let request = UNNotificationRequest(identifier: UUID().uuidString, content: content, trigger: nil)
        UNUserNotificationCenter.current().add(request)
    }
    
    private func handleNewCommentNotification(_ event: RealtimeEvent) {
        let content = UNMutableNotificationContent()
        content.title = "New Comment"
        content.body = event.payload?["text"] as? String ?? "Someone replied to the post"
        content.sound = .default
        content.userInfo = ["commentId": event.objectId, "postId": event.resourceId]
        
        let request = UNNotificationRequest(identifier: UUID().uuidString, content: content, trigger: nil)
        UNUserNotificationCenter.current().add(request)
    }
}

struct RealtimeEvent: Codable {
    let eventType: String
    let resourceId: String
    let objectId: String
    let payload: [String: Any]?
    let publishedAt: String
}
```

#### Android (Kotlin)

```kotlin
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import org.json.JSONObject
import com.google.gson.Gson

class RealtimeManager(context: Context) {
    private val context = context
    private val apiKey = "da2-xxxxxxxxxxxxxxxxxxxxxxxxxx"
    private val endpoint = "wss://<your-appsync-id>.appsync-api.us-east-1.amazonaws.com/event"
    private var webSocket: WebSocket? = null
    private val gson = Gson()
    
    fun subscribeToEvents() {
        val client = OkHttpClient()
        val request = Request.Builder()
            .url(endpoint)
            .addHeader("x-api-key", apiKey)
            .addHeader("Content-Type", "application/json")
            .build()
        
        webSocket = client.newWebSocket(request, object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                try {
                    val event = gson.fromJson(text, RealtimeEvent::class.java)
                    handleEvent(event)
                } catch (e: Exception) {
                    Log.e("RealtimeManager", "Failed to parse event: $e")
                }
            }
            
            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                Log.e("RealtimeManager", "WebSocket error: ${t.message}")
            }
        })
    }
    
    private fun handleEvent(event: RealtimeEvent) {
        when (event.eventType) {
            "new_post" -> handleNewPostNotification(event)
            "new_comment" -> handleNewCommentNotification(event)
        }
    }
    
    private fun handleNewPostNotification(event: RealtimeEvent) {
        val title = "New Post"
        val body = (event.payload as? Map<*, *>)?.get("title")?.toString() ?: "New post added"
        
        showNotification(title, body, mapOf(
            "postId" to event.objectId,
            "roomId" to event.resourceId
        ))
    }
    
    private fun handleNewCommentNotification(event: RealtimeEvent) {
        val title = "New Comment"
        val body = (event.payload as? Map<*, *>)?.get("text")?.toString() ?: "Someone replied"
        
        showNotification(title, body, mapOf(
            "commentId" to event.objectId,
            "postId" to event.resourceId
        ))
    }
    
    private fun showNotification(title: String, body: String, data: Map<String, String>) {
        val notificationManager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        
        val notification = NotificationCompat.Builder(context, "channel_id")
            .setContentTitle(title)
            .setContentText(body)
            .setSmallIcon(R.drawable.ic_notification)
            .setAutoCancel(true)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .build()
        
        notificationManager.notify(System.currentTimeMillis().toInt(), notification)
    }
}

data class RealtimeEvent(
    val eventType: String,
    val resourceId: String,
    val objectId: String,
    val payload: Any?,
    val publishedAt: String
)
```

#### React Native / Expo

```javascript
import * as Notifications from 'expo-notifications';

class RealtimeManager {
    constructor(apiKey) {
        this.apiKey = apiKey;
        this.endpoint = 'wss://<your-appsync-id>.appsync-api.us-east-1.amazonaws.com/event';
        this.webSocket = null;
    }
    
    subscribeToEvents() {
        this.webSocket = new WebSocket(this.endpoint, {
            headers: {
                'x-api-key': this.apiKey,
                'Content-Type': 'application/json'
            }
        });
        
        this.webSocket.onopen = () => {
            console.log('WebSocket connected');
        };
        
        this.webSocket.onmessage = (event) => {
            try {
                const realtimeEvent = JSON.parse(event.data);
                this.handleEvent(realtimeEvent);
            } catch (error) {
                console.error('Failed to parse event:', error);
            }
        };
        
        this.webSocket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
        
        this.webSocket.onclose = () => {
            console.log('WebSocket disconnected');
            // Reconnect after delay
            setTimeout(() => this.subscribeToEvents(), 3000);
        };
    }
    
    handleEvent(event) {
        switch (event.eventType) {
            case 'new_post':
                this.handleNewPostNotification(event);
                break;
            case 'new_comment':
                this.handleNewCommentNotification(event);
                break;
        }
    }
    
    async handleNewPostNotification(event) {
        await Notifications.scheduleNotificationAsync({
            content: {
                title: 'New Post',
                body: event.payload?.title || 'A new post has been added',
                data: {
                    postId: event.objectId,
                    roomId: event.resourceId
                },
                sound: 'default'
            },
            trigger: null
        });
    }
    
    async handleNewCommentNotification(event) {
        await Notifications.scheduleNotificationAsync({
            content: {
                title: 'New Comment',
                body: event.payload?.text || 'Someone replied to the post',
                data: {
                    commentId: event.objectId,
                    postId: event.resourceId
                },
                sound: 'default'
            },
            trigger: null
        });
    }
    
    disconnect() {
        if (this.webSocket) {
            this.webSocket.close();
        }
    }
}

// Usage
const realtimeManager = new RealtimeManager('da2-xxxxxxxxxxxxxxxxxxxxxxxxxx');
realtimeManager.subscribeToEvents();
```

---

## Event Schema

### AppSync Event API (Event-based, not GraphQL)

### Event Types

#### New Post Event

**When:** A user creates a new post in a room

**Event Type:** `new_post`

**Resource ID:** `room_id` (the room where the post was created)

**Object ID:** `post_id`

**Payload Example:**
```json
{
  "eventType": "new_post",
  "resourceId": "672d45c5f1234567890abcde",
  "objectId": "672d45c6f1234567890abcdf",
  "payload": {
    "id": "672d45c6f1234567890abcdf",
    "title": "Check out this amazing feature",
    "content": "I just discovered an awesome feature that could help us...",
    "author": {
      "id": "672d45c0f1234567890abcd0",
      "name": "John Doe",
      "email": "john@example.com"
    },
    "roomId": "672d45c5f1234567890abcde",
    "attachments": [],
    "createdAt": "2026-10-17T15:04:14.754868Z",
    "updatedAt": "2026-10-17T15:04:14.754868Z"
  },
  "publishedAt": "2026-10-17T15:04:14.754868Z"
}
```

#### New Comment Event

**When:** A user adds a comment to a post

**Event Type:** `new_comment`

**Resource ID:** `post_id` (the post being commented on)

**Object ID:** `comment_id`

**Payload Example:**
```json
{
  "eventType": "new_comment",
  "resourceId": "672d45c6f1234567890abcdf",
  "objectId": "672d45c8f1234567890abce1",
  "payload": {
    "id": "672d45c8f1234567890abce1",
    "text": "This is a great idea! Let's discuss this more.",
    "author": {
      "id": "672d45c1f1234567890abcd1",
      "name": "Jane Smith",
      "email": "jane@example.com"
    },
    "postId": "672d45c6f1234567890abcdf",
    "createdAt": "2026-10-17T15:05:22.123456Z",
    "updatedAt": "2026-10-17T15:05:22.123456Z"
  },
  "publishedAt": "2026-10-17T15:05:22.123456Z"
}
```

---

## Implementation Examples

### Complete Event API Setup - iOS (Swift)

```swift
import URLSession
import UserNotifications

class AppSyncRealtimeManager: NSObject, UNUserNotificationCenterDelegate {
    static let shared = AppSyncRealtimeManager()
    
    private let endpoint = "wss://<your-appsync-id>.appsync-api.us-east-1.amazonaws.com/event"
    private let apiKey = "da2-xxxxxxxxxxxxxxxxxxxxxxxxxx"
    private var webSocketTask: URLSessionWebSocketTask?
    
    override init() {
        super.init()
        setupNotifications()
    }
    
    private func setupNotifications() {
        UNUserNotificationCenter.current().delegate = self
        UNUserNotificationCenter.current().requestAuthorization(
            options: [.alert, .sound, .badge]
        ) { granted, _ in
            if granted {
                DispatchQueue.main.async {
                    UIApplication.shared.registerForRemoteNotifications()
                }
            }
        }
    }
    
    func connect() {
        var request = URLRequest(url: URL(string: endpoint)!)
        request.setValue(apiKey, forHTTPHeaderField: "x-api-key")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        let urlSession = URLSession(configuration: .default)
        webSocketTask = urlSession.webSocketTask(with: request)
        webSocketTask?.resume()
        
        print("AppSync realtime WebSocket connecting...")
        receiveMessage()
    }
    
    private func receiveMessage() {
        webSocketTask?.receive { [weak self] result in
            switch result {
            case .success(let message):
                switch message {
                case .string(let text):
                    if let data = text.data(using: .utf8),
                       let event = try? JSONDecoder().decode(RealtimeEvent.self, from: data) {
                        self?.handleEvent(event)
                    }
                case .data(let data):
                    if let event = try? JSONDecoder().decode(RealtimeEvent.self, from: data) {
                        self?.handleEvent(event)
                    }
                @unknown default:
                    break
                }
                self?.receiveMessage()
            case .failure(let error):
                print("WebSocket error: \(error)")
                // Reconnect after delay
                DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
                    self?.connect()
                }
            }
        }
    }
    
    private func handleEvent(_ event: RealtimeEvent) {
        switch event.eventType {
        case "new_post":
            handleNewPostNotification(event)
        case "new_comment":
            handleNewCommentNotification(event)
        default:
            break
        }
    }
    
    private func handleNewPostNotification(_ event: RealtimeEvent) {
        let content = UNMutableNotificationContent()
        content.title = "New Post"
        content.body = event.payload?["title"] as? String ?? "A new post has been added"
        content.sound = .default
        content.badge = NSNumber(value: UIApplication.shared.applicationIconBadgeNumber + 1)
        content.userInfo = [
            "postId": event.objectId,
            "roomId": event.resourceId
        ]
        
        let request = UNNotificationRequest(identifier: UUID().uuidString, content: content, trigger: nil)
        UNUserNotificationCenter.current().add(request)
    }
    
    private func handleNewCommentNotification(_ event: RealtimeEvent) {
        let content = UNMutableNotificationContent()
        content.title = "New Comment"
        content.body = event.payload?["text"] as? String ?? "Someone replied to the post"
        content.sound = .default
        content.userInfo = [
            "commentId": event.objectId,
            "postId": event.resourceId
        ]
        
        let request = UNNotificationRequest(identifier: UUID().uuidString, content: content, trigger: nil)
        UNUserNotificationCenter.current().add(request)
    }
    
    func disconnect() {
        webSocketTask?.cancel(with: .goingAway, reason: nil)
    }
}

struct RealtimeEvent: Codable {
    let eventType: String
    let resourceId: String
    let objectId: String
    let payload: [String: AnyCodable]?
    let publishedAt: String
}

// Helper to decode dynamic JSON objects
enum AnyCodable: Codable {
    case string(String)
    case number(Double)
    case bool(Bool)
    case dict([String: AnyCodable])
    case array([AnyCodable])
    
    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        
        if let value = try? container.decode(String.self) {
            self = .string(value)
        } else if let value = try? container.decode(Double.self) {
            self = .number(value)
        } else if let value = try? container.decode(Bool.self) {
            self = .bool(value)
        } else if let value = try? container.decode([String: AnyCodable].self) {
            self = .dict(value)
        } else if let value = try? container.decode([AnyCodable].self) {
            self = .array(value)
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Cannot decode AnyCodable")
        }
    }
    
    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        switch self {
        case .string(let value):
            try container.encode(value)
        case .number(let value):
            try container.encode(value)
        case .bool(let value):
            try container.encode(value)
        case .dict(let value):
            try container.encode(value)
        case .array(let value):
            try container.encode(value)
        }
    }
}

// Usage in AppDelegate
class AppDelegate: UIResponder, UIApplicationDelegate {
    func applicationDidFinishLaunching(_ application: UIApplication) {
        AppSyncRealtimeManager.shared.connect()
    }
}
```

### Complete Event API Setup - Android (Kotlin)

Already provided above in the Mobile Subscription section.

### Complete Event API Setup - React Native

Already provided above in the Mobile Subscription section.
        
        content.userInfo = [
            "postId": post.objectId,
            "roomId": post.resourceId
        ]
        
        let request = UNNotificationRequest(
            identifier: UUID().uuidString,
            content: content,
            trigger: nil
        )
    
    private func handleNewCommentNotification(_ event: RealtimeEvent) {
        let content = UNMutableNotificationContent()
        content.title = "New Comment"
        content.body = event.payload?["text"] as? String ?? "Someone replied to the post"
        content.sound = .default
        
        content.userInfo = [
            "commentId": event.objectId,
            "postId": event.resourceId
        ]
        
        let request = UNNotificationRequest(
            identifier: UUID().uuidString,
            content: content,
            trigger: nil
        )
        
        UNUserNotificationCenter.current().add(request) { error in
            if let error = error {
                print("Failed to schedule notification: \(error)")
            }
        }
    }
    
    // Handle notification tap
    func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        let userInfo = response.notification.request.content.userInfo
        
        if let postId = userInfo["postId"] as? String {
            // Navigate to post detail
            NotificationCenter.default.post(
                name: NSNotification.Name("navigateToPost"),
                object: nil,
                userInfo: ["postId": postId]
            )
        }
        
        completionHandler()
    }
}
```

---

## Troubleshooting

### Common Issues

#### 1. Connection Refused / Timeout
**Error:** `Connection refused` or `Request timeout`

**Causes:**
- AppSync endpoint is incorrect (should end with `/event`, not `/graphql`)
- API Key is expired or invalid
- AWS region mismatch

**Solution:**
```bash
# Verify endpoint
curl -H "x-api-key: da2-xxxxx" \
  https://<your-appsync-id>.appsync-api.us-east-1.amazonaws.com/event
```

#### 2. 400 BadRequest / Missing fields
**Error:** `{"errors": [{"message": "Missing events"}]}`

**Cause:** Payload structure doesn't match AppSync schema

**Solution:** Ensure payload has all required fields:
- `eventType` (string)
- `resourceId` (string)
- `objectId` (string)
- `payload` (JSON)
- `publishedAt` (ISO datetime)

#### 3. 401 Unauthorized
**Error:** `401 Unauthorized`

**Causes:**
- API Key is invalid or expired
- Wrong authorization header format

**Solution:**
```javascript
// Correct header format
headers: {
  "x-api-key": "da2-xxxxxxxxxxxxxxxxxxxxxxxxxx",
  "Content-Type": "application/json"
}
```

#### 4. Subscription not receiving events
**Causes:**
- Mobile client is subscribing to wrong resource ID
- Backend is disabled (`APPSYNC_ENABLED=false`)
- Event type doesn't match subscription

**Solution:**
- Verify `APPSYNC_ENABLED=true` on backend
- Check backend logs: `grep "appsync event published" logs`
- Verify mobile client is subscribing to correct room/post ID

#### 5. High latency / Events arriving late
**Cause:** Network or AWS region issues

**Solution:**
- Check AWS region matches endpoint
- Verify mobile network connectivity
- Monitor CloudWatch metrics in AppSync console

---

## Best Practices

### 1. Always Refetch Data

Don't rely only on event payload. When you receive an alert:

```swift
// ❌ Don't do this (event payload might be incomplete)
displayPost(from: event.payload)

// ✅ Do this (fetch fresh data from server)
if let postId = event.objectId {
    fetchPost(id: postId) { post in
        displayPost(post)
    }
}
```

### 2. Handle Connection Interruptions

```swift
// iOS example
func subscribeToRoom(_ roomId: String, retryCount: Int = 0) {
    let maxRetries = 3
    
    appSyncClient?.subscribe(subscription: ...) { result, _, error in
        if let error = error {
            if retryCount < maxRetries {
                print("Subscription failed, retrying...")
                DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                    self.subscribeToRoom(roomId, retryCount: retryCount + 1)
                }
            } else {
                print("Subscription failed after \(maxRetries) attempts")
            }
            return
        }
        // Handle subscription...
    }
}
```

### 3. Unsubscribe When Not Needed

```swift
// Clean up subscriptions when user leaves a room
func unsubscribeFromRoom() {
    subscriptions.forEach { watcher in
        watcher.cancel()
    }
    subscriptions.removeAll()
}
```

### 4. Rate Limit Local Notifications

```swift
// Don't bombard user with notifications
private var lastNotificationTime: Date?
private let notificationDebounceInterval: TimeInterval = 1.0

private func shouldShowNotification() -> Bool {
    let now = Date()
    if let lastTime = lastNotificationTime,
       now.timeIntervalSince(lastTime) < notificationDebounceInterval {
        return false
    }
    lastNotificationTime = now
    return true
}
```

### 5. Log Events for Debugging

```swift
private func logEvent(_ event: RealtimeEvent) {
    let formatter = ISO8601DateFormatter()
    let timestamp = formatter.string(from: Date())
    
    print("""
    [AppSync] \(timestamp)
    Type: \(event.eventType)
    Resource: \(event.resourceId)
    Object: \(event.objectId)
    Published: \(event.publishedAt)
    """)
}
```

---

## Support & Questions

If you encounter issues:

1. **Check backend logs:**
   ```bash
   docker logs temp_backend_api | grep appsync
   ```

2. **Check AWS AppSync CloudWatch metrics:**
   - AWS Console → AppSync → Your API → Monitoring

3. **Test endpoint connectivity:**
   ```bash
   curl -v -H "x-api-key: <KEY>" \
     https://<appsync-id>.appsync-api.us-east-1.amazonaws.com/event
   ```

4. **Review backend code:**
   - [appsync_publisher.go](/internal/service/appsync_publisher.go)
   - [post_handler.go](/internal/handler/post_handler.go)
   - [comment_handler.go](/internal/handler/comment_handler.go)

