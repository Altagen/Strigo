# Strigo Documentation

Welcome to the Strigo documentation! Strigo is a flexible SDK/JDK version manager that downloads and manages multiple SDK versions from custom repositories.

## 📚 Documentation Index

### Getting Started
- [Installation & Quick Start](../README.md) - Main project README
- [Configuration Guide](CONFIGURATION.md) - How to configure Strigo
- [Custom Patterns](CUSTOM_PATTERNS.md) - Defining version extraction patterns

### Advanced Topics
- [Architecture Overview](ARCHITECTURE.md) - Project structure and design
- [Testing Guide](TESTING.md) - Running and writing tests

## 🎯 Quick Navigation

### For Users
1. **First time setup**: See [Configuration Guide](CONFIGURATION.md)
2. **Using with Nexus**: See [Configuration Guide - Registries](CONFIGURATION.md#registries)
3. **Custom version patterns**: See [Custom Patterns](CUSTOM_PATTERNS.md)

### For Developers
1. **Understanding the codebase**: See [Architecture](ARCHITECTURE.md)
2. **Running tests**: See [Testing Guide](TESTING.md)

## 🔧 Common Tasks

### Install a JDK
```bash
# List available versions
strigo available jdk temurin

# Install specific version
strigo install jdk temurin 11.0.24_8

# List installed versions
strigo list jdk
```

### Configure a Custom Registry
See [Configuration Guide - Custom Registries](CONFIGURATION.md#custom-registries)

### Add Support for a New SDK Distribution
See [Custom Patterns - Adding Patterns](CUSTOM_PATTERNS.md#adding-new-patterns)

## 📖 Documentation Structure

```
docs/
├── README.md              # This file - documentation index
├── ARCHITECTURE.md        # Project architecture and design
├── CONFIGURATION.md       # Configuration reference
├── CUSTOM_PATTERNS.md     # Version pattern syntax
└── TESTING.md            # Testing guide
```

## 🤝 Getting Help

- **Bug reports**: [GitHub Issues](https://github.com/yourusername/strigo/issues)
- **Feature requests**: [GitHub Issues](https://github.com/yourusername/strigo/issues)
- **Questions**: Check existing documentation or open a discussion

## 📝 License

This project is licensed under the same license as Strigo. See the main [LICENSE](../LICENSE) file for details.
