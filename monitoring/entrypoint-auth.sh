#!/bin/bash
set -e

# Function to process URL and extract components
process_url() {
    local url="$1"
    # Remove the protocol prefix to parse the URL
    local without_protocol="${url#*://}"
    
    # Check if URL contains authentication
    if [[ "$without_protocol" == *"@"* ]]; then
        # Extract auth and host parts
        local auth_part="${without_protocol%%@*}"
        local host_part="${without_protocol#*@}"
        local protocol="${url%%://*}"
        
        # Return clean URL without auth
        echo "${protocol}://${host_part}"
    else
        # Return URL as-is
        echo "$url"
    fi
}

# Function to replace placeholders in config with environment variables
process_config() {
    local config_file="/config/chainpulse.toml"
    local processed_file="/tmp/chainpulse.toml"
    
    if [ -f "$config_file" ]; then
        cp "$config_file" "$processed_file"
        
        # Process WebSocket URLs - remove embedded authentication
        if [ -n "$COSMOS_WS_URL" ]; then
            COSMOS_WS_CLEAN=$(process_url "$COSMOS_WS_URL")
            sed -i "s|\${COSMOS_WS_URL}|$COSMOS_WS_CLEAN|g" "$processed_file"
        fi
        if [ -n "$OSMOSIS_WS_URL" ]; then
            OSMOSIS_WS_CLEAN=$(process_url "$OSMOSIS_WS_URL")
            sed -i "s|\${OSMOSIS_WS_URL}|$OSMOSIS_WS_CLEAN|g" "$processed_file"
        fi
        if [ -n "$NEUTRON_WS_URL" ]; then
            NEUTRON_WS_CLEAN=$(process_url "$NEUTRON_WS_URL")
            sed -i "s|\${NEUTRON_WS_URL}|$NEUTRON_WS_CLEAN|g" "$processed_file"
        fi
        if [ -n "$NOBLE_WS_URL" ]; then
            NOBLE_WS_CLEAN=$(process_url "$NOBLE_WS_URL")
            sed -i "s|\${NOBLE_WS_URL}|$NOBLE_WS_CLEAN|g" "$processed_file"
        fi
        
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