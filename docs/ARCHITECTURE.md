# Strigo Architecture

This document describes the architecture and design of Strigo.

## Table of Contents

- [Overview](#overview)
- [Project Structure](#project-structure)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Design Decisions](#design-decisions)
- [Adding New Features](#adding-new-features)

## Overview

Strigo is a modular SDK version manager designed with the following principles:

- **Separation of Concerns**: Each package has a single, well-defined responsibility
- **Registry-Agnostic**: Support for multiple registry types (currently Nexus, extensible)
- **Pattern-Based**: Flexible version extraction using regex patterns
- **Cache-Aware**: Efficient download management with optional caching

## Project Structure

```
strigo/
├── cmd/                    # CLI commands (Cobra)
│   ├── available.go       # List available versions
│   ├── clean.go           # Clean cache
│   ├── install.go         # Install SDK versions
│   ├── json.go            # JSON output
│   ├── list.go            # List installed versions
│   └── root.go            # Root command setup
│
├── config/                 # Configuration management
│   └── config.go          # TOML config loading & validation
│
├── repository/             # Registry abstraction layer
│   ├── fetcher.go         # Generic asset fetching
│   ├── nexus.go           # Nexus-specific implementation with pagination
│   ├── types.go           # Common types (SDKAsset, etc.)
│   └── version/           # Version extraction
│       ├── parser.go      # Pattern-based version parsing
│       └── extractor.go   # Version extraction logic
│
├── downloader/             # Download & installation
│   ├── manager.go         # Orchestrates download process
│   ├── extract.go         # Archive extraction
│   ├── types.go           # Download options
│   ├── network/
│   │   └── client.go      # HTTP client with auth
│   ├── cache/
│   │   └── manager.go     # Cache management
│   ├── core/
│   │   ├── types.go       # Core types
│   │   └── validation.go  # Disk space validation
│   └── jdk/
│       └── certificates.go # JDK certificate management
│
├── logging/                # Structured logging
│   └── logger.go          # Logger implementation
│
├── tests/                  # Test suites
│   ├── unit/              # Unit tests
│   └── integration/       # Integration tests
│       ├── test-e2e-real-sdks.sh    # E2E tests with real SDKs
│       ├── test-regression-e2e.sh   # Regression tests
│       ├── test-pagination-e2e.sh   # Pagination tests
│       ├── run-integration-tests.sh # Mock integration tests
│       ├── strigo-e2e-test.toml     # E2E test configuration
│       └── strigo-test-variants.toml # Variant repositories example
│
├── examples/               # Configuration examples
│   ├── strigo-basic.toml           # Minimal configuration
│   ├── strigo-with-certificates.toml # With custom certificates
│   ├── strigo-multi-sdk.toml       # Multiple SDK types
│   └── README.md                    # Examples documentation
│
└── strigo-patterns.toml    # Version extraction patterns
```

## Core Components

### 1. Command Layer (`cmd/`)

**Responsibility**: CLI interface using Cobra framework

```
User → Cobra Command → Handler Function → Core Logic
```

**Key Files:**
- `root.go`: Sets up Cobra app, loads config, initializes logging
- `install.go`: Handles SDK installation workflow
- `available.go`: Lists versions from registry
- `list.go`: Lists locally installed versions

**Example Flow (install command):**
```go
// cmd/install.go
func install(cmd *cobra.Command, args []string) {
    sdkType, distribution, version := args[0], args[1], args[2]
    handleInstall(sdkType, distribution, version)  // Delegates to handler
}
```

### 2. Configuration (`config/`)

**Responsibility**: Load, validate, and provide configuration

```
TOML File → Unmarshal → Validate → Config Struct
```

**Key Features:**
- Path expansion (~ → home directory)
- Environment variable support (`STRIGO_CONFIG_PATH`)
- Validation of required fields
- Clear error messages for common issues (duplicate keys, missing files)

**Important Functions:**
- `LoadConfig()`: Main entry point
- `Validate()`: Checks configuration validity
- `EnsureDirectoriesExist()`: Creates required directories

### 3. Repository Layer (`repository/`)

**Responsibility**: Abstract registry access and version extraction

```
SDK Request → Registry API → Parse Response → Extract Versions → Return Assets
```

**Key Components:**

#### `fetcher.go` - Generic Interface
```go
type SDKAsset struct {
    Version     string
    DownloadUrl string
    Path        string
}

func FetchAvailableVersions(
    sdkRepo SDKRepository,
    registry Registry,
    versionFilter string,
    removeDisplay bool,
    patternsFile string,
) ([]SDKAsset, error)
```

#### `nexus.go` - Nexus Implementation
- Constructs Nexus API URLs with continuationToken pagination
- Handles authentication (HTTP Basic Auth)
- Parses JSON responses across multiple pages
- Client-side filtering by path prefix (strict `HasPrefix` matching)
- Accumulates all items before processing (handles >100 items per repository)

#### `version/` - Version Extraction
- **`parser.go`**: Loads and manages regex patterns from external file
- **`extractor.go`**: Extracts versions from file paths
- Patterns loaded from `strigo-patterns.toml` (configurable via `patterns_file` setting)

**Pattern Matching Flow:**
```
File Path → Try Each Pattern → Match? → Extract Version Groups → Build Version String
```

### 4. Downloader Layer (`downloader/`)

**Responsibility**: Download, extract, and configure SDKs

```
Download URL → Cache Check → Download → Validate → Extract → Configure
```

**Key Components:**

#### `manager.go` - Orchestrator
```go
func (m *Manager) DownloadAndExtract(opts DownloadOptions) error {
    // 1. Check file size & disk space
    // 2. Prepare cache directory
    // 3. Download file (with progress)
    // 4. Extract archive
    // 5. Clean cache (if configured)
    // 6. Configure certificates (JDK only)
}
```

#### `network/client.go` - HTTP Client
- HTTP Basic Auth support
- Progress reporting
- File size checking

#### `cache/manager.go` - Cache Management
- Creates cache directory structure
- Cleanup based on `keep_cache` setting

#### `jdk/certificates.go` - JDK Certificate Management
- **Optional**: Only runs if `custom_certificates` configured
- **Non-destructive**: Adds certificates to existing keystore (preserves JDK default CAs)
- **Auto-detection**: Finds `cacerts` automatically (Java 8: `jre/lib/security/cacerts`, Java 11+: `lib/security/cacerts`)
- **Backup**: Creates `cacerts.original` before modification
- **Format**: PEM files with explicit aliases for security audits
- **Configurable password**: Supports custom keystore passwords or password-less PKCS12
- Uses `keystore-go` library for safe keystore manipulation

### 5. Logging (`logging/`)

**Responsibility**: Structured logging with levels

**Features:**
- Log levels: DEBUG, INFO, WARN, ERROR
- Color-coded output
- File and stdout logging
- Pre-initialization logging (`PreLog`) for early errors

## Data Flow

### Installation Flow

```
User Command: strigo install jdk temurin 11.0.24_8
    ↓
1. Parse Arguments (cmd/install.go)
    ↓
2. Load Config (config/config.go)
    ↓
3. Fetch Available Versions (repository/fetcher.go)
    ├→ Call Nexus API (repository/nexus.go)
    ├→ Parse Response
    └→ Extract Versions (repository/version/extractor.go)
    ↓
4. Find Matching Version
    ↓
5. Download & Extract (downloader/manager.go)
    ├→ Download File (downloader/network/client.go)
    ├→ Cache File (downloader/cache/manager.go)
    ├→ Extract Archive (downloader/extract.go)
    └→ Configure Certificates (downloader/jdk/certificates.go)
    ↓
6. Report Success
```

### Available Versions Flow

```
User Command: strigo available jdk temurin
    ↓
1. Parse Arguments (cmd/available.go)
    ↓
2. Load Config
    ↓
3. Fetch All Versions (repository/fetcher.go)
    ├→ Call Registry API
    ├→ Filter by path prefix
    └→ Extract versions from all matching files
    ↓
4. Group by Major Version
    ↓
5. Display (Formatted Output)
```

## Design Decisions

### Why Patterns File?

**Problem**: Different SDK distributions use different version naming schemes:
- Temurin: `11.0.24_8`
- Corretto: `8u442b06`  or `8.442.06.1`
- Zulu: `17.0.5-17.38.21`

**Solution**: Regex patterns in TOML file
- Extensible: Add new distributions without code changes
- User-customizable: Override with `patterns_file` config
- Type-specific: Different patterns for JDK, Node, etc.

### Why Repository Abstraction?

**Problem**: Need to support multiple registry types (Nexus, Artifactory, custom APIs)

**Solution**: `repository/fetcher.go` provides generic interface
- Registry-specific logic in separate files (`nexus.go`, future: `artifactory.go`)
- Common types (`SDKAsset`) used throughout
- Easy to add new registry types

### Why Separate Network Layer?

**Problem**: HTTP operations (auth, progress, retries) are complex

**Solution**: Dedicated `network/client.go`
- Encapsulates HTTP logic
- Reusable across different download scenarios
- Easy to add features (retries, timeouts, proxies)

### Why Cache Management?

**Problem**: SDK files are large (100MB+), re-downloading is slow

**Solution**: Optional cache with cleanup
- Configurable with `keep_cache` setting
- Structured: `cache/{sdk_type}/{distribution}/{version}/`
- Automatic cleanup after successful installation

### Why Certificate Management for JDK?

**Problem**: JDK ships with outdated CA certificates

**Solution**: Symlink to system certificates
- JDK always uses up-to-date system certs
- No manual certificate updates needed
- Transparent to Java applications

## Adding New Features

### Adding a New Registry Type

1. **Create registry implementation** in `repository/`
   ```go
   // repository/artifactory.go
   func FetchFromArtifactory(config Registry, sdkRepo SDKRepository) ([]Item, error) {
       // Implement Artifactory-specific API calls
   }
   ```

2. **Update `FetchAvailableVersions()`** in `repository/fetcher.go`
   ```go
   switch registry.Type {
   case "nexus":
       return FetchFromNexus(...)
   case "artifactory":
       return FetchFromArtifactory(...)
   }
   ```

3. **Update config validation** in `config/config.go`
   ```go
   if registry.Type != "nexus" && registry.Type != "artifactory" {
       return fmt.Errorf("unsupported registry type: %s", registry.Type)
   }
   ```

### Adding a New SDK Type

1. **Add SDK type** to config:
   ```toml
   [sdk_types]
   python = { type = "python", install_dir = "pythons" }
   ```

2. **Add patterns** for version extraction:
   ```toml
   [[patterns]]
   id = "python"
   name = "Python"
   sdk_type = "python"
   path_regex = 'python/(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)'
   ```

3. **Add post-installation logic** (if needed) in `cmd/install.go`:
   ```go
   if sdkType == "python" {
       // Configure Python-specific settings
   }
   ```

### Adding a New Command

1. **Create command file** in `cmd/`:
   ```go
   // cmd/use.go
   var useCmd = &cobra.Command{
       Use:   "use [type] [distribution] [version]",
       Short: "Switch to a specific SDK version",
       Run:   use,
   }
   ```

2. **Register command** in `cmd/root.go`:
   ```go
   func init() {
       rootCmd.AddCommand(useCmd)
   }
   ```

3. **Implement handler**:
   ```go
   func use(cmd *cobra.Command, args []string) {
       // Implementation
   }
   ```

## Testing Strategy

See [Testing Guide](TESTING.md) for details on:
- Unit tests (`tests/unit/`)
- Integration tests (`tests/integration/`)
- E2E tests with real SDKs

## See Also

- [Configuration Guide](CONFIGURATION.md) - Configuration reference
- [Custom Patterns](CUSTOM_PATTERNS.md) - Pattern syntax
- [Testing Guide](TESTING.md) - Running tests
