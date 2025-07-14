#!/bin/bash

# Test runner script for Relayooor

set -e

echo "========================================="
echo "Running Relayooor Test Suite"
echo "========================================="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests in a directory
run_tests() {
    local dir=$1
    local name=$2
    
    echo -e "${YELLOW}Running $name tests...${NC}"
    
    if [ -d "$dir" ]; then
        cd "$dir"
        
        if [ -f "go.mod" ]; then
            # Go tests
            if go test ./... -v -count=1; then
                echo -e "${GREEN}✓ $name tests passed${NC}"
                ((PASSED_TESTS++))
            else
                echo -e "${RED}✗ $name tests failed${NC}"
                ((FAILED_TESTS++))
            fi
        elif [ -f "package.json" ]; then
            # Node/Vue tests
            if [ -f "vitest.config.ts" ] || [ -f "vitest.config.js" ]; then
                if npm test; then
                    echo -e "${GREEN}✓ $name tests passed${NC}"
                    ((PASSED_TESTS++))
                else
                    echo -e "${RED}✗ $name tests failed${NC}"
                    ((FAILED_TESTS++))
                fi
            fi
        fi
        
        cd - > /dev/null
        ((TOTAL_TESTS++))
    else
        echo -e "${YELLOW}Skipping $name - directory not found${NC}"
    fi
    
    echo
}

# 1. API Backend Tests
echo -e "${YELLOW}Phase 1: API Backend Tests${NC}"
echo "==============================="
run_tests "api" "Simple API Backend"

# 2. Relayer Middleware Tests
echo -e "${YELLOW}Phase 2: Relayer Middleware Tests${NC}"
echo "===================================="
run_tests "relayer-middleware/api" "Relayer Middleware API"

# 3. Frontend Tests
echo -e "${YELLOW}Phase 3: Frontend Tests${NC}"
echo "========================="
run_tests "webapp" "Vue.js Frontend"

# 4. Integration Tests
echo -e "${YELLOW}Phase 4: Integration Tests${NC}"
echo "============================"
if [ -d "tests/integration" ]; then
    echo "Running integration tests..."
    cd tests/integration
    
    # Check if services are running
    SERVICES_UP=true
    
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${YELLOW}Warning: API backend not running on port 8080${NC}"
        SERVICES_UP=false
    fi
    
    if ! curl -s http://localhost:5173 > /dev/null 2>&1; then
        echo -e "${YELLOW}Warning: Frontend not running on port 5173${NC}"
        SERVICES_UP=false
    fi
    
    if [ "$SERVICES_UP" = true ]; then
        if npm test; then
            echo -e "${GREEN}✓ Integration tests passed${NC}"
            ((PASSED_TESTS++))
        else
            echo -e "${RED}✗ Integration tests failed${NC}"
            ((FAILED_TESTS++))
        fi
        ((TOTAL_TESTS++))
    else
        echo -e "${YELLOW}Skipping integration tests - services not running${NC}"
        echo "Run 'docker-compose up' to start services for integration testing"
    fi
    
    cd - > /dev/null
fi

echo

# 5. Test Coverage Report
echo -e "${YELLOW}Phase 5: Test Coverage${NC}"
echo "======================="

# Go coverage
if command -v go &> /dev/null; then
    echo "Generating Go test coverage..."
    
    # API coverage
    if [ -d "api" ]; then
        cd api
        go test ./... -coverprofile=coverage.out
        go tool cover -html=coverage.out -o coverage.html
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        echo -e "API Backend Coverage: ${GREEN}$COVERAGE${NC}"
        cd - > /dev/null
    fi
    
    # Middleware coverage
    if [ -d "relayer-middleware/api" ]; then
        cd relayer-middleware/api
        go test ./... -coverprofile=coverage.out
        go tool cover -html=coverage.out -o coverage.html
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        echo -e "Relayer Middleware Coverage: ${GREEN}$COVERAGE${NC}"
        cd - > /dev/null
    fi
fi

# Vue coverage
if [ -d "webapp" ] && [ -f "webapp/package.json" ]; then
    cd webapp
    if npm run test:coverage > /dev/null 2>&1; then
        echo -e "Frontend Coverage: ${GREEN}Generated in webapp/coverage${NC}"
    fi
    cd - > /dev/null
fi

echo

# 6. Security Scan
echo -e "${YELLOW}Phase 6: Security Scan${NC}"
echo "======================="

# Check for Go vulnerabilities
if command -v govulncheck &> /dev/null; then
    echo "Scanning Go dependencies for vulnerabilities..."
    govulncheck ./... || echo -e "${YELLOW}Install govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest${NC}"
else
    echo -e "${YELLOW}govulncheck not found - skipping Go vulnerability scan${NC}"
fi

# Check for npm vulnerabilities
if [ -d "webapp" ]; then
    echo "Scanning npm dependencies for vulnerabilities..."
    cd webapp
    npm audit || true
    cd - > /dev/null
fi

echo

# Final Summary
echo "========================================="
echo -e "${YELLOW}Test Summary${NC}"
echo "========================================="
echo -e "Total Test Suites: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ] && [ $TOTAL_TESTS -gt 0 ]; then
    echo
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo
    echo -e "${RED}✗ Some tests failed. Please review the output above.${NC}"
    exit 1
fi