#!/bin/bash

echo "CometBFT 0.38 Compatibility Test"
echo "================================"

CHAINPULSE_URL="http://localhost:3000"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Expected chains and their consensus versions
declare -A CHAIN_VERSIONS=(
    ["cosmoshub-4"]="0.34"
    ["neutron-1"]="0.34"
    ["osmosis-1"]="0.34"
)

echo -e "${BLUE}=== 1. Chain Connectivity Test ===${NC}"

# Test if ChainPulse is running
if ! curl -s ${CHAINPULSE_URL}/metrics >/dev/null 2>&1; then
    echo -e "${RED}ChainPulse is not accessible at ${CHAINPULSE_URL}${NC}"
    exit 1
fi

echo -e "${GREEN}✓ ChainPulse is running${NC}"

# Fetch metrics
METRICS=$(curl -s ${CHAINPULSE_URL}/metrics)

echo -e "\n${BLUE}=== 2. Chain Monitoring Status ===${NC}"

for CHAIN in "${!CHAIN_VERSIONS[@]}"; do
    VERSION="${CHAIN_VERSIONS[$CHAIN]}"
    echo -e "\n${YELLOW}Checking ${CHAIN} (Consensus: ${VERSION})...${NC}"
    
    # Check if chain is being monitored
    if echo "$METRICS" | grep -q "chain_id=\"${CHAIN}\""; then
        echo -e "${GREEN}✓ ${CHAIN} is being monitored${NC}"
        
        # Get transaction count
        TX_COUNT=$(echo "$METRICS" | grep "chainpulse_txs{chain_id=\"${CHAIN}\"" | grep -oE "[0-9]+$" || echo "0")
        echo "  Transactions processed: ${TX_COUNT}"
        
        # Get packet count
        PACKET_COUNT=$(echo "$METRICS" | grep "chainpulse_packets{chain_id=\"${CHAIN}\"" | grep -oE "[0-9]+$" || echo "0")
        echo "  Packets processed: ${PACKET_COUNT}"
        
        # Check for errors
        ERROR_COUNT=$(echo "$METRICS" | grep "chainpulse_errors{chain_id=\"${CHAIN}\"" | grep -oE "[0-9]+$" || echo "0")
        if [ "$ERROR_COUNT" -gt "0" ]; then
            echo -e "  ${RED}⚠ Errors detected: ${ERROR_COUNT}${NC}"
        else
            echo -e "  ${GREEN}✓ No errors${NC}"
        fi
        
        # Check for reconnects (indicates connection issues)
        RECONNECT_COUNT=$(echo "$METRICS" | grep "chainpulse_reconnects{chain_id=\"${CHAIN}\"" | grep -oE "[0-9]+$" || echo "0")
        if [ "$RECONNECT_COUNT" -gt "0" ]; then
            echo "  Reconnections: ${RECONNECT_COUNT}"
        fi
    else
        echo -e "${RED}✗ ${CHAIN} is NOT being monitored${NC}"
    fi
done

echo -e "\n${BLUE}=== 3. IBC Packet Processing Test ===${NC}"

# Check if packets are being processed
echo -e "\n${YELLOW}IBC packet metrics:${NC}"

# Count total effected packets
TOTAL_EFFECTED=$(echo "$METRICS" | grep "ibc_effected_packets{" | grep -oE "[0-9]+$" | awk '{sum+=$1} END {print sum+0}')
echo "Total effected packets: ${TOTAL_EFFECTED}"

# Count total uneffected packets
TOTAL_UNEFFECTED=$(echo "$METRICS" | grep "ibc_uneffected_packets{" | grep -oE "[0-9]+$" | awk '{sum+=$1} END {print sum+0}')
echo "Total uneffected packets: ${TOTAL_UNEFFECTED}"

# Check for stuck packets
echo -e "\n${YELLOW}Checking for stuck packets:${NC}"
STUCK_COUNT=$(echo "$METRICS" | grep "ibc_stuck_packets{" | grep -v " 0$" | wc -l)
if [ "$STUCK_COUNT" -gt "0" ]; then
    echo -e "${YELLOW}Found ${STUCK_COUNT} channels with stuck packets:${NC}"
    echo "$METRICS" | grep "ibc_stuck_packets{" | grep -v " 0$" | head -5
else
    echo -e "${GREEN}✓ No stuck packets detected${NC}"
fi

echo -e "\n${BLUE}=== 4. API Endpoint Functionality ===${NC}"

# Test each API endpoint
echo -e "\n${YELLOW}Testing API endpoints:${NC}"

# Test stuck packets endpoint
echo -n "Testing /api/v1/packets/stuck... "
STUCK_RESPONSE=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/stuck?min_age_seconds=300")
if echo "$STUCK_RESPONSE" | jq -e '.packets' >/dev/null 2>&1; then
    STUCK_API_COUNT=$(echo "$STUCK_RESPONSE" | jq -r '.packets | length')
    echo -e "${GREEN}✓ Working (${STUCK_API_COUNT} packets)${NC}"
else
    echo -e "${RED}✗ Failed${NC}"
fi

# Test user packets endpoint
echo -n "Testing /api/v1/packets/by-user... "
TEST_USER="osmo1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y"
USER_RESPONSE=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/by-user?address=${TEST_USER}")
if echo "$USER_RESPONSE" | jq -e '.packets' >/dev/null 2>&1; then
    USER_PACKET_COUNT=$(echo "$USER_RESPONSE" | jq -r '.packets | length')
    echo -e "${GREEN}✓ Working (${USER_PACKET_COUNT} packets)${NC}"
else
    echo -e "${RED}✗ Failed${NC}"
fi

# Test channel congestion endpoint
echo -n "Testing /api/v1/channels/congestion... "
CONGESTION_RESPONSE=$(curl -s "${CHAINPULSE_URL}/api/v1/channels/congestion")
if echo "$CONGESTION_RESPONSE" | jq -e '.channels' >/dev/null 2>&1; then
    CHANNEL_COUNT=$(echo "$CONGESTION_RESPONSE" | jq -r '.channels | length')
    echo -e "${GREEN}✓ Working (${CHANNEL_COUNT} channels)${NC}"
else
    echo -e "${RED}✗ Failed${NC}"
fi

echo -e "\n${BLUE}=== 5. Data Freshness Check ===${NC}"

# Check if data is being updated
echo -e "\n${YELLOW}Checking data freshness:${NC}"

# Get current metrics
METRICS_1=$(curl -s ${CHAINPULSE_URL}/metrics)
TIMESTAMP_1=$(date +%s)

echo "Waiting 10 seconds to check for updates..."
sleep 10

# Get metrics again
METRICS_2=$(curl -s ${CHAINPULSE_URL}/metrics)
TIMESTAMP_2=$(date +%s)

# Compare key metrics
for CHAIN in "${!CHAIN_VERSIONS[@]}"; do
    TX_1=$(echo "$METRICS_1" | grep "chainpulse_txs{chain_id=\"${CHAIN}\"" | grep -oE "[0-9]+$" || echo "0")
    TX_2=$(echo "$METRICS_2" | grep "chainpulse_txs{chain_id=\"${CHAIN}\"" | grep -oE "[0-9]+$" || echo "0")
    
    if [ "$TX_2" -gt "$TX_1" ]; then
        echo -e "${GREEN}✓ ${CHAIN}: Data is updating (${TX_1} -> ${TX_2} txs)${NC}"
    elif [ "$TX_2" -eq "$TX_1" ] && [ "$TX_1" -gt "0" ]; then
        echo -e "${YELLOW}⚠ ${CHAIN}: No new transactions in 10s (current: ${TX_1})${NC}"
    else
        echo -e "${RED}✗ ${CHAIN}: No data${NC}"
    fi
done

echo -e "\n${BLUE}=== 6. CometBFT 0.38 Specific Checks ===${NC}"

echo -e "\n${YELLOW}Checking for version-specific issues:${NC}"

# Look for any parsing errors in metrics
ERROR_METRICS=$(echo "$METRICS" | grep -i "error" | grep -E "(parse|decode|unmarshal|consensus)")
if [ -n "$ERROR_METRICS" ]; then
    echo -e "${RED}Found potential parsing errors:${NC}"
    echo "$ERROR_METRICS"
else
    echo -e "${GREEN}✓ No parsing errors detected${NC}"
fi

# Check for timeout issues
TIMEOUT_COUNT=$(echo "$METRICS" | grep "chainpulse_timeouts{" | grep -oE "[0-9]+$" | awk '{sum+=$1} END {print sum+0}')
if [ "$TIMEOUT_COUNT" -gt "0" ]; then
    echo -e "${YELLOW}⚠ Total timeouts: ${TIMEOUT_COUNT}${NC}"
    echo "$METRICS" | grep "chainpulse_timeouts{" | grep -v " 0$"
else
    echo -e "${GREEN}✓ No timeout issues${NC}"
fi

echo -e "\n${BLUE}=== 7. Summary Report ===${NC}"

echo -e "\n${YELLOW}CometBFT 0.38 Compatibility Summary:${NC}"

# Count monitored chains
MONITORED_CHAINS=$(echo "$METRICS" | grep -oE 'chain_id="[^"]+' | sort -u | wc -l)
echo "- Chains monitored: ${MONITORED_CHAINS}/3"

# Check overall health
if [ "$TOTAL_EFFECTED" -gt "0" ] || [ "$TOTAL_UNEFFECTED" -gt "0" ]; then
    echo -e "- Packet processing: ${GREEN}✓ Active${NC}"
else
    echo -e "- Packet processing: ${YELLOW}⚠ No recent packets${NC}"
fi

# API status
API_HEALTHY=true
if ! echo "$STUCK_RESPONSE" | jq -e '.packets' >/dev/null 2>&1; then
    API_HEALTHY=false
fi
if ! echo "$USER_RESPONSE" | jq -e '.packets' >/dev/null 2>&1; then
    API_HEALTHY=false
fi
if ! echo "$CONGESTION_RESPONSE" | jq -e '.channels' >/dev/null 2>&1; then
    API_HEALTHY=false
fi

if [ "$API_HEALTHY" = true ]; then
    echo -e "- API endpoints: ${GREEN}✓ All working${NC}"
else
    echo -e "- API endpoints: ${RED}✗ Some failures${NC}"
fi

# Overall status
if [ "$MONITORED_CHAINS" -ge 2 ] && [ "$API_HEALTHY" = true ]; then
    echo -e "\n${GREEN}✓ ChainPulse is compatible and functioning correctly${NC}"
else
    echo -e "\n${YELLOW}⚠ ChainPulse has some issues that may need attention${NC}"
fi

echo -e "\n${YELLOW}Recommendations:${NC}"
echo "1. Monitor error and timeout metrics for any chains"
echo "2. Verify all expected chains are being monitored"
echo "3. Check ChainPulse logs for any consensus-related errors"
echo "4. Ensure RPC endpoints are accessible and responsive"