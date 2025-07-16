#!/bin/bash

echo "Packet Clearing Use Case Tests"
echo "=============================="

CHAINPULSE_URL="http://localhost:3001"
API_BACKEND_URL="http://localhost:8080"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test addresses for different chains
OSMO_ADDRESS="osmo1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep88n0y"
COSMOS_ADDRESS="cosmos1clpqr4nrk4khgkxj78fcwwh6dl3uw4ep6dqkuf"
NEUTRON_ADDRESS="neutron1clpqr4nrk4khgkxj78fcwwh6dl3uw4epz87gc"

echo -e "${BLUE}=== 1. User Packet Discovery ===${NC}"

# Test finding packets for users on different chains
for CHAIN_ADDR in "$OSMO_ADDRESS:Osmosis" "$COSMOS_ADDRESS:Cosmos Hub" "$NEUTRON_ADDRESS:Neutron"; do
    IFS=':' read -r ADDR CHAIN_NAME <<< "$CHAIN_ADDR"
    echo -e "\n${YELLOW}Testing packets for ${CHAIN_NAME} user: ${ADDR}${NC}"
    
    # As sender
    SENDER_PACKETS=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/by-user?address=${ADDR}&role=sender")
    if echo "$SENDER_PACKETS" | jq -e '.packets | length' >/dev/null 2>&1; then
        COUNT=$(echo "$SENDER_PACKETS" | jq -r '.packets | length')
        echo -e "${GREEN}✓ Found ${COUNT} packets as sender${NC}"
        if [ "$COUNT" -gt 0 ]; then
            echo "$SENDER_PACKETS" | jq -r '.packets[0] | "  First packet: Chain=\(.chain_id), Channel=\(.src_channel), Seq=\(.sequence), Status=\(.status // "unknown")"'
        fi
    else
        echo -e "  No packets found as sender"
    fi
    
    # As receiver
    RECEIVER_PACKETS=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/by-user?address=${ADDR}&role=receiver")
    if echo "$RECEIVER_PACKETS" | jq -e '.packets | length' >/dev/null 2>&1; then
        COUNT=$(echo "$RECEIVER_PACKETS" | jq -r '.packets | length')
        echo -e "${GREEN}✓ Found ${COUNT} packets as receiver${NC}"
    else
        echo -e "  No packets found as receiver"
    fi
done

echo -e "\n${BLUE}=== 2. Stuck Packet Analysis ===${NC}"

# Get stuck packets with different age thresholds
echo -e "\n${YELLOW}Analyzing stuck packets by age:${NC}"
declare -A STUCK_BY_AGE
for MINUTES in 5 15 30 60 120; do
    SECONDS=$((MINUTES * 60))
    RESPONSE=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/stuck?min_age_seconds=${SECONDS}")
    
    if echo "$RESPONSE" | jq -e '.packets' >/dev/null 2>&1; then
        COUNT=$(echo "$RESPONSE" | jq -r '.packets | length')
        STUCK_BY_AGE[$MINUTES]=$COUNT
        echo "Stuck > ${MINUTES} minutes: ${COUNT} packets"
        
        # Show details for the oldest threshold
        if [ "$MINUTES" -eq 5 ] && [ "$COUNT" -gt 0 ]; then
            echo -e "\n${YELLOW}Sample stuck packets:${NC}"
            echo "$RESPONSE" | jq -r '.packets[:3][] | "  Chain: \(.chain_id), Channel: \(.src_channel), Seq: \(.sequence), Age: \(.age_seconds)s, Amount: \(.amount // "unknown") \(.denom // "")"'
        fi
    fi
done

# Check for patterns in stuck packets
echo -e "\n${YELLOW}Stuck packet patterns:${NC}"
STUCK_RESPONSE=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/stuck?min_age_seconds=300")
if echo "$STUCK_RESPONSE" | jq -e '.packets' >/dev/null 2>&1; then
    # Group by channel
    echo -e "\nBy channel:"
    echo "$STUCK_RESPONSE" | jq -r '.packets | group_by(.src_channel) | .[] | "\(.[0].src_channel): \(length) packets"'
    
    # Group by chain
    echo -e "\nBy chain:"
    echo "$STUCK_RESPONSE" | jq -r '.packets | group_by(.chain_id) | .[] | "\(.[0].chain_id): \(length) packets"'
fi

echo -e "\n${BLUE}=== 3. Channel Congestion Impact ===${NC}"

CONGESTION=$(curl -s "${CHAINPULSE_URL}/api/v1/channels/congestion")
if echo "$CONGESTION" | jq -e '.channels' >/dev/null 2>&1; then
    echo -e "\n${YELLOW}Congested channels:${NC}"
    echo "$CONGESTION" | jq -r '.channels[] | select(.stuck_count > 0) | "Channel: \(.src_channel) -> \(.dst_channel), Stuck: \(.stuck_count), Oldest: \(.oldest_stuck_age_seconds // 0)s"' | head -10
    
    # Calculate clearing priorities
    echo -e "\n${YELLOW}Clearing priority (by value):${NC}"
    echo "$CONGESTION" | jq -r '.channels[] | select(.stuck_count > 0) | . as $ch | .total_value | to_entries[] | "\($ch.src_channel): \(.value) \(.key)"' | sort -rn -k2 | head -10
else
    echo "No congestion data available"
fi

echo -e "\n${BLUE}=== 4. Packet Clearing Simulation ===${NC}"

# Find clearable packets
echo -e "\n${YELLOW}Identifying clearable packets:${NC}"
CLEARABLE=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/stuck?min_age_seconds=1800&limit=10")

if echo "$CLEARABLE" | jq -e '.packets | length' >/dev/null 2>&1; then
    PACKET_COUNT=$(echo "$CLEARABLE" | jq -r '.packets | length')
    echo "Found ${PACKET_COUNT} packets stuck for > 30 minutes"
    
    # Calculate clearing costs
    if [ "$PACKET_COUNT" -gt 0 ]; then
        BASE_GAS=100000
        PER_PACKET_GAS=10000
        TOTAL_GAS=$((BASE_GAS + (PACKET_COUNT * PER_PACKET_GAS)))
        
        echo -e "\n${YELLOW}Clearing cost estimation:${NC}"
        echo "Base gas: ${BASE_GAS}"
        echo "Per packet gas: ${PER_PACKET_GAS}"
        echo "Total packets: ${PACKET_COUNT}"
        echo "Total gas needed: ${TOTAL_GAS}"
        
        # Show packet IDs for clearing
        echo -e "\n${YELLOW}Packet identifiers for clearing:${NC}"
        echo "$CLEARABLE" | jq -r '.packets[] | {chain: .chain_id, channel: .src_channel, sequence: .sequence}' | jq -s '.'
    fi
else
    echo "No clearable packets found"
fi

echo -e "\n${BLUE}=== 5. Performance Metrics ===${NC}"

# Get overall metrics
METRICS=$(curl -s ${CHAINPULSE_URL}/metrics)

echo -e "\n${YELLOW}Relayer performance summary:${NC}"
# Calculate success rates by relayer
echo "$METRICS" | grep -E "ibc_(effected|uneffected)_packets" | \
    awk -F'[{}"]' '
    /effected/ {effected[$8]+=$NF}
    /uneffected/ {uneffected[$8]+=$NF}
    END {
        for (signer in effected) {
            total = effected[signer] + uneffected[signer]
            if (total > 0) {
                rate = (effected[signer] / total) * 100
                printf "%-60s Success: %.1f%% (%d/%d)\n", 
                    substr(signer, 1, 60), rate, effected[signer], total
            }
        }
    }' | sort -k3 -rn | head -10

echo -e "\n${BLUE}=== 6. Integration Test Results ===${NC}"

# Test the full integration flow
echo -e "\n${YELLOW}Testing full clearing flow integration:${NC}"

# 1. Check if API backend is running
if curl -s ${API_BACKEND_URL}/health >/dev/null 2>&1; then
    echo -e "${GREEN}✓ API backend is healthy${NC}"
    
    # 2. Test user packet retrieval through backend
    USER_TRANSFERS=$(curl -s "${API_BACKEND_URL}/api/user/${OSMO_ADDRESS}/transfers")
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ User transfers endpoint working${NC}"
        if echo "$USER_TRANSFERS" | jq -e '.[0]' >/dev/null 2>&1; then
            echo "  Found transfers for user"
        fi
    fi
    
    # 3. Test stuck packet endpoint
    BACKEND_STUCK=$(curl -s "${API_BACKEND_URL}/api/packets/stuck")
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Stuck packets endpoint working${NC}"
    fi
else
    echo -e "${RED}✗ API backend not available${NC}"
fi

echo -e "\n${BLUE}=== 7. Data Quality Assessment ===${NC}"

echo -e "\n${YELLOW}Checking data completeness:${NC}"

# Check for missing data
echo "$METRICS" | awk '
/chainpulse_errors{/ {errors[$0]++}
/chainpulse_timeouts{/ {timeouts[$0]++}
END {
    if (length(errors) > 0) print "Found " length(errors) " error metrics"
    if (length(timeouts) > 0) print "Found " length(timeouts) " timeout metrics"
}'

# Verify expected chains are present
echo -e "\n${YELLOW}Chain coverage:${NC}"
for CHAIN in "cosmoshub-4" "neutron-1" "osmosis-1"; do
    if echo "$METRICS" | grep -q "chain_id=\"${CHAIN}\""; then
        echo -e "${GREEN}✓ ${CHAIN} is being monitored${NC}"
    else
        echo -e "${RED}✗ ${CHAIN} missing from metrics${NC}"
    fi
done

echo -e "\n${GREEN}Packet clearing scenario tests complete!${NC}"

# Summary
echo -e "\n${BLUE}=== Summary ===${NC}"
echo "1. User packet discovery: Tested for 3 chains"
echo "2. Stuck packets: Analyzed by age and patterns"
echo "3. Channel congestion: Identified high-value clearing opportunities"
echo "4. Clearing simulation: Calculated costs for clearable packets"
echo "5. Performance metrics: Analyzed relayer success rates"
echo "6. Integration: Verified API backend connectivity"
echo "7. Data quality: Checked completeness and coverage"