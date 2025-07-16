#!/bin/bash
# Script to update API with dynamic chain list from chainpulse

# Get list of chains from chainpulse metrics
CHAINS=$(curl -s http://localhost:3001/metrics | grep -E "^chainpulse_packets{" | sed -E 's/.*chain_id="([^"]+)".*/\1/' | sort -u)

echo "Found chains: $CHAINS"

# Create a dynamic chain configuration
cat > /tmp/chains.json << EOF
{
  "chains": [
EOF

first=true
for chain in $CHAINS; do
    if [ "$first" = true ]; then
        first=false
    else
        echo "," >> /tmp/chains.json
    fi
    
    # Get chain display name
    case $chain in
        "cosmoshub-4") name="Cosmos Hub" ;;
        "osmosis-1") name="Osmosis" ;;
        "neutron-1") name="Neutron" ;;
        "noble-1") name="Noble" ;;
        "akash-1") name="Akash" ;;
        "stargaze-1") name="Stargaze" ;;
        "juno-1") name="Juno" ;;
        "stride-1") name="Stride" ;;
        *) name="$chain" ;;
    esac
    
    # Get metrics for this chain
    packets=$(curl -s http://localhost:3001/metrics | grep "chainpulse_packets{chain_id=\"$chain\"}" | awk '{print $2}' || echo "0")
    txs=$(curl -s http://localhost:3001/metrics | grep "chainpulse_txs{chain_id=\"$chain\"}" | awk '{print $2}' || echo "0")
    errors=$(curl -s http://localhost:3001/metrics | grep "chainpulse_errors{chain_id=\"$chain\"}" | awk '{print $2}' || echo "0")
    
    cat >> /tmp/chains.json << EOF
    {
      "chainId": "$chain",
      "chainName": "$name",
      "totalTxs": ${txs:-0},
      "totalPackets": ${packets:-0},
      "errors": ${errors:-0},
      "status": "connected",
      "autoDetected": true
    }
EOF
done

echo "  ]" >> /tmp/chains.json
echo "}" >> /tmp/chains.json

echo "Generated chain config:"
cat /tmp/chains.json | jq .

# TODO: Send this to API endpoint to update chains dynamically