#!/bin/bash
set -e

echo "Initializing Hermes with authentication..."

# Create Hermes directory if it doesn't exist
mkdir -p /home/hermes/.hermes

# Copy config to writable location
cp /config/config.toml /home/hermes/.hermes/config.toml

# The credentials are already URL-encoded in the .env file
# We just need to ensure they're not double-encoded
if [ -n "$RPC_USERNAME" ] && [ -n "$RPC_PASSWORD" ]; then
    echo "Configuring RPC and WebSocket authentication for Hermes..."
    
    # For configs that don't have authentication yet, add it
    # But skip lines that already have authentication (contain @)
    
    # For RPC addresses (but not lines containing grpc_addr or already containing @)
    sed -i "/grpc_addr/!s|rpc_addr = 'https://\([^@']*\)'|rpc_addr = 'https://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    sed -i "/grpc_addr/!s|rpc_addr = 'http://\([^@']*\)'|rpc_addr = 'http://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    
    # For WebSocket URLs in event_source (but not already containing @)
    sed -i "s|url = 'wss://\([^@']*\)'|url = 'wss://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    sed -i "s|url = 'ws://\([^@']*\)'|url = 'ws://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /home/hermes/.hermes/config.toml
    
    echo "Authentication configured for RPC and WebSocket endpoints"
else
    echo "No RPC credentials provided, using config as-is"
fi

echo "Starting Hermes..."
exec hermes "$@"