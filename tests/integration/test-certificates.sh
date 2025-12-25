#!/bin/bash

set -e  # Exit on error

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DOWNLOAD_DIR="/tmp/strigo-e2e-downloads"
TEST_INSTALL_DIR="/tmp/strigo-e2e-test-sdks"
TEST_CACHE_DIR="/tmp/strigo-e2e-test-cache"
TEST_CERT_DIR="/tmp/strigo-test-certs"

# Test configuration
CONFIG_FILE="$PROJECT_ROOT/strigo-e2e-test.toml"
BINARY="$PROJECT_ROOT/strigo-new"

# Nexus configuration
NEXUS_CONTAINER_NAME="strigo-nexus-e2e-test"
NEXUS_PASSWORD="admin"

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Check if Nexus is already running
check_nexus() {
    if podman ps --format "{{.Names}}" | grep -q "^${NEXUS_CONTAINER_NAME}$"; then
        print_info "Nexus container already running"
        return 0
    else
        print_error "Nexus container not found. Please run test-e2e-real-sdks.sh first to set up Nexus"
        exit 1
    fi
}

# Wait for Nexus to be ready
wait_for_nexus() {
    print_info "Waiting for Nexus to be ready..."
    for i in {1..60}; do
        if curl -s -o /dev/null -w "%{http_code}" http://localhost:8081 | grep -q "200\|403"; then
            print_success "Nexus is ready!"
            return 0
        fi
        echo -n "."
        sleep 2
    done
    print_error "Nexus did not become ready in time"
    exit 1
}

# Build Strigo
build_strigo() {
    print_header "Building Strigo"
    cd "$PROJECT_ROOT"
    go build -o "$BINARY" .
    print_success "Strigo built successfully"
}

# Setup test certificate
setup_test_certificate() {
    print_header "Setting up Test Certificate"

    # Certificate should already exist from previous test
    if [ ! -f "$TEST_CERT_DIR/test-cert.pem" ]; then
        print_error "Test certificate not found. Please run test-e2e-real-sdks.sh first"
        exit 1
    fi

    print_success "Test certificate found: $TEST_CERT_DIR/test-cert.pem"
}

# Clean test directories
clean_test_dirs() {
    print_header "Cleaning Test Directories"
    rm -rf "$TEST_INSTALL_DIR" "$TEST_CACHE_DIR"
    mkdir -p "$TEST_INSTALL_DIR" "$TEST_CACHE_DIR"
    print_success "Test directories cleaned"
}

# Test JDK certificate injection
test_jdk_certificates() {
    print_header "Testing JDK Certificate Injection"

    # Test multiple Java versions
    local distributions=("temurin:11" "corretto:8" "zulu:17")

    for dist_ver in "${distributions[@]}"; do
        IFS=':' read -r dist version <<< "$dist_ver"

        print_info "Test: Installing $dist JDK $version with certificates..."

        # Get available version (first one, filtering out debug and headers)
        local available_version=$("$BINARY" --config "$CONFIG_FILE" available jdk "$dist" 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+(_[0-9]+|b[0-9]+|u[0-9]+)?' | head -n 1)

        if [ -z "$available_version" ]; then
            print_error "No $dist versions available, skipping"
            continue
        fi

        print_info "Installing version: $available_version"

        # Install JDK
        "$BINARY" --config "$CONFIG_FILE" install jdk "$dist" "$available_version"

        # Find the installed JDK directory
        local jdk_path=$(find "$TEST_INSTALL_DIR/jdks/$dist/$available_version" -mindepth 1 -maxdepth 1 -type d | head -n 1)

        if [ -z "$jdk_path" ]; then
            print_error "Could not find installed JDK directory"
            exit 1
        fi

        print_info "JDK installed at: $jdk_path"

        # Detect cacerts path (Java 8 vs 11+)
        local cacerts_path=""
        if [ -f "$jdk_path/jre/lib/security/cacerts" ]; then
            cacerts_path="$jdk_path/jre/lib/security/cacerts"
            print_info "Detected Java 8 cacerts path"
        elif [ -f "$jdk_path/lib/security/cacerts" ]; then
            cacerts_path="$jdk_path/lib/security/cacerts"
            print_info "Detected Java 11+ cacerts path"
        else
            print_error "Could not find cacerts file"
            exit 1
        fi

        print_info "Cacerts path: $cacerts_path"

        # Verify backup was created
        if [ -f "$cacerts_path.original" ]; then
            print_success "Backup file created: $cacerts_path.original"
        else
            print_error "Backup file not found"
            exit 1
        fi

        # Verify certificate was injected
        print_info "Verifying certificate injection with keytool..."

        local keytool_path=""
        if [ -f "$jdk_path/bin/keytool" ]; then
            keytool_path="$jdk_path/bin/keytool"
        else
            print_error "keytool not found"
            exit 1
        fi

        if "$keytool_path" -list -keystore "$cacerts_path" -storepass changeit -alias strigo-test-ca > /dev/null 2>&1; then
            print_success "Certificate 'strigo-test-ca' found in keystore!"
        else
            print_error "Certificate 'strigo-test-ca' not found in keystore"
            "$keytool_path" -list -keystore "$cacerts_path" -storepass changeit | grep -i strigo || true
            exit 1
        fi

        # Verify default CAs are preserved
        local cert_count=$("$keytool_path" -list -keystore "$cacerts_path" -storepass changeit 2>&1 | grep -ic "trustedCertEntry" || true)
        print_info "Total certificates in keystore: $cert_count"

        if [ "$cert_count" -gt 50 ]; then
            print_success "Default CA certificates preserved (found $cert_count certs)"
        else
            print_error "Default CA certificates may have been lost (only $cert_count certs found)"
            exit 1
        fi

        print_success "Test passed for $dist JDK $available_version"
        echo ""
    done
}

# Test Node.js certificate configuration
test_node_certificates() {
    print_header "Testing Node.js Certificate Configuration"

    print_info "Installing Node.js with certificate configuration..."

    # Get available Node.js version
    local available_version=$("$BINARY" --config "$CONFIG_FILE" available node nodejs 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -n 1)

    if [ -z "$available_version" ]; then
        print_error "No Node.js versions available"
        exit 1
    fi

    print_info "Installing Node.js version: $available_version"

    # Install Node.js with certificate
    "$BINARY" --config "$CONFIG_FILE" install node nodejs "$available_version" --node-extra-ca-certs "$TEST_CERT_DIR/test-cert.pem"

    print_success "Node.js installed with certificate configuration"

    # Verify metadata file was created
    local metadata_path="$TEST_INSTALL_DIR/nodes/nodejs/$available_version/.strigo-metadata.json"

    if [ -f "$metadata_path" ]; then
        print_success "Metadata file created: $metadata_path"

        # Check if certificate path is in metadata
        if grep -q "node_extra_ca_certs" "$metadata_path"; then
            print_success "Certificate path found in metadata"
            print_info "Metadata content:"
            cat "$metadata_path" | jq . || cat "$metadata_path"
        else
            print_error "Certificate path not found in metadata"
            exit 1
        fi
    else
        print_error "Metadata file not found"
        exit 1
    fi

    # Test 'use' command without --set-env
    print_info "Testing 'strigo use' command (display mode)..."
    local use_output=$("$BINARY" --config "$CONFIG_FILE" use node nodejs "$available_version" 2>&1)

    if echo "$use_output" | grep -q "NODE_EXTRA_CA_CERTS"; then
        print_success "'strigo use' displays NODE_EXTRA_CA_CERTS correctly"
        echo "$use_output" | grep "NODE_EXTRA_CA_CERTS"
    else
        print_error "'strigo use' does not display NODE_EXTRA_CA_CERTS"
        echo "$use_output"
        exit 1
    fi

    # Test 'use' command with --set-env
    print_info "Testing 'strigo use --set-env' command..."

    # Create a temporary RC file for testing
    local temp_rc=$(mktemp)

    # Use a custom config that points to our temp RC file
    local temp_config=$(mktemp)
    cat "$CONFIG_FILE" > "$temp_config"
    echo "shell_config_path = \"$temp_rc\"" >> "$temp_config"

    "$BINARY" --config "$temp_config" use node nodejs "$available_version" --set-env

    if [ -f "$temp_rc" ]; then
        print_success "Shell configuration file created"

        if grep -q "NODE_EXTRA_CA_CERTS" "$temp_rc"; then
            print_success "NODE_EXTRA_CA_CERTS added to shell configuration"
            print_info "Shell configuration content:"
            cat "$temp_rc"
        else
            print_error "NODE_EXTRA_CA_CERTS not found in shell configuration"
            cat "$temp_rc"
            exit 1
        fi
    else
        print_error "Shell configuration file not created"
        exit 1
    fi

    # Clean up temp files
    rm -f "$temp_rc" "$temp_config"

    print_success "Node.js certificate configuration test passed"
}

# Main test flow
main() {
    print_header "STRIGO CERTIFICATE MANAGEMENT TESTS"
    echo "Testing certificate injection for JDK and Node.js"
    echo ""

    # Check prerequisites
    check_nexus
    wait_for_nexus

    # Setup
    build_strigo
    setup_test_certificate
    clean_test_dirs

    # Run tests
    test_jdk_certificates
    test_node_certificates

    # Summary
    print_header "Test Summary"
    echo ""
    print_success "╔════════════════════════════════════════════════╗"
    print_success "║                                                ║"
    print_success "║  ✅  ALL CERTIFICATE TESTS PASSED!            ║"
    print_success "║                                                ║"
    print_success "║  Validated:                                    ║"
    print_success "║  - JDK certificate injection (keystore)        ║"
    print_success "║  - JDK default CAs preservation               ║"
    print_success "║  - JDK backup creation                        ║"
    print_success "║  - Node.js metadata storage                   ║"
    print_success "║  - Node.js 'use' command display             ║"
    print_success "║  - Node.js 'use --set-env' configuration     ║"
    print_success "║                                                ║"
    print_success "╚════════════════════════════════════════════════╝"
    echo ""

    print_info "Test artifacts:"
    print_info "   - Installed SDKs: $TEST_INSTALL_DIR"
    print_info "   - Cache: $TEST_CACHE_DIR"
    print_info ""
    print_info "To clean up test artifacts, run:"
    print_info "   rm -rf $TEST_INSTALL_DIR $TEST_CACHE_DIR"
}

# Run tests
main "$@"
