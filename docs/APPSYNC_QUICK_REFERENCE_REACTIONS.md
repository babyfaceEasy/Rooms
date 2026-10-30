# AppSync Reaction Events - Quick Reference

## Single Subscription Covers Everything

```javascript
// One subscription = all reaction events
subscribe(channel: "default/channel")
  ↓ Receives all: post_validated, post_validation_removed, post_respected, post_respect_removed
```

## What Events Look Like (Real Examples)

### Event 1: post_validated (User Validates Post)
```json
{
  "eventType": "post_validated",
  "payload": "{\"postId\":\"6adc9a1305fb43cdf96d7631\",\"reactorUserId\":\"6a80ab36948abf3b1594c4c6\",\"reactorName\":\"Jane Doe\",\"reactorProfilePicture\":\"https://bouk-room-dev.s3.amazonaws.com/profiles/jane.jpg\",\"totalValidationCount\":5}"
}
```
**To parse:** `JSON.parse(event.payload)`  
**Action:** Update counter to 5, show "Jane Doe validated your post"

### Event 2: post_validation_removed (User Removes Validation)
```json
{
  "eventType": "post_validation_removed",
  "payload": "{\"postId\":\"6adc9a1305fb43cdf96d7631\",\"reactorUserId\":\"6a80ab36948abf3b1594c4c6\",\"reactorName\":\"Jane Doe\",\"reactorProfilePicture\":\"https://bouk-room-dev.s3.amazonaws.com/profiles/jane.jpg\",\"totalValidationCount\":4}"
}
```
**To parse:** `JSON.parse(event.payload)`  
**Action:** Update counter to 4, remove from reactions

### Event 3: post_respected (User Respects Post)
```json
{
  "eventType": "post_respected",
  "payload": "{\"postId\":\"6adc9a1305fb43cdf96d7631\",\"reactorUserId\":\"6a80ab3695abf3b1594c4d22\",\"reactorName\":\"John Smith\",\"reactorProfilePicture\":\"https://bouk-room-dev.s3.amazonaws.com/profiles/john.jpg\",\"totalRespectCount\":8}"
}
```
**To parse:** `JSON.parse(event.payload)`  
**Action:** Update counter to 8, show "John Smith respected your post"

### Event 4: post_respect_removed (User Removes Respect)
```json
{
  "eventType": "post_respect_removed",
  "payload": "{\"postId\":\"6adc9a1305fb43cdf96d7631\",\"reactorUserId\":\"6a80ab3695abf3b1594c4d22\",\"reactorName\":\"John Smith\",\"reactorProfilePicture\":\"https://bouk-room-dev.s3.amazonaws.com/profiles/john.jpg\",\"totalRespectCount\":7}"
}
```
**To parse:** `JSON.parse(event.payload)`  
**Action:** Update counter to 7, remove from reactions

## 4 Event Types

### 1. `post_validated` → User validated a post
```json
{
  "eventType": "post_validated",
  "payload": {
    "reactorName": "Jane Doe",
    "reactorProfilePicture": "https://...",
    "totalValidationCount": 5
  }
}
```
**Action:** Show notification, increment counter

---

### 2. `post_validation_removed` → User removed validation
```json
{
  "eventType": "post_validation_removed",
  "payload": {
    "reactorName": "Jane Doe",
    "totalValidationCount": 4
  }
}
```
**Action:** Decrement counter, remove from reaction list

---

### 3. `post_respected` → User respected a post
```json
{
  "eventType": "post_respected",
  "payload": {
    "reactorName": "John Smith",
    "reactorProfilePicture": "https://...",
    "totalRespectCount": 8
  }
}
```
**Action:** Show notification, increment counter

---

### 4. `post_respect_removed` → User removed respect
```json
{
  "eventType": "post_respect_removed",
  "payload": {
    "reactorName": "John Smith",
    "totalRespectCount": 7
  }
}
```
**Action:** Decrement counter, remove from reaction list

---

## Minimal React Native Example

```javascript
import { API } from 'aws-amplify';

API.graphql(
  graphqlOperation(`
    subscription {
      onRealtimeEvent(channel: "default/channel") {
        eventType
        payload
      }
    }
  `)
).subscribe({
  next: (data) => {
    const { eventType, payload } = data.value.data.onRealtimeEvent;
    
    if (eventType.includes('validation')) {
      updateValidationCount(payload.totalValidationCount);
    } else if (eventType.includes('respect')) {
      updateRespectCount(payload.totalRespectCount);
    }
    
    showNotification(`${payload.reactorName} ${eventType.replace(/_/g, ' ')}`);
  }
});
```

## Minimal iOS Example

```swift
appSyncClient.subscribe(
  subscription: OnRealtimeEventSubscription(channel: "default/channel")
) { result in
  guard let event = try? result.get().data?.onRealtimeEvent else { return }
  
  let eventType = event.eventType ?? ""
  let payload = event.payload as? [String: Any] ?? [:]
  
  if eventType.contains("validation") {
    updateValidationCount(Int(payload["totalValidationCount"] as? String ?? "0") ?? 0)
  } else if eventType.contains("respect") {
    updateRespectCount(Int(payload["totalRespectCount"] as? String ?? "0") ?? 0)
  }
  
  if let name = payload["reactorName"] as? String {
    showNotification(title: "Reaction", message: "\(name) \(eventType)")
  }
}
```

## Minimal Android Example

```kotlin
appSyncClient.subscribe(
  OnRealtimeEventSubscription.builder().channel("default/channel").build()
).enqueue(object : ApolloSubscriptionCall.Callback<OnRealtimeEventSubscription.Data> {
  override fun onResponse(response: Response<OnRealtimeEventSubscription.Data>) {
    val event = response.data()?.onRealtimeEvent() ?: return
    val eventType = event.eventType() ?: ""
    val payload = event.payload() as? Map<String, Any> ?: return
    
    when {
      eventType.contains("validation") -> {
        updateValidationCount((payload["totalValidationCount"] as? Number)?.toInt() ?: 0)
      }
      eventType.contains("respect") -> {
        updateRespectCount((payload["totalRespectCount"] as? Number)?.toInt() ?: 0)
      }
    }
    
    showNotification("${payload["reactorName"]} $eventType")
  }
})
```

## Event Payloads at a Glance

| Event | Key Fields | Use |
|-------|-----------|-----|
| `post_validated` | reactorName, reactorProfilePicture, **totalValidationCount** | Update validation counter to this value |
| `post_validation_removed` | reactorName, **totalValidationCount** | Update validation counter to this value |
| `post_respected` | reactorName, reactorProfilePicture, **totalRespectCount** | Update respect counter to this value |
| `post_respect_removed` | reactorName, **totalRespectCount** | Update respect counter to this value |

## Best Practices (TL;DR)

✅ **DO:**
- Subscribe to `default/channel` once in your feed/room screen
- Use `totalValidationCount` and `totalRespectCount` directly—don't fetch again
- Unsubscribe when leaving the screen
- Handle null profile pictures with fallback avatar
- Deduplicate events using timestamp if needed

❌ **DON'T:**
- Make separate API calls to fetch post updates
- Treat `totalValidationCount` as an increment—it's the NEW count
- Keep multiple subscriptions open
- Assume profile picture is always present

## Event Flow

```
User taps "Validate"
  ↓
Backend HTTP response (immediate)
  ↓
Backend publishes to AppSync (background, ~100-500ms)
  ↓
Mobile client receives WebSocket event
  ↓
UI updates with new count and notification
```

## Testing

```bash
# Trigger a validation
curl -X POST http://localhost:8080/api/v1/posts/{postID}/validate \
  -H "Authorization: Bearer {token}"

# Watch your mobile logs for "post_validated" event within 500ms
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Events not arriving | Check AppSync logs, verify subscription active, test HTTP endpoint first |
| Duplicate events | Use `publishedAt` timestamp to deduplicate |
| Events delayed | Normal latency is 100-500ms; check network and AppSync metrics |
| Profile picture missing | Fallback to default avatar; URL may be empty string |
| Wrong event channel | All events use `default/channel`—no room-specific channels |

## Configuration

Backend has no additional configuration needed. React to these environment variables:

| Var | Value | Meaning |
|-----|-------|---------|
| APPSYNC_ENABLED | true | Reactions publish to AppSync |
| APPSYNC_ENDPOINT | https://...appsync-api.us-east-1.amazonaws.com/event | AppSync Event API URL |
| APPSYNC_API_KEY | da2-... | API key for publishing |

All configured in backend already. Mobile just needs AppSync SDK setup with same region/endpoint.
