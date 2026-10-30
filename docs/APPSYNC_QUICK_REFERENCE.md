# AWS AppSync Event API - Quick Reference

Quick setup guide for mobile engineers integrating AWS AppSync Event API.

---

## Configuration

### Backend Provides

Your backend publishes events to AppSync with this configuration:

```
Endpoint:    https://<appsync-id>.appsync-api.us-east-1.amazonaws.com/event
Region:      us-east-1
Auth Mode:   API_KEY
API Key:     da2-xxxxxxxxxxxxxxxxxxxxxxxxxx
```

**Important:** Endpoint ends with `/event`, not `/graphql` — this is the Event API, not GraphQL.

---

## Event Schema

### Event Structure

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
    "createdAt": "2026-10-18T..."
  },
  "publishedAt": "2026-10-18T..."
}
```

### Event Types

| Event Type | Fired When | Resource ID | Object ID |
|---|---|---|---|
| `new_post` | User creates a post in a room | `room_id` | `post_id` |
| `new_comment` | User comments on a post | `post_id` | `comment_id` |

---

## Mobile Integration

### iOS (Swift)

```swift
import URLSession
import UserNotifications

class RealtimeManager {
    let endpoint = "wss://<appsync-id>.appsync-api.us-east-1.amazonaws.com/event"
    let apiKey = "da2-xxxxxxxxxxxxxxxxxxxxxxxxxx"
    var webSocket: URLSessionWebSocketTask?
    
    func connect() {
        var request = URLRequest(url: URL(string: endpoint)!)
        request.setValue(apiKey, forHTTPHeaderField: "x-api-key")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        webSocket = URLSession.shared.webSocketTask(with: request)
        webSocket?.resume()
        receive()
    }
    
    private func receive() {
        webSocket?.receive { [weak self] result in
            switch result {
            case .success(let message):
                switch message {
                case .string(let text):
                    if let data = text.data(using: .utf8),
                       let event = try? JSONDecoder().decode(RealtimeEvent.self, from: data) {
                        self?.handleEvent(event)
                    }
                default: break
                }
                self?.receive()
            case .failure:
                // Reconnect after delay
                DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
                    self?.connect()
                }
            }
        }
    }
    
    private func handleEvent(_ event: RealtimeEvent) {
        let content = UNMutableNotificationContent()
        
        switch event.eventType {
        case "new_post":
            content.title = "New Post"
            content.body = event.payload?["title"] as? String ?? "New post added"
        case "new_comment":
            content.title = "New Comment"
            content.body = event.payload?["text"] as? String ?? "Someone replied"
        default: return
        }
        
        content.userInfo = ["eventId": event.objectId]
        let request = UNNotificationRequest(identifier: UUID().uuidString, content: content, trigger: nil)
        UNUserNotificationCenter.current().add(request)
    }
    
    func disconnect() {
        webSocket?.cancel(with: .goingAway, reason: nil)
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

### Android (Kotlin)

```kotlin
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import org.json.JSONObject

class RealtimeManager(context: Context) {
    private val endpoint = "wss://<appsync-id>.appsync-api.us-east-1.amazonaws.com/event"
    private val apiKey = "da2-xxxxxxxxxxxxxxxxxxxxxxxxxx"
    private var webSocket: WebSocket? = null
    
    fun connect() {
        val client = OkHttpClient()
        val request = Request.Builder()
            .url(endpoint)
            .addHeader("x-api-key", apiKey)
            .addHeader("Content-Type", "application/json")
            .build()
        
        webSocket = client.newWebSocket(request, object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                try {
                    val event = JSONObject(text)
                    handleEvent(event)
                } catch (e: Exception) {
                    Log.e("RealtimeManager", "Parse error: $e")
                }
            }
            
            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                // Reconnect after delay
                Handler(Looper.getMainLooper()).postDelayed({ connect() }, 3000)
            }
        })
    }
    
    private fun handleEvent(event: JSONObject) {
        val eventType = event.getString("eventType")
        val objectId = event.getString("objectId")
        val payload = event.getJSONObject("payload")
        
        val title = when (eventType) {
            "new_post" -> "New Post"
            "new_comment" -> "New Comment"
            else -> return
        }
        
        val body = when (eventType) {
            "new_post" -> payload.optString("title", "New post added")
            "new_comment" -> payload.optString("text", "Someone replied")
            else -> return
        }
        
        showNotification(title, body, objectId)
    }
    
    private fun showNotification(title: String, body: String, eventId: String) {
        val notificationManager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        
        val notification = NotificationCompat.Builder(context, "channel_id")
            .setContentTitle(title)
            .setContentText(body)
            .setSmallIcon(R.drawable.ic_notification)
            .setAutoCancel(true)
            .build()
        
        notificationManager.notify(System.currentTimeMillis().toInt(), notification)
    }
    
    fun disconnect() {
        webSocket?.close(1000, "Done")
    }
}
```

### React Native / Expo

```javascript
import * as Notifications from 'expo-notifications';

class RealtimeManager {
    constructor(apiKey) {
        this.endpoint = 'wss://<appsync-id>.appsync-api.us-east-1.amazonaws.com/event';
        this.apiKey = apiKey;
        this.webSocket = null;
    }
    
    connect() {
        this.webSocket = new WebSocket(this.endpoint, {
            headers: {
                'x-api-key': this.apiKey,
                'Content-Type': 'application/json'
            }
        });
        
        this.webSocket.onmessage = (event) => {
            try {
                const realtimeEvent = JSON.parse(event.data);
                this.handleEvent(realtimeEvent);
            } catch (error) {
                console.error('Parse error:', error);
            }
        };
        
        this.webSocket.onerror = () => {
            // Reconnect after delay
            setTimeout(() => this.connect(), 3000);
        };
    }
    
    async handleEvent(event) {
        const { eventType, objectId, payload } = event;
        
        let title, body;
        
        if (eventType === 'new_post') {
            title = 'New Post';
            body = payload?.title || 'New post added';
        } else if (eventType === 'new_comment') {
            title = 'New Comment';
            body = payload?.text || 'Someone replied';
        } else {
            return;
        }
        
        await Notifications.scheduleNotificationAsync({
            content: {
                title,
                body,
                data: { eventId: objectId },
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
const manager = new RealtimeManager('da2-xxxxxxxxxxxxxxxxxxxxxxxxxx');
manager.connect();
```

---

## Important Notes

- **WebSocket Protocol**: Use `wss://` (secure WebSocket) for production
- **API Key**: Store safely; use IAM authentication in production
- **Channel**: Backend publishes to `default/channel` for event routing
- **Reconnection**: Implement exponential backoff on connection failure
- **Payload**: Each `payload` field contains event-specific data
- **Publishing Format**: Backend sends `{"channel": "default/channel", "events": [stringified JSON]}`

---

## Troubleshooting

| Issue | Cause | Solution |
|---|---|---|
| Connection refused | Wrong endpoint (should be `/event`) | Use `https://<id>.appsync-api.us-east-1.amazonaws.com/event` |
| 401 Unauthorized | Invalid or expired API key | Verify key in AWS AppSync dashboard |
| Timeout | Network or firewall issue | Check WebSocket connectivity; use WSCat to test |
| No events received | Not subscribed properly | Ensure WebSocket is open and listening for messages |

---

## Testing

Test event publishing with curl:

```bash
curl -X POST https://<appsync-id>.appsync-api.us-east-1.amazonaws.com/event \
  -H "x-api-key: da2-xxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "default/channel",
    "events": [
      "{\"eventType\":\"new_post\",\"resourceId\":\"room123\",\"objectId\":\"post456\",\"payload\":{\"title\":\"Test Post\"},\"publishedAt\":\"2026-10-18T00:00:00Z\"}"
    ]
  }'
```

Expected response:
```json
{
  "successful": [{"identifier": "...", "index": 0}],
  "failed": []
}
```

---

## For More Details

See [APPSYNC_MOBILE_INTEGRATION.md](./APPSYNC_MOBILE_INTEGRATION.md) for comprehensive implementation guide with production considerations.
