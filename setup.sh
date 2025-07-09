#!/bin/bash
# Setup script for IBC Relayer Dashboard

echo "Setting up IBC Relayer Dashboard..."

# Add relayer repositories as submodules
if [ ! -d "hermes/.git" ]; then
    echo "Adding Hermes as submodule..."
    git submodule add https://github.com/informalsystems/hermes.git hermes
fi

if [ ! -d "relayer/.git" ]; then
    echo "Adding Go relayer as submodule..."
    git submodule add https://github.com/cosmos/relayer.git relayer
fi

# Initialize and update submodules
git submodule update --init --recursive

# Create config directories if they don't exist
mkdir -p config/hermes config/relayer

# Copy example configs if available
if [ -f "hermes/config.toml" ] && [ ! -f "config/hermes/config.toml" ]; then
    echo "Copying example Hermes config..."
    cp hermes/config.toml config/hermes/config.toml.example
fi

if [ -f "relayer/examples/config_EXAMPLE.yaml" ] && [ ! -f "config/relayer/config.yaml" ]; then
    echo "Copying example Go relayer config..."
    cp relayer/examples/config_EXAMPLE.yaml config/relayer/config.yaml.example
fi

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    echo "Creating .env file from example..."
    cp .env.example .env
    echo "Please edit .env file with your configuration"
fi

echo "Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your configuration"
echo "2. Place your relayer configs in config/hermes/ and config/relayer/"
echo "3. Run 'docker-compose up -d' to start the dashboard"
echo "4. Access the dashboard at http://localhost:8080"