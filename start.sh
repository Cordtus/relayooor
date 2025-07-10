#!/bin/bash

echo "🚀 Starting Relayooor IBC Monitoring Stack"
echo "========================================="

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Creating .env file from .env.example..."
    cp .env.example .env
    echo "✅ Created .env file - please edit it with your configuration"
fi

# Create necessary directories
echo "Creating directories..."
mkdir -p logs/relayer
mkdir -p config/keys

# Check if we have keys for the relayer
if [ -z "$(ls -A config/keys 2>/dev/null)" ]; then
    echo ""
    echo "⚠️  WARNING: No keys found in config/keys/"
    echo "For the relayer to work, you'll need to add keys for both chains:"
    echo "  - cosmos-key for cosmoshub-4"
    echo "  - osmosis-key for osmosis-1"
    echo ""
    echo "You can add keys later using:"
    echo "  docker exec -it relayooor-relayer-1 hermes keys add --chain cosmoshub-4 --mnemonic-file /path/to/mnemonic"
    echo "  docker exec -it relayooor-relayer-1 hermes keys add --chain osmosis-1 --mnemonic-file /path/to/mnemonic"
    echo ""
fi

# Start the services
echo "Starting Docker services..."
docker-compose up -d

# Wait for services to be ready
echo ""
echo "Waiting for services to start..."
sleep 10

# Check service status
echo ""
echo "Service Status:"
echo "==============="

# Function to check if service is running
check_service() {
    local service=$1
    local port=$2
    local url=$3
    
    if curl -s -o /dev/null -w "%{http_code}" $url | grep -q "200\|301\|302"; then
        echo "✅ $service is running on port $port"
    else
        echo "⚠️  $service may not be ready yet on port $port"
    fi
}

check_service "Web App" 80 "http://localhost"
check_service "API Backend" 8080 "http://localhost:8080/health"
check_service "Hermes REST API" 3000 "http://localhost:3000/version"
check_service "Hermes Telemetry" 3001 "http://localhost:3001/metrics"
check_service "Chainpulse Metrics" 3002 "http://localhost:3002/metrics"
check_service "Prometheus" 9090 "http://localhost:9090"
check_service "Grafana" 3003 "http://localhost:3003"

echo ""
echo "🎉 Stack is starting up!"
echo ""
echo "Access points:"
echo "=============="
echo "📊 Web Dashboard: http://localhost"
echo "📈 Grafana: http://localhost:3003 (admin/admin)"
echo "🔍 Prometheus: http://localhost:9090"
echo "🌐 API: http://localhost:8080"
echo "📡 Hermes API: http://localhost:3000"
echo ""
echo "Monitoring:"
echo "==========="
echo "Cosmos Hub <> Osmosis (channel-0 <> channel-141)"
echo ""
echo "Useful commands:"
echo "==============="
echo "View logs: docker-compose logs -f [service]"
echo "Stop all: docker-compose down"
echo "Restart service: docker-compose restart [service]"
echo ""
echo "Services: relayer, chainpulse, prometheus, grafana, api-backend, webapp"