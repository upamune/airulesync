# 🔄 aisyncrule

[![Go Report Card](https://goreportcard.com/badge/github.com/upamune/airulesync)](https://goreportcard.com/report/github.com/upamune/airulesync)
[![GitHub Release](https://img.shields.io/github/v/release/upamune/airulesync)](https://github.com/upamune/airulesync/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Synchronize AI coding tool rule files across directories with path adjustments.

## 🌟 Features

- ✅ Sync AI tool rule files between any directories (parent-to-child, child-to-parent, siblings, or cross-repository)
- ✅ Automatically adjust relative paths in rule files
- ✅ Flexible configuration with YAML
- ✅ Auto-detection of rule files
- ✅ Dry-run simulation mode

## 🚀 Installation

### Using Go

```bash
go install github.com/upamune/airulesync@latest
```

### Using Homebrew

```bash
brew upamune/tap/airulesync
```

### Binary Releases

Download the appropriate binary for your platform from the [releases page](https://github.com/upamune/airulesync/releases).

## 📖 Usage

### Quick Start

```bash
# Auto-generate a config file
airulesync init .

# Edit .airulesync.yaml if needed

# Run a dry-run simulation
airulesync sync --dry-run

# Perform the actual synchronization
airulesync sync
```

### Commands

- `airulesync sync` - Synchronizes rule files according to configuration
- `airulesync init [dir]` - Scans directory and generates a configuration file
- `airulesync version` - Displays version information
- `airulesync help` - Displays help information

### Flags

#### Global Flags
- `--config, -c` - Path to config file (default: `.airulesync.yaml`)
- `--verbose, -v` - Enable verbose output
- `--help, -h` - Display help information

#### Sync Command Flags
- `--dry-run, -d` - Simulate execution without applying changes

## ⚙️ Configuration

airulesync uses a YAML configuration file to define source and target directories, files to sync, and sync options. The configuration file includes helpful header comments for editor integration.

### Example Configuration

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/upamune/airulesync/refs/heads/main/schema.json
# vim: set ts=2 sw=2 tw=0 fo=cnqoj

# Source directories containing rule files to sync
source_dirs:
  - path: "./src/main-project"
    files:
      - ".clinerules"
      - pattern: ".cursor/rules/**/*.mdc"
        adjust_paths: true
      - pattern: ".roomodes"
        adjust_paths: false
    ignore_files:
      - ".cursor/rules/private/*.mdc"

# Target directories to sync to
target_dirs:
  - path: "./src/sub-project-a"
  - path: "./src/sub-project-b"
  - path: "../other-repo/src/component"
    external: true
```

### Configuration Reference

#### Source Directories

- `path`: Directory path containing rule files to sync
- `overwrite`: Whether to overwrite existing files (default: true)
- `files`: List of files to synchronize
  - Simple format: `".clinerules"` (uses default settings)
  - Detailed format:
    - `pattern`: File pattern (supports glob patterns)
    - `adjust_paths`: Whether to adjust relative paths in file (default: true)
    - `overwrite`: Whether to overwrite existing files (default: true)
- `ignore_files`: List of files to ignore (supports glob patterns)

#### Target Directories

- `path`: Directory path to sync files to
- `external`: Flag for targets outside the current repository (optional)
- `ignore_files`: List of files to ignore (supports glob patterns)

## 📝 Path Adjustment

airulesync handles path adjustments based on the relationship between source and target directories:

- **Parent to Child**: Adjusts paths for use in subdirectories
- **Child to Parent**: Adjusts paths for use in parent directories 
- **Sibling to Sibling**: Computes correct relative paths between siblings
- **Cross-Repository**: Handles external repository targets with appropriate warnings

### Path Detection Patterns

airulesync detects and adjusts various path formats:

- Import/require statements in various languages
- JSON/YAML path references
- File path references in configuration files
- Markdown links and references
- HTML href and src attributes
- General file paths with common extensions

### Development Commands

For developers contributing to the project:

- `make schema` - Generates the JSON Schema for configuration validation

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📜 License

This project is licensed under the MIT License - see the LICENSE file for details.