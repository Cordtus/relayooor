#!/bin/bash

echo "=== Testing Enriched Packet System ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. Test basic search endpoint
echo -e "${YELLOW}1. Testing basic packet search...${NC}"
curl -s "http://localhost:3000/api/packets/search?limit=2" | jq '.'
echo ""

# 2. Search by wallet address
echo -e "${YELLOW}2. Searching packets by wallet address...${NC}"
WALLET="osmo1dxl3f4fh07de8qcdl6fqe64fraplst8zpx2807"
echo "Searching for wallet: $WALLET"
curl -s "http://localhost:3000/api/packets/search?sender=$WALLET&limit=5" | jq '.'
echo ""

# 3. Search by chain
echo -e "${YELLOW}3. Searching packets by chain...${NC}"
echo "Searching for Osmosis packets..."
curl -s "http://localhost:3000/api/packets/search?chain_id=osmosis-1&limit=3" | jq '.'
echo ""

# 4. Get first stuck packet for enrichment demo
echo -e "${YELLOW}4. Getting a stuck packet for enrichment demo...${NC}"
STUCK_PACKET=$(curl -s "http://localhost:3000/api/packets/stuck?limit=1" | jq -r '.[0]')
if [ "$STUCK_PACKET" != "null" ]; then
    PACKET_ID=$(echo "$STUCK_PACKET" | jq -r '.id')
    echo -e "${GREEN}Found stuck packet: $PACKET_ID${NC}"
    echo "$STUCK_PACKET" | jq '.'
    
    # 5. Demonstrate enrichment data structure
    echo ""
    echo -e "${YELLOW}5. Enrichment would add the following data:${NC}"
    
    CHAIN_ID=$(echo "$STUCK_PACKET" | jq -r '.sourceChain')
    CHANNEL_ID=$(echo "$STUCK_PACKET" | jq -r '.channelId')
    DENOM=$(echo "$STUCK_PACKET" | jq -r '.denom')
    AMOUNT=$(echo "$STUCK_PACKET" | jq -r '.amount')
    
    cat <<EOF | jq '.'
{
  "original_packet": {
    "id": "$PACKET_ID",
    "chain": "$CHAIN_ID",
    "channel": "$CHANNEL_ID"
  },
  "enriched_data": {
    "chain_info": {
      "source": {
        "chain_name": "$([ "$CHAIN_ID" == "osmosis-1" ] && echo "Osmosis" || echo "$CHAIN_ID")",
        "rpc_endpoint": "https://rpc.osmosis.zone",
        "current_height": "~40183650"
      },
      "destination": {
        "chain_name": "Cosmos Hub",
        "rpc_endpoint": "https://rpc.cosmos.quokkastake.io",
        "hermes_connected": true
      }
    },
    "token_info": {
      "denom": "$DENOM",
      "symbol": "$(echo "$DENOM" | grep -o 'uatom\|uosmo\|uusdc' | sed 's/uatom/ATOM/;s/uosmo/OSMO/;s/uusdc/USDC/' || echo "UNKNOWN")",
      "decimals": 6,
      "is_ibc_token": $(echo "$DENOM" | grep -q "transfer/" && echo "true" || echo "false"),
      "amount_human": "$(echo "scale=6; $AMOUNT / 1000000" | bc) tokens"
    },
    "clearing_info": {
      "can_clear": true,
      "estimated_gas": 150000,
      "estimated_fee_uatom": "3750",
      "hermes_status": "$([ "$CHAIN_ID" == "cosmoshub-4" ] && echo "READY" || echo "NOT_CONFIGURED")",
      "clearing_command": "hermes tx packet-recv --dst-chain cosmoshub-4 --src-chain $CHAIN_ID --src-port transfer --src-channel $CHANNEL_ID"
    },
    "routing_info": {
      "source_channel": "$CHANNEL_ID",
      "destination_channel": "$([ "$CHANNEL_ID" == "channel-0" ] && echo "channel-141" || echo "unknown")",
      "channel_state": "OPEN",
      "known_route": $([ "$CHANNEL_ID" == "channel-0" ] && echo "true" || echo "false")
    }
  }
}
EOF
    
else
    echo -e "${BLUE}No stuck packets found${NC}"
fi

# 6. Demonstrate combined search capabilities
echo ""
echo -e "${YELLOW}6. Demonstrating combined search capabilities...${NC}"
echo "We can search by multiple criteria:"
echo ""
echo "a) By sender AND chain:"
curl -s "http://localhost:3000/api/packets/search?sender=cosmos17zvwa39gecl7x4mfkpz0aarnhz2npme930y52z&chain_id=osmosis-1" | jq '. | {found: .total, example: .packets[0]}'
echo ""

echo "b) By token type (ATOM transfers):"
curl -s "http://localhost:3000/api/packets/search?denom=uatom&limit=3" | jq '. | {found: .total, packets: [.packets[] | {id, amount, denom, source: .sourceChain}]}'
echo ""

echo "c) By age (packets stuck > 1 hour):"
curl -s "http://localhost:3000/api/packets/search?min_age_seconds=3600&limit=3" | jq '. | {found: .total, packets: [.packets[] | {id, stuck_duration: .stuckDuration, chain: .sourceChain}]}'

echo ""
echo -e "${GREEN}=== Enrichment Demo Complete ===${NC}"
echo ""
echo "Summary of capabilities:"
echo "1. Search by wallet (sender OR receiver)"
echo "2. Search by chain (source OR destination)"
echo "3. Search by channel"
echo "4. Search by token/denom"
echo "5. Search by stuck duration"
echo "6. Combine any search criteria"
echo "7. Enrich with chain info, token details, clearing requirements"
echo "8. Provide actionable clearing instructions"