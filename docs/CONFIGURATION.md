# Strigo Configuration Guide

This guide explains how to configure Strigo using the `strigo.toml` configuration file.

## Table of Contents

- [Configuration File Location](#configuration-file-location)
- [General Settings](#general-settings)
- [SDK Types](#sdk-types)
- [Registries](#registries)
- [SDK Repositories](#sdk-repositories)
- [Complete Example](#complete-example)
- [Advanced Configuration](#advanced-configuration)

## Configuration File Location

Strigo looks for `strigo.toml` in the following order:

1. Path specified by `STRIGO_CONFIG_PATH` environment variable
2. `strigo.toml` in the current directory
3. `~/.config/strigo/strigo.toml` (future)

### Using a Custom Config File

```bash
export STRIGO_CONFIG_PATH=/path/to/custom-config.toml
strigo available jdk temurin
```

Or for a single command:

```bash
STRIGO_CONFIG_PATH=/path/to/custom-config.toml strigo install jdk temurin 11.0.24_8
```

## General Settings

The `[general]` section configures global Strigo behavior.

```toml
[general]
log_level = "info"              # Log verbosity: debug, info, warn, error
sdk_install_dir = "~/.sdks"     # Where to install SDKs
cache_dir = "~/.cache/strigo"   # Download cache directory
log_path = ""                   # Optional: log file path (empty = stdout only)
keep_cache = false              # Keep downloaded files after installation
shell_config_path = ""          # Optional: shell config file to update
patterns_file = ""              # Optional: custom patterns file (empty = use builtin)

# JDK-specific settings
jdk_security_path = "lib/security/cacerts"           # Path to cacerts in JDK
system_cacerts_path = "/etc/ssl/certs/ca-certificates.crt"  # System CA certificates
```

### Path Expansion

- Paths starting with `~` are expanded to the user's home directory
- Relative paths are resolved from the current working directory

### Log Levels

- `debug`: Verbose output for troubleshooting
- `info`: Standard informational messages (default)
- `warn`: Warnings only
- `error`: Errors only

### Certificate Management (JDK only)

Strigo automatically configures JDK installations to use system certificates:

- `jdk_security_path`: Relative path to `cacerts` file within JDK installation
- `system_cacerts_path`: Absolute path to system CA certificates file

On installation, Strigo:
1. Removes the default JDK `cacerts` file
2. Creates a symlink to the system certificates

This ensures JDKs use up-to-date system certificates.

## SDK Types

Define SDK types and their installation subdirectories.

```toml
[sdk_types]
jdk = {
    type = "jdk",
    install_dir = "jdks"     # SDKs installed in: {sdk_install_dir}/jdks/
}
node = {
    type = "node",
    install_dir = "nodes"    # SDKs installed in: {sdk_install_dir}/nodes/
}
```

### Installation Structure

With the above configuration and `sdk_install_dir = "~/.sdks"`:

```
~/.sdks/
├── jdks/
│   ├── temurin/
│   │   ├── 11.0.24_8/
│   │   └── 17.0.5_8/
│   └── corretto/
│       └── 8u442b06/
└── nodes/
    └── nodejs/
        └── 20.18.2/
```

## Registries

Registries define where Strigo fetches SDK metadata and files.

### Nexus Registry

```toml
[registries]
nexus = {
    type = "nexus",
    api_url = "http://localhost:8081/service/rest/v1/assets?repository={repository}",
    username = "admin",      # Optional: for authenticated access
    password = "admin"       # Optional: for authenticated access
}
```

**URL Template Variables:**
- `{repository}`: Replaced with the repository name from SDK configuration

**Authentication:**
- If `username` and `password` are provided, Strigo uses HTTP Basic Auth
- Omit both fields for anonymous access

### Multiple Registries

You can define multiple registries:

```toml
[registries]
production = {
    type = "nexus",
    api_url = "https://nexus.company.com/service/rest/v1/assets?repository={repository}",
    username = "ci-user",
    password = "secret"
}

development = {
    type = "nexus",
    api_url = "http://localhost:8081/service/rest/v1/assets?repository={repository}"
}
```

## SDK Repositories

SDK repositories link distributions to registries and define where to find them.

```toml
[sdk_repositories]
temurin = {
    registry = "nexus",                    # Registry name from [registries]
    repository = "raw",                    # Repository name in the registry
    type = "jdk",                          # SDK type from [sdk_types]
    path = "jdk/adoptium/temurin"          # Path prefix in repository
}

corretto = {
    registry = "nexus",
    repository = "raw",
    type = "jdk",
    path = "jdk/amazon/corretto"
}

nodejs = {
    registry = "nexus",
    repository = "raw",
    type = "node",
    path = "node"
}
```

### Path Structure in Repository

Files must follow this structure in your repository:

```
{path}/{major_version}/{version_dir}/{filename}
```

**Example for Temurin JDK 11:**
- Path: `jdk/adoptium/temurin`
- Major version: `11`
- Version directory: `jdk-11.0.24_8`
- Full path: `jdk/adoptium/temurin/11/jdk-11.0.24_8/OpenJDK11U-jdk_x64_linux_hotspot_11.0.24_8.tar.gz`

**Example for Corretto JDK 8 (legacy):**
- Path: `jdk/amazon/corretto`
- Major version: `8`
- Version directory: `jdk-8u442b06`
- Full path: `jdk/amazon/corretto/8/jdk-8u442b06/amazon-corretto-8.442.06.1-linux-x64.tar.gz`

### Version Pattern Matching

Strigo uses pattern files to extract version numbers from paths. See [Custom Patterns](CUSTOM_PATTERNS.md) for details.

## Complete Example

Here's a full configuration file:

```toml
[general]
log_level = "info"
sdk_install_dir = "~/.sdks"
cache_dir = "~/.cache/strigo"
log_path = ""
keep_cache = false
shell_config_path = ""
patterns_file = ""

# JDK certificate configuration
jdk_security_path = "lib/security/cacerts"
system_cacerts_path = "/etc/ssl/certs/ca-certificates.crt"

[sdk_types]
jdk = { type = "jdk", install_dir = "jdks" }
node = { type = "node", install_dir = "nodes" }

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

zulu = {
    registry = "nexus",
    repository = "raw",
    type = "jdk",
    path = "jdk/azul/zulu"
}

mandrel = {
    registry = "nexus",
    repository = "raw",
    type = "jdk",
    path = "jdk/graalvm/mandrel"
}

nodejs = {
    registry = "nexus",
    repository = "raw",
    type = "node",
    path = "node"
}
```

## Advanced Configuration

### Custom Patterns File

To use a custom patterns file instead of the built-in patterns:

```toml
[general]
patterns_file = "/path/to/custom-patterns.toml"
```

Or use an environment variable:

```bash
export STRIGO_PATTERNS_PATH=/path/to/custom-patterns.toml
```

See [Custom Patterns](CUSTOM_PATTERNS.md) for pattern file format.

### Environment Variable Override

You can override the patterns file path:

```bash
STRIGO_PATTERNS_PATH=./my-patterns.toml strigo available jdk temurin
```

### Shell Integration (Future Feature)

Configure `shell_config_path` to automatically update your shell configuration:

```toml
[general]
shell_config_path = "~/.bashrc"
```

This will add SDK bin directories to your PATH automatically.

## Troubleshooting

### Duplicate Keys Error

If you see:

```
❌ Failed to parse config file: The following key was defined twice: sdk_repositories.temurin
💡 Hint: You have duplicate keys in your TOML file. Each repository/registry name must be unique.
```

**Solution**: Check your TOML file for duplicate section keys. Each repository name must be unique.

### Path Not Found Errors

If SDKs aren't found:

1. Verify `path` in `[sdk_repositories]` matches your repository structure
2. Check registry `api_url` is correct
3. Ensure authentication credentials are valid (if required)
4. Use `log_level = "debug"` to see detailed path resolution

### Certificate Errors (JDK)

If Java applications can't connect to HTTPS sites:

1. Verify `system_cacerts_path` points to your system CA bundle
2. Check the file exists: `ls -la /etc/ssl/certs/ca-certificates.crt`
3. Ensure the symlink was created: `ls -la ~/.sdks/jdks/temurin/11.0.24_8/*/lib/security/cacerts`

## See Also

- [Custom Patterns](CUSTOM_PATTERNS.md) - Define version extraction patterns
- [Architecture](ARCHITECTURE.md) - How Strigo works internally
- [Testing Guide](TESTING.md) - Testing your configuration
