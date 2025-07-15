#!/bin/bash
# Script to debug Neutron chain connection issues

echo "=== Debugging Neutron Chain Connection ==="
echo ""

# Check Neutron RPC status
echo "1. Checking Neutron RPC status..."
curl -s -u 'skip:p01kachu?!' https://neutron-1-skip-rpc.polkachu.com/status | jq -r '.result.node_info | {chain_id: .network, version: .version, protocol_version: .protocol_version}'

echo ""
echo "2. Checking ABCI info..."
curl -s -u 'skip:p01kachu?!' https://neutron-1-skip-rpc.polkachu.com/abci_info | jq -r '.result.response'

echo ""
echo "3. Checking current chainpulse metrics for Neutron..."
curl -s http://localhost:3001/metrics | grep -E "neutron-1"

echo ""
echo "4. Checking chainpulse logs for Neutron errors (last 10)..."
docker logs relayooor-chainpulse-1 2>&1 | grep -i "neutron" | tail -10

echo ""
echo "5. Testing WebSocket connection directly..."
echo "Note: This requires wscat. Install with: npm install -g wscat"
echo "Command to test: wscat -c 'wss://skip:p01kachu%3F%21@neutron-1-skip-rpc.polkachu.com/websocket'"

echo ""
echo "=== Recommendations ==="
echo "Based on the version info above:"
echo "- If version starts with '0.34', use comet_version = \"0.34\""
echo "- If version starts with '0.37', use comet_version = \"0.37\""
echo "- If version starts with '0.38', use comet_version = \"0.38\""
echo ""
echo "Current config uses: $(grep -A1 'neutron-1' /Users/cordt/repos/relayooor/config/chainpulse-selected.toml | grep comet_version)"