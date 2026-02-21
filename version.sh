#!/bin/bash

# This script generates the version based on Git tags
# It should be run as part of the build process

set -e

# Get the latest tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0-beta")

# Remove 'v' prefix if present
VERSION=${LATEST_TAG#v}

# If no tags exist, use default version
default_version="1.0.0-beta"

# Output the version
if [ -z "$VERSION" ] || [ "$VERSION" = "" ]; then
    echo "$default_version"
else
    echo "$VERSION"
fi