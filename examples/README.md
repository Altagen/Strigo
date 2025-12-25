# Strigo Configuration Examples

This directory contains example configuration files to help you get started with Strigo.

## Available Examples

### 1. Basic Configuration (`strigo-basic.toml`)

Minimal configuration for JDK management with Temurin and Corretto distributions.

**Use case**: Getting started with Strigo, simple JDK version management.

```bash
cp examples/strigo-basic.toml strigo.toml
# Edit strigo.toml with your Nexus URL and credentials
strigo available jdk temurin
```

### 2. Custom Certificates (`strigo-with-certificates.toml`)

Configuration with custom CA certificate injection into JDK keystores.

**Use case**: Corporate environments with internal CAs, custom PKI infrastructure.

**Features**:
- Inject multiple custom certificates with explicit aliases
- Non-destructive: preserves default JDK CA certificates
- Automatic backup of original keystore
- Configurable keystore password

```bash
cp examples/strigo-with-certificates.toml strigo.toml
# Edit certificate paths and aliases
strigo install jdk temurin 17.0.13_11
```

**Important**: Replace certificate paths and aliases with your own:
```toml
custom_certificates = [
    { path = "/path/to/your/root-ca.pem", alias = "my-root-ca" },
    { path = "/path/to/your/intermediate-ca.pem", alias = "my-intermediate-ca" }
]
```

### 3. Multi-SDK Configuration (`strigo-multi-sdk.toml`)

Full configuration with multiple JDK distributions and Node.js.

**Use case**: Development environments requiring multiple SDK types.

**Includes**:
- 4 JDK distributions: Temurin, Corretto, Zulu, Mandrel
- Node.js support
- Organized directory structure

```bash
cp examples/strigo-multi-sdk.toml strigo.toml
strigo available jdk      # List all JDK distributions
strigo available node     # List Node.js versions
```

## Configuration Guidelines

### Required Fields

All configuration files must include:
- `patterns_file`: Path to version extraction patterns (use `strigo-patterns.toml` from project root)
- At least one `[sdk_types]` definition
- At least one `[registries]` definition
- At least one `[sdk_repositories]` definition

### Optional Certificate Configuration

If you need to inject custom certificates into JDKs:

1. **Use explicit aliases** for security audits and tracking
2. **One certificate per file** (PEM format only)
3. **Password configuration**:
   - Default: `"changeit"` (standard JDK password)
   - Custom: Set `jdk_cacerts_password` to your password
   - Password-less (Java 17+): Use `jdk_cacerts_password = ""`

### Path Expansion

- `~` expands to user home directory: `~/.sdks` → `/home/user/.sdks`
- Relative paths are resolved from working directory
- Absolute paths are used as-is

## Testing Your Configuration

After creating your configuration file:

```bash
# Test basic connectivity
strigo available jdk

# Install a test JDK
strigo install jdk temurin 17.0.13_11

# Verify certificate injection (if configured)
keytool -list -keystore ~/.sdks/jdks/temurin/17.0.13_11/*/lib/security/cacerts -storepass changeit | grep your-alias

# List installed SDKs
strigo list
```

## Environment Variables

Override configuration paths:

```bash
# Use custom config file
export STRIGO_CONFIG_PATH=/path/to/custom-config.toml

# Use custom patterns file
export STRIGO_PATTERNS_PATH=/path/to/custom-patterns.toml

# Run command with custom config (one-time)
STRIGO_CONFIG_PATH=./my-config.toml strigo install jdk temurin 17.0.13_11
```

## See Also

- [Configuration Guide](../docs/CONFIGURATION.md) - Detailed configuration reference
- [Custom Patterns](../docs/CUSTOM_PATTERNS.md) - Version extraction patterns
- [Architecture](../docs/ARCHITECTURE.md) - How Strigo works internally
