#!/bin/bash
# Script to check chain compatibility with chainpulse
# Identifies chains that might have custom protobuf implementations

echo "=== Chain Compatibility Checker ==="
echo ""
echo "This script helps identify chains that might have compatibility issues with chainpulse"
echo "due to custom protobuf implementations, vote extensions, or other ABCI++ features."
echo ""

# Function to check a chain
check_chain() {
    local chain_name=$1
    local rpc_url=$2
    local auth_user=$3
    local auth_pass=$4
    
    echo "Checking $chain_name..."
    
    # Get chain info
    if [ -n "$auth_user" ] && [ -n "$auth_pass" ]; then
        STATUS=$(curl -s -u "$auth_user:$auth_pass" "$rpc_url/status" 2>/dev/null)
        ABCI=$(curl -s -u "$auth_user:$auth_pass" "$rpc_url/abci_info" 2>/dev/null)
    else
        STATUS=$(curl -s "$rpc_url/status" 2>/dev/null)
        ABCI=$(curl -s "$rpc_url/abci_info" 2>/dev/null)
    fi
    
    # Parse results
    if [ -n "$STATUS" ]; then
        VERSION=$(echo "$STATUS" | jq -r '.result.node_info.version // "unknown"')
        PROTOCOL=$(echo "$STATUS" | jq -r '.result.node_info.protocol_version.block // "unknown"')
        APP_VERSION=$(echo "$ABCI" | jq -r '.result.response.version // "unknown"')
        
        echo "  Version: $VERSION"
        echo "  Protocol Block Version: $PROTOCOL"
        echo "  App Version: $APP_VERSION"
        
        # Check for known compatibility issues
        if [[ "$PROTOCOL" == "11" ]] && [[ "$VERSION" =~ ^0\.38 ]]; then
            echo "  ⚠️  WARNING: Uses protocol version 11 with CometBFT 0.38 (may have vote extensions)"
        fi
        
        if [[ "$APP_VERSION" =~ "slinky" ]] || [[ "$APP_VERSION" =~ "oracle" ]]; then
            echo "  ⚠️  WARNING: Detected oracle implementation (likely uses vote extensions)"
        fi
        
        # Check if it's a known chain with custom features
        case "$chain_name" in
            "Neutron"|"neutron-1")
                echo "  ⚠️  KNOWN ISSUE: Uses Slinky oracle with vote extensions"
                ;;
            "dYdX"|"dydx-mainnet-1")
                echo "  ⚠️  POTENTIAL ISSUE: Custom implementation, may have compatibility issues"
                ;;
            "Injective"|"injective-1")
                echo "  ⚠️  POTENTIAL ISSUE: Custom Tendermint fork, may require special handling"
                ;;
        esac
    else
        echo "  ❌ Failed to connect to RPC"
    fi
    
    echo ""
}

# Check common chains
echo "=== Checking Common Chains ==="
echo ""

# Chains we're already monitoring
check_chain "Cosmos Hub" "https://cosmoshub-4-skip-rpc.polkachu.com" "skip" "p01kachu?!"
check_chain "Osmosis" "https://osmosis-1-skip-rpc.polkachu.com" "skip" "p01kachu?!"
check_chain "Neutron" "https://neutron-1-skip-rpc.polkachu.com" "skip" "p01kachu?!"
check_chain "Noble" "https://noble-1-skip-rpc.polkachu.com" "skip" "p01kachu?!"

# Additional chains to check
echo "=== Checking Additional Chains ==="
echo ""
check_chain "Akash" "https://akash-rpc.polkachu.com" "" ""
check_chain "Stride" "https://stride-rpc.polkachu.com" "" ""
check_chain "Stargaze" "https://stargaze-rpc.polkachu.com" "" ""
check_chain "Juno" "https://juno-rpc.polkachu.com" "" ""

echo "=== Recommendations ==="
echo ""
echo "1. Chains with protocol version 11 and CometBFT 0.38+ may use vote extensions"
echo "2. Chains with oracle implementations (Slinky, etc.) likely use vote extensions"
echo "3. Custom chain implementations may require special handling in chainpulse"
echo "4. Test each chain thoroughly before adding to production monitoring"
echo ""
echo "For chains with known issues:"
echo "- Document the specific error in chain-integration-troubleshooting.md"
echo "- Consider opening an issue with the chainpulse repository"
echo "- Work with the chain's development team for RPC compatibility"