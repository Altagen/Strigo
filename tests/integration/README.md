# Strigo Integration Tests

This directory contains integration tests for Strigo with different configurations.

## 📋 Test Scripts

### 1. `test-e2e-real-sdks.sh` ⭐ **Recommended for Comprehensive Testing**

**End-to-end tests with real JDK/SDK files from official sources**

This script performs complete production-ready testing with real SDK downloads and installations.

**Features:**
- ✅ Downloads real SDK files (~500MB total)
- ✅ Automated Nexus setup (EULA + admin password)
- ✅ Tests complete installation workflow
- ✅ Validates certificate management
- ✅ Verifies cache cleanup
- ✅ Tests binary functionality (`java -version`)
- ✅ **No hardcoded user paths** - works on any system
- ✅ Automatic cleanup of all resources

**Tested SDKs:**
- Eclipse Temurin JDK 11 (modern format)
- Amazon Corretto JDK 8 (legacy format)
- Azul Zulu JDK 17
- Mandrel 24 (GraalVM-based)
- Node.js 20 LTS

**Usage:**
```bash
cd tests/integration
./test-e2e-real-sdks.sh

# ⚠️ IMPORTANT: Cleanup is MANUAL (not automatic)
# This allows you to inspect Nexus and test artifacts after the tests
# Nexus UI: http://localhost:8081 (admin/admin)

# When you're done inspecting, cleanup with:
./test-e2e-real-sdks.sh cleanup
```

**Duration:** 5-10 minutes (network-dependent)

---

### 2. `run-integration-tests.sh` - Lightweight Mock Tests

**Fast integration tests with small dummy files**

For quick validation using small dummy files instead of real SDK downloads.

**Features:**
- ✅ Small dummy files (~1KB each)
- ✅ Fast execution (~2-3 minutes)
- ✅ Tests pattern matching
- ✅ Validates `available` command
- ❌ Does not test installation
- ❌ Does not test binaries

**Usage:**
```bash
cd tests/integration
./run-integration-tests.sh

# Cleanup
./run-integration-tests.sh cleanup
```

---

## 🚀 Quick Start

**For comprehensive validation (recommended):**
```bash
cd tests/integration
./test-e2e-real-sdks.sh

# Nexus UI will be available at: http://localhost:8081 (admin/admin)
# Cleanup manually when done: ./test-e2e-real-sdks.sh cleanup
```

**For quick pattern validation:**
```bash
cd tests/integration
./run-integration-tests.sh
```

## 📦 Prerequisites

- **Podman** (for running Nexus container)
- **Internet connection** (for downloading SDK files)
- **~2GB free disk space** (for real SDK tests)
- **Go 1.21+** (to build Strigo)

## 🧪 What Gets Tested

### End-to-End Tests (`test-e2e-real-sdks.sh`)

✅ **Download & Upload**: Real SDKs from official sources
✅ **Nexus Setup**: Automated EULA acceptance + admin password
✅ **Authentication**: HTTP Basic Auth with Nexus
✅ **Available Command**: List versions for all distributions
✅ **Install Command**: Complete installation with extraction
✅ **Pattern Matching**: All supported formats (modern + legacy)
✅ **Certificate Management**: Automatic symlink creation for JDKs
✅ **Cache Management**: Automatic cleanup when `keep_cache=false`
✅ **Binary Functionality**: Execute `java -version` and `node --version`
✅ **List Command**: Display installed distributions

### Mock Tests (`run-integration-tests.sh`)

✅ **Connection to Nexus**: Real HTTP calls (not mocked)
✅ **Version listing**: Pattern detection
✅ **Version filtering**: Major version filters
❌ Installation not tested

## 📊 Test Output Example

```
========================================
STRIGO END-TO-END INTEGRATION TESTS
========================================
Testing with: Temurin, Corretto, Azul Zulu, Mandrel, Node.js
Using real SDK files from official sources
Automated: Nexus setup (EULA + admin password)

========================================
STEP 1/10: Starting Nexus Container
========================================
✅ Nexus container started

========================================
STEP 2/10: Waiting for Nexus to Initialize
========================================
✅ Nexus is ready!

[... downloads, uploads, installations ...]

╔════════════════════════════════════════════════╗
║                                                ║
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
║                                                ║
╚════════════════════════════════════════════════╝
```

## 🗂️ Test Environment Isolation

All tests use temporary directories under `/tmp` to ensure:
- ✅ **No interference** with production installations
- ✅ **No hardcoded user paths** - works on any system
- ✅ **Easy cleanup** - everything in `/tmp`
- ✅ **Parallel execution** - multiple runs won't conflict

**Test directories:**
- `/tmp/strigo-e2e-downloads/` - Downloaded SDK files
- `/tmp/strigo-e2e-test-sdks/` - Installed SDKs
- `/tmp/strigo-e2e-test-cache/` - Download cache
- `/tmp/strigo-test/` - Mock test files (lightweight script)

## 🛠️ Troubleshooting

### Port 8081 already in use
```bash
# Cleanup existing Nexus
./test-e2e-real-sdks.sh cleanup

# Or check what's using the port
sudo netstat -tlnp | grep 8081
sudo lsof -i :8081
```

### Nexus container already exists
```bash
podman stop strigo-nexus-e2e-test
podman rm strigo-nexus-e2e-test
```

### Downloads failing
- Check internet connection
- CDNs might be rate-limiting, try again later
- Verify URLs in script are still valid
- GitHub releases might have moved

### Permission errors
```bash
chmod +x tests/integration/*.sh
```

### Nexus takes too long
- First startup can take 2-3 minutes
- Check logs: `podman logs strigo-nexus-e2e-test`
- Wait for "Started Sonatype Nexus" message

## 🔄 CI/CD Integration

Both scripts are CI-ready with proper exit codes and automatic cleanup:

```yaml
# GitHub Actions example
- name: Run comprehensive tests
  run: |
    cd tests/integration
    ./test-e2e-real-sdks.sh

# Or for quick validation
- name: Run quick tests
  run: |
    cd tests/integration
    ./run-integration-tests.sh
```

## 📝 Development

### Adding a new SDK to tests

**In `test-e2e-real-sdks.sh`:**

1. Add download in `download_real_sdks()`:
```bash
# New SDK
print_info "Downloading New SDK..."
curl -L "https://example.com/sdk.tar.gz" \
    -o newsdk.tar.gz 2>/dev/null &
PID_NEWSDK=$!
```

2. Add upload in `upload_sdks_to_nexus()`:
```bash
# New SDK upload
print_info "Uploading New SDK..."
curl -u "$NEXUS_USER:$NEXUS_PASSWORD" --upload-file newsdk.tar.gz \
    "$NEXUS_URL/repository/$NEXUS_REPO/sdk/vendor/product/version/file.tar.gz"
```

3. Add to configuration in `create_test_config()`:
```toml
newsdk = {
    registry = "nexus",
    repository = "raw",
    type = "sdk",
    path = "sdk/vendor/product"
}
```

4. Add test in `run_tests()`:
```bash
print_info "Test: Installing New SDK..."
STRIGO_CONFIG_PATH="$config" ./strigo install sdk newsdk "1.0.0"
```

### Adding a new pattern

Update `repository/version/patterns/builtin.toml` with your pattern, then test:

```bash
./test-e2e-real-sdks.sh
```

## 🎯 Best Practices

✅ Always use `/tmp` for test data
✅ Avoid hardcoded paths with usernames
✅ Use dynamic version extraction
✅ Implement cleanup in trap handlers
✅ Provide clear progress messages
✅ Test with real files for production validation
✅ Use dummy files for quick pattern validation

## 📚 Related Documentation

- Unit tests: `../unit/`
- Patterns: `../../repository/version/patterns/builtin.toml`
- Main docs: `../../README.md`
- Configuration: `../../strigo.toml`

## 🔒 Security Notes

- Tests use temporary admin password (`admin/admin`)
- Nexus runs in isolated container
- All credentials are local to test environment
- No sensitive data persists after cleanup

## 📄 License

Same as main Strigo project.
