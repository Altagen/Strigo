# Strigo Testing Guide

This guide covers all aspects of testing Strigo, from unit tests to end-to-end integration tests.

## Table of Contents

- [Test Organization](#test-organization)
- [Running Tests](#running-tests)
- [Unit Tests](#unit-tests)
- [Integration Tests](#integration-tests)
- [Writing Tests](#writing-tests)
- [CI/CD Integration](#cicd-integration)

## Test Organization

```
tests/
├── unit/                          # Unit tests
│   ├── cache_test.go             # Cache management tests
│   ├── case_insensitive_test.go  # Case-insensitive matching
│   ├── extractor_test.go         # Version extraction tests
│   ├── install_test.go           # Installation path tests
│   ├── inventory_test.go         # Inventory management tests
│   ├── nexus_mock_test.go        # Nexus client tests
│   ├── parser_test.go            # Pattern parser tests
│   └── strigopatterns.toml       # Test patterns
│
└── integration/                   # Integration tests
    ├── README.md                 # Integration test documentation
    ├── test-e2e-real-sdks.sh     # E2E tests with real SDK files
    └── run-integration-tests.sh   # Mock integration tests
```

## Running Tests

### Quick Test (Unit Tests Only)

```bash
# Run all unit tests
go test ./tests/unit/... -v

# Run specific test
go test ./tests/unit/ -run TestExtractVersion -v

# Run with coverage
go test ./tests/unit/... -cover
```

### Integration Tests (Lightweight)

Fast validation using small dummy files:

```bash
cd tests/integration
./run-integration-tests.sh

# Cleanup
./run-integration-tests.sh cleanup
```

**Duration**: ~2-3 minutes
**Network**: Required (downloads small files)
**Disk Space**: ~10MB

### End-to-End Tests (Comprehensive)

Production-ready testing with real SDK downloads:

```bash
cd tests/integration
./test-e2e-real-sdks.sh

# ⚠️ Manual cleanup required when done:
./test-e2e-real-sdks.sh cleanup
```

**Duration**: ~5-10 minutes
**Network**: Required (downloads ~500MB)
**Disk Space**: ~2GB
**Nexus UI**: http://localhost:8081 (admin/admin)

## Unit Tests

### What's Tested

#### Version Extraction (`extractor_test.go`)

Tests the pattern-based version extraction from file paths.

```go
func TestExtractVersion(t *testing.T) {
    tests := []struct {
        name        string
        path        string
        sdkType     string
        wantVersion string
        wantPattern string
    }{
        {
            name:        "Temurin modern format",
            path:        "/jdk/adoptium/temurin/11/jdk-11.0.24_8/file.tar.gz",
            sdkType:     "jdk",
            wantVersion: "11.0.24_8",
            wantPattern: "temurin",
        },
        // More test cases...
    }
}
```

**Patterns Tested:**
- Temurin (modern format): `11.0.24_8`
- Corretto (legacy format): `8u442b06`
- Corretto (modern format): `8.442.06.1`
- Zulu: `17.0.5-17.38.21`
- Mandrel: `24.1.1.0`
- Node.js: `20.18.2`

#### Pattern Parser (`parser_test.go`)

Tests loading and compiling regex patterns from TOML files.

```go
func TestLoadPatterns(t *testing.T) {
    patterns, err := version.LoadPatterns("strigopatterns.toml")
    assert.NoError(t, err)
    assert.NotEmpty(t, patterns)
}
```

#### Cache Management (`cache_test.go`)

Tests cache directory creation and cleanup.

```go
func TestCacheManager(t *testing.T) {
    // Test cache directory creation
    // Test cleanup with keep_cache = true/false
}
```

#### Installation Paths (`install_test.go`)

Tests installation path generation.

```go
func TestGetInstallPath(t *testing.T) {
    path, err := GetInstallPath(cfg, "jdk", "temurin", "11.0.24_8")
    assert.NoError(t, err)
    assert.Equal(t, "/path/to/.sdks/jdks/temurin/11.0.24_8", path)
}
```

### Running Specific Tests

```bash
# Test version extraction only
go test ./tests/unit/ -run TestExtractVersion -v

# Test parser only
go test ./tests/unit/ -run TestLoadPatterns -v

# Test cache management
go test ./tests/unit/ -run TestCacheManager -v

# Run all tests with verbose output
go test ./tests/unit/... -v

# Generate coverage report
go test ./tests/unit/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Integration Tests

### Mock Integration Tests

**Script**: `tests/integration/run-integration-tests.sh`

**What It Tests:**
- ✅ Nexus container startup
- ✅ Repository creation
- ✅ File uploads (small dummy files)
- ✅ Pattern matching
- ✅ `available` command
- ❌ **Does NOT test**: Installation, binaries

**Use Case**: Quick validation during development

```bash
cd tests/integration
./run-integration-tests.sh

# Check logs
tail -f /tmp/strigo-integration-test.log

# Cleanup
./run-integration-tests.sh cleanup
```

### End-to-End Tests

**Script**: `tests/integration/test-e2e-real-sdks.sh`

**What It Tests:**
- ✅ Download real SDKs (~500MB) from official sources
- ✅ Automated Nexus setup (EULA + admin password)
- ✅ Complete installation workflow
- ✅ Certificate management (JDK symlinks)
- ✅ Cache cleanup validation
- ✅ Binary functionality (`java -version`, `node --version`)
- ✅ All SDK patterns (modern + legacy formats)

**SDKs Tested:**
1. Eclipse Temurin JDK 11 (modern format)
2. Amazon Corretto JDK 8 (legacy format)
3. Azul Zulu JDK 17
4. Mandrel 24 (GraalVM-based)
5. Node.js 20 LTS

**Test Steps:**

```
1. Start Nexus container
2. Wait for Nexus initialization
3. Configure Nexus (EULA + admin password)
4. Create raw repository
5. Download real SDK files in parallel
6. Upload SDKs to Nexus
7. Build Strigo
8. Create test configuration
9. Run tests:
   - List available versions (5 distributions)
   - Install each distribution
   - Verify certificate management
   - Verify cache cleanup
   - Test installed binaries
   - List installed distributions
10. Display summary (cleanup is MANUAL)
```

**Output Example:**

```
╔════════════════════════════════════════════════╗
║  ✅  ALL TESTS PASSED SUCCESSFULLY!           ║
║                                                ║
║  Tested:                                       ║
║  - Temurin JDK 11 (modern format)              ║
║  - Corretto JDK 8 (legacy format)              ║
║  - Azul Zulu JDK 17                            ║
║  - Mandrel 24 (GraalVM-based)                  ║
║  - Node.js 20 LTS                              ║
║                                                ║
║  Validated:                                    ║
║  - Authentication (HTTP Basic Auth)            ║
║  - Pattern matching (all formats)              ║
║  - Installation & extraction                   ║
║  - Certificate management (symlinks)           ║
║  - Cache management                            ║
║  - Binary functionality                        ║
║  - Nexus auto-setup (EULA + admin)             ║
╚════════════════════════════════════════════════╝
```

### Inspecting Test Results

After E2E tests complete, you can inspect:

**Nexus UI:**
```
URL: http://localhost:8081
Username: admin
Password: admin
```

Browse to verify:
- Repository structure
- Uploaded files
- Asset metadata

**Test Artifacts:**
```bash
# Installed SDKs
ls -la /tmp/strigo-e2e-test-sdks/

# Downloaded files
ls -la /tmp/strigo-e2e-downloads/

# Cache directory
ls -la /tmp/strigo-e2e-test-cache/
```

**Manual Cleanup:**
```bash
./tests/integration/test-e2e-real-sdks.sh cleanup
```

## Writing Tests

### Unit Test Template

```go
package unit

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestYourFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
        {
            name:     "invalid input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := YourFunction(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### Adding a New Pattern Test

1. Add test case to `tests/unit/extractor_test.go`:

```go
{
    name:        "My new SDK format",
    path:        "/sdk/vendor/product/1/sdk-1.2.3/file.tar.gz",
    sdkType:     "jdk",
    wantVersion: "1.2.3",
    wantPattern: "mynewsdk",
},
```

2. Add pattern to `tests/unit/strigopatterns.toml`:

```toml
[[patterns]]
id = "mynewsdk"
name = "My New SDK"
sdk_type = "jdk"
path_regex = '/sdk/vendor/product/(?P<major>\d+)/sdk-(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)/'
version_format = "{major}.{minor}.{patch}"
```

3. Run test:

```bash
go test ./tests/unit/ -run TestExtractVersion -v
```

### Adding SDK to E2E Tests

Edit `tests/integration/test-e2e-real-sdks.sh`:

1. **Add download** in `download_real_sdks()`:

```bash
# New SDK
print_info "Downloading New SDK..."
curl -L "https://example.com/sdk.tar.gz" \
    -o newsdk.tar.gz 2>/dev/null &
PID_NEWSDK=$!
```

2. **Add upload** in `upload_sdks_to_nexus()`:

```bash
# New SDK upload
print_info "Uploading New SDK..."
curl -u "$NEXUS_USER:$NEXUS_PASSWORD" --upload-file newsdk.tar.gz \
    "$NEXUS_URL/repository/$NEXUS_REPO/sdk/vendor/product/version/file.tar.gz"
```

3. **Add config** in `create_test_config()`:

```toml
newsdk = {
    registry = "nexus",
    repository = "raw",
    type = "jdk",
    path = "sdk/vendor/product"
}
```

4. **Add test** in `run_tests()`:

```bash
print_info "Test: Installing New SDK..."
STRIGO_CONFIG_PATH="$config" ./strigo install jdk newsdk "1.0.0"
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run unit tests
        run: go test ./tests/unit/... -v

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Podman
        run: sudo apt-get install -y podman

      - name: Run integration tests
        run: |
          cd tests/integration
          ./run-integration-tests.sh
```

### Local CI Simulation

```bash
# Run what CI runs
go test ./tests/unit/... -v
cd tests/integration && ./run-integration-tests.sh
```

## Troubleshooting Tests

### Unit Tests Failing

**Pattern not matching:**
- Check regex syntax in pattern file
- Verify test path matches pattern
- Use online regex tester: https://regex101.com/

**File not found:**
- Ensure test files exist in `tests/unit/`
- Check working directory

### Integration Tests Failing

**Port 8081 in use:**
```bash
# Find and kill process
sudo lsof -i :8081
sudo kill -9 <PID>

# Or cleanup old containers
podman stop strigo-nexus-e2e-test
podman rm strigo-nexus-e2e-test
```

**Downloads failing:**
- Check internet connection
- Verify URLs in script are valid
- Try manual download to test

**Nexus won't start:**
```bash
# Check Podman logs
podman logs strigo-nexus-e2e-test

# Increase wait time in script (line 62)
local max_attempts=120  # Instead of 60
```

### Permission Errors

```bash
# Make scripts executable
chmod +x tests/integration/*.sh

# Check file ownership
ls -la tests/integration/
```

## Test Coverage

Generate coverage report:

```bash
# Unit test coverage
go test ./tests/unit/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Open in browser
xdg-open coverage.html  # Linux
open coverage.html      # macOS
```

## See Also

- [Architecture](ARCHITECTURE.md) - Understanding the codebase
- [Integration Test README](../tests/integration/README.md) - Detailed integration test docs
