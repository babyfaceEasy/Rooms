# AppSync Real-Time Reaction Events

## Overview

Post reactions (validations and respects) now trigger real-time AppSync events, allowing mobile clients to receive instant notifications when users interact with posts.

This guide explains how to subscribe to and handle reaction events on mobile platforms.

---

## Event Types

### 1. `post_validated`
Triggered when a user validates a post.

**Payload:**
```json
{
  "eventType": "post_validated",
  "resourceId": "{postID}",
  "objectId": "{reactorUserID}",
  "payload": {
    "postId": "6adc9a1305fb43cdf96d7631",
    "reactorUserId": "6a80ab36948abf3b1594c4c6",
    "reactorName": "Jane Doe",
    "reactorProfilePicture": "https://bouk-room-dev.s3.amazonaws.com/jane-doe-profile.jpg",
    "totalValidationCount": 5
  },
  "publishedAt": "2026-10-28T09:31:23.057+01:00"
}
```

**Use Case:**
- Show a notification: "Jane Doe validated your post"
- Update the validation count in real-time
- Add the reactor's avatar to the validation counter UI

---

### 2. `post_validation_removed`
Triggered when a user removes their validation from a post.

**Payload:**
```json
{
  "eventType": "post_validation_removed",
  "resourceId": "{postID}",
  "objectId": "{reactorUserID}",
  "payload": {
    "postId": "6adc9a1305fb43cdf96d7631",
    "reactorUserId": "6a80ab36948abf3b1594c4c6",
    "reactorName": "Jane Doe",
    "reactorProfilePicture": "https://bouk-room-dev.s3.amazonaws.com/jane-doe-profile.jpg",
    "totalValidationCount": 4
  },
  "publishedAt": "2026-10-28T09:31:24.123+01:00"
}
```

**Use Case:**
- Update the validation count in real-time
- Remove the reactor's avatar from the validation counter UI
- Optionally show a notification: "Jane Doe removed their validation"

---

### 3. `post_respected`
Triggered when a user respects a post.

**Payload:**
```json
{
  "eventType": "post_respected",
  "resourceId": "{postID}",
  "objectId": "{reactorUserID}",
  "payload": {
    "postId": "6adc9a1305fb43cdf96d7631",
    "reactorUserId": "6a80ab36948abf3b1594c4c6",
    "reactorName": "John Smith",
    "reactorProfilePicture": "https://bouk-room-dev.s3.amazonaws.com/john-smith-profile.jpg",
    "totalRespectCount": 8
  },
  "publishedAt": "2026-10-28T09:31:25.456+01:00"
}
```

**Use Case:**
- Show a notification: "John Smith respected your post"
- Update the respect count in real-time
- Add the reactor's avatar to the respect counter UI

---

### 4. `post_respect_removed`
Triggered when a user removes their respect from a post.

**Payload:**
```json
{
  "eventType": "post_respect_removed",
  "resourceId": "{postID}",
  "objectId": "{reactorUserID}",
  "payload": {
    "postId": "6adc9a1305fb43cdf96d7631",
    "reactorUserId": "6a80ab36948abf3b1594c4c6",
    "reactorName": "John Smith",
    "reactorProfilePicture": "https://bouk-room-dev.s3.amazonaws.com/john-smith-profile.jpg",
    "totalRespectCount": 7
  },
  "publishedAt": "2026-10-28T09:31:26.789+01:00"
}
```

**Use Case:**
- Update the respect count in real-time
- Remove the reactor's avatar from the respect counter UI
- Optionally show a notification: "John Smith removed their respect"

---

## Sample Events - What to Expect

Here are complete, real-world sample events that your mobile app will receive. Use these to test your subscription and know exactly what to search for in logs.

### Sample 1: User Validates a Post

**When to expect:** User taps the "Validate" button on a post

**WebSocket message your app receives:**
```json
{
  "type": "data",
  "id": "1",
  "payload": {
    "data": {
      "onRealtimeEvent": {
        "eventType": "post_validated",
        "resourceId": "6adc9a1305fb43cdf96d7631",
        "objectId": "6a80ab36948abf3b1594c4c6",
        "payload": "{\"postId\":\"6adc9a1305fb43cdf96d7631\",\"reactorUserId\":\"6a80ab36948abf3b1594c4c6\",\"reactorName\":\"Jane Doe\",\"reactorProfilePicture\":\"https://bouk-room-dev.s3.amazonaws.com/profiles/6a80ab36948abf3b1594c4c6.jpg\",\"totalValidationCount\":5}",
        "publishedAt": "2026-10-28T14:23:45.123Z"
      }
    }
  }
}
```

**What to search for in logs/debugging:**
- `eventType` = `"post_validated"`
- `"totalValidationCount": 5` (the new count)
- `"reactorName": "Jane Doe"` (who validated)
- Post ID in `resourceId`

**Code to handle:**
```javascript
if (event.eventType === 'post_validated') {
  const payload = JSON.parse(event.payload);
  console.log(`${payload.reactorName} validated post ${payload.postId}`);
  updateValidationCount(payload.postId, payload.totalValidationCount);
  showNotification(`${payload.reactorName} validated your post`);
}
```

---

### Sample 2: User Removes Validation

**When to expect:** User taps the "Remove Validation" button on a post they already validated

**WebSocket message your app receives:**
```json
{
  "type": "data",
  "id": "2",
  "payload": {
    "data": {
      "onRealtimeEvent": {
        "eventType": "post_validation_removed",
        "resourceId": "6adc9a1305fb43cdf96d7631",
        "objectId": "6a80ab36948abf3b1594c4c6",
        "payload": "{\"postId\":\"6adc9a1305fb43cdf96d7631\",\"reactorUserId\":\"6a80ab36948abf3b1594c4c6\",\"reactorName\":\"Jane Doe\",\"reactorProfilePicture\":\"https://bouk-room-dev.s3.amazonaws.com/profiles/6a80ab36948abf3b1594c4c6.jpg\",\"totalValidationCount\":4}",
        "publishedAt": "2026-10-28T14:24:12.456Z"
      }
    }
  }
}
```

**What to search for in logs/debugging:**
- `eventType` = `"post_validation_removed"`
- `"totalValidationCount": 4` (decreased from previous)
- `"reactorName": "Jane Doe"` (who removed validation)
- Notice: Same post ID, count went DOWN

**Code to handle:**
```javascript
if (event.eventType === 'post_validation_removed') {
  const payload = JSON.parse(event.payload);
  console.log(`${payload.reactorName} removed validation from post ${payload.postId}`);
  updateValidationCount(payload.postId, payload.totalValidationCount);
  removeReactorAvatar(payload.postId, payload.reactorUserId);
}
```

---

### Sample 3: User Respects a Post

**When to expect:** User taps the "Respect" button on a post

**WebSocket message your app receives:**
```json
{
  "type": "data",
  "id": "3",
  "payload": {
    "data": {
      "onRealtimeEvent": {
        "eventType": "post_respected",
        "resourceId": "6adc9a1305fb43cdf96d7631",
        "objectId": "6a80ab3695abf3b1594c4d22",
        "payload": "{\"postId\":\"6adc9a1305fb43cdf96d7631\",\"reactorUserId\":\"6a80ab3695abf3b1594c4d22\",\"reactorName\":\"John Smith\",\"reactorProfilePicture\":\"https://bouk-room-dev.s3.amazonaws.com/profiles/6a80ab3695abf3b1594c4d22.jpg\",\"totalRespectCount\":8}",
        "publishedAt": "2026-10-28T14:25:33.789Z"
      }
    }
  }
}
```

**What to search for in logs/debugging:**
- `eventType` = `"post_respected"`
- `"totalRespectCount": 8` (the new count)
- `"reactorName": "John Smith"` (who respected)
- Notice: Field is `totalRespectCount` NOT `totalValidationCount`

**Code to handle:**
```javascript
if (event.eventType === 'post_respected') {
  const payload = JSON.parse(event.payload);
  console.log(`${payload.reactorName} respected post ${payload.postId}`);
  updateRespectCount(payload.postId, payload.totalRespectCount);
  showNotification(`${payload.reactorName} respected your post`);
  addReactorAvatar(payload.postId, payload.reactorName, payload.reactorProfilePicture);
}
```

---

### Sample 4: User Removes Respect

**When to expect:** User taps the "Remove Respect" button on a post they already respected

**WebSocket message your app receives:**
```json
{
  "type": "data",
  "id": "4",
  "payload": {
    "data": {
      "onRealtimeEvent": {
        "eventType": "post_respect_removed",
        "resourceId": "6adc9a1305fb43cdf96d7631",
        "objectId": "6a80ab3695abf3b1594c4d22",
        "payload": "{\"postId\":\"6adc9a1305fb43cdf96d7631\",\"reactorUserId\":\"6a80ab3695abf3b1594c4d22\",\"reactorName\":\"John Smith\",\"reactorProfilePicture\":\"https://bouk-room-dev.s3.amazonaws.com/profiles/6a80ab3695abf3b1594c4d22.jpg\",\"totalRespectCount\":7}",
        "publishedAt": "2026-10-28T14:26:01.234Z"
      }
    }
  }
}
```

**What to search for in logs/debugging:**
- `eventType` = `"post_respect_removed"`
- `"totalRespectCount": 7` (decreased from previous)
- `"reactorName": "John Smith"` (who removed respect)
- Notice: Count went DOWN

**Code to handle:**
```javascript
if (event.eventType === 'post_respect_removed') {
  const payload = JSON.parse(event.payload);
  console.log(`${payload.reactorName} removed respect from post ${payload.postId}`);
  updateRespectCount(payload.postId, payload.totalRespectCount);
  removeReactorAvatar(payload.postId, payload.reactorUserId);
}
```

---

## Real-World Scenario

Here's what happens in a complete user flow:

**Initial state:**
- Post has 3 validations, 5 respects

**User 1 validates the post:**
```json
{ "eventType": "post_validated", "payload": { ..., "totalValidationCount": 4 } }
```
→ Other clients update: 3 → 4 validations

**User 2 respects the post:**
```json
{ "eventType": "post_respected", "payload": { ..., "totalRespectCount": 6 } }
```
→ Other clients update: 5 → 6 respects

**User 1 removes validation:**
```json
{ "eventType": "post_validation_removed", "payload": { ..., "totalValidationCount": 3 } }
```
→ Other clients update: 4 → 3 validations

**User 2 removes respect:**
```json
{ "eventType": "post_respect_removed", "payload": { ..., "totalRespectCount": 5 } }
```
→ Other clients update: 6 → 5 respects

**Final state:**
- Post has 3 validations, 5 respects (back to original)

---

## Debugging: What to Look For

### Missing Events?

**Check backend logs (debug level):**
```bash
curl -X POST http://localhost:8080/api/v1/posts/{postID}/validate \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json"

# Look for these logs:
# ✓ "appsync event payload details"
# ✓ "appsync event request sending"
# ✓ "appsync event published"
```

### Verify Event Structure

**All events have this structure:**
```json
{
  "type": "data",
  "payload": {
    "data": {
      "onRealtimeEvent": {
        "eventType": "string",  // post_validated, post_validation_removed, etc.
        "resourceId": "string", // Post ID
        "objectId": "string",   // Reactor User ID
        "payload": "string",    // JSON stringified - MUST PARSE THIS
        "publishedAt": "string" // ISO timestamp
      }
    }
  }
}
```

**Key: The `payload` field is a JSON STRING, not an object. You must parse it:**
```javascript
const event = received_data.payload.data.onRealtimeEvent;
const payload = JSON.parse(event.payload); // ← MUST PARSE

// Now access fields:
console.log(payload.reactorName);
console.log(payload.totalValidationCount);
```

### Test Commands

**Test validation endpoint:**
```bash
POST http://localhost:8080/api/v1/posts/6adc9a1305fb43cdf96d7631/validate
Authorization: Bearer {your_token}

# Response should be 200 OK with updated post data
# Event should arrive in your mobile app within 500ms
```

**Test respect endpoint:**
```bash
POST http://localhost:8080/api/v1/posts/6adc9a1305fb43cdf96d7631/respect
Authorization: Bearer {your_token}

# Response should be 200 OK with updated post data
# Event should arrive in your mobile app within 500ms
```

---

## Channel Configuration

All events are published to the **default channel**:
```
default/channel
```

This means:
- **One subscription covers all events** from your entire application
- When a post is validated anywhere in any room, all connected clients receive the event
- Mobile clients should subscribe once and handle all reaction event types

---

## Mobile Implementation

### React Native / JavaScript

Subscribe to all reaction events:

```javascript
import { API, graphqlOperation } from 'aws-amplify';

// Subscribe to reaction events (post_validated, post_validation_removed, post_respected, post_respect_removed)
const reactionSubscription = API.graphql(
  graphqlOperation(`
    subscription OnReactionEvent {
      onRealtimeEvent(channel: "default/channel") {
        eventType
        resourceId
        objectId
        payload
        publishedAt
      }
    }
  `)
).subscribe({
  next: (data) => {
    const event = data.value.data.onRealtimeEvent;
    handleReactionEvent(event);
  },
  error: (err) => {
    console.error('Subscription error:', err);
  },
  complete: () => {
    console.log('Subscription closed');
  }
});

// Handle different reaction event types
function handleReactionEvent(event) {
  switch (event.eventType) {
    case 'post_validated':
      handlePostValidated(event);
      break;
    case 'post_validation_removed':
      handlePostValidationRemoved(event);
      break;
    case 'post_respected':
      handlePostRespected(event);
      break;
    case 'post_respect_removed':
      handlePostRespectRemoved(event);
      break;
  }
}

function handlePostValidated(event) {
  const { payload } = event;
  console.log(`${payload.reactorName} validated post ${payload.postId}`);
  console.log(`Total validations: ${payload.totalValidationCount}`);
  
  // Update UI: increment validation count
  updatePostValidationCount(payload.postId, payload.totalValidationCount);
  
  // Show notification
  showNotification({
    title: 'Post Validated',
    message: `${payload.reactorName} validated your post`,
    image: payload.reactorProfilePicture
  });
}

function handlePostRespected(event) {
  const { payload } = event;
  console.log(`${payload.reactorName} respected post ${payload.postId}`);
  console.log(`Total respects: ${payload.totalRespectCount}`);
  
  // Update UI: increment respect count
  updatePostRespectCount(payload.postId, payload.totalRespectCount);
  
  // Show notification
  showNotification({
    title: 'Post Respected',
    message: `${payload.reactorName} respected your post`,
    image: payload.reactorProfilePicture
  });
}

// Cleanup on unmount
reactionSubscription.unsubscribe();
```

### Swift (iOS)

Use AWS AppSync SDK or Amplify for Swift:

```swift
import AWSAppSync
import Combine

class ReactionEventListener {
    var subscription: AnyCancellable?
    
    func subscribeToReactions() {
        let subscription = appSyncClient.subscribe(
            subscription: OnRealtimeEventSubscription(channel: "default/channel")
        ) { result in
            switch result {
            case .success(let graphQLResult):
                if let event = graphQLResult.data?.onRealtimeEvent {
                    self.handleReactionEvent(event)
                }
            case .failure(let error):
                print("Subscription error: \(error)")
            }
        }
    }
    
    private func handleReactionEvent(_ event: OnRealtimeEventSubscription.Data.OnRealtimeEvent) {
        let eventType = event.eventType ?? ""
        let payload = event.payload as? [String: Any] ?? [:]
        
        switch eventType {
        case "post_validated":
            handlePostValidated(payload: payload)
        case "post_validation_removed":
            handlePostValidationRemoved(payload: payload)
        case "post_respected":
            handlePostRespected(payload: payload)
        case "post_respect_removed":
            handlePostRespectRemoved(payload: payload)
        default:
            break
        }
    }
    
    private func handlePostValidated(payload: [String: Any]) {
        guard let postId = payload["postId"] as? String,
              let reactorName = payload["reactorName"] as? String,
              let totalCount = payload["totalValidationCount"] as? Int else {
            return
        }
        
        print("\(reactorName) validated post \(postId)")
        print("Total validations: \(totalCount)")
        
        // Update UI on main thread
        DispatchQueue.main.async {
            self.updatePostValidationCount(postId: postId, count: totalCount)
            self.showNotification(
                title: "Post Validated",
                message: "\(reactorName) validated your post"
            )
        }
    }
    
    private func handlePostRespected(payload: [String: Any]) {
        guard let postId = payload["postId"] as? String,
              let reactorName = payload["reactorName"] as? String,
              let totalCount = payload["totalRespectCount"] as? Int else {
            return
        }
        
        print("\(reactorName) respected post \(postId)")
        print("Total respects: \(totalCount)")
        
        // Update UI on main thread
        DispatchQueue.main.async {
            self.updatePostRespectCount(postId: postId, count: totalCount)
            self.showNotification(
                title: "Post Respected",
                message: "\(reactorName) respected your post"
            )
        }
    }
}
```

### Kotlin (Android)

Using AWS AppSync SDK for Android:

```kotlin
import com.amazonaws.mobileconnectors.appsync.AWSAppSyncClient
import com.apollographql.apollo.api.Response

class ReactionEventListener(private val appSyncClient: AWSAppSyncClient) {
    
    fun subscribeToReactions() {
        appSyncClient.subscribe(
            OnRealtimeEventSubscription.builder()
                .channel("default/channel")
                .build()
        )
        .enqueue(object : ApolloSubscriptionCall.Callback<OnRealtimeEventSubscription.Data> {
            override fun onResponse(response: Response<OnRealtimeEventSubscription.Data>) {
                response.data()?.onRealtimeEvent()?.let { event ->
                    handleReactionEvent(event)
                }
            }

            override fun onFailure(e: ApolloException) {
                Log.e("ReactionListener", "Subscription failed", e)
            }

            override fun onCompleted() {
                Log.d("ReactionListener", "Subscription completed")
            }
        })
    }

    private fun handleReactionEvent(event: OnRealtimeEventSubscription.OnRealtimeEvent) {
        when (event.eventType()) {
            "post_validated" -> handlePostValidated(event)
            "post_validation_removed" -> handlePostValidationRemoved(event)
            "post_respected" -> handlePostRespected(event)
            "post_respect_removed" -> handlePostRespectRemoved(event)
        }
    }

    private fun handlePostValidated(event: OnRealtimeEventSubscription.OnRealtimeEvent) {
        val payload = event.payload() as? Map<String, Any> ?: return
        val postId = payload["postId"] as? String ?: return
        val reactorName = payload["reactorName"] as? String ?: return
        val totalCount = (payload["totalValidationCount"] as? Number)?.toInt() ?: 0

        Log.d("ReactionListener", "$reactorName validated post $postId")
        Log.d("ReactionListener", "Total validations: $totalCount")

        // Update UI on main thread
        Handler(Looper.getMainLooper()).post {
            updatePostValidationCount(postId, totalCount)
            showNotification(
                title = "Post Validated",
                message = "$reactorName validated your post"
            )
        }
    }

    private fun handlePostRespected(event: OnRealtimeEventSubscription.OnRealtimeEvent) {
        val payload = event.payload() as? Map<String, Any> ?: return
        val postId = payload["postId"] as? String ?: return
        val reactorName = payload["reactorName"] as? String ?: return
        val totalCount = (payload["totalRespectCount"] as? Number)?.toInt() ?: 0

        Log.d("ReactionListener", "$reactorName respected post $postId")
        Log.d("ReactionListener", "Total respects: $totalCount")

        // Update UI on main thread
        Handler(Looper.getMainLooper()).post {
            updatePostRespectCount(postId, totalCount)
            showNotification(
                title = "Post Respected",
                message = "$reactorName respected your post"
            )
        }
    }
}
```

---

## Data Flow

### Validation Flow

```
User taps "Validate" button
    ↓
POST /api/v1/posts/{id}/validate (HTTP)
    ↓
Backend updates post.Validations array
    ↓
Backend publishes "post_validated" to AppSync
    ↓
AppSync routes event to all subscribers on "default/channel"
    ↓
All mobile clients receive the event
    ↓
UI updates: Show notification, increment counter, add reactor avatar
```

### Respect Flow

```
User taps "Respect" button
    ↓
POST /api/v1/posts/{id}/respect (HTTP)
    ↓
Backend updates post.Respects array
    ↓
Backend publishes "post_respected" to AppSync
    ↓
AppSync routes event to all subscribers on "default/channel"
    ↓
All mobile clients receive the event
    ↓
UI updates: Show notification, increment counter, add reactor avatar
```

---

## Key Properties for Mobile Implementation

### Validation Event
- **postId** — MongoDB ObjectID of the post
- **reactorUserId** — MongoDB ObjectID of the user who validated
- **reactorName** — Full name of the validator (for display)
- **reactorProfilePicture** — S3 URL to profile picture (can be empty)
- **totalValidationCount** — New total count after this reaction

### Respect Event
- **postId** — MongoDB ObjectID of the post
- **reactorUserId** — MongoDB ObjectID of the user who respected
- **reactorName** — Full name of the respecter (for display)
- **reactorProfilePicture** — S3 URL to profile picture (can be empty)
- **totalRespectCount** — New total count after this reaction

---

## Best Practices

### 1. Handle Null/Empty Values
Profile pictures may be empty strings or null. Always check before using:

```javascript
const profilePic = payload.reactorProfilePicture || 'default-avatar.png';
```

### 2. Idempotent Updates
The same event might arrive twice (at-least-once delivery). Use the `publishedAt` timestamp to deduplicate:

```javascript
const eventKey = `${event.eventType}-${payload.postId}-${payload.reactorUserId}-${event.publishedAt}`;
if (processedEvents.has(eventKey)) {
  return; // Already processed
}
processedEvents.add(eventKey);
```

### 3. Update Counts Immediately
Don't make an additional API call to fetch the updated post. Trust the `totalValidationCount` and `totalRespectCount` from the event:

```javascript
// ✅ DO: Use the count from the event
updateValidationCount(payload.postId, payload.totalValidationCount);

// ❌ DON'T: Make another API call
// const updatedPost = await getPost(payload.postId);
// updateValidationCount(updatedPost.validations.length);
```

### 4. Manage Connection Lifecycle
Subscribe when entering a room/feed, unsubscribe when leaving:

```javascript
// Subscribe
useEffect(() => {
  const sub = subscribeToReactions();
  return () => sub.unsubscribe();
}, []);
```

### 5. Batch Updates for High-Traffic Posts
If a post receives many reactions quickly, batch updates to avoid excessive re-renders:

```javascript
let updateQueue = [];
let updateTimeout;

function handleReactionEvent(event) {
  updateQueue.push(event);
  
  clearTimeout(updateTimeout);
  updateTimeout = setTimeout(() => {
    batchUpdateUI(updateQueue);
    updateQueue = [];
  }, 100); // Wait 100ms to batch updates
}
```

### 6. Handle Validation Count vs Respect Count
Use the correct field name:
- Validations → `totalValidationCount`
- Respects → `totalRespectCount`

```javascript
const isValidation = event.eventType.includes('validation');
const newCount = isValidation 
  ? event.payload.totalValidationCount 
  : event.payload.totalRespectCount;
```

---

## Troubleshooting

### Events Not Arriving

1. **Verify subscription is active:**
   - Check logs for subscription errors
   - Ensure user is authenticated before subscribing

2. **Verify AppSync configuration:**
   - Confirm `APPSYNC_ENABLED=true` in backend env
   - Confirm `APPSYNC_API_KEY` is set correctly
   - Check AppSync dashboard for publish/subscribe metrics

3. **Check network connectivity:**
   - WebSocket connection to AppSync should be established
   - Look for connection errors in logs

### Duplicate Events

- Use `publishedAt` timestamp for deduplication
- Track processed events in a Set with TTL cleanup

### High Latency

- Events should arrive within 100-500ms
- If latency is high, check network latency and AppSync dashboard metrics
- Consider batching multiple reactions in high-traffic scenarios

### Profile Picture Not Loading

- Profile picture URL may be empty or invalid
- Always provide a fallback/default avatar
- URL is hosted on S3; ensure CORS headers are correct

---

## Testing with Postman

Test the validation endpoint:

```bash
POST /api/v1/posts/{postID}/validate
Authorization: Bearer {accessToken}
```

Expected response:
```json
{
  "status": 200,
  "message": "post validated successfully",
  "data": {
    "id": "6adc9a1305fb43cdf96d7631",
    "validations_count": 5,
    ...
  }
}
```

Then check your mobile client logs or dashboard to see the `post_validated` event arrive within a few seconds.

---

## Summary

| Event | Trigger | Use Case |
|-------|---------|----------|
| `post_validated` | User validates | Show notification, update count |
| `post_validation_removed` | User removes validation | Update count, remove avatar |
| `post_respected` | User respects | Show notification, update count |
| `post_respect_removed` | User removes respect | Update count, remove avatar |

All events use the **default/channel** for routing, so a single AppSync subscription covers all reactions from all rooms and posts.

The payload includes reactor name and profile picture for rich notifications, and the total count for real-time UI updates.
