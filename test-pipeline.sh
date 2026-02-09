#!/bin/bash

# Argus Pipeline Test Script
# This script tests the full pipeline: API → RabbitMQ → Worker → API

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Argus Pipeline Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Check Docker infrastructure
echo -e "${YELLOW}[1/7] Checking Docker infrastructure...${NC}"
if ! docker ps | grep -q "infra-rabbitmq-1\|infra-db-1"; then
    echo -e "${RED}✗ Docker containers not running!${NC}"
    echo "   Run: cd infra && docker compose up -d"
    exit 1
fi
echo -e "${GREEN}✓ Docker containers running${NC}"
echo ""

# Step 2: Check if API is running
echo -e "${YELLOW}[2/7] Checking API health...${NC}"
API_URL="http://localhost:8080"
if ! curl -s -f "${API_URL}/health" > /dev/null; then
    echo -e "${RED}✗ API not responding!${NC}"
    echo "   Make sure API is running: cd backend && go run ./cmd/api"
    exit 1
fi
HEALTH_RESPONSE=$(curl -s "${API_URL}/health")
if echo "$HEALTH_RESPONSE" | grep -q '"ok":true'; then
    echo -e "${GREEN}✓ API is healthy${NC}"
else
    echo -e "${RED}✗ API health check failed${NC}"
    exit 1
fi
echo ""

# Step 3: Check if Worker is running (by checking if it processes events)
echo -e "${YELLOW}[3/7] Verifying worker is ready...${NC}"
echo -e "${BLUE}   (Worker should be running in another terminal)${NC}"
echo ""

# Step 4: Get initial delivery count
echo -e "${YELLOW}[4/7] Getting initial state...${NC}"
INITIAL_COUNT=$(curl -s "${API_URL}/deliveries" | python3 -c "import sys, json; data = json.load(sys.stdin); print(len(data))" 2>/dev/null || echo "0")
echo -e "${BLUE}   Initial deliveries: ${INITIAL_COUNT}${NC}"
echo ""

# Step 5: Publish test event via API
echo -e "${YELLOW}[5/7] Publishing event via API endpoint...${NC}"
API_RESPONSE=$(curl -s -X POST "${API_URL}/debug/publish")
EVENT_ID=$(echo "$API_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('event_id', ''))" 2>/dev/null || echo "")
if [ -n "$EVENT_ID" ]; then
    echo -e "${GREEN}✓ Event published via API${NC}"
    echo -e "${BLUE}   Event ID: ${EVENT_ID}${NC}"
else
    echo -e "${RED}✗ Failed to publish event${NC}"
    exit 1
fi
echo ""

# Step 6: Publish test event via CLI
echo -e "${YELLOW}[6/7] Publishing event via CLI tool...${NC}"
cd backend
CLI_OUTPUT=$(go run ./cmd/cli -source="test-script" -title="Pipeline Test Event" 2>&1)
if echo "$CLI_OUTPUT" | grep -q "✓ Published event"; then
    echo -e "${GREEN}✓ Event published via CLI${NC}"
    CLI_EVENT_ID=$(echo "$CLI_OUTPUT" | grep "event_id=" | sed 's/.*event_id=\([a-f0-9]*\).*/\1/' | head -1)
    echo -e "${BLUE}   Event ID: ${CLI_EVENT_ID}${NC}"
else
    echo -e "${RED}✗ Failed to publish event via CLI${NC}"
    echo "$CLI_OUTPUT"
    exit 1
fi
cd ..
echo ""

# Step 7: Wait for processing and verify
echo -e "${YELLOW}[7/7] Waiting for events to be processed...${NC}"
sleep 3

FINAL_COUNT=$(curl -s "${API_URL}/deliveries" | python3 -c "import sys, json; data = json.load(sys.stdin); print(len(data))" 2>/dev/null || echo "0")
DELIVERED_COUNT=$(curl -s "${API_URL}/deliveries" | python3 -c "import sys, json; data = json.load(sys.stdin); print(sum(1 for d in data if d.get('status') == 'delivered'))" 2>/dev/null || echo "0")

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Test Results${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Initial deliveries:  ${INITIAL_COUNT}"
echo -e "Final deliveries:     ${FINAL_COUNT}"
echo -e "Delivered events:     ${DELIVERED_COUNT}"
echo ""

# Show recent deliveries
echo -e "${YELLOW}Recent deliveries:${NC}"
curl -s "${API_URL}/deliveries" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for d in data[:5]:
    status_color = '\033[0;32m' if d['status'] == 'delivered' else '\033[1;33m'
    print(f\"  {status_color}{d['status']:10}\033[0m | {d['source']:15} | {d['title'][:30]}\")
" 2>/dev/null || echo "  (Could not parse deliveries)"

echo ""

# Final verification
if [ "$DELIVERED_COUNT" -gt "$INITIAL_COUNT" ]; then
    echo -e "${GREEN}✓ Pipeline test PASSED!${NC}"
    echo -e "${GREEN}  Events are flowing through the system correctly.${NC}"
    exit 0
else
    echo -e "${RED}✗ Pipeline test FAILED!${NC}"
    echo -e "${RED}  Expected more delivered events.${NC}"
    exit 1
fi
