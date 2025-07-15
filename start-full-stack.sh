#!/bin/bash

# Full Stack Deployment Script for Relayooor
# This script starts all services in the correct order with health checks

set -e

echo "Starting Relayooor Full Stack Deployment..."
echo "========================================="

# Function to check if a service is healthy
check_service() {
    local service=$1
    local port=$2
    local max_attempts=30
    local attempt=0
    
    echo -n "Waiting for $service on port $port..."
    while ! nc -z localhost $port 2>/dev/null; do
        attempt=$((attempt + 1))
        if [ $attempt -eq $max_attempts ]; then
            echo " FAILED"
            echo "Service $service failed to start on port $port"
            exit 1
        fi
        echo -n "."
        sleep 2
    done
    echo " OK"
}

# Clean up any existing containers
echo "Cleaning up existing containers..."
docker-compose -f docker-compose.full.yml down

# Build all services
echo "Building all services..."
docker-compose -f docker-compose.full.yml build

# Start database first
echo "Starting PostgreSQL..."
docker-compose -f docker-compose.full.yml up -d postgres
check_service "PostgreSQL" 5432

# Start monitoring services
echo "Starting Chainpulse..."
docker-compose -f docker-compose.full.yml up -d chainpulse
check_service "Chainpulse API" 3000

echo "Starting Prometheus..."
docker-compose -f docker-compose.full.yml up -d prometheus
check_service "Prometheus" 9090

# Start API backend
echo "Starting API Backend..."
docker-compose -f docker-compose.full.yml up -d api-backend
check_service "API Backend" 8080

# Start webapp
echo "Starting Web Application..."
docker-compose -f docker-compose.full.yml up -d webapp
check_service "Web Application" 80

# Start optional services
echo "Starting Grafana..."
docker-compose -f docker-compose.full.yml up -d grafana
check_service "Grafana" 3003

echo ""
echo "Full stack deployment complete!"
echo "================================"
echo "Services available at:"
echo "  - Web Application: http://localhost"
echo "  - API Backend: http://localhost:8080"
echo "  - Chainpulse API: http://localhost:3000"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3003 (admin/admin)"
echo ""
echo "To view logs: docker-compose -f docker-compose.full.yml logs -f [service-name]"
echo "To stop all services: docker-compose -f docker-compose.full.yml down"