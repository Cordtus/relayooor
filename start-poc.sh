#!/bin/bash

echo "Starting IBC Monitoring POC..."

# Check if chainpulse is built
if [ ! -d "chainpulse" ]; then
    echo "Cloning chainpulse..."
    git clone https://github.com/informalsystems/chainpulse.git
fi

# Create directories for Grafana provisioning
mkdir -p config/grafana-poc/dashboards
mkdir -p config/grafana-poc/datasources

# Start services
docker-compose -f docker-compose-poc.yml up -d

echo ""
echo "IBC Monitoring POC started!"
echo ""
echo "Services:"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - Prometheus: http://localhost:9090"
echo "  - Chainpulse metrics: http://localhost:3001/metrics"
echo ""
echo "Monitoring paths:"
echo "  - osmosis-1 <> cosmoshub-4"
echo "  - osmosis-1 <> pacific-1"
echo ""
echo "To view logs: docker-compose -f docker-compose-poc.yml logs -f"
echo "To stop: docker-compose -f docker-compose-poc.yml down"