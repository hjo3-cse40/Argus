#!/bin/bash

# Test script for the migration command
# This script verifies the migration command works correctly

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Migration Command Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Check if database is running
echo -e "${YELLOW}[1/4] Checking database availability...${NC}"
if ! docker ps | grep -q "infra-db-1"; then
    echo -e "${RED}✗ Database container not running!${NC}"
    echo "   Run: cd infra && docker compose up -d db"
    exit 1
fi
echo -e "${GREEN}✓ Database container is running${NC}"
echo ""

# Step 2: Set environment variables
echo -e "${YELLOW}[2/4] Setting environment variables...${NC}"
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=argus
export DB_PASSWORD=argus
export DB_NAME=argus
export ENV=dev
echo -e "${GREEN}✓ Environment variables set${NC}"
echo ""

# Step 3: Build the migration command
echo -e "${YELLOW}[3/4] Building migration command...${NC}"
cd backend
go build -o ../migrate ./cmd/migrate
cd ..
echo -e "${GREEN}✓ Migration command built${NC}"
echo ""

# Step 4: Run migrations
echo -e "${YELLOW}[4/4] Running migrations...${NC}"
./migrate
MIGRATION_EXIT_CODE=$?

echo ""
if [ $MIGRATION_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ Migration test PASSED!${NC}"
    echo -e "${GREEN}  Migrations completed successfully.${NC}"
    exit 0
else
    echo -e "${RED}✗ Migration test FAILED!${NC}"
    echo -e "${RED}  Migration command exited with code ${MIGRATION_EXIT_CODE}${NC}"
    exit 1
fi
