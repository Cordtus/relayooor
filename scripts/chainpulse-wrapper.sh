#!/bin/bash
# Chainpulse wrapper script with authentication support

# Load environment variables
if [ -f /config/.env ]; then
    export $(cat /config/.env | grep -v '^#' | xargs)
fi

# Function to add authentication to WebSocket URLs
add_auth_to_url() {
    local url=$1
    local user=$2
    local pass=$3
    
    if [ -n "$user" ] && [ -n "$pass" ]; then
        # Replace wss:// with wss://user:pass@
        echo "$url" | sed "s|wss://|wss://${user}:${pass}@|"
    else
        echo "$url"
    fi
}

# Create a temporary config with substituted values
CONFIG_FILE="/config/chainpulse.toml"
TEMP_CONFIG="/tmp/chainpulse-processed.toml"

# Process the config file, substituting environment variables
envsubst < "$CONFIG_FILE" > "$TEMP_CONFIG"

# Additional processing for authentication headers
if [ -n "$COSMOS_API_TOKEN" ]; then
    # Add authentication headers if needed
    sed -i "s/\${COSMOS_API_TOKEN}/$COSMOS_API_TOKEN/g" "$TEMP_CONFIG"
fi

# Run chainpulse with the processed config
exec /app/chainpulse --config "$TEMP_CONFIG" "$@"