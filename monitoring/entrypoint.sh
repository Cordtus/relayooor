#!/bin/bash
set -e

# Function to replace placeholders in config with environment variables
process_config() {
    local config_file="/config/chainpulse.toml"
    local processed_file="/tmp/chainpulse.toml"
    
    if [ -f "$config_file" ]; then
        cp "$config_file" "$processed_file"
        
        # Replace WebSocket URLs with authenticated versions if credentials provided
        if [ -n "$RPC_USERNAME" ] && [ -n "$RPC_PASSWORD" ]; then
            echo "Configuring RPC authentication..."
            # Replace ws:// and wss:// URLs with authenticated versions
            sed -i "s|ws://\([^/]*\)|ws://${RPC_USERNAME}:${RPC_PASSWORD}@\1|g" "$processed_file"
            sed -i "s|wss://\([^/]*\)|wss://${RPC_USERNAME}:${RPC_PASSWORD}@\1|g" "$processed_file"
        fi
        
        # Replace any environment variable placeholders
        while IFS='=' read -r name value; do
            if [[ $name =~ ^CHAIN_ ]]; then
                sed -i "s|\${$name}|$value|g" "$processed_file"
            fi
        done < <(env)
        
        # Use the processed config
        exec chainpulse --config "$processed_file"
    else
        echo "Error: Config file not found at /config/chainpulse.toml"
        exit 1
    fi
}

# If running chainpulse command, process the config
if [ "$1" = "chainpulse" ]; then
    process_config
else
    # Otherwise, run the command as-is
    exec "$@"
fi