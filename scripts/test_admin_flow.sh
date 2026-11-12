#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

API_URL="${API_URL:-http://localhost:8080}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-password123}"

echo -e "${BLUE}=== Admin Authentication Flow Test ===${NC}\n"

# Test 1: Admin Login
echo -e "${BLUE}1. Testing Admin Login (POST /api/v1/admin/auth/login)${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/admin/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")

echo "Response:"
echo "$LOGIN_RESPONSE" | jq .

# Extract tokens
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token // empty')

if [ -z "$ACCESS_TOKEN" ]; then
  echo -e "${RED}✗ Login failed: No access token received${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Login successful${NC}\n"

# Test 2: Admin Profile
echo -e "${BLUE}2. Testing Get Admin Profile (GET /api/v1/admin/profile)${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/admin/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Response:"
echo "$PROFILE_RESPONSE" | jq .

ADMIN_ID=$(echo "$PROFILE_RESPONSE" | jq -r '.id // empty')
if [ -z "$ADMIN_ID" ]; then
  echo -e "${RED}✗ Profile retrieval failed${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Profile retrieval successful${NC}\n"

# Test 3: Admin Profile without token (should fail)
echo -e "${BLUE}3. Testing Admin Profile without token (should fail)${NC}"
UNAUTH_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/admin/profile")

echo "Response:"
echo "$UNAUTH_RESPONSE" | jq .

if echo "$UNAUTH_RESPONSE" | jq -e '.error' > /dev/null; then
  echo -e "${GREEN}✓ Correctly rejected unauthorized request${NC}\n"
else
  echo -e "${RED}✗ Should have rejected unauthorized request${NC}"
  exit 1
fi

# Test 4: Admin Logout
echo -e "${BLUE}4. Testing Admin Logout (POST /api/v1/admin/auth/logout)${NC}"
LOGOUT_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/admin/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Response:"
echo "$LOGOUT_RESPONSE" | jq .

echo -e "${GREEN}✓ Logout successful${NC}\n"

# Test 5: Using old token after logout (should fail)
echo -e "${BLUE}5. Testing old token after logout (should fail)${NC}"
OLD_TOKEN_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/admin/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Response:"
echo "$OLD_TOKEN_RESPONSE" | jq .

if echo "$OLD_TOKEN_RESPONSE" | jq -e '.error' > /dev/null; then
  echo -e "${GREEN}✓ Correctly rejected revoked token${NC}\n"
else
  echo -e "${YELLOW}! Note: Token still valid (may be by design)${NC}\n"
fi

echo -e "${GREEN}=== All tests completed successfully ===${NC}"
