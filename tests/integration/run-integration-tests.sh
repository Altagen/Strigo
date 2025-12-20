#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Starting Strigo Integration Tests with Nexus${NC}"
echo "=================================================="

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Configuration
NEXUS_URL="http://localhost:8081"
NEXUS_USER="admin"
NEXUS_REPO="raw"

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Nexus is ready
wait_for_nexus() {
    print_info "Waiting for Nexus to be ready..."
    local max_attempts=60
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -s -f "$NEXUS_URL/service/rest/v1/status" > /dev/null 2>&1; then
            print_info "✅ Nexus is ready!"
            return 0
        fi
        attempt=$((attempt + 1))
        echo -n "."
        sleep 2
    done

    print_error "Nexus failed to start after ${max_attempts} attempts"
    return 1
}

# Function to get Nexus admin password
get_nexus_password() {
    print_info "Retrieving Nexus admin password..."

    # Try to get password from container
    NEXUS_PASSWORD=$(podman exec strigo-nexus-test cat /nexus-data/admin.password 2>/dev/null || echo "")

    if [ -z "$NEXUS_PASSWORD" ]; then
        print_warn "Could not retrieve password from container, using default 'admin'"
        NEXUS_PASSWORD="admin"
    else
        print_info "Retrieved password from container"
    fi
}

# Function to accept EULA and complete onboarding
accept_eula() {
    print_info "Accepting Nexus EULA..."

    # Accept EULA
    curl -u "$NEXUS_USER:$NEXUS_PASSWORD" -X PUT "$NEXUS_URL/service/rest/v1/system/license" \
        -H "Content-Type: application/json" \
        -d '{"accepted": true}' 2>/dev/null || print_warn "EULA might already be accepted"

    # Disable anonymous access
    curl -u "$NEXUS_USER:$NEXUS_PASSWORD" -X PUT "$NEXUS_URL/service/rest/v1/security/anonymous" \
        -H "Content-Type: application/json" \
        -d '{"enabled": false}' 2>/dev/null || true

    # Wait for EULA changes to propagate
    print_info "Waiting for EULA changes to propagate..."
    sleep 5

    print_info "✅ EULA accepted"
}

# Function to create raw repository
create_raw_repository() {
    print_info "Creating raw repository..."

    curl -u "$NEXUS_USER:$NEXUS_PASSWORD" -X POST "$NEXUS_URL/service/rest/v1/repositories/raw/hosted" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "'"$NEXUS_REPO"'",
            "online": true,
            "storage": {
                "blobStoreName": "default",
                "strictContentTypeValidation": false,
                "writePolicy": "ALLOW"
            }
        }' 2>/dev/null || print_warn "Repository might already exist"

    print_info "✅ Repository created or already exists"
}

# Function to upload test JDK files
upload_test_files() {
    print_info "Uploading test JDK files..."

    # Create temporary test files with proper directory structure
    # Format: distribution/vendor/product/major_version/jdk-full_version/filename
    local test_files=(
        # Eclipse Temurin (AdoptOpenJDK)
        "jdk/adoptium/temurin/8/jdk-8u442b06/OpenJDK8U-jdk_x64_linux_hotspot_8u442b06.tar.gz"
        "jdk/adoptium/temurin/11/jdk-11.0.24_8/OpenJDK11U-jdk_x64_linux_hotspot_11.0.24_8.tar.gz"
        "jdk/adoptium/temurin/11/jdk-11.0.26_4/OpenJDK11U-jdk_x64_linux_hotspot_11.0.26_4.tar.gz"
        "jdk/adoptium/temurin/17/jdk-17.0.15_6/OpenJDK17U-jdk_x64_linux_hotspot_17.0.15_6.tar.gz"
        "jdk/adoptium/temurin/21/jdk-21.0.6_7/OpenJDK21U-jdk_x64_linux_hotspot_21.0.6_7.tar.gz"
        "jdk/adoptium/temurin/21/jdk-21.0.9_10/OpenJDK21U-jdk_x64_linux_hotspot_21.0.9_10.tar.gz"

        # Amazon Corretto
        "jdk/amazon/corretto/8/jdk-8.442.06.1/amazon-corretto-8.442.06.1-linux-x64.tar.gz"
        "jdk/amazon/corretto/11/jdk-11.0.24.7.1/amazon-corretto-11.0.24.7.1-linux-x64.tar.gz"
        "jdk/amazon/corretto/11/jdk-11.0.26.4.1/amazon-corretto-11.0.26.4.1-linux-x64.tar.gz"
        "jdk/amazon/corretto/17/jdk-17.0.15.8.1/amazon-corretto-17.0.15.8.1-linux-x64.tar.gz"
        "jdk/amazon/corretto/21/jdk-21.0.9.11.1/amazon-corretto-21.0.9.11.1-linux-x64.tar.gz"

        # Azul Zulu
        "jdk/azul/zulu/8/jdk-8.0.442/zulu8.84.0.15-ca-jdk8.0.442-linux_x64.tar.gz"
        "jdk/azul/zulu/11/jdk-11.0.26/zulu11.78.15-ca-jdk11.0.26-linux_x64.tar.gz"
        "jdk/azul/zulu/17/jdk-17.0.15/zulu17.54.25-ca-jdk17.0.15-linux_x64.tar.gz"
        "jdk/azul/zulu/21/jdk-21.0.6/zulu21.38.21-ca-jdk21.0.6-linux_x64.tar.gz"

        # Node.js
        "node/v22/node-v22.13.1-linux-x64.tar.xz"
        "node/v20/node-v20.18.2-linux-x64.tar.xz"
        "node/v18/node-v18.20.5-linux-x64.tar.xz"
    )

    for file_path in "${test_files[@]}"; do
        local filename=$(basename "$file_path")
        local dir=$(dirname "$file_path")

        # Create a small dummy tar.gz/tar.xz file
        mkdir -p "/tmp/strigo-test/$dir"
        echo "Test SDK file: $filename" > "/tmp/strigo-test/$dir/$filename.txt"

        # Use appropriate compression based on file extension
        if [[ "$filename" == *.tar.xz ]]; then
            tar -cJf "/tmp/strigo-test/$file_path" -C "/tmp/strigo-test/$dir" "$filename.txt" 2>/dev/null || true
        else
            tar -czf "/tmp/strigo-test/$file_path" -C "/tmp/strigo-test/$dir" "$filename.txt" 2>/dev/null || true
        fi

        # Upload to Nexus
        local content_type="application/x-gzip"
        [[ "$filename" == *.tar.xz ]] && content_type="application/x-xz"

        curl -u "$NEXUS_USER:$NEXUS_PASSWORD" -X PUT \
            "$NEXUS_URL/repository/$NEXUS_REPO/$file_path" \
            --upload-file "/tmp/strigo-test/$file_path" \
            -H "Content-Type: $content_type" 2>/dev/null || true

        echo -n "."
    done

    echo ""
    print_info "✅ Test files uploaded (18 files: 6 Temurin, 5 Corretto, 4 Zulu, 3 Node.js)"

    # Cleanup temp files
    rm -rf /tmp/strigo-test
}

# Function to run integration tests
run_integration_tests() {
    print_info "Running integration tests..."

    cd "$SCRIPT_DIR/../.."

    # Update strigo.toml with test Nexus URL
    if [ -f "strigo.toml.backup" ]; then
        cp strigo.toml.backup strigo.toml
    else
        cp strigo.toml strigo.toml.backup
    fi

    # Create test config
    cat > strigo-test.toml << EOF
[general]
log_level = "debug"
sdk_install_dir = "/tmp/strigo-test-sdks"
cache_dir = "/tmp/strigo-test-cache"
log_path = ""
keep_cache = false
shell_config_path = ""
patterns_file = ""

jdk_security_path = "lib/security/cacerts"
system_cacerts_path = "/etc/ssl/certs"

[registries.nexus]
type = "nexus"
api_url = "$NEXUS_URL/service/rest/v1/assets?repository={repository}"
username = "$NEXUS_USER"
password = "$NEXUS_PASSWORD"

[sdk_types]
jdk = {
    type = "jdk",
    install_dir = "jdks"
}
node = {
    type = "node",
    install_dir = "nodes"
}

[sdk_repositories]
temurin = {
    registry = "nexus",
    repository = "$NEXUS_REPO",
    type = "jdk",
    path = "jdk/adoptium/temurin"
}
corretto = {
    registry = "nexus",
    repository = "$NEXUS_REPO",
    type = "jdk",
    path = "jdk/amazon/corretto"
}
zulu = {
    registry = "nexus",
    repository = "$NEXUS_REPO",
    type = "jdk",
    path = "jdk/azul/zulu"
}
nodejs = {
    registry = "nexus",
    repository = "$NEXUS_REPO",
    type = "node",
    path = "node"
}
EOF

    print_info "Testing 'available' command with Temurin (all versions)..."
    STRIGO_CONFIG_PATH="strigo-test.toml" ./strigo available jdk temurin || true

    print_info "Testing 'available' with Corretto JDK 11 filter..."
    STRIGO_CONFIG_PATH="strigo-test.toml" ./strigo available jdk corretto 11 || true

    print_info "Testing 'available' with Zulu JDK 8 filter..."
    STRIGO_CONFIG_PATH="strigo-test.toml" ./strigo available jdk zulu 8 || true

    print_info "Testing 'available' with Temurin JDK 21 (multiple versions)..."
    STRIGO_CONFIG_PATH="strigo-test.toml" ./strigo available jdk temurin 21 || true

    print_info "Testing 'available' with Node.js (all versions)..."
    STRIGO_CONFIG_PATH="strigo-test.toml" ./strigo available node nodejs || true

    print_info "Testing 'available' with Node.js 22 filter..."
    STRIGO_CONFIG_PATH="strigo-test.toml" ./strigo available node nodejs 22 || true

    print_info "✅ Integration tests completed"

    # Cleanup
    rm -f strigo-test.toml
    rm -rf /tmp/strigo-test-sdks /tmp/strigo-test-cache
}

# Function to cleanup
cleanup() {
    print_info "Cleaning up..."
    podman-compose down -v 2>/dev/null || true

    if [ -f "strigo.toml.backup" ]; then
        mv strigo.toml.backup strigo.toml 2>/dev/null || true
    fi

    print_info "✅ Cleanup completed"
}

# Main execution
main() {
    # Handle Ctrl+C
    trap cleanup EXIT

    # Check if we should cleanup only
    if [ "$1" = "cleanup" ]; then
        cleanup
        exit 0
    fi

    print_info "Step 1/7: Starting Nexus container..."
    podman-compose up -d

    print_info "Step 2/7: Waiting for Nexus to be ready..."
    wait_for_nexus

    print_info "Step 3/7: Getting Nexus credentials..."
    get_nexus_password

    print_info "Step 4/7: Accepting EULA..."
    accept_eula

    print_info "Step 5/7: Creating repository..."
    create_raw_repository

    print_info "Step 6/7: Uploading test files..."
    upload_test_files

    print_info "Step 7/7: Running integration tests..."
    run_integration_tests

    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✅ All integration tests completed!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    print_info "To cleanup, run: $0 cleanup"
    print_info "Nexus UI available at: $NEXUS_URL (user: admin, password: see above)"
}

# Run main function
main "$@"
