#!/bin/bash

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR/.."

echo "╔════════════════════════════════════════╗"
echo "║   Temp Backend - Local Dev Setup       ║"
echo "╚════════════════════════════════════════╝"
echo ""

# Step 1: Check Docker
echo "1️⃣  Checking Docker..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker."
    exit 1
fi
echo "✅ Docker found"
echo ""

# Step 2: Check .env file
echo "2️⃣  Checking .env file..."
if [ ! -f ".env" ]; then
    echo "⚠️  .env file not found. Creating from .env.example..."
    cp .env.example .env
    echo "✅ .env created (adjust values if needed)"
else
    echo "✅ .env file found"
fi
echo ""

# Step 3: Start Docker services
echo "3️⃣  Starting Docker services (MongoDB, MinIO)..."
docker-compose up -d mongodb minio createbuckets

echo "⏳ Waiting for services to be healthy..."
for i in {1..30}; do
    if docker-compose exec -T mongodb mongosh --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
        echo "✅ MongoDB is healthy"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ MongoDB failed to start"
        exit 1
    fi
    sleep 1
done

for i in {1..30}; do
    if docker-compose exec -T minio mc ready local >/dev/null 2>&1; then
        echo "✅ MinIO is healthy"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ MinIO failed to start"
        exit 1
    fi
    sleep 1
done
echo ""

# Step 4: Build the application
echo "4️⃣  Building application..."
go build -o temp_backend cmd/api/main.go 2>&1 | grep -v "^#"
echo "✅ Build successful"
echo ""

# Step 5: Show information
echo "════════════════════════════════════════"
echo "✅ Setup complete!"
echo "════════════════════════════════════════"
echo ""
echo "📋 Service Information:"
echo "  MongoDB:      mongodb://admin:secret@localhost:27017"
echo "  MinIO API:    http://localhost:9000"
echo "  MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"
echo ""
echo "🚀 To start the API server, run:"
echo "   go run cmd/api/main.go"
echo ""
echo "📖 For testing guide, see:"
echo "   cat LOCAL_ADMIN_TESTING_GUIDE.md"
echo ""
echo "🧪 To run tests:"
echo "   ./scripts/test_admin_endpoints.sh"
echo ""
