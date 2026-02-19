#!/bin/bash

# Simple Argus Demo Script
# Run this in your Proxmox container to show a test demo

echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║                                                        ║"
echo "║              🎉 Hello Argus! 🎉                        ║"
echo "║                                                        ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Show system info
echo "System Information:"
echo "   Hostname: $(hostname)"
echo "   Date: $(date)"
echo "   Uptime: $(uptime -p 2>/dev/null || uptime)"
echo ""

# Show container info if available
if [ -f /etc/pve/.version ]; then
    echo "Proxmox Container Info:"
    echo "   Container ID: $(hostname | grep -o '[0-9]*' || echo 'N/A')"
    echo "   OS: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2 2>/dev/null || echo 'Unknown')"
    echo ""
fi

# Show network info
echo "Network:"
echo "   IP Address: $(hostname -I 2>/dev/null | awk '{print $1}' || echo 'N/A')"
echo ""

# Show disk usage
echo "Disk Usage:"
df -h / | tail -1 | awk '{print "   Used: " $3 " / " $2 " (" $5 ")"}'
echo ""

# Success message
echo "Argus container is running successfully!"
echo ""

# echo "Next steps:"
# echo "   - Install Go: apt-get update && apt-get install -y golang-go"
# echo "   - Clone your Argus repo and run the pipeline tests"
# echo "   - Or run: ./test-pipeline.sh (if you have the code)"
# echo ""

