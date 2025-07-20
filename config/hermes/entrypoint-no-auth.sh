#!/bin/bash
set -e

echo "Initializing Hermes (no authentication)..."

# Create Hermes directory if it doesn't exist
mkdir -p /home/hermes/.hermes

# Copy config to writable location
cp /config/config.toml /home/hermes/.hermes/config.toml

echo "Starting Hermes..."
exec hermes "$@"