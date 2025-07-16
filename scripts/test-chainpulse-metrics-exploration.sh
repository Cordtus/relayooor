#!/bin/bash

echo "ChainPulse Metrics Deep Dive"
echo "============================"

CHAINPULSE_URL="http://localhost:3000"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Fetch all metrics
echo -e "${YELLOW}Fetching all metrics from ChainPulse...${NC}"
METRICS=$(curl -s ${CHAINPULSE_URL}/metrics)

if [ -z "$METRICS" ]; then
    echo -e "${RED}Failed to fetch metrics!${NC}"
    exit 1
fi

# Save metrics to file for analysis
echo "$METRICS" > chainpulse-metrics-snapshot.txt
echo -e "${GREEN}✓ Metrics saved to chainpulse-metrics-snapshot.txt${NC}"

# 1. Chain Monitoring Status
echo -e "\n${BLUE}=== 1. Chain Monitoring Status ===${NC}"
echo -e "\n${YELLOW}Active chains being monitored:${NC}"
echo "$METRICS" | grep "chainpulse_chains" | grep -v "^#"

echo -e "\n${YELLOW}Transactions processed per chain:${NC}"
echo "$METRICS" | grep "chainpulse_txs{" | sort

echo -e "\n${YELLOW}Packets processed per chain:${NC}"
echo "$METRICS" | grep "chainpulse_packets{" | sort

echo -e "\n${YELLOW}Reconnection attempts per chain:${NC}"
echo "$METRICS" | grep "chainpulse_reconnects{" | sort

echo -e "\n${YELLOW}Errors per chain:${NC}"
echo "$METRICS" | grep "chainpulse_errors{" | sort

# 2. IBC Packet Metrics
echo -e "\n${BLUE}=== 2. IBC Packet Metrics ===${NC}"

echo -e "\n${YELLOW}Effected (successful) packets:${NC}"
echo "$METRICS" | grep "ibc_effected_packets{" | head -10

echo -e "\n${YELLOW}Uneffected (failed) packets:${NC}"
echo "$METRICS" | grep "ibc_uneffected_packets{" | head -10

echo -e "\n${YELLOW}Stuck packets by channel:${NC}"
echo "$METRICS" | grep "ibc_stuck_packets{" | grep -v " 0$" || echo "No stuck packets found"

echo -e "\n${YELLOW}Frontrun events:${NC}"
echo "$METRICS" | grep "ibc_frontrun_counter{" | head -5 || echo "No frontrun events found"

# 3. Channel Analysis
echo -e "\n${BLUE}=== 3. Channel Analysis ===${NC}"

echo -e "\n${YELLOW}Extracting unique channels...${NC}"
CHANNELS=$(echo "$METRICS" | grep -E "ibc_(effected|uneffected)_packets" | \
    sed -n 's/.*src_channel="\([^"]*\)".*dst_channel="\([^"]*\)".*/\1 -> \2/p' | \
    sort | uniq)

echo "Active channels:"
echo "$CHANNELS" | head -20

# 4. Signer/Relayer Analysis
echo -e "\n${BLUE}=== 4. Relayer Performance ===${NC}"

echo -e "\n${YELLOW}Top relayers by effected packets:${NC}"
echo "$METRICS" | grep "ibc_effected_packets{" | \
    sed -n 's/.*signer="\([^"]*\)".*} \([0-9]*\)/\2 \1/p' | \
    sort -rn | head -10 | \
    awk '{printf "%-60s %s packets\n", $2, $1}'

echo -e "\n${YELLOW}Relayers with failed packets:${NC}"
echo "$METRICS" | grep "ibc_uneffected_packets{" | \
    grep -v " 0$" | \
    sed -n 's/.*signer="\([^"]*\)".*} \([0-9]*\)/\2 \1/p' | \
    sort -rn | head -10 | \
    awk '{printf "%-60s %s failed\n", $2, $1}'

# 5. Port Analysis
echo -e "\n${BLUE}=== 5. IBC Port Usage ===${NC}"

echo -e "\n${YELLOW}Active ports:${NC}"
echo "$METRICS" | grep -E "ibc_(effected|uneffected)_packets" | \
    sed -n 's/.*src_port="\([^"]*\)".*/\1/p' | \
    sort | uniq -c | sort -rn

# 6. Memo Field Analysis
echo -e "\n${BLUE}=== 6. Memo Field Analysis ===${NC}"

echo -e "\n${YELLOW}Unique memo fields (relayer identification):${NC}"
echo "$METRICS" | grep -E "ibc_(effected|uneffected)_packets" | \
    sed -n 's/.*memo="\([^"]*\)".*/\1/p' | \
    sort | uniq -c | sort -rn | head -10

# 7. API Endpoint Testing
echo -e "\n${BLUE}=== 7. API Endpoint Data Quality ===${NC}"

echo -e "\n${YELLOW}Testing stuck packets API with different time windows:${NC}"
for MINUTES in 5 15 30 60; do
    SECONDS=$((MINUTES * 60))
    COUNT=$(curl -s "${CHAINPULSE_URL}/api/v1/packets/stuck?min_age_seconds=${SECONDS}" | \
        jq -r '.packets | length' 2>/dev/null || echo "0")
    echo "Packets stuck for more than ${MINUTES} minutes: ${COUNT}"
done

echo -e "\n${YELLOW}Testing channel congestion data:${NC}"
CONGESTION=$(curl -s "${CHAINPULSE_URL}/api/v1/channels/congestion")
if echo "$CONGESTION" | jq '.' >/dev/null 2>&1; then
    echo "$CONGESTION" | jq -r '.channels[] | "\(.src_channel) -> \(.dst_channel): \(.stuck_count) stuck, oldest: \(.oldest_stuck_age_seconds // 0)s"' 2>/dev/null || echo "No congestion data"
else
    echo "No valid congestion data returned"
fi

# 8. Data Consistency Check
echo -e "\n${BLUE}=== 8. Data Consistency Checks ===${NC}"

echo -e "\n${YELLOW}Comparing metric totals:${NC}"
TOTAL_EFFECTED=$(echo "$METRICS" | grep "ibc_effected_packets{" | \
    grep -oE "[0-9]+$" | awk '{sum+=$1} END {print sum}')
TOTAL_UNEFFECTED=$(echo "$METRICS" | grep "ibc_uneffected_packets{" | \
    grep -oE "[0-9]+$" | awk '{sum+=$1} END {print sum}')

echo "Total effected packets: ${TOTAL_EFFECTED:-0}"
echo "Total uneffected packets: ${TOTAL_UNEFFECTED:-0}"
if [ -n "$TOTAL_EFFECTED" ] && [ -n "$TOTAL_UNEFFECTED" ] && [ "$TOTAL_EFFECTED" -gt 0 ]; then
    SUCCESS_RATE=$(echo "scale=2; $TOTAL_EFFECTED * 100 / ($TOTAL_EFFECTED + $TOTAL_UNEFFECTED)" | bc)
    echo "Overall success rate: ${SUCCESS_RATE}%"
fi

# 9. CometBFT 0.38 Specific Checks
echo -e "\n${BLUE}=== 9. CometBFT 0.38 Compatibility ===${NC}"

echo -e "\n${YELLOW}Checking for v0.38 specific metrics or errors:${NC}"
# Look for any errors related to consensus or block parsing
echo "$METRICS" | grep -i "error" | grep -E "(consensus|block|height)" || echo "No consensus-related errors found"

# Check if all chains are reporting data
echo -e "\n${YELLOW}Chain data freshness:${NC}"
for CHAIN in "cosmoshub-4" "neutron-1" "osmosis-1"; do
    LAST_TX=$(echo "$METRICS" | grep "chainpulse_txs{chain_id=\"${CHAIN}\"" | grep -oE "[0-9]+$")
    LAST_PACKET=$(echo "$METRICS" | grep "chainpulse_packets{chain_id=\"${CHAIN}\"" | grep -oE "[0-9]+$")
    echo "${CHAIN}: Txs=${LAST_TX:-0}, Packets=${LAST_PACKET:-0}"
done

# 10. Summary Report
echo -e "\n${BLUE}=== 10. Summary Report ===${NC}"

echo -e "\n${GREEN}ChainPulse Metrics Analysis Complete!${NC}"
echo -e "\nKey Findings:"
echo "- Monitoring $(echo "$METRICS" | grep -c "chain_id=") chains"
echo "- Tracking $(echo "$CHANNELS" | wc -l) active IBC channels"
echo "- $(echo "$METRICS" | grep "ibc_effected_packets{" | grep -c "signer=") unique relayers detected"
echo "- Metrics snapshot saved to: chainpulse-metrics-snapshot.txt"

echo -e "\n${YELLOW}Recommended next steps:${NC}"
echo "1. Review chainpulse-metrics-snapshot.txt for detailed metrics"
echo "2. Check for any stuck packets that need clearing"
echo "3. Monitor relayer performance metrics"
echo "4. Verify all expected chains are reporting data"