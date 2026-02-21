#!/bin/bash

# Build script for vibewatch
# This script automates the build process with proper versioning

set -e

# Get version from git tags
VERSION=$(./version.sh)

echo "Building vibewatch version $VERSION..."

# Build with version embedded
go build -ldflags "-X main.version=$VERSION" -o vibewatch .

echo "Build complete: vibewatch version $VERSION"

# Create tarball
echo "Creating distribution tarball..."
tar -czvf vibewatch-$VERSION.tar.gz vibewatch

echo "Distribution package created: vibewatch-$VERSION.tar.gz"

# Calculate checksum
echo "Calculating checksum..."
shasum -a 256 vibewatch-$VERSION.tar.gz

echo "Build process complete!"