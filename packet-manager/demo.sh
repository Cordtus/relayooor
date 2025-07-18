#!/bin/bash

# Packet Manager Demonstration Script
# This script demonstrates how to test and verify the packet manager capabilities

echo "═══════════════════════════════════════════════════════════════"
echo "       IBC Packet Manager - Demonstration & Testing"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Check if services are running
echo "1. Checking services status..."
echo "───────────────────────────────────────────────────────────────"

check_service() {
    local service=$1
    local port=$2
    local url=$3
    
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "200\|404"; then
        echo "✅ $service is running on port $port"
    else
        echo "❌ $service is not accessible on port $port"
    fi
}

check_service "Packet Manager UI" 5174 "http://localhost:5174"
check_service "API Backend" 3000 "http://localhost:3000/api/health"
check_service "Chainpulse Metrics" 3001 "http://localhost:3001"
check_service "Hermes REST API" 5185 "http://localhost:5185/version"
check_service "Hermes Metrics" 3010 "http://localhost:3010/metrics"

echo ""
echo "2. Querying stuck packets from different chains..."
echo "───────────────────────────────────────────────────────────────"

# Function to query stuck packets
query_stuck_packets() {
    local chain=$1
    echo ""
    echo "📊 Querying $chain..."
    
    # Query stuck packets
    result=$(curl -s "http://localhost:3000/api/packets/stuck?chain=$chain" | jq -r 'length')
    
    if [ "$result" -gt 0 ]; then
        echo "  Found $result stuck packets on $chain"
        
        # Show sample packet details
        curl -s "http://localhost:3000/api/packets/stuck?chain=$chain" | jq -r '.[0:2] | .[] | "  - Channel: \(.channelId), Sequence: \(.sequence), Stuck for: \(.stuckDuration)"' 2>/dev/null
    else
        echo "  No stuck packets found on $chain"
    fi
}

# Query each chain
for chain in cosmoshub-4 osmosis-1 noble-1 stride-1 jackal-1 axelar-dojo-1; do
    query_stuck_packets $chain
done

echo ""
echo "3. Verifying Hermes metrics..."
echo "───────────────────────────────────────────────────────────────"

# Check Hermes metrics
if curl -s http://localhost:3010/metrics > /dev/null 2>&1; then
    echo "✅ Hermes metrics endpoint is accessible"
    
    # Count pending packets
    pending=$(curl -s http://localhost:3010/metrics | grep "hermes_pending_packets" | grep -v "#" | wc -l)
    echo "  Found metrics for $pending channel configurations"
else
    echo "❌ Hermes metrics endpoint is not accessible"
fi

echo ""
echo "4. Cross-validation example..."
echo "───────────────────────────────────────────────────────────────"

# Find a chain with stuck packets
for chain in osmosis-1 axelar-dojo-1 jackal-1; do
    count=$(curl -s "http://localhost:3000/api/packets/stuck?chain=$chain" | jq -r 'length')
    if [ "$count" -gt 0 ]; then
        echo "Example: $chain has $count stuck packets"
        
        # Get channel details
        channel=$(curl -s "http://localhost:3000/api/packets/stuck?chain=$chain" | jq -r '.[0].channelId')
        echo "  Channel with stuck packets: $channel"
        
        # Try to find corresponding Hermes metrics
        echo "  Checking Hermes metrics for validation..."
        hermes_data=$(curl -s http://localhost:3010/metrics | grep "hermes_pending_packets" | grep "$chain" | grep "$channel" || echo "  No matching Hermes metrics found")
        
        if [ -n "$hermes_data" ]; then
            echo "  $hermes_data"
        fi
        
        break
    fi
done

echo ""
echo "5. Testing packet clearing (simulated)..."
echo "───────────────────────────────────────────────────────────────"

# Simulate clearing a packet
echo "Simulating packet clear request..."
echo "curl -X POST http://localhost:3000/api/relayer/hermes/clear -d '{\"chain\":\"osmosis-1\",\"channel\":\"channel-0\",\"sequences\":[1234]}'"
echo "Note: With the simple API, this is simulated and won't actually clear packets"

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "                    Testing Complete!"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "📝 How to use the Packet Manager:"
echo ""
echo "1. Open http://localhost:5174 in your browser"
echo "2. Select a chain (try Osmosis or Axelar for more stuck packets)"
echo "3. Select a channel"
echo "4. Click 'Query Packets' to see both:"
echo "   - Chainpulse data (left panel) - actual stuck packets"
echo "   - Hermes metrics (right panel) - pending packet counts"
echo "5. Use 'Clear' or 'Clear All Packets' buttons"
echo ""
echo "🔍 What to look for:"
echo "- Chains with high stuck packet counts (Osmosis, Axelar often have more)"
echo "- Cross-validation between Chainpulse and Hermes data"
echo "- Channel-specific packet backlogs"
echo "- Packet details including sender/receiver when available"
echo ""
echo "⚡ Common problematic channels:"
echo "- Osmosis channel-0 (to Cosmos Hub)"
echo "- Osmosis channel-750 (to Noble)"
echo "- Axelar channel-208 (to Osmosis)"
echo "- Jackal channel-2 (to Osmosis)"
echo ""