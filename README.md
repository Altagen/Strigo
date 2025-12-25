# Strigo - SDK Version Manager

Strigo is a lightweight and efficient CLI tool for managing SDKs locally. Install, remove, and switch between multiple versions of Java Development Kits (JDKs) and other SDKs with ease.

![Strigo Logo](assets/img/strigo.jpeg)

---

## 🚀 Quick Start

### Installation

#### From Source
```bash
git clone https://github.com/your-username/strigo.git
cd strigo
task build  # or: go build -o bin/strigo
```

#### From Releases
Download the appropriate binary from the [releases page](https://github.com/Altagen/Strigo/releases).

### Basic Usage

```bash
# List available JDK versions
strigo available jdk temurin

# Install a specific version
strigo install jdk temurin 17.0.13_11

# List installed SDKs
strigo list

# Switch to a version
strigo use jdk temurin 17.0.13_11

# Remove a version
strigo remove jdk temurin 17.0.13_11
```

---

## ✨ Features

- **Multiple Distributions**: Supports Temurin, Corretto, Zulu, Mandrel, and more
- **Nexus Integration**: Fetch SDKs from your Nexus repository
- **Custom Certificates**: Inject corporate CA certificates into JDK keystores (optional)
- **Flexible Configuration**: TOML-based configuration with pattern matching
- **Shell Integration**: Automatic environment variable management
- **Cross-Platform**: Linux and macOS support (amd64 and arm64)
- **Pagination Support**: Handle large repositories with 100+ SDK versions
- **SBOM Included**: Each release includes a Software Bill of Materials for security audits

---

## ⚙️ Configuration

Create a `strigo.toml` configuration file:

```toml
[general]
sdk_install_dir = "~/.sdks"
cache_dir = "~/.cache/strigo"
patterns_file = "strigo-patterns.toml"

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
```

**For complete configuration options and examples:**
- 📖 [Configuration Guide](docs/CONFIGURATION.md) - Detailed configuration reference
- 📁 [Configuration Examples](examples/) - Ready-to-use configuration files

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [Configuration Guide](docs/CONFIGURATION.md) | Detailed configuration options, certificate management, troubleshooting |
| [Custom Patterns](docs/CUSTOM_PATTERNS.md) | Define version extraction patterns for custom repositories |
| [Architecture](docs/ARCHITECTURE.md) | Internal architecture and design decisions |
| [Testing Guide](docs/TESTING.md) | Running tests and validating your setup |
| [Examples](examples/) | Configuration examples for different use cases |

---

## 🔒 Security

### Custom Certificates (Optional)

Strigo can inject custom CA certificates into JDK installations for corporate environments:

```toml
[general]
custom_certificates = [
    { path = "/etc/ssl/corporate/root-ca.pem", alias = "corporate-root-ca" }
]
```

**Features:**
- Non-destructive: Preserves default JDK CA certificates
- Automatic backup: Creates `cacerts.original` before modification
- Auto-detection: Finds `cacerts` location automatically (Java 8 vs Java 11+)

See [Configuration Guide](docs/CONFIGURATION.md#certificate-management-jdk-only) for details.

### Software Bill of Materials (SBOM)

Each release includes a comprehensive SBOM in CycloneDX format (`sbom.json`):
- Complete dependency list with exact versions
- Security audit and compliance support
- Available in all releases

---

## 📋 Command Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `strigo available [type] [distribution]` | List available SDK versions |
| `strigo install <type> <distribution> <version>` | Install a specific SDK version |
| `strigo list` | List installed SDK versions |
| `strigo use <type> <distribution> <version>` | Switch to a specific SDK version |
| `strigo remove <type> <distribution> <version>` | Remove an installed SDK version |
| `strigo clean` | Remove invalid environment configurations |

### Global Flags

| Flag | Description |
|------|-------------|
| `--config <path>` | Use custom configuration file |
| `--json` | Output in JSON format |
| `--json-logs` | Enable JSON-formatted logging |
| `--help, -h` | Show help information |

**For complete command reference and examples, see [Configuration Guide](docs/CONFIGURATION.md).**

---

## 🔧 Environment Variables

Strigo manages environment variables for different SDK types:

```bash
# Automatically configure environment
strigo use jdk temurin 17.0.13_11 --set-env

# Remove environment configuration
strigo use jdk --unset
```

**Managed Variables:**
- **Java**: `JAVA_HOME`, `PATH`
- **Node.js**: `NODE_HOME`, `NPM_CONFIG_PREFIX`, `PATH`

**For detailed environment management, see [README - Environment Variables](docs/CONFIGURATION.md).**

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Follow Go best practices
2. Add tests for new functionality
3. Update documentation
4. Submit a pull request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🔗 Links

- [Issue Tracker](https://github.com/Altagen/Strigo/issues)
- [Releases](https://github.com/Altagen/Strigo/releases)
- [Documentation](docs/)
