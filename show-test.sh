#!/bin/bash

# Quick test visualization script
# Shows the current state of the pipeline

API_URL="http://localhost:8080"

echo "╔════════════════════════════════════════════════════════╗"
echo "║          Argus Pipeline Test Results                   ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Check API
echo "🔍 Checking API..."
if curl -s -f "${API_URL}/health" > /dev/null; then
    echo "   ✅ API is healthy"
else
    echo "   ❌ API not responding"
    exit 1
fi
echo ""

# Check Docker
echo "🐳 Checking infrastructure..."
if docker ps | grep -q "infra-rabbitmq-1"; then
    echo "   ✅ RabbitMQ running"
else
    echo "   ❌ RabbitMQ not running"
fi
if docker ps | grep -q "infra-db-1"; then
    echo "   ✅ Postgres running"
else
    echo "   ❌ Postgres not running"
fi
echo ""

# Show deliveries
echo "📦 Current Deliveries:"
echo ""
DELIVERIES=$(curl -s "${API_URL}/deliveries" 2>/dev/null)

if [ -n "$DELIVERIES" ]; then
    echo "$DELIVERIES" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    total = len(data)
    delivered = sum(1 for d in data if d.get('status') == 'delivered')
    queued = sum(1 for d in data if d.get('status') == 'queued')
    
    print(f'   Total: {total}')
    print(f'   ✅ Delivered: {delivered}')
    print(f'   ⏳ Queued: {queued}')
    print('')
    print('   Recent events:')
    print('   ──────────────────────────────────────────────────────')
    for d in data[:5]:
        status_icon = '✅' if d['status'] == 'delivered' else '⏳'
        source = d.get('source', 'unknown')[:12]
        title = d.get('title', '')[:35]
        event_id = d.get('event_id', '')[:8]
        print(f'   {status_icon} [{source:12}] {title:35} ({event_id}...)')
except:
    print('   (Could not parse deliveries)')
" 2>/dev/null || echo "   (No deliveries or parsing error)"
else
    echo "   (No deliveries found)"
fi

echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║  To test: Run ./test-pipeline.sh                      ║"
echo "╚════════════════════════════════════════════════════════╝"
