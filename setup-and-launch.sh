#!/bin/bash

# Relayooor Setup and Launch Script
# This script handles the complete setup and launch process for the Relayooor application

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

print_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"
}

# Check if Docker is running
check_docker() {
    print_status "Checking Docker status..."
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker Desktop and try again."
        exit 1
    fi
    print_status "Docker is running ✓"
}

# Check for required environment variables
check_env() {
    print_status "Checking environment variables..."
    
    if [ ! -f .env ]; then
        print_warning ".env file not found. Creating from template..."
        if [ -f .env.example ]; then
            cp .env.example .env
            print_status "Created .env file. Please update with your RPC credentials."
        else
            cat > .env << EOF
# RPC Authentication (Skip endpoints)
RPC_USERNAME=skip
RPC_PASSWORD=your_password_here

# Optional: Override default WebSocket URLs
# COSMOS_WS_URL=wss://cosmos-rpc.polkachu.com/websocket
# OSMOSIS_WS_URL=wss://osmosis-rpc.polkachu.com/websocket
# NEUTRON_WS_URL=wss://neutron-rpc.polkachu.com/websocket
# NOBLE_WS_URL=wss://noble-rpc.polkachu.com/websocket

# Grafana admin credentials
GF_ADMIN_USER=admin
GF_ADMIN_PASSWORD=admin
EOF
            print_warning "Created .env file with defaults. Please update RPC_PASSWORD!"
        fi
    fi
    
    # Source the .env file
    source .env
    
    # Check critical variables
    if [ -z "$RPC_USERNAME" ] || [ -z "$RPC_PASSWORD" ]; then
        print_error "RPC_USERNAME and RPC_PASSWORD must be set in .env file"
        exit 1
    fi
    
    print_status "Environment variables loaded ✓"
}

# Clean up any existing containers
cleanup() {
    print_status "Cleaning up existing containers..."
    docker-compose down --remove-orphans > /dev/null 2>&1 || true
    print_status "Cleanup complete ✓"
}

# Build all services
build_services() {
    print_status "Building services..."
    
    # Build with progress output
    if ! docker-compose build --progress=plain; then
        print_error "Build failed. Check the error messages above."
        exit 1
    fi
    
    print_status "All services built successfully ✓"
}

# Start all services
start_services() {
    print_status "Starting services..."
    
    if ! docker-compose up -d; then
        print_error "Failed to start services"
        exit 1
    fi
    
    print_status "All services started ✓"
}

# Wait for services to be healthy
wait_for_services() {
    print_status "Waiting for services to be ready..."
    
    # Wait for API
    print_status "Waiting for API..."
    local retries=30
    while [ $retries -gt 0 ]; do
        if curl -s http://localhost:3000/health > /dev/null 2>&1; then
            print_status "API is ready ✓"
            break
        fi
        retries=$((retries - 1))
        sleep 2
    done
    
    if [ $retries -eq 0 ]; then
        print_error "API failed to start"
        docker-compose logs api-backend
        exit 1
    fi
    
    # Wait for Chainpulse
    print_status "Waiting for Chainpulse..."
    retries=30
    while [ $retries -gt 0 ]; do
        if curl -s http://localhost:3001/metrics > /dev/null 2>&1; then
            print_status "Chainpulse is ready ✓"
            break
        fi
        retries=$((retries - 1))
        sleep 2
    done
    
    if [ $retries -eq 0 ]; then
        print_error "Chainpulse failed to start"
        docker-compose logs chainpulse
        exit 1
    fi
    
    # Give webapp a moment to start
    sleep 5
    
    # Check webapp
    if curl -sI http://localhost:80 | grep -q "200 OK"; then
        print_status "Web application is ready ✓"
    else
        print_warning "Web application may not be fully ready yet"
    fi
}

# Display service status
show_status() {
    print_status "Service Status:"
    echo ""
    echo "  📊 Web Application:    http://localhost"
    echo "  🔌 API Backend:        http://localhost:3000"
    echo "  📈 Chainpulse Metrics: http://localhost:3001/metrics"
    echo "  🗄️  PostgreSQL:         localhost:5432"
    echo "  💾 Redis:              localhost:6379"
    echo "  📊 Prometheus:         http://localhost:9090"
    echo "  📈 Grafana:            http://localhost:3003 (admin/admin)"
    echo ""
    
    # Show container status
    print_status "Container Status:"
    docker-compose ps
}

# Display logs
show_logs() {
    print_status "Recent logs from all services:"
    docker-compose logs --tail=20
}

# Main execution
main() {
    echo "🚀 Relayooor Setup and Launch Script"
    echo "===================================="
    echo ""
    
    # Parse command line arguments
    case "${1:-}" in
        "stop")
            print_status "Stopping all services..."
            docker-compose down
            print_status "All services stopped ✓"
            exit 0
            ;;
        "restart")
            print_status "Restarting all services..."
            docker-compose restart
            wait_for_services
            show_status
            exit 0
            ;;
        "logs")
            docker-compose logs -f
            exit 0
            ;;
        "status")
            show_status
            exit 0
            ;;
        "clean")
            print_warning "This will remove all containers and volumes!"
            read -p "Are you sure? (y/N) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                docker-compose down -v
                print_status "Clean complete ✓"
            fi
            exit 0
            ;;
    esac
    
    # Full setup and launch sequence
    check_docker
    check_env
    cleanup
    build_services
    start_services
    wait_for_services
    
    echo ""
    print_status "🎉 Relayooor is running!"
    echo ""
    show_status
    
    echo ""
    print_status "Useful commands:"
    echo "  ./setup-and-launch.sh stop     - Stop all services"
    echo "  ./setup-and-launch.sh restart  - Restart all services"
    echo "  ./setup-and-launch.sh logs     - View logs"
    echo "  ./setup-and-launch.sh status   - Show service status"
    echo "  ./setup-and-launch.sh clean    - Remove all data and containers"
}

# Run main function
main "$@"