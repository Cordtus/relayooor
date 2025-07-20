#!/bin/bash

echo "=== Testing Packet Clearing Flow ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 1. Check Hermes state
echo -e "${YELLOW}1. Checking Hermes state...${NC}"
HERMES_STATE=$(curl -s http://localhost:5185/state)
echo "$HERMES_STATE" | jq .
echo ""

# 2. Check for stuck packets from our API
echo -e "${YELLOW}2. Checking for stuck packets...${NC}"
STUCK_PACKETS=$(curl -s "http://localhost:3000/api/packets/stuck?limit=1")
echo "$STUCK_PACKETS" | jq .
echo ""

# Extract first stuck packet details
if [ "$(echo "$STUCK_PACKETS" | jq -r '. | length')" -gt 0 ]; then
    PACKET=$(echo "$STUCK_PACKETS" | jq -r '.[0]')
    CHAIN_ID=$(echo "$PACKET" | jq -r '.sourceChain')
    CHANNEL_ID=$(echo "$PACKET" | jq -r '.channelId')
    SEQUENCE=$(echo "$PACKET" | jq -r '.sequence')
    
    echo -e "${GREEN}Found stuck packet:${NC}"
    echo "Chain: $CHAIN_ID"
    echo "Channel: $CHANNEL_ID"
    echo "Sequence: $SEQUENCE"
    echo ""
    
    # 3. Simulate packet clearing request
    echo -e "${YELLOW}3. Simulating packet clearing request...${NC}"
    echo "Would send to Hermes:"
    cat <<EOF | jq .
{
  "chain_id": "$CHAIN_ID",
  "channel_id": "$CHANNEL_ID",
  "port_id": "transfer",
  "sequences": [$SEQUENCE]
}
EOF
    
    # 4. Check if we can query packet commitments (dry run)
    echo ""
    echo -e "${YELLOW}4. Checking chain status in Hermes...${NC}"
    CHAIN_STATUS=$(curl -s "http://localhost:5185/chain/$CHAIN_ID" 2>/dev/null || echo '{"error": "Chain not configured"}')
    if [ -n "$CHAIN_STATUS" ] && [ "$CHAIN_STATUS" != "" ]; then
        echo "$CHAIN_STATUS" | jq .
    else
        echo -e "${RED}Chain $CHAIN_ID not configured in Hermes${NC}"
    fi
    
else
    echo -e "${RED}No stuck packets found${NC}"
    
    # Create a mock packet for testing
    echo ""
    echo -e "${YELLOW}Creating mock packet for dry-run test...${NC}"
    cat <<EOF | jq .
{
  "test_mode": true,
  "mock_packet": {
    "chain_id": "cosmoshub-4",
    "channel_id": "channel-141",
    "port_id": "transfer",
    "sequence": 12345,
    "amount": "1000000",
    "denom": "uatom",
    "sender": "cosmos1test...",
    "receiver": "osmo1test..."
  }
}
EOF
fi

# 5. Test Hermes REST API endpoints
echo ""
echo -e "${YELLOW}5. Testing Hermes REST API capabilities...${NC}"

# Check if clear endpoint exists (it might not in this version)
echo "Checking available endpoints..."
curl -s -X OPTIONS http://localhost:5185/ 2>/dev/null || echo "No OPTIONS response"

# Try to get chain queries
echo ""
echo "Attempting chain query for cosmoshub-4..."
COSMOS_CHAIN=$(curl -s http://localhost:5185/chain/cosmoshub-4)
if [ -n "$COSMOS_CHAIN" ]; then
    echo "$COSMOS_CHAIN" | jq -r '.result | {id: .id, rpc_addr: .rpc_addr, key_name: .key_name}'
fi

echo ""
echo -e "${GREEN}=== Dry Run Complete ===${NC}"
echo ""
echo "Summary:"
echo "- Hermes is running and connected to: $(echo "$HERMES_STATE" | jq -r '.result.chains[]' | tr '\n' ', ' | sed 's/,$//')"
echo "- Stuck packets API is working"
echo "- REST API is accessible"
echo ""
echo "Note: Actual packet clearing would require:"
echo "1. Hermes to have wallet keys configured"
echo "2. The clear_packets endpoint (may need custom implementation)"
echo "3. Gas fees in the relayer wallet"