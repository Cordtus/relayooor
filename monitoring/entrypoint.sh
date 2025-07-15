#!/bin/bash
set -e

# Function to replace placeholders in config with environment variables
process_config() {
    local config_file="/config/chainpulse.toml"
    local processed_file="/tmp/chainpulse.toml"
    
    if [ -f "$config_file" ]; then
        cp "$config_file" "$processed_file"
        
        # Replace authentication placeholders
        if [ -n "$RPC_USERNAME" ] && [ -n "$RPC_PASSWORD" ]; then
            echo "Configuring RPC authentication..."
            # Replace the placeholders in the config
            sed -i "s|\${RPC_USERNAME}|$RPC_USERNAME|g" "$processed_file"
            sed -i "s|\${RPC_PASSWORD}|$RPC_PASSWORD|g" "$processed_file"
        else
            echo "Warning: RPC authentication credentials not provided"
            # Remove the authentication lines if no credentials
            sed -i '/username = "\${RPC_USERNAME}"/d' "$processed_file"
            sed -i '/password = "\${RPC_PASSWORD}"/d' "$processed_file"
        fi
        
        # Replace WebSocket URL placeholders (strip auth since we use username/password fields)
        if [ -n "$COSMOS_WS_URL" ]; then
            # Remove any authentication from the URL
            COSMOS_WS_CLEAN=$(echo "$COSMOS_WS_URL" | sed 's|://[^@]*@|://|')
            sed -i "s|\${COSMOS_WS_URL}|$COSMOS_WS_CLEAN|g" "$processed_file"
        fi
        if [ -n "$OSMOSIS_WS_URL" ]; then
            OSMOSIS_WS_CLEAN=$(echo "$OSMOSIS_WS_URL" | sed 's|://[^@]*@|://|')
            sed -i "s|\${OSMOSIS_WS_URL}|$OSMOSIS_WS_CLEAN|g" "$processed_file"
        fi
        if [ -n "$NEUTRON_WS_URL" ]; then
            NEUTRON_WS_CLEAN=$(echo "$NEUTRON_WS_URL" | sed 's|://[^@]*@|://|')
            sed -i "s|\${NEUTRON_WS_URL}|$NEUTRON_WS_CLEAN|g" "$processed_file"
        fi
        if [ -n "$NOBLE_WS_URL" ]; then
            NOBLE_WS_CLEAN=$(echo "$NOBLE_WS_URL" | sed 's|://[^@]*@|://|')
            sed -i "s|\${NOBLE_WS_URL}|$NOBLE_WS_CLEAN|g" "$processed_file"
        fi
        
        # Replace any other environment variable placeholders
        while IFS='=' read -r name value; do
            if [[ $name =~ ^CHAIN_ ]]; then
                sed -i "s|\${$name}|$value|g" "$processed_file"
            fi
        done < <(env)
        
        # Debug: Show the processed config
        echo "Processed configuration:"
        cat "$processed_file"
        
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