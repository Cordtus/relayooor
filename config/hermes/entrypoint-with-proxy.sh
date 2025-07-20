#!/bin/bash
set -e

echo "Initializing Hermes with authentication proxy..."

# Create Hermes directory if it doesn't exist
mkdir -p /home/hermes/.hermes

# Copy config to writable location
cp /config/config.toml /home/hermes/.hermes/config.toml

# If authentication is enabled, modify URLs to use local proxy
if [ -n "$RPC_USERNAME" ] && [ -n "$RPC_PASSWORD" ]; then
    echo "Starting authentication proxy..."
    
    # Install proxy dependencies
    cd /config
    npm install --silent
    
    # Start proxy in background
    export PROXY_PORT=8545
    node /config/auth-proxy.js &
    PROXY_PID=$!
    
    # Wait for proxy to start
    sleep 2
    
    echo "Modifying config to use authentication proxy..."
    
    # Replace RPC URLs to use proxy
    # Format: https://domain.com -> http://localhost:8545/https://domain.com
    sed -i "s|rpc_addr = 'https://\([^']*\)'|rpc_addr = 'http://localhost:8545/https://\1'|g" /home/hermes/.hermes/config.toml
    
    # Replace WebSocket URLs to use proxy
    # Format: wss://domain.com -> ws://localhost:8545/wss://domain.com
    sed -i "s|url = 'wss://\([^']*\)'|url = 'ws://localhost:8545/wss://\1'|g" /home/hermes/.hermes/config.toml
    
    echo "Authentication proxy configured"
fi

echo "Starting Hermes..."
exec hermes "$@"