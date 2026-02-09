#!/bin/bash
# Build script for syncer
# Creates release binaries for all supported platforms

set -e

echo "Building syncer for multiple platforms..."

# Create releases directory
mkdir -p releases

# Build for Windows
echo "Building for Windows (AMD64)..."
cd syncer
GOOS=windows GOARCH=amd64 go build -o ../releases/syncer-windows-amd64.exe ./cmd/syncer
cd ..

# Build for Linux
echo "Building for Linux (AMD64)..."
cd syncer
GOOS=linux GOARCH=amd64 go build -o ../releases/syncer-linux-amd64 ./cmd/syncer
cd ..

# Build for macOS Intel
echo "Building for macOS (Intel)..."
cd syncer
GOOS=darwin GOARCH=amd64 go build -o ../releases/syncer-darwin-amd64 ./cmd/syncer
cd ..

# Build for macOS Apple Silicon
echo "Building for macOS (Apple Silicon)..."
cd syncer
GOOS=darwin GOARCH=arm64 go build -o ../releases/syncer-darwin-arm64 ./cmd/syncer
cd ..

echo ""
echo "Build complete! Binaries are in the 'releases/' directory:"
ls -lh releases/
