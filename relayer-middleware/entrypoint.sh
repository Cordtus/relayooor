#!/bin/bash
set -e

echo "Initializing IBC Relayer Middleware..."

# Function to replace auth credentials in URLs
replace_auth_in_url() {
    local url=$1
    local username=$2
    local password=$3
    
    if [[ -n "$username" ]] && [[ -n "$password" ]]; then
        # Replace or add auth to URL
        if [[ "$url" =~ ^(https?://)(.*)$ ]]; then
            echo "${BASH_REMATCH[1]}${username}:${password}@${BASH_REMATCH[2]}"
        else
            echo "$url"
        fi
    else
        echo "$url"
    fi
}

# Setup Hermes configuration with auth if provided
if [ -f "/config/hermes/config.toml" ]; then
    cp /config/hermes/config.toml /root/.hermes/config.toml
    
    # Replace RPC URLs with authenticated versions if credentials are provided
    if [ -n "$RPC_USERNAME" ] && [ -n "$RPC_PASSWORD" ]; then
        echo "Configuring RPC authentication for Hermes..."
        # This is a simplified example - in production, use proper TOML parsing
        sed -i "s|rpc_addr = 'http://\([^']*\)'|rpc_addr = 'http://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /root/.hermes/config.toml
        sed -i "s|grpc_addr = 'http://\([^']*\)'|grpc_addr = 'http://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /root/.hermes/config.toml
    fi
fi

# Setup legacy Hermes configuration
if [ -f "/config/hermes-legacy/config.toml" ]; then
    # Create separate config directory for legacy version
    mkdir -p /root/.hermes-legacy
    cp /config/hermes-legacy/config.toml /root/.hermes-legacy/config.toml
    
    if [ -n "$RPC_USERNAME" ] && [ -n "$RPC_PASSWORD" ]; then
        sed -i "s|rpc_addr = 'http://\([^']*\)'|rpc_addr = 'http://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /root/.hermes-legacy/config.toml
        sed -i "s|grpc_addr = 'http://\([^']*\)'|grpc_addr = 'http://${RPC_USERNAME}:${RPC_PASSWORD}@\1'|g" /root/.hermes-legacy/config.toml
    fi
fi

# Setup Go relayer configuration with auth
if [ -f "/config/relayer/config.yaml" ]; then
    cp /config/relayer/config.yaml /root/.relayer/config/config.yaml
    
    if [ -n "$RPC_USERNAME" ] && [ -n "$RPC_PASSWORD" ]; then
        echo "Configuring RPC authentication for Go relayer..."
        # This would need proper YAML parsing in production
        # For now, this is a placeholder
    fi
fi

# Initialize relayer keys if provided
if [ -d "/config/keys" ]; then
    echo "Importing keys..."
    # Import keys for each relayer
    # This would be customized based on your key management strategy
fi

# Start specific relayer based on environment variable
case "$ACTIVE_RELAYER" in
    "hermes")
        echo "Starting Hermes relayer..."
        supervisorctl start hermes
        ;;
    "hermes-legacy")
        echo "Starting legacy Hermes relayer..."
        supervisorctl start hermes-legacy
        ;;
    "rly")
        echo "Starting Go relayer..."
        supervisorctl start rly
        ;;
    "rly-legacy")
        echo "Starting legacy Go relayer..."
        supervisorctl start rly-legacy
        ;;
    *)
        echo "No active relayer specified. Set ACTIVE_RELAYER environment variable."
        echo "Available options: hermes, hermes-legacy, rly, rly-legacy"
        ;;
esac

# Execute the main command
exec "$@"