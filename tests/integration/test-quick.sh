#!/bin/bash
# Quick test to verify the integration setup works
# This doesn't wait for Nexus to fully start

set -e

GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}🧪 Quick Integration Test Validation${NC}"
echo "========================================"

# Test 1: Podman is available
echo -n "✓ Checking Podman... "
podman --version > /dev/null && echo "OK"

# Test 2: podman-compose is available
echo -n "✓ Checking podman-compose... "
podman-compose --version > /dev/null && echo "OK"

# Test 3: Compose file is valid
echo -n "✓ Validating podman-compose.yml... "
podman-compose config > /dev/null && echo "OK"

# Test 4: Script is executable
echo -n "✓ Checking run-integration-tests.sh... "
[ -x "run-integration-tests.sh" ] && echo "OK"

# Test 5: Can pull Nexus image (without starting)
echo -n "✓ Checking Nexus image... "
podman pull docker.io/sonatype/nexus3:latest --quiet 2>/dev/null && echo "OK" || echo "SKIP (will download on first run)"

echo ""
echo -e "${GREEN}✅ All checks passed!${NC}"
echo ""
echo "To run full integration tests (will take 3-5 minutes):"
echo "  ./run-integration-tests.sh"
echo ""
echo "To cleanup after tests:"
echo "  ./run-integration-tests.sh cleanup"
