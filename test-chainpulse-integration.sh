#!/bin/bash

echo "Testing Chainpulse Integration..."
echo "================================"

CHAINPULSE_URL="http://localhost:3000"
API_BACKEND_URL="http://localhost:8080"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test 1: Prometheus metrics
echo -e "\n${YELLOW}1. Testing Chainpulse Prometheus metrics:${NC}"
METRICS=$(curl -s ${CHAINPULSE_URL}/metrics)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Metrics endpoint accessible${NC}"
    echo "Sample metrics:"
    echo "$METRICS" | grep -E "(chainpulse_chains|ibc_stuck_packets|ibc_effected_packets)" | head -10
else
    echo -e "${RED}✗ Failed to access metrics endpoint${NC}"
fi

# Test 2: Stuck packets API
echo -e "\n${YELLOW}2. Testing stuck packets API:${NC}"
STUCK_PACKETS=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/stuck?min_age_seconds=300")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Stuck packets API accessible${NC}"
    echo "$STUCK_PACKETS" | jq '.' 2>/dev/null || echo "$STUCK_PACKETS"
else
    echo -e "${RED}✗ Failed to access stuck packets API${NC}"
fi

# Test 3: User packets API
echo -e "\n${YELLOW}3. Testing packets by user API:${NC}"
# Test with a sample Osmosis address
TEST_ADDRESS="osmo1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y"
USER_PACKETS=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/by-user?address=${TEST_ADDRESS}&role=sender")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ User packets API accessible${NC}"
    echo "$USER_PACKETS" | jq '.' 2>/dev/null || echo "$USER_PACKETS"
else
    echo -e "${RED}✗ Failed to access user packets API${NC}"
fi

# Test 4: Channel congestion
echo -e "\n${YELLOW}4. Testing channel congestion API:${NC}"
CONGESTION=$(curl -s "${CHAINPULSE_URL}/api/v1/channels/congestion")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Channel congestion API accessible${NC}"
    echo "$CONGESTION" | jq '.' 2>/dev/null || echo "$CONGESTION"
else
    echo -e "${RED}✗ Failed to access channel congestion API${NC}"
fi

# Test 5: Specific packet details (example)
echo -e "\n${YELLOW}5. Testing specific packet details API:${NC}"
# This will likely 404 unless we have actual packet data
PACKET_DETAILS=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/osmosis-1/channel-0/1")
echo "Response: $PACKET_DETAILS"

# Test 6: API backend integration
echo -e "\n${YELLOW}6. Testing API backend integration:${NC}"
if curl -s ${API_BACKEND_URL}/health >/dev/null 2>&1; then
    echo -e "${GREEN}✓ API backend is running${NC}"
    
    # Test stuck packets through backend
    BACKEND_STUCK=$(curl -s ${API_BACKEND_URL}/api/packets/stuck)
    echo "Backend stuck packets response:"
    echo "$BACKEND_STUCK" | jq '.' 2>/dev/null || echo "$BACKEND_STUCK"
else
    echo -e "${RED}✗ API backend not running on port 8080${NC}"
fi

# Test 7: Analyze metrics for CometBFT 0.38 compatibility
echo -e "\n${YELLOW}7. Analyzing metrics for chain versions:${NC}"
if [ -n "$METRICS" ]; then
    echo "Checking monitored chains:"
    echo "$METRICS" | grep -E "chainpulse_chains|chainpulse_txs{chain_id" | head -10
    
    echo -e "\nPacket metrics by chain:"
    echo "$METRICS" | grep "ibc_effected_packets" | grep -E "(cosmoshub-4|neutron-1|osmosis-1)" | head -5
fi

echo -e "\n${GREEN}Integration test complete!${NC}"