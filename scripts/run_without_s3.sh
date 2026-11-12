#!/bin/bash
set -e

# Script to run the application without S3/MinIO

echo "🚀 Starting temp_backend without S3..."
echo ""

# Check if MongoDB is running
echo "✓ Checking MongoDB..."
if ! docker-compose ps mongodb | grep -q "healthy"; then
    echo "  MongoDB not running. Starting..."
    docker-compose up -d mongodb
    echo "  Waiting for MongoDB to be healthy..."
    sleep 5
fi
echo "  ✓ MongoDB is ready"
echo ""

# Set up environment
echo "📝 Configuring environment..."
export S3_ENABLED=false
export MONGO_URI=mongodb://admin:secret@localhost:27017/temp_backend?authSource=admin
export MONGO_DATABASE=temp_backend
export APP_NAME=temp_backend
export PORT=8080
export APP_ENV=development
export LOG_LEVEL=info
export JWT_SECRET=your-256-bit-secret-key-min-32-characters-long-here-12345
export ACCESS_TOKEN_TTL=1h
export REFRESH_TOKEN_TTL=168h
export ADMIN_ACCESS_TOKEN_TTL=30m
export ADMIN_REFRESH_TOKEN_TTL=168h
export ADMIN_MAX_LOGIN_ATTEMPTS=5
export ADMIN_LOCKOUT_DURATION=15m
export ADMIN_SETUP_TOKEN_EXPIRY=24h
export TOKEN_RESEND_COOLDOWN=2m
export SENDGRID_API_KEY=
export SENDGRID_SENDER_EMAIL=noreply@example.com
export SENDGRID_ENABLED=false
export APPSYNC_ENABLED=false
export AWS_REGION=us-east-1
export RATE_LIMIT_GLOBAL_MAX=100
export RATE_LIMIT_GLOBAL_WINDOW=1m
export RATE_LIMIT_AUTH_MAX=20
export RATE_LIMIT_AUTH_WINDOW=1m
export RATE_LIMIT_EMAIL_MAX=5
export RATE_LIMIT_EMAIL_WINDOW=1h
export FIREBASE_ENABLED=false
export NOTIFICATION_ACTIVITY_COOLDOWN_MINUTES=30
export DEVICE_TOKEN_EXPIRY_DAYS=90
export MAX_REPORTS_PER_DAY=10
export AUTO_SOFT_DELETE_REPORT_THRESHOLD=15

echo "  ✓ Environment configured"
echo ""

echo "🎯 Starting application on http://localhost:8080"
echo "   File storage is DISABLED - file uploads will fail"
echo ""
echo "To stop the server, press Ctrl+C"
echo ""

# Run the application
go run cmd/api/main.go
