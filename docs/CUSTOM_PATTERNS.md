# Guide: Adding Custom Patterns in Strigo

**Date:** 2025-12-18

---

## 📋 Overview

Strigo uses a **user-editable configuration file** to define version detection patterns: `strigopatterns.toml`

This file:
- ✅ Is automatically created on first run
- ✅ Contains all default patterns (15+ JDK distributions, Node.js, Python, etc.)
- ✅ Can be edited to add your own distributions
- ✅ **Is NOT overwritten** during Strigo updates

---

## 📍 File Location

### Default
```bash
./strigopatterns.toml
```

The file is created in the same directory as `strigo.toml` (working directory).

### Customizing the Location

You can specify a custom location in `strigo.toml`:

```toml
[general]
patterns_file = "/path/to/your/strigopatterns.toml"
```

Or use the `STRIGO_PATTERNS_PATH` environment variable:

```bash
export STRIGO_PATTERNS_PATH="$HOME/.config/strigo/strigopatterns.toml"
strigo available jdk
```

**Priority order:**
1. `STRIGO_PATTERNS_PATH` environment variable (highest)
2. `patterns_file` setting in `strigo.toml`
3. Default: `strigopatterns.toml` in current directory

---

## 🚀 First Use

### 1. Launch Strigo

On first run, the `strigopatterns.toml` file is created automatically:

```bash
$ strigo available jdk temurin

# The strigopatterns.toml file is created with all default patterns
```

### 2. Verify the File

```bash
$ ls -la strigopatterns.toml
-rw-r--r-- 1 user user 9.2K Dec 18 08:27 strigopatterns.toml
```

### 3. View the Content

```bash
$ head -30 strigopatterns.toml
```

---

## ✏️ Adding a Custom Pattern

### Pattern Structure

```toml
[[patterns]]
name = "provider-name"           # Unique identifier
type = "jdk"                      # SDK type (jdk, node, python, etc.)
description = "..."               # Human-readable description
patterns = [                      # Array of regex patterns (tried in order)
    "(?i)pattern1...",            # Case-insensitive pattern
    "(?i)pattern2...",            # Alternative pattern
]
```

### Example 1: Adding a Custom JDK Provider

```toml
# Add this to the end of strigopatterns.toml

[[patterns]]
name = "mycompany-jdk"
type = "jdk"
description = "MyCompany Custom JDK Build"
patterns = [
    "(?i)mycompany-jdk-(\\d+\\.\\d+\\.\\d+)",
    "(?i)custom-java-(\\d+\\.\\d+\\.\\d+_\\d+)",
]
```

**Then**, configure the repository in `strigo.toml`:

```toml
[sdk_repositories.mycompany]
registry = "nexus"
repository = "raw"
type = "jdk"
path = "jdk/mycompany"
```

**Test**:
```bash
$ strigo available jdk mycompany
```

### Example 2: Adding Ruby Support

```toml
[[patterns]]
name = "ruby"
type = "ruby"
description = "Ruby programming language"
patterns = [
    "(?i)ruby-(\\d+\\.\\d+\\.\\d+)",
    "(?i)ruby_(\\d+\\.\\d+\\.\\d+)_linux",
]
```

Configuration in `strigo.toml`:

```toml
[sdk_types.ruby]
type = "ruby"
install_dir = "~/.strigo/ruby"

[sdk_repositories.ruby-official]
registry = "nexus"
repository = "raw"
type = "ruby"
path = "ruby/official"
```

### Example 3: Pattern with Complex Versions

```toml
[[patterns]]
name = "graalvm-enterprise"
type = "jdk"
description = "GraalVM Enterprise Edition"
patterns = [
    # Format: graalvm-ee-java17-22.3.1
    "(?i)graalvm-ee-java\\d+-(\\d+\\.\\d+\\.\\d+)",

    # Format: graalvm-ee-17.0.11+7.1
    "(?i)graalvm-ee-(\\d+\\.\\d+\\.\\d+\\+\\d+(?:\\.\\d+)?)",

    # Format with build number
    "(?i)graalvm-enterprise-(\\d+\\.\\d+\\.\\d+)-b(\\d+)",
]
```

---

## 🎯 Best Practices

### 1. Use (?i) for Case-Insensitive Matching

✅ **GOOD:**
```toml
patterns = [
    "(?i)MyProvider-jdk-(\\d+\\.\\d+\\.\\d+)",
]
```

Matches:
- `MyProvider-jdk-11.0.26`
- `myprovider-jdk-11.0.26`
- `MYPROVIDER-JDK-11.0.26`

❌ **BAD:**
```toml
patterns = [
    "MyProvider-jdk-(\\d+\\.\\d+\\.\\d+)",  # Missing (?i)
]
```

Only matches: `MyProvider-jdk-11.0.26`

### 2. Pattern Order

Patterns are tried **in order** as they appear. Put the most specific patterns first:

✅ **GOOD:**
```toml
patterns = [
    "(?i)custom-jdk-enterprise-(\\d+\\.\\d+\\.\\d+)",  # Specific
    "(?i)custom-jdk-(\\d+\\.\\d+\\.\\d+)",             # Generic
]
```

### 3. Testing Patterns

Use an online tool to test your regex patterns:
- https://regex101.com/
- https://regexr.com/

**Test example:**
```
Regex: (?i)mycompany-jdk-(\d+\.\d+\.\d+)

Test string: MyCompany-JDK-17.0.11-linux-x64.tar.gz
Match: 17.0.11
```

### 4. Capture Groups

Use **ONE CAPTURE GROUP** `()` to extract the version:

✅ **GOOD:**
```toml
"(?i)provider-jdk-(\\d+\\.\\d+\\.\\d+)"
#                  ^^^^^^^^^^^^^^^^^^^^
#                  Capture group = extracted version
```

❌ **BAD:**
```toml
"(?i)(provider)-(jdk)-(\\d+\\.\\d+\\.\\d+)"
#     ^^^^^^^^^  ^^^^^  ^^^^^^^^^^^^^^^^^^^^
#     3 groups - only the last one will be used!
```

---

## 🔍 Real Examples

### Complete Example: Adding OpenLogic OpenJDK Support

**1. Add the pattern in `strigopatterns.toml`:**

```toml
[[patterns]]
name = "openlogic"
type = "jdk"
description = "OpenLogic OpenJDK"
patterns = [
    "(?i)openlogic-openjdk-(\\d+\\.\\d+\\.\\d+\\+\\d+)",
    "(?i)openlogic-jdk_(\\d+\\.\\d+\\.\\d+)",
]
```

**2. Configure in `strigo.toml`:**

```toml
[sdk_repositories.openlogic]
registry = "nexus"
repository = "raw"
type = "jdk"
path = "jdk/openlogic"
```

**3. Test:**

```bash
$ strigo available jdk openlogic
Available versions for openlogic (jdk):
  - 21.0.11+9
  - 17.0.15+8
  - 11.0.26+7
```

---

## 🛠️ Debugging

### Problem: Pattern Not Detecting Versions

**1. Enable debug logs:**

```bash
export STRIGO_LOG_LEVEL=DEBUG
strigo available jdk your-provider
```

**2. Check the logs:**

The logs show which patterns are being tried:

```
DEBUG: 🔍 Extracting version from path: /jdk/provider/custom-jdk-17.0.11.tar.gz
DEBUG:    Trying pattern 'temurin': NO MATCH
DEBUG:    Trying pattern 'corretto': NO MATCH
DEBUG:    Trying pattern 'your-pattern': MATCH! version=17.0.11
```

### Problem: Wrong Version Extracted

**Example:** Pattern `(?i)jdk-(\\d+)` extracts `17` instead of `17.0.11`

**Solution:** Add the rest of the version in the capture group:

```toml
# BEFORE (bad)
"(?i)jdk-(\\d+)"

# AFTER (good)
"(?i)jdk-(\\d+\\.\\d+\\.\\d+)"
```

### Problem: File Not Found

```
Error: failed to read patterns file strigo-patterns.toml: no such file or directory
```

**Solution:**

Copy the example patterns file from the Strigo repository:
```bash
# From the Strigo project root
cp strigo-patterns.toml /path/to/your/config/directory/

# Or download from GitHub
wget https://raw.githubusercontent.com/your-org/strigo/main/strigo-patterns.toml
```

Then configure the path in your `strigo.toml`:
```toml
[general]
patterns_file = "/path/to/your/strigo-patterns.toml"
```

---

## 📚 Default Patterns

### JDK (15+ distributions)

- ✅ Temurin (AdoptOpenJDK)
- ✅ Amazon Corretto
- ✅ Azul Zulu
- ✅ GraalVM (CE & Oracle)
- ✅ Mandrel
- ✅ BellSoft Liberica
- ✅ Oracle JDK
- ✅ SAP Machine
- ✅ Microsoft OpenJDK
- ✅ IBM Semeru
- ✅ Alibaba Dragonwell
- ✅ Eclipse OpenJ9

### Other SDKs

- ✅ Node.js
- ✅ Python
- ✅ Go
- ✅ Rust
- ✅ .NET
- ✅ Maven
- ✅ Gradle

---

## ⚠️ Important Notes

1. **Restart Required**
   - After modifying `strigopatterns.toml`, rerun the strigo command
   - The file is read once at startup

2. **Backup Recommended**
   ```bash
   cp strigopatterns.toml strigopatterns.toml.backup
   ```

3. **No Automatic Deletion**
   - The `strigopatterns.toml` file is **NEVER deleted** by Strigo
   - Your modifications are preserved during updates

4. **Validation**
   - If the file contains TOML errors, Strigo will display an error message
   - Use an online TOML validator: https://www.toml-lint.com/

---

## 🔗 References

- **Regex syntax:** https://github.com/google/re2/wiki/Syntax
- **TOML syntax:** https://toml.io/en/
- **Reference file:** `strigo-patterns.toml` (in the Strigo repository)

---

## 💡 Need Help?

If you have questions or want to share your custom patterns:

1. Open an issue on GitHub
2. Share your pattern so it can be added to the default patterns!

**Contributing:** Good patterns can be added to the `builtin.toml` file to benefit all users.
