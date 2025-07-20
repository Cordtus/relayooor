#!/bin/bash
set -e

echo "Initializing Hermes with authentication..."

# Create Hermes directory if it doesn't exist
mkdir -p /home/hermes/.hermes

# Copy config to writable location
cp /config/config.toml /home/hermes/.hermes/config.toml

# Replace RPC and WebSocket URLs with authenticated versions if credentials are provided
# Note: gRPC does not require authentication
if [ -n "$RPC_USERNAME" ] && [ -n "$RPC_PASSWORD" ]; then
    echo "Configuring RPC and WebSocket authentication for Hermes..."
    
    # URL-encode function for special characters
    url_encode() {
        local string="${1}"
        local strlen=${#string}
        local encoded=""
        local pos c o
        
        for (( pos=0 ; pos<strlen ; pos++ )); do
            c=${string:$pos:1}
            case "$c" in
                [-_.~a-zA-Z0-9] ) o="${c}" ;;
                * ) printf -v o '%%%02x' "'$c" ;;
            esac
            encoded+="${o}"
        done
        echo "${encoded}"
    }
    
    # URL-encode the credentials
    ENCODED_USERNAME=$(url_encode "$RPC_USERNAME")
    ENCODED_PASSWORD=$(url_encode "$RPC_PASSWORD")
    
    echo "Credentials URL-encoded for use in configuration"
    
    # Now escape for sed (much simpler with URL-encoded strings)
    ESCAPED_USERNAME="$ENCODED_USERNAME"
    ESCAPED_PASSWORD="$ENCODED_PASSWORD"
    
    # Use sed to add authentication to URLs - but only for rpc_addr and WebSocket URLs, NOT grpc_addr
    # For RPC addresses (but not lines containing grpc_addr)
    # Match everything in the URL after the protocol
    sed -i "/grpc_addr/!s|rpc_addr = 'https://\([^@']*\)'|rpc_addr = 'https://${ESCAPED_USERNAME}:${ESCAPED_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    sed -i "/grpc_addr/!s|rpc_addr = 'http://\([^@']*\)'|rpc_addr = 'http://${ESCAPED_USERNAME}:${ESCAPED_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    
    # For WebSocket URLs in event_source
    sed -i "s|url = 'wss://\([^@']*\)'|url = 'wss://${ESCAPED_USERNAME}:${ESCAPED_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    sed -i "s|url = 'ws://\([^@']*\)'|url = 'ws://${ESCAPED_USERNAME}:${ESCAPED_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    
    echo "Authentication configured for RPC and WebSocket endpoints"
fi

echo "Starting Hermes..."
exec hermes "$@"