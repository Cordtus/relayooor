#!/bin/bash
set -e

echo "Initializing Hermes with authentication..."

# Copy config to writable location
cp /config/config.toml /home/hermes/.hermes/config.toml

# Replace RPC and WebSocket URLs with authenticated versions if credentials are provided
# Note: gRPC does not require authentication
if [ -n "$RPC_USERNAME" ] && [ -n "$RPC_PASSWORD" ]; then
    echo "Configuring RPC and WebSocket authentication for Hermes..."
    
    # Handle RPC URLs (both http and https)
    sed -i "s|rpc_addr = 'https://\([^']*\)'|rpc_addr = 'https://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    sed -i "s|rpc_addr = 'http://\([^']*\)'|rpc_addr = 'http://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    
    # Handle WebSocket URLs
    sed -i "s|url = 'wss://\([^']*\)'|url = 'wss://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    sed -i "s|url = 'ws://\([^']*\)'|url = 'ws://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    
    echo "Authentication configured for RPC and WebSocket endpoints"
fi

echo "Starting Hermes..."
exec hermes "$@"