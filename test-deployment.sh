#!/bin/bash

# Test script for individual service deployment

set -e

echo "Testing Relayooor Service Deployment"
echo "===================================="

# Function to test a service
test_service() {
    local service=$1
    local description=$2
    
    echo ""
    echo "Testing $description..."
    echo "------------------------"
    
    # Build the service
    echo "Building $service..."
    if docker-compose -f docker-compose.full.yml build --no-cache $service; then
        echo "✓ Build successful"
    else
        echo "✗ Build failed for $service"
        return 1
    fi
    
    # Start the service
    echo "Starting $service..."
    if docker-compose -f docker-compose.full.yml up -d $service; then
        echo "✓ Container started"
        
        # Check if running
        sleep 5
        if docker-compose -f docker-compose.full.yml ps | grep -q "$service.*Up"; then
            echo "✓ Service is running"
            
            # Show logs
            echo "Recent logs:"
            docker-compose -f docker-compose.full.yml logs --tail=10 $service
        else
            echo "✗ Service failed to stay running"
            docker-compose -f docker-compose.full.yml logs $service
        fi
    else
        echo "✗ Failed to start $service"
        return 1
    fi
}

# Test each service
echo "1. Testing PostgreSQL Database..."
test_service postgres "PostgreSQL Database"

echo ""
echo "2. Testing Chainpulse..."
test_service chainpulse "Chainpulse Monitoring"

echo ""
echo "3. Testing API Backend..."
test_service api-backend "API Backend Server"

echo ""
echo "4. Testing Web Application..."
test_service webapp "Vue.js Web Application"

echo ""
echo "===================================="
echo "Deployment test complete!"
echo ""
echo "Running services:"
docker-compose -f docker-compose.full.yml ps

echo ""
echo "To stop all services: docker-compose -f docker-compose.full.yml down"