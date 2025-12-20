#!/usr/bin/env bash
set -euo pipefail

# Test Strigo Binary in Isolation
# This script builds the strigo binary and tests it in a completely isolated environment
# to ensure it works correctly when distributed to users.

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Get project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ISOLATED_DIR="/tmp/strigo-isolated-test-$$"

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Strigo Isolated Binary Test${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Step 1: Build the binary
echo -e "${GREEN}[1/8]${NC} Building strigo binary..."
cd "$PROJECT_ROOT"
go build -o strigo ./main.go
echo "✅ Binary built successfully"
echo ""

# Step 2: Create isolated test directory
echo -e "${GREEN}[2/8]${NC} Creating isolated test directory: $ISOLATED_DIR"
mkdir -p "$ISOLATED_DIR"
cd "$ISOLATED_DIR"
echo "✅ Isolated directory created"
echo ""

# Step 3: Copy binary and patterns file
echo -e "${GREEN}[3/8]${NC} Copying binary and patterns file..."
cp "$PROJECT_ROOT/strigo" ./
cp "$PROJECT_ROOT/strigo-patterns.toml" ./
chmod +x strigo
echo "✅ Files copied"
echo ""

# Step 4: Create test configuration
echo -e "${GREEN}[4/8]${NC} Creating test configuration..."
cat > strigo.toml << 'EOF'
[general]
log_level = "info"
sdk_install_dir = "/tmp/strigo-isolated-sdks"
cache_dir = "/tmp/strigo-isolated-cache"
patterns_file = "strigo-patterns.toml"
jdk_security_path = "lib/security/cacerts"
system_cacerts_path = "/etc/ssl/certs/ca-certificates.crt"

[sdk_types]
jdk = { type = "jdk", install_dir = "jdks" }

[registries]
nexus = {
    type = "nexus",
    api_url = "http://localhost:8081/service/rest/v1/assets?repository={repository}",
    username = "admin",
    password = "admin"
}

[sdk_repositories]
temurin = {
    registry = "nexus",
    repository = "raw",
    type = "jdk",
    path = "jdk/adoptium/temurin"
}
corretto = {
    registry = "nexus",
    repository = "raw",
    type = "jdk",
    path = "jdk/amazon/corretto"
}
EOF
echo "✅ Configuration created"
echo ""

# Step 5: Test basic commands
echo -e "${GREEN}[5/8]${NC} Testing basic commands..."
echo ""

echo "→ Testing: ./strigo --help"
./strigo --help > /dev/null
echo "✅ Help command works"

echo "→ Testing: ./strigo available"
./strigo available > /dev/null
echo "✅ Available SDK types listed"

echo "→ Testing: ./strigo available jdk"
./strigo available jdk > /dev/null
echo "✅ JDK distributions listed"

echo ""

# Step 6: Test with different configuration methods
echo -e "${GREEN}[6/8]${NC} Testing configuration methods..."
echo ""

# Test with --config flag
echo "→ Testing: ./strigo --config strigo.toml list jdk"
./strigo --config strigo.toml list jdk > /dev/null
echo "✅ Config flag works"

# Test with environment variable
echo "→ Testing: STRIGO_CONFIG_PATH=strigo.toml ./strigo list jdk"
STRIGO_CONFIG_PATH=strigo.toml ./strigo list jdk > /dev/null
echo "✅ Config env var works"

# Test with --patterns flag override
echo "→ Testing: ./strigo --patterns strigo-patterns.toml available jdk"
./strigo --patterns strigo-patterns.toml available jdk > /dev/null
echo "✅ Patterns flag works"

# Test with patterns env var
echo "→ Testing: STRIGO_PATTERNS_PATH=strigo-patterns.toml ./strigo available jdk"
STRIGO_PATTERNS_PATH=strigo-patterns.toml ./strigo available jdk > /dev/null
echo "✅ Patterns env var works"

echo ""

# Step 7: Test if Nexus is available and do E2E test
echo -e "${GREEN}[7/8]${NC} Testing with Nexus (if available)..."
if curl -s http://localhost:8081/ > /dev/null 2>&1; then
    echo "✓ Nexus is running, performing E2E test..."
    echo ""

    echo "→ Listing available Temurin versions..."
    ./strigo available jdk temurin
    echo ""

    # Get first available version
    VERSION=$(./strigo available jdk temurin 2>/dev/null | grep "✅" | head -1 | awk '{print $2}')

    if [ -n "$VERSION" ]; then
        echo "→ Installing JDK version: $VERSION"
        ./strigo install jdk temurin "$VERSION"
        echo ""

        echo "→ Listing installed JDKs..."
        ./strigo list jdk temurin
        echo ""

        echo "✅ E2E test completed successfully"
    else
        echo "⚠️  No versions found in Nexus, skipping install test"
    fi
else
    echo "⚠️  Nexus not running, skipping E2E test"
    echo "   To run full E2E test, start Nexus with: podman-compose up -d"
fi
echo ""

# Step 8: Cleanup
echo -e "${GREEN}[8/8]${NC} Cleaning up..."
rm -rf /tmp/strigo-isolated-sdks /tmp/strigo-isolated-cache
echo "✅ Test artifacts cleaned up"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✅ All isolated tests passed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Test directory (not deleted): $ISOLATED_DIR"
echo "To manually test, run:"
echo "  cd $ISOLATED_DIR"
echo "  ./strigo --help"
echo ""
echo "To clean up:"
echo "  rm -rf $ISOLATED_DIR"
echo ""
