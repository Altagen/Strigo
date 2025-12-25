#!/bin/bash

# End-to-End tests for Nexus pagination and filtering
# Tests: installation, listing, pagination, filtering, certificates

set -e

STRIGO_CONFIG="strigo-e2e-test.toml"
STRIGO_BIN="./strigo"

echo "🧪 Starting E2E Tests for Pagination & Filtering"
echo "================================================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass_count=0
fail_count=0

pass() {
    echo -e "${GREEN}✅ PASS${NC}: $1"
    ((pass_count++))
}

fail() {
    echo -e "${RED}❌ FAIL${NC}: $1"
    ((fail_count++))
}

info() {
    echo -e "${YELLOW}ℹ️  INFO${NC}: $1"
}

# Test 1: List versions with pagination (>100 items)
echo ""
echo "TEST 1: Pagination - List Temurin versions (>100 items in Nexus)"
echo "-------------------------------------------------------------------"
OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG STRIGO_LOG_LEVEL=DEBUG $STRIGO_BIN available jdk temurin 2>&1)

# Check pagination logs
if echo "$OUTPUT" | grep -q "Fetching page 1"; then
    pass "Page 1 fetched"
else
    fail "Page 1 not found in logs"
fi

if echo "$OUTPUT" | grep -q "Fetching page 2"; then
    pass "Page 2 fetched (pagination triggered)"
else
    fail "Page 2 not found (pagination not working?)"
fi

if echo "$OUTPUT" | grep -q "Pagination complete. Total items: 113"; then
    pass "All 113 items retrieved"
else
    fail "Total items count incorrect"
fi

# Check versions are displayed
if echo "$OUTPUT" | grep -q "8u432b06"; then
    pass "Java 8 version found"
else
    fail "Java 8 version missing"
fi

if echo "$OUTPUT" | grep -q "17.0.13_11"; then
    pass "Java 17 version found"
else
    fail "Java 17 version missing"
fi

# Test 2: Filtering - Ensure only Temurin versions are shown
echo ""
echo "TEST 2: Filtering - Verify only Temurin versions (not Corretto, Zulu, etc.)"
echo "-----------------------------------------------------------------------------"
OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN available jdk temurin 2>&1)

# Should NOT contain Corretto versions
if echo "$OUTPUT" | grep -q "8.472.08.1"; then
    fail "Corretto version found (filtering broken!)"
else
    pass "Corretto versions correctly filtered out"
fi

# Should NOT contain Zulu versions
if echo "$OUTPUT" | grep -q "zulu"; then
    fail "Zulu version found (filtering broken!)"
else
    pass "Zulu versions correctly filtered out"
fi

# Test 3: Verify deduplication (no duplicate versions)
echo ""
echo "TEST 3: Deduplication - Ensure no duplicate versions displayed"
echo "---------------------------------------------------------------"
OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN available jdk temurin 2>&1)

# Count occurrences of version 17.0.13_11 (uploaded 11 times: original + v1-v10)
COUNT=$(echo "$OUTPUT" | grep -o "17.0.13_11" | wc -l)
if [ "$COUNT" -eq 1 ]; then
    pass "Version 17.0.13_11 appears only once (deduplication works)"
else
    fail "Version 17.0.13_11 appears $COUNT times (should be 1)"
fi

# Test 4: Installation from paginated results
echo ""
echo "TEST 4: Installation - Install JDK from paginated results"
echo "----------------------------------------------------------"
STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN install jdk temurin 17.0.13_11 --force

if [ $? -eq 0 ]; then
    pass "JDK 17 installation successful"
else
    fail "JDK 17 installation failed"
fi

# Verify installation directory exists
if [ -d "/tmp/strigo-e2e-test-sdks/jdks/temurin/17.0.13_11" ]; then
    pass "JDK installation directory created"
else
    fail "JDK installation directory not found"
fi

# Test 5: Certificate injection (regression test)
echo ""
echo "TEST 5: Certificate Injection - Verify custom cert injected"
echo "------------------------------------------------------------"
CACERTS_PATH=$(find /tmp/strigo-e2e-test-sdks/jdks/temurin/17.0.13_11 -name cacerts | head -1)

if [ -z "$CACERTS_PATH" ]; then
    fail "cacerts file not found"
else
    info "Found cacerts at: $CACERTS_PATH"

    # Check if custom certificate is in the keystore
    CERT_COUNT=$(keytool -list -keystore "$CACERTS_PATH" -storepass changeit 2>/dev/null | grep -i "strigo-test-ca" | wc -l)

    if [ "$CERT_COUNT" -ge 1 ]; then
        pass "Custom certificate 'strigo-test-ca' found in keystore"
    else
        fail "Custom certificate not found in keystore"
    fi

    # Check backup was created
    if [ -f "${CACERTS_PATH}.original" ]; then
        pass "Backup cacerts.original created"
    else
        fail "Backup cacerts.original not found"
    fi
fi

# Test 6: List all distributions
echo ""
echo "TEST 6: List Distributions - Verify all configured distributions available"
echo "---------------------------------------------------------------------------"
OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN available jdk 2>&1)

for dist in temurin corretto zulu mandrel; do
    if echo "$OUTPUT" | grep -q "$dist"; then
        pass "Distribution '$dist' found"
    else
        fail "Distribution '$dist' missing"
    fi
done

# Test 7: Verify Corretto versions (separate test for filtering)
echo ""
echo "TEST 7: Corretto Versions - Verify Corretto filtering works independently"
echo "--------------------------------------------------------------------------"
OUTPUT=$(STRIGO_CONFIG_PATH=$STRIGO_CONFIG $STRIGO_BIN available jdk corretto 2>&1)

if echo "$OUTPUT" | grep -q "8.472.08.1"; then
    pass "Corretto 8 version found"
else
    fail "Corretto 8 version missing"
fi

# Should NOT contain Temurin versions
if echo "$OUTPUT" | grep -q "8u432b06"; then
    fail "Temurin version found in Corretto results (filtering broken!)"
else
    pass "Temurin versions correctly filtered out from Corretto"
fi

# Summary
echo ""
echo "================================================"
echo "📊 Test Summary"
echo "================================================"
echo -e "${GREEN}Passed: $pass_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"
echo ""

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
