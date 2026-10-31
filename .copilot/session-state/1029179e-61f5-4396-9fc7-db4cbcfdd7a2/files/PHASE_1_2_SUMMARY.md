# FCM Push Notifications - Phase 1 & 2 Implementation Summary

## ✅ Phase 1: Dependencies & Configuration

### Dependencies Added
- **Firebase Admin SDK**: `firebase.google.com/go/v4 v4.22.0`
  - Provides FCM send capabilities and authentication
  - Automatically pulls in necessary Google Cloud dependencies

### Configuration Enhancements

**Updated** `config/config.go`:

```go
// Firebase Configuration
Firebase struct {
    ServiceAccountKeyPath string  // Path to Firebase service account JSON
    Enabled               bool    // Enable/disable Firebase integration
}

// Notification Configuration
Notifications struct {
    ActivityCooldownMinutes int  // Default: 30 minutes
    DeviceTokenExpiryDays  int   // Default: 90 days
}
```

**New environment variables** (added to `.env.example`):
- `FIREBASE_ENABLED` - Toggle Firebase integration (default: true)
- `FIREBASE_SERVICE_ACCOUNT_KEY_PATH` - Path to service account key JSON
- `NOTIFICATION_ACTIVITY_COOLDOWN_MINUTES` - Activity cooldown window (default: 30)
- `DEVICE_TOKEN_EXPIRY_DAYS` - Device token expiry threshold (default: 90)

---

## ✅ Phase 2: Domain Models

### 1. Extended `Notification` Entity
**File**: `internal/domain/notification.go`

**New enums**:
- `NotificationType` - Added, Joined, Activity
- `NotificationStatus` - Pending, Sent, Failed

**New fields**:
- `DeviceTokenID` - Reference to the device token used
- `PatchID` - Reference to the patch (for patch-related notifications)
- `NotificationType` - Enum specifying the type
- `Status` - Delivery status tracking
- `SentAt` - Timestamp when notification was sent
- `ErrorMessage` - Captures any FCM errors
- `FCMMessageID` - Firebase's message ID for tracking

**Backward compatible**: Existing fields (`PostID`, `Reason`) retained for legacy report notifications.

### 2. New `DeviceToken` Entity
**File**: `internal/domain/device_token.go`

Represents an FCM device token for a user:
```go
type DeviceToken struct {
    ID        ObjectID              // Unique identifier
    UserID    ObjectID              // User who owns this device
    Token     string                // FCM registration token
    Platform  DeviceTokenPlatform   // ios, android, or web
    IsActive  bool                  // Can be deactivated
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Platforms supported**: iOS, Android, Web

### 3. New `ActivityCooldown` Entity
**File**: `internal/domain/activity_cooldown.go`

Tracks 30-minute activity windows per (user, patch) pair:
```go
type ActivityCooldown struct {
    ID                     ObjectID   // Unique identifier
    UserID                 ObjectID   // User receiving notifications
    PatchID                ObjectID   // Patch with activity
    LastNotificationSentAt *time.Time // When we last sent notification
    LastActivityAt         *time.Time // When activity last occurred
    LastCheckedAt          *time.Time // When user last viewed patch
    CreatedAt              time.Time
    UpdatedAt              time.Time
}
```

---

## 📋 What's Ready

✅ Dependencies installed and locked  
✅ Configuration structure supports Firebase auth  
✅ Domain models with proper BSON tags for MongoDB  
✅ All code compiles without errors  
✅ Type-safe enum support for platforms, statuses, and types  

---

## 🚀 Next Steps (Phase 3)

Repository layer implementation:
- `DeviceTokenRepository` - CRUD operations for device tokens
- Extended `NotificationRepository` - Status and error tracking
- `ActivityCooldownRepository` - Track activity windows per (user, patch)

---

## 📝 Notes

- Firebase service account key should be added via environment variable (never committed to git)
- DeviceTokens are automatically marked inactive after 90 days (configurable)
- Activity cooldown logic will be implemented in Phase 4 (Service layer)
