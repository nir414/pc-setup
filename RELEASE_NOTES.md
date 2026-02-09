# Syncer v0.1.0 Release

## Overview
This is the first release of the pc-setup syncer tool - a backup and synchronization utility.

## Features
- Backup functionality for syncing data
- Status checking
- Synchronization operations

## Build Artifacts
The following pre-built binaries are available:

- **Windows (AMD64)**: `syncer-windows-amd64.exe`
- **Linux (AMD64)**: `syncer-linux-amd64`
- **macOS (Intel)**: `syncer-darwin-amd64`
- **macOS (Apple Silicon)**: `syncer-darwin-arm64`

## Usage
Run the syncer with one of the following commands:
```
syncer backup   # Create a backup
syncer status   # Check sync status
syncer sync     # Perform synchronization
```

## Configuration
Place your `sync.toml` configuration file in the working directory.

## Notes
- Requires Go 1.22 or later to build from source
- Built with Go 1.24.12
