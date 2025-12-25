#!/bin/bash

# E2E Regression Tests - Verify core Strigo functionality still works
# Focus: installation, certificates, use command, list, remove

set -e

STRIGO_CONFIG="strigo-e2e-test.toml"
STRIGO_BIN="./strigo"
TEST_SDK_DIR="/tmp/strigo-e2e-test-sdks"

echo "🧪 Strigo E2E Regression Tests"
echo "========================================"
echo ""

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass_count=0
fail_count=0

pass() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
    ((pass_count++))
}

fail() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    ((fail_count++))
    exit 1  # Stop on first failure for easier debugging
}

section() {
    echo ""
    echo "================================================"
    echo "$1"
    echo "================================================"
}

# Cleanup before tests
section "🧹 Cleanup"
rm -rf "$TEST_SDK_DIR"
echo "Cleaned up test SDK directory"

# TEST 1: Available command with pagination
section "TEST 1: List Available Versions (with pagination)"
OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN available jdk temurin 2>&1)

if echo "$OUTPUT" | grep -q "8u432b06"; then
    pass "Temurin Java 8 listed"
else
    fail "Temurin Java 8 NOT listed"
fi

if echo "$OUTPUT" | grep -q "17.0.13_11"; then
    pass "Temurin Java 17 listed"
else
    fail "Temurin Java 17 NOT listed"
fi

# TEST 2: Install JDK 17
section "TEST 2: Install JDK 17 (Temurin)"
STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN install jdk temurin 17.0.13_11 --force 2>&1 | tee /tmp/install.log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    pass "Installation command succeeded"
else
    fail "Installation command failed"
fi

# Verify installation directory
JDK_DIR=$(find "$TEST_SDK_DIR/jdks/temurin" -maxdepth 2 -type d -name "*17*" | head -1)
if [ -n "$JDK_DIR" ] && [ -d "$JDK_DIR" ]; then
    pass "JDK 17 directory created: $JDK_DIR"
else
    fail "JDK 17 directory NOT found"
fi

# Verify java binary exists and works
JAVA_BIN=$(find "$JDK_DIR" -name java -type f | head -1)
if [ -n "$JAVA_BIN" ] && [ -x "$JAVA_BIN" ]; then
    pass "Java binary found and executable"

    VERSION=$("$JAVA_BIN" -version 2>&1 | head -1)
    echo "   Java version: $VERSION"

    if echo "$VERSION" | grep -q "17"; then
        pass "Java version is 17"
    else
        fail "Java version is NOT 17: $VERSION"
    fi
else
    fail "Java binary NOT found or not executable"
fi

# TEST 3: Certificate injection verification
section "TEST 3: Certificate Injection"
CACERTS=$(find "$JDK_DIR" -name cacerts -type f | head -1)

if [ -n "$CACERTS" ] && [ -f "$CACERTS" ]; then
    pass "cacerts file found: $CACERTS"

    # Check if it's a valid keystore (not a symlink)
    if [ -L "$CACERTS" ]; then
        fail "cacerts is a symlink (old destructive behavior!)"
    else
        pass "cacerts is a regular file (not symlink)"
    fi

    # Verify custom certificate
    if keytool -list -keystore "$CACERTS" -storepass changeit 2>/dev/null | grep -qi "strigo-test-ca"; then
        pass "Custom certificate 'strigo-test-ca' found in keystore"
    else
        fail "Custom certificate NOT found in keystore"
    fi

    # Verify original CA certificates are still present
    CERT_COUNT=$(keytool -list -keystore "$CACERTS" -storepass changeit 2>/dev/null | grep -c "trustedCertEntry")
    echo "   Total certificates in keystore: $CERT_COUNT"

    if [ "$CERT_COUNT" -ge 140 ]; then
        pass "Original CA certificates preserved (count: $CERT_COUNT)"
    else
        fail "CA certificates missing (count: $CERT_COUNT, expected >=140)"
    fi

    # Check backup was created
    if [ -f "${CACERTS}.original" ]; then
        pass "Backup cacerts.original created"
    else
        fail "Backup cacerts.original NOT found"
    fi
else
    fail "cacerts file NOT found"
fi

# TEST 4: Install another JDK (Java 8)
section "TEST 4: Install JDK 8 (Temurin)"
STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN install jdk temurin 8u432b06 --force >/dev/null 2>&1

if [ $? -eq 0 ]; then
    pass "JDK 8 installation succeeded"
else
    fail "JDK 8 installation failed"
fi

JDK8_DIR=$(find "$TEST_SDK_DIR/jdks/temurin" -maxdepth 2 -type d -name "*8*" | head -1)
if [ -n "$JDK8_DIR" ] && [ -d "$JDK8_DIR" ]; then
    pass "JDK 8 directory created"
else
    fail "JDK 8 directory NOT found"
fi

# TEST 5: List installed SDKs
section "TEST 5: List Installed SDKs"
OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN list 2>&1)

if echo "$OUTPUT" | grep -q "temurin"; then
    pass "'strigo list' shows installed JDKs"
else
    fail "'strigo list' does NOT show installed JDKs"
fi

if echo "$OUTPUT" | grep -q "17.0.13_11"; then
    pass "JDK 17 listed"
else
    fail "JDK 17 NOT listed"
fi

if echo "$OUTPUT" | grep -q "8u432b06"; then
    pass "JDK 8 listed"
else
    fail "JDK 8 NOT listed"
fi

# TEST 6: Use command (switch JDK)
section "TEST 6: Use Command (switch to JDK 17)"
STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN use jdk temurin 17.0.13_11 >/dev/null 2>&1

if [ $? -eq 0 ]; then
    pass "'strigo use' command succeeded"
else
    fail "'strigo use' command failed"
fi

# Verify shell config was updated (if shell config path is set)
# Note: This might not work in non-interactive shell, so we just check the command succeeded

# TEST 7: Install Corretto (different distribution)
section "TEST 7: Install Corretto JDK 17 (different distribution)"
STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN available jdk corretto >/dev/null 2>&1

if [ $? -eq 0 ]; then
    pass "Corretto versions available"
else
    fail "Corretto versions NOT available"
fi

STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN install jdk corretto 17.0.13.11.1 --force >/dev/null 2>&1

if [ $? -eq 0 ]; then
    pass "Corretto 17 installation succeeded"
else
    fail "Corretto 17 installation failed"
fi

CORRETTO_DIR=$(find "$TEST_SDK_DIR/jdks/corretto" -maxdepth 2 -type d -name "*17*" | head -1)
if [ -n "$CORRETTO_DIR" ] && [ -d "$CORRETTO_DIR" ]; then
    pass "Corretto 17 directory created"
else
    fail "Corretto 17 directory NOT found"
fi

# TEST 8: Remove JDK
section "TEST 8: Remove JDK 8"
STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN remove jdk temurin 8u432b06 --force >/dev/null 2>&1

if [ $? -eq 0 ]; then
    pass "'strigo remove' command succeeded"
else
    fail "'strigo remove' command failed"
fi

# Verify JDK 8 is removed
if [ ! -d "$JDK8_DIR" ]; then
    pass "JDK 8 directory removed"
else
    fail "JDK 8 directory still exists after remove"
fi

# Verify JDK 17 is still present
if [ -d "$JDK_DIR" ]; then
    pass "JDK 17 still present (not removed)"
else
    fail "JDK 17 was incorrectly removed"
fi

# TEST 9: Pagination stress test (verify all items retrieved)
section "TEST 9: Pagination - Verify All Items Retrieved"
OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG STRIGO_LOG_LEVEL=DEBUG $STRIGO_BIN available jdk temurin 2>&1)

if echo "$OUTPUT" | grep -q "Fetching page 2"; then
    pass "Pagination triggered (>100 items)"
else
    fail "Pagination NOT triggered"
fi

if echo "$OUTPUT" | grep -q "Total items: 113"; then
    pass "All 113 items retrieved from Nexus"
else
    fail "Not all items retrieved from Nexus"
fi

# Verify only unique versions are shown (no duplicates)
VERSION_COUNT=$(echo "$OUTPUT" | grep -o "17.0.13_11" | wc -l)
if [ "$VERSION_COUNT" -eq 1 ]; then
    pass "No duplicate versions (deduplication works)"
else
    fail "Duplicate versions found (count: $VERSION_COUNT)"
fi

# TEST 10: Filtering accuracy
section "TEST 10: Filtering - Verify Cross-Distribution Isolation"
TEMURIN_OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN available jdk temurin 2>&1)
CORRETTO_OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN available jdk corretto 2>&1)

# Temurin should NOT have Corretto versions
if echo "$TEMURIN_OUTPUT" | grep -q "8.472.08.1"; then
    fail "Corretto version leaked into Temurin results"
else
    pass "Temurin results clean (no Corretto versions)"
fi

# Corretto should NOT have Temurin versions
if echo "$CORRETTO_OUTPUT" | grep -q "8u432b06"; then
    fail "Temurin version leaked into Corretto results"
else
    pass "Corretto results clean (no Temurin versions)"
fi

# Summary
section "📊 Test Summary"
echo -e "${GREEN}Passed: $pass_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"
echo ""

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}✅ All regression tests passed!${NC}"
    echo ""
    echo "Core Strigo functionality verified:"
    echo "  • List available versions (with pagination)"
    echo "  • Install JDKs (multiple distributions)"
    echo "  • Certificate injection (non-destructive)"
    echo "  • List installed SDKs"
    echo "  • Use command (switch JDK)"
    echo "  • Remove JDKs"
    echo "  • Filtering (distribution isolation)"
    echo ""
    exit 0
else
    echo -e "${RED}❌ Regression detected!${NC}"
    exit 1
fi
