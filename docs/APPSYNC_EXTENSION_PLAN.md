# AppSync Real-time Notifications - Extension Plan

## Current State Analysis

### ✅ What's Already Working
1. **New Post Events** (`new_post`)
   - Published in: `PostHandler.CreatePost()` → `appSyncPublisher.PublishNewPost()`
   - Triggered when: User creates a post in a room
   - Channel: `default/channel`
   - Payload: Full post object (title, content, author, reactions counts, etc.)

2. **New Comment Events** (`new_comment`)
   - Published in: `CommentHandler.CreateComment()` → `appSyncPublisher.PublishCommentCreated()`
   - Triggered when: User adds a comment to a post
   - Channel: `default/channel`
   - Payload: Full comment object (text, author, timestamps)

3. **Infrastructure**
   - Generic `publish()` method that handles all event types
   - Proper error handling and logging
   - Non-blocking design (API continues even if AppSync fails)
   - SSE (Server-Sent Events) already integrated as backup

---

## What Needs to Be Added

### Feature 1: Post Reactions (Validation & Respect)

**Current Status:** No real-time notifications for reactions
- `PostHandler.ValidatePost()` - ❌ No AppSync call
- `PostHandler.RemoveValidation()` - ❌ No AppSync call
- `PostHandler.RespectPost()` - ❌ No AppSync call
- `PostHandler.RemoveRespect()` - ❌ No AppSync call

**What to Add:**
- 4 new event types: `post_validated`, `post_validation_removed`, `post_respected`, `post_respect_removed`
- 4 new publisher methods in `AppSyncRealtimePublisher`
- 4 new calls in `PostHandler` after reactions succeed

---

## Proposed Event Types & Structure

### Event 1: Post Validated (`post_validated`)
```
eventType:   "post_validated"
resourceId:  "post_id"
objectId:    "validation_user_id"
payload:     {
  post_id: "...",
  validated_by_user_id: "...",
  validated_by_user_name: "...",
  total_validations_count: 5
}
publishedAt: timestamp
```

### Event 2: Post Validation Removed (`post_validation_removed`)
```
eventType:   "post_validation_removed"
resourceId:  "post_id"
objectId:    "user_id"
payload:     {
  post_id: "...",
  removed_by_user_id: "...",
  removed_by_user_name: "...",
  total_validations_count: 4
}
publishedAt: timestamp
```

### Event 3: Post Respected (`post_respected`)
```
eventType:   "post_respected"
resourceId:  "post_id"
objectId:    "respect_user_id"
payload:     {
  post_id: "...",
  respected_by_user_id: "...",
  respected_by_user_name: "...",
  total_respects_count: 3
}
publishedAt: timestamp
```

### Event 4: Post Respect Removed (`post_respect_removed`)
```
eventType:   "post_respect_removed"
resourceId:  "post_id"
objectId:    "user_id"
payload:     {
  post_id: "...",
  removed_by_user_id: "...",
  removed_by_user_name: "...",
  total_respects_count: 2
}
publishedAt: timestamp
```

---

## Implementation Plan (Before Writing Code)

### Phase 1: Extend AppSyncRealtimePublisher
1. Add 4 new public methods:
   - `PublishPostValidated(postID, userID, userData)`
   - `PublishPostValidationRemoved(postID, userID, userData)`
   - `PublishPostRespected(postID, userID, userData)`
   - `PublishPostRespectRemoved(postID, userID, userData)`

2. These will call the generic `publish()` with appropriate eventType, resourceId, objectId

### Phase 2: Integrate into PostHandler
1. After successful `ValidatePost()` call:
   - Get user details (name) for payload
   - Get updated post (to get total validation count)
   - Call `appSyncPublisher.PublishPostValidated()`

2. After successful `RemoveValidation()` call:
   - Similar pattern with `PublishPostValidationRemoved()`

3. Repeat for RespectPost and RemoveRespect methods

### Phase 3: Documentation Update
1. Update APPSYNC_MOBILE_INTEGRATION.md with new event types
2. Update APPSYNC_QUICK_REFERENCE.md with examples
3. Document payload structure for each event

---

## Questions for You

Before I write code, I need clarification on:

### Q1: Event Granularity
Should the reaction events include:
- ✅ Just the reaction user info (lightweight)?
- 📋 OR the full updated post with all reactions (heavy)?
- 📋 OR just the count change (minimal)?

**Impact:** Affects payload size and mobile app logic

### Q2: Channel Strategy
Current: All events go to `default/channel`

Should we:
- ✅ Keep using `default/channel` for all (current approach)?
- 📋 OR use room-specific channels (e.g., `room/{roomID}/channel`)?
- 📋 OR use post-specific channels (e.g., `post/{postID}/channel`)?

**Impact:** Affects how mobile clients subscribe; room-specific would reduce noise

### Q3: Payload User Data
When a user validates/respects a post, do you want:
- ✅ Just the user ID who made the reaction (lightweight)?
- 📋 OR user ID + name + profile picture (heavier, better UX)?

**Impact:** Affects notification display on mobile

### Q4: Real-time vs. Batch
For updates to reactions, should we:
- ✅ Send a separate event each time (what I proposed)?
- 📋 OR batch them and send periodic updates?

**Impact:** Mobile app receives more events but has real-time data

### Q5: Comment Reactions
Do comments need reactions like posts?
- ❌ No, just comment creation is enough
- ✅ Yes, add like/respect to comments too

**Impact:** Would add 8 more event types

---

## Architecture Notes

### What Stays the Same
- Generic `publish()` method signature
- Event payload structure (eventType, resourceId, objectId, payload, publishedAt)
- HTTP POST to AppSync Event API
- Error handling and logging patterns

### What Changes
- AppSyncRealtimePublisher gains new public methods
- PostHandler gains new AppSync calls in 4 methods
- Documentation expands with 4+ new event types

### Data Flow
```
User Action (React to Post)
    ↓
PostHandler.RespectPost() / ValidatePost()
    ↓
postService.RespectPost() / ValidatePost()
    ↓
Repository update (AddRespect / AddValidation)
    ↓
Handler fetches updated data (user info, new counts)
    ↓
appSyncPublisher.PublishPostRespected() [NEW]
    ↓
AppSync Event API receives event
    ↓
Mobile clients receive via WebSocket (if subscribed)
    ↓
Mobile shows notification/updates UI
```

---

## Risk Assessment

### Low Risk
- Adding new methods to AppSyncRealtimePublisher (isolated)
- Non-blocking design (failures don't affect API response)

### Medium Risk
- Need to fetch user data for payload (small performance impact)
- Need to fetch updated post counts (minor DB query)

### Mitigation
- Keep all AppSync calls in goroutines (async)
- Add timeouts to prevent blocking
- Log failures without failing the API request

---

## Summary

To add real-time notifications for comments and reactions, we need to:

1. ✅ Add 4 new event types (post_validated, post_validation_removed, post_respected, post_respect_removed)
2. ✅ Add 4 new publisher methods to AppSyncRealtimePublisher
3. ✅ Add 4 publisher calls to PostHandler reaction methods
4. ✅ Update documentation with new events

**Before writing code, please answer the 5 questions above to confirm:**
- Event payload granularity (Q1)
- Channel strategy (Q2)
- User data detail level (Q3)
- Real-time vs batch (Q4)
- Comment reactions support (Q5)

