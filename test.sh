#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}Go Backend Boilerplate - Test${NC}"
echo -e "${BLUE}================================${NC}\n"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed!${NC}"
    echo -e "${YELLOW}Install Go from: https://golang.org/dl${NC}\n"
    exit 1
fi

echo -e "${GREEN}✓ Go is installed${NC} ($(go version))\n"

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file not found${NC}"
    echo -e "${BLUE}ℹ️  Creating .env from env.example...${NC}"
    cp env.example .env
    
    # Auto-generate secrets
    if command -v openssl &> /dev/null; then
        ACCESS_SECRET=$(openssl rand -base64 32)
        REFRESH_SECRET=$(openssl rand -base64 32)
        
        # Update .env file
        sed -i.bak "s|your-access-token-secret-change-me-in-production|${ACCESS_SECRET}|g" .env
        sed -i.bak "s|your-refresh-token-secret-change-me-in-production|${REFRESH_SECRET}|g" .env
        rm .env.bak
        
        echo -e "${GREEN}✓ .env file created and secrets generated${NC}\n"
    else
        echo -e "${YELLOW}⚠️  openssl not found, please update secrets in .env manually${NC}\n"
    fi
else
    echo -e "${GREEN}✓ .env file exists${NC}\n"
fi

# Download dependencies
echo -e "${BLUE}📦 Downloading dependencies...${NC}"
go mod download
echo -e "${GREEN}✓ Dependencies downloaded${NC}\n"

# Build test
echo -e "${BLUE}🔨 Testing build...${NC}"
if go build -o /tmp/go-boilerplate-test cmd/api/main.go; then
    echo -e "${GREEN}✓ Build successful${NC}\n"
    rm /tmp/go-boilerplate-test
else
    echo -e "${RED}❌ Build failed!${NC}\n"
    exit 1
fi

# Start server in background
echo -e "${BLUE}🚀 Starting server...${NC}"
go run cmd/api/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Health check
echo -e "\n${BLUE}🏥 Testing health endpoint...${NC}"
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)

if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo -e "${GREEN}✓ Health endpoint working!${NC}"
    echo -e "Response: ${HEALTH_RESPONSE}\n"
else
    echo -e "${RED}❌ Health endpoint not working!${NC}\n"
    kill $SERVER_PID
    exit 1
fi

# Register test
echo -e "${BLUE}👤 Testing register endpoint...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}')

if echo "$REGISTER_RESPONSE" | grep -q "access_token"; then
    echo -e "${GREEN}✓ Register endpoint working!${NC}"
    echo -e "Response (first 100 chars): $(echo $REGISTER_RESPONSE | cut -c1-100)...\n"
    
    # Extract access token
    ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    
    # Test protected endpoint
    echo -e "${BLUE}🔐 Testing protected endpoint (/v1/auth/me)...${NC}"
    ME_RESPONSE=$(curl -s http://localhost:8080/v1/auth/me \
      -H "Authorization: Bearer $ACCESS_TOKEN")
    
    if echo "$ME_RESPONSE" | grep -q "test@example.com"; then
        echo -e "${GREEN}✓ Protected endpoint working!${NC}"
        echo -e "Response: ${ME_RESPONSE}\n"
    else
        echo -e "${RED}❌ Protected endpoint not working!${NC}\n"
    fi
else
    echo -e "${RED}❌ Register endpoint not working!${NC}"
    echo -e "Response: ${REGISTER_RESPONSE}\n"
fi

# Stop server
echo -e "${YELLOW}🛑 Stopping server...${NC}"
kill $SERVER_PID
sleep 1

echo -e "\n${BLUE}================================${NC}"
echo -e "${GREEN}✅ Test completed!${NC}"
echo -e "${BLUE}================================${NC}\n"

echo -e "${YELLOW}To run server manually:${NC}"
echo -e "  ${BLUE}go run cmd/api/main.go${NC}\n"

echo -e "${YELLOW}Or use Makefile:${NC}"
echo -e "  ${BLUE}make run${NC}\n"
