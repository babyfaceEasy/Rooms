# AppSync Reaction Events - Sample Messages

This document shows exactly what messages your mobile app will receive from AppSync. Use these as reference to debug and test your subscription handling.

---

## Table of Contents

1. [Event 1: post_validated](#event-1-post_validated)
2. [Event 2: post_validation_removed](#event-2-post_validation_removed)
3. [Event 3: post_respected](#event-3-post_respected)
4. [Event 4: post_respect_removed](#event-4-post_respect_removed)
5. [Real-World Scenario](#real-world-scenario)
6. [Debugging Checklist](#debugging-checklist)
7. [Field Reference](#field-reference)

---

## Event 1: post_validated

**Scenario:** User taps the "Validate" button on a post  
**Time to receive:** ~100-500ms after the button is tapped

### Complete WebSocket Message

```json
{
  "type": "data",
  "id": "1",
  "payload": {
    "data": {
      "onRealtimeEvent": {
        "__typename": "RealtimeEvent",
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

### What Each Field Means

```
type: "data"
├─ payload.data.onRealtimeEvent
│  ├─ eventType: "post_validated"           ← Event type to switch on
│  ├─ resourceId: "6adc9a1305fb43cdf96d7631" ← Post ID (for routing)
│  ├─ objectId: "6a80ab36948abf3b1594c4c6"   ← Reactor user ID
│  ├─ payload: "{JSON_STRING}"               ← ⚠️ MUST PARSE THIS
│  │  └─ After JSON.parse():
│  │     ├─ postId: "6adc9a1305fb43cdf96d7631"
│  │     ├─ reactorUserId: "6a80ab36948abf3b1594c4c6"
│  │     ├─ reactorName: "Jane Doe"
│  │     ├─ reactorProfilePicture: "https://..."
│  │     └─ totalValidationCount: 5           ← Use this to update UI
│  └─ publishedAt: "2026-10-28T14:23:45.123Z"
```

### How to Process (JavaScript)

```javascript
// Receive message from AppSync subscription
function handleReactionEvent(message) {
  const event = message.payload.data.onRealtimeEvent;
  
  // Parse the payload (it's a JSON string, not an object)
  const payload = JSON.parse(event.payload);
  
  // Now use the fields
  if (event.eventType === 'post_validated') {
    console.log(`Post: ${payload.postId}`);
    console.log(`Reactor: ${payload.reactorName}`);
    console.log(`New validation count: ${payload.totalValidationCount}`);
    
    // Update UI
    updateValidationCounter(payload.postId, payload.totalValidationCount);
    showNotification(`${payload.reactorName} validated your post`);
    addReactorAvatar(payload.postId, payload.reactorName, payload.reactorProfilePicture);
  }
}
```

### What to Search For (Debugging)

When your app doesn't receive events, check:

1. **In backend logs:** `msg="appsync event published" event_type="post_validated"`
2. **In mobile logs:** Look for `"post_validated"` in received messages
3. **Check:** Is `totalValidationCount` present and a number?
4. **Check:** Can you `JSON.parse(event.payload)` without errors?

### Variations

The same event structure appears multiple times, with different data:

```javascript
// Same structure, different post
{
  resourceId: "6adc9a1305fb43cdf96d7999",  // Different post
  payload: "{...\"postId\":\"6adc9a1305fb43cdf96d7999\"...}"
}

// Same post, different reactor
{
  objectId: "6a80ab36948abf3b1594c4d99",   // Different user
  payload: "{...\"reactorUserId\":\"6a80ab36948abf3b1594c4d99\",\"reactorName\":\"John Doe\"...}"
}

// Same post, more validations
{
  payload: "{...\"totalValidationCount\":6}  // Count increased (5 → 6)
}
```

---

## Event 2: post_validation_removed

**Scenario:** User taps "Remove Validation" on a post they validated  
**Time to receive:** ~100-500ms after the button is tapped

### Complete WebSocket Message

```json
{
  "type": "data",
  "id": "2",
  "payload": {
    "data": {
      "onRealtimeEvent": {
        "__typename": "RealtimeEvent",
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

### Key Differences from Event 1

```diff
- eventType: "post_validated"
+ eventType: "post_validation_removed"

- totalValidationCount: 5  (increased)
+ totalValidationCount: 4  (decreased)
```

### How to Process

```javascript
if (event.eventType === 'post_validation_removed') {
  const payload = JSON.parse(event.payload);
  
  // The count decreased
  console.log(`${payload.reactorName} removed validation from post ${payload.postId}`);
  console.log(`Validation count now: ${payload.totalValidationCount}`);
  
  // Update UI
  updateValidationCounter(payload.postId, payload.totalValidationCount);
  removeReactorAvatar(payload.postId, payload.reactorUserId);
}
```

---

## Event 3: post_respected

**Scenario:** User taps the "Respect" button on a post  
**Time to receive:** ~100-500ms after the button is tapped

### Complete WebSocket Message

```json
{
  "type": "data",
  "id": "3",
  "payload": {
    "data": {
      "onRealtimeEvent": {
        "__typename": "RealtimeEvent",
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

### Key Differences from Validation Events

```diff
- eventType: "post_validated"
+ eventType: "post_respected"

- payload.totalValidationCount
+ payload.totalRespectCount   ← DIFFERENT FIELD NAME!

+ objectId: "6a80ab3695abf3b1594c4d22"  ← Different reactor (John Smith not Jane Doe)
```

⚠️ **Important:** The field is `totalRespectCount` NOT `totalValidationCount`

### How to Process

```javascript
if (event.eventType === 'post_respected') {
  const payload = JSON.parse(event.payload);
  
  console.log(`${payload.reactorName} respected post ${payload.postId}`);
  console.log(`Respect count now: ${payload.totalRespectCount}`);  // ← Use this field
  
  // Update UI
  updateRespectCounter(payload.postId, payload.totalRespectCount);  // ← totalRespectCount
  showNotification(`${payload.reactorName} respected your post`);
  addReactorAvatar(payload.postId, payload.reactorName, payload.reactorProfilePicture);
}
```

---

## Event 4: post_respect_removed

**Scenario:** User taps "Remove Respect" on a post they respected  
**Time to receive:** ~100-500ms after the button is tapped

### Complete WebSocket Message

```json
{
  "type": "data",
  "id": "4",
  "payload": {
    "data": {
      "onRealtimeEvent": {
        "__typename": "RealtimeEvent",
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

### Key Differences

```diff
- eventType: "post_respected"
+ eventType: "post_respect_removed"

- totalRespectCount: 8  (increased)
+ totalRespectCount: 7  (decreased)
```

### How to Process

```javascript
if (event.eventType === 'post_respect_removed') {
  const payload = JSON.parse(event.payload);
  
  console.log(`${payload.reactorName} removed respect from post ${payload.postId}`);
  console.log(`Respect count now: ${payload.totalRespectCount}`);
  
  // Update UI
  updateRespectCounter(payload.postId, payload.totalRespectCount);
  removeReactorAvatar(payload.postId, payload.reactorUserId);
}
```

---

## Real-World Scenario

Here's what happens when multiple users react to a post in sequence:

### Initial State
- Post 6adc9a1305fb43cdf96d7631
- Validation count: 3
- Respect count: 5

### Step 1: Jane Doe validates
```javascript
Event:
{
  eventType: "post_validated",
  payload: {
    reactorName: "Jane Doe",
    totalValidationCount: 4  // 3 → 4
  }
}

App updates: validation counter 3 → 4
```

### Step 2: John Smith respects
```javascript
Event:
{
  eventType: "post_respected",
  payload: {
    reactorName: "John Smith",
    totalRespectCount: 6  // 5 → 6
  }
}

App updates: respect counter 5 → 6
```

### Step 3: Jane Doe removes validation
```javascript
Event:
{
  eventType: "post_validation_removed",
  payload: {
    reactorName: "Jane Doe",
    totalValidationCount: 3  // 4 → 3
  }
}

App updates: validation counter 4 → 3
```

### Step 4: John Smith removes respect
```javascript
Event:
{
  eventType: "post_respect_removed",
  payload: {
    reactorName: "John Smith",
    totalRespectCount: 5  // 6 → 5
  }
}

App updates: respect counter 6 → 5
```

### Final State
- Validation count: 3 (back to original)
- Respect count: 5 (back to original)

---

## Debugging Checklist

### Event Not Arriving?

**Step 1: Verify subscription is active**
```javascript
// Check if you're subscribed
const sub = API.graphql(graphqlOperation(`
  subscription {
    onRealtimeEvent(channel: "default/channel") {
      eventType
      payload
    }
  }
`)).subscribe({
  next: (data) => console.log('Received:', data),
  error: (err) => console.error('Subscription error:', err)
});
```

**Step 2: Check backend logs**
```bash
# Look for these messages in CloudWatch:
"msg": "appsync event published"
"msg": "appsync event payload details"

# If you see these, event was published successfully
# Check your mobile for received message
```

**Step 3: Test the HTTP endpoint**
```bash
curl -X POST http://localhost:8080/api/v1/posts/{postID}/validate \
  -H "Authorization: Bearer {token}"

# If 200 OK, HTTP endpoint works
# Event should arrive at mobile within 500ms
```

**Step 4: Check AppSync metrics**
- Go to AWS Console → AppSync → Event API
- Look for "Publish Requests" metric
- Should increase when you trigger reactions

### Event Received but Wrong Format?

**Common mistake: Forgetting to parse payload**
```javascript
// ❌ WRONG
const count = event.payload.totalValidationCount;  // undefined

// ✅ CORRECT
const payload = JSON.parse(event.payload);
const count = payload.totalValidationCount;  // 5
```

**Common mistake: Using wrong field name**
```javascript
// ❌ For respect events, using validation field
const count = payload.totalValidationCount;  // undefined

// ✅ Use correct field
const count = payload.totalRespectCount;  // 8
```

---

## Field Reference

### Top-level Event Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Always `"data"` |
| `payload.data.onRealtimeEvent.eventType` | string | One of: `post_validated`, `post_validation_removed`, `post_respected`, `post_respect_removed` |
| `payload.data.onRealtimeEvent.resourceId` | string | MongoDB ObjectID of the post |
| `payload.data.onRealtimeEvent.objectId` | string | MongoDB ObjectID of the reactor (user) |
| `payload.data.onRealtimeEvent.payload` | string | ⚠️ JSON-stringified - must call `JSON.parse()` |
| `payload.data.onRealtimeEvent.publishedAt` | string | ISO timestamp when event was published |

### Payload Fields (After JSON.parse)

**For validation events (`post_validated`, `post_validation_removed`):**

| Field | Type | Description |
|-------|------|-------------|
| `postId` | string | MongoDB ObjectID of the post |
| `reactorUserId` | string | MongoDB ObjectID of the user who validated |
| `reactorName` | string | Full name of the reactor |
| `reactorProfilePicture` | string | S3 URL to profile picture (may be empty) |
| `totalValidationCount` | number | New total validation count (use this to update UI) |

**For respect events (`post_respected`, `post_respect_removed`):**

| Field | Type | Description |
|-------|------|-------------|
| `postId` | string | MongoDB ObjectID of the post |
| `reactorUserId` | string | MongoDB ObjectID of the user who respected |
| `reactorName` | string | Full name of the reactor |
| `reactorProfilePicture` | string | S3 URL to profile picture (may be empty) |
| `totalRespectCount` | number | New total respect count (use this to update UI) |

---

## Tips for Mobile Implementation

### 1. Always Parse the Payload
```javascript
const event = message.payload.data.onRealtimeEvent;
const payload = JSON.parse(event.payload);  // Don't skip this!
```

### 2. Use the Correct Count Field
```javascript
if (event.eventType.includes('validation')) {
  const count = payload.totalValidationCount;  // For validation events
} else if (event.eventType.includes('respect')) {
  const count = payload.totalRespectCount;     // For respect events
}
```

### 3. Handle Empty Profile Pictures
```javascript
const avatar = payload.reactorProfilePicture || 'default-avatar.png';
// Profile picture may be empty string, so check
```

### 4. Filter Events for Current Post
```javascript
// Only update UI if event is for the post currently on screen
if (payload.postId === currentPostId) {
  updateCounter(payload.totalValidationCount);
}
```

### 5. Use Event Type to Determine Action
```javascript
const actions = {
  'post_validated': () => increment(payload.totalValidationCount),
  'post_validation_removed': () => decrement(payload.totalValidationCount),
  'post_respected': () => increment(payload.totalRespectCount),
  'post_respect_removed': () => decrement(payload.totalRespectCount)
};

actions[event.eventType]?.();
```

---

## Quick Copy-Paste Handler

```javascript
function handleReactionEvent(message) {
  try {
    const event = message.payload.data.onRealtimeEvent;
    const payload = JSON.parse(event.payload);
    
    // Determine which counter to update
    const isValidation = event.eventType.includes('validation');
    const count = isValidation ? payload.totalValidationCount : payload.totalRespectCount;
    
    // Only update if event is for current post
    if (payload.postId !== currentPostId) return;
    
    // Update UI based on event type
    switch(event.eventType) {
      case 'post_validated':
        updateValidationCounter(count);
        showNotification(`${payload.reactorName} validated your post`);
        addAvatar(payload.reactorName, payload.reactorProfilePicture);
        break;
        
      case 'post_validation_removed':
        updateValidationCounter(count);
        removeAvatar(payload.reactorUserId);
        break;
        
      case 'post_respected':
        updateRespectCounter(count);
        showNotification(`${payload.reactorName} respected your post`);
        addAvatar(payload.reactorName, payload.reactorProfilePicture);
        break;
        
      case 'post_respect_removed':
        updateRespectCounter(count);
        removeAvatar(payload.reactorUserId);
        break;
    }
  } catch(error) {
    console.error('Error handling reaction event:', error);
  }
}
```

---

## Need Help?

- Check the full guide: [APPSYNC_REACTION_EVENTS.md](./APPSYNC_REACTION_EVENTS.md)
- Quick reference: [APPSYNC_QUICK_REFERENCE_REACTIONS.md](./APPSYNC_QUICK_REFERENCE_REACTIONS.md)
- Implementation details: [APPSYNC_REACTION_IMPLEMENTATION_SUMMARY.md](../APPSYNC_REACTION_IMPLEMENTATION_SUMMARY.md)
