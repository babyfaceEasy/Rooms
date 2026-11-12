#!/bin/bash

set -e

API_URL="http://localhost:8080"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="password123"

echo "════════════════════════════════════════════════════"
echo "     Admin Endpoints Testing Script"
echo "════════════════════════════════════════════════════"
echo ""

# Check if API is running
echo "🔍 Checking if API is running..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ API is not running!"
    echo "Start it with: go run cmd/api/main.go"
    exit 1
fi
echo "✅ API is running"
echo ""

# Step 1: Login
echo "1️⃣  Testing Admin Login..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/admin/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$ADMIN_EMAIL\",
    \"password\": \"$ADMIN_PASSWORD\"
  }")

ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token' 2>/dev/null || echo "null")
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.refresh_token' 2>/dev/null || echo "null")

if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ Login failed!"
  echo "Response: $LOGIN_RESPONSE"
  echo ""
  echo "💡 Tip: Make sure the admin user exists in MongoDB"
  echo "Use: mongosh 'mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin'"
  echo "Then create admin: db.admins.insertOne({...})"
  exit 1
fi

echo "✅ Login successful!"
echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo ""

# Step 2: Get Admin Profile
echo "2️⃣  Testing Get Admin Profile..."
PROFILE_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/admin/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json")

ADMIN_ID=$(echo $PROFILE_RESPONSE | jq -r '.id' 2>/dev/null || echo "null")
ADMIN_NAME=$(echo $PROFILE_RESPONSE | jq -r '.name' 2>/dev/null || echo "null")
ADMIN_ROLE=$(echo $PROFILE_RESPONSE | jq -r '.role' 2>/dev/null || echo "null")

if [ "$ADMIN_ID" = "null" ] || [ -z "$ADMIN_ID" ]; then
  echo "❌ Get profile failed!"
  echo "Response: $PROFILE_RESPONSE"
  exit 1
fi

echo "✅ Get admin profile successful!"
echo "   Admin ID: $ADMIN_ID"
echo "   Admin Name: $ADMIN_NAME"
echo "   Admin Role: $ADMIN_ROLE"
echo ""

# Step 3: Verify sensitive fields are removed
echo "3️⃣  Verifying sensitive fields are removed..."
PASSWORD_HASH=$(echo $PROFILE_RESPONSE | jq -r '.password_hash' 2>/dev/null || echo "null")
SETUP_TOKEN=$(echo $PROFILE_RESPONSE | jq -r '.setup_token' 2>/dev/null || echo "null")

if [ "$PASSWORD_HASH" != "null" ] && [ -n "$PASSWORD_HASH" ] && [ "$PASSWORD_HASH" != "" ]; then
  echo "❌ Password hash should be removed!"
  exit 1
fi

if [ "$SETUP_TOKEN" != "null" ] && [ -n "$SETUP_TOKEN" ] && [ "$SETUP_TOKEN" != "" ]; then
  echo "❌ Setup token should be removed!"
  exit 1
fi

echo "✅ Sensitive fields properly removed!"
echo ""

# Step 4: Test invalid token
echo "4️⃣  Testing with invalid token..."
INVALID_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/admin/profile" \
  -H "Authorization: Bearer invalid_token_12345" \
  -H "Content-Type: application/json")

ERROR=$(echo $INVALID_RESPONSE | jq -r '.error' 2>/dev/null || echo "null")

if [ "$ERROR" = "null" ] || [ -z "$ERROR" ]; then
  echo "❌ Should have returned error for invalid token!"
  exit 1
fi

echo "✅ Invalid token properly rejected!"
echo "   Error: $ERROR"
echo ""

# Step 5: Test missing auth header
echo "5️⃣  Testing without authorization header..."
NO_AUTH_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/admin/profile" \
  -H "Content-Type: application/json")

ERROR=$(echo $NO_AUTH_RESPONSE | jq -r '.error' 2>/dev/null || echo "null")

if [ "$ERROR" = "null" ] || [ -z "$ERROR" ]; then
  echo "❌ Should have returned error for missing auth!"
  exit 1
fi

echo "✅ Missing auth header properly rejected!"
echo "   Error: $ERROR"
echo ""

# Display full profile response
echo "════════════════════════════════════════════════════"
echo "✅ All tests passed!"
echo "════════════════════════════════════════════════════"
echo ""
echo "📊 Admin Profile Response:"
echo $PROFILE_RESPONSE | jq '.' 2>/dev/null || echo $PROFILE_RESPONSE
echo ""
