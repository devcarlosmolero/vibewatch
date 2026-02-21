#!/bin/bash

# Release script for vibewatch
# This script automates the release process including tagging, building, packaging,
# Codeberg release creation, and Homebrew formula updates

set -e

# Default values
CODEBERG_API_TOKEN="${CODEBERG_API_TOKEN:-}"
UPDATE_HOMEBREW=true
HOMEBREW_REPO_PATH="${HOMEBREW_REPO_PATH:-}"

  # Parse command line arguments
  while getopts "v:" opt; do
    case $opt in
    v)
      VERSION="$OPTARG"
      ;;
    *)
      echo "Usage: $0 -v <version>" >&2
      exit 1
      ;;
    esac
  done

# Get environment variables
CODEBERG_API_TOKEN="${CODEBERG_API_TOKEN:-}"
HOMEBREW_REPO_PATH="${HOMEBREW_REPO_PATH:-}"

if [ -z "$VERSION" ]; then
  echo "Error: Version not specified. Use -v flag."
  echo "Usage: $0 -v <version> [-u]"
  echo ""
  echo "Options:"
  echo "  -v <version>          Version number for the release (required)"
  echo ""
  echo "Environment Variables:"
  echo "  CODEBERG_API_TOKEN     Codeberg API token for automated release creation"
  echo "  HOMEBREW_REPO_PATH     Path to Homebrew repository (required)"
  echo ""
  echo "Examples:"
  echo "  # Basic release (manual upload to Codeberg)"
  echo "  ./release.sh -v 1.1.0"
  echo ""
  echo "  # Full automation (Codeberg + Homebrew)"
  echo "  export CODEBERG_API_TOKEN='your_token'"
  echo "  export HOMEBREW_REPO_PATH='~/Desktop/Git/homebrew'"
  echo "  ./release.sh -v 1.1.0"
  exit 1
fi

echo "Starting release process for vibewatch version $VERSION..."

# Step 1: Create git tag
echo "Creating git tag $VERSION..."
git tag "$VERSION"
echo "Tag created: $VERSION"

# Step 2: Build binary with version embedded
echo "Building vibewatch binary..."
go build -ldflags "-X main.version=$VERSION" -o vibewatch .
echo "Binary built: vibewatch"

# Step 3: Create tarball
echo "Creating distribution tarball..."
tar -czvf vibewatch-$VERSION.tar.gz vibewatch
echo "Distribution package created: vibewatch-$VERSION.tar.gz"

# Step 4: Calculate and output checksum
echo "Calculating SHA-256 checksum..."
SHA256=$(shasum -a 256 vibewatch-$VERSION.tar.gz | awk '{print $1}')
echo "SHA-256 checksum: $SHA256"

# Step 5: Push tag to remote
echo "Pushing tag to remote..."
git push origin "$VERSION"
echo "Tag pushed: $VERSION"

# Step 6: Create Codeberg release and upload assets
if [ -z "$CODEBERG_API_TOKEN" ]; then
  echo "Error: CODEBERG_API_TOKEN environment variable not set."
  echo "Please set the environment variable and try again."
  exit 1
fi

echo "Creating Codeberg release..."

  # Create release via API
  RELEASE_RESPONSE=$(curl -s -X POST \
    -H "Authorization: token $CODEBERG_API_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"tag_name\":\"$VERSION\",\"name\":\"vibewatch $VERSION\",\"body\":\"Automated release $VERSION\"}" \
    "https://codeberg.org/api/v1/repos/devcarlosmolero/vibewatch/releases")

  # Extract release ID from response
  RELEASE_ID=$(echo "$RELEASE_RESPONSE" | grep -o '"id":[0-9]*' | cut -d: -f2)
  
  if [ -z "$RELEASE_ID" ]; then
    echo "Error: Failed to create release via Codeberg API"
    echo "Response: $RELEASE_RESPONSE"
    exit 1
  fi
  
  echo "Release created with ID: $RELEASE_ID"
  
  # Upload tarball as release asset
  echo "Uploading release asset..."
  UPLOAD_RESPONSE=$(curl -s -X POST \
    -H "Authorization: token $CODEBERG_API_TOKEN" \
    -F "attachment=@vibewatch-$VERSION.tar.gz" \
    -F "name=vibewatch-$VERSION.tar.gz" \
    "https://codeberg.org/api/v1/repos/devcarlosmolero/vibewatch/releases/$RELEASE_ID/assets")
  
  if echo "$UPLOAD_RESPONSE" | grep -q '"name":"vibewatch-'$VERSION'.tar.gz"'; then
    echo "Asset uploaded successfully!"
    
    # Extract download URL
    DOWNLOAD_URL=$(echo "$UPLOAD_RESPONSE" | grep -o '"browser_download_url":"[^"]*"' | cut -d\" -f4)
    echo "Download URL: $DOWNLOAD_URL"
    
    # Clean up tarball after successful upload
    echo "Cleaning up tarball..."
    rm -f vibewatch-$VERSION.tar.gz
  else
    echo "Error: Failed to upload asset"
    echo "Response: $UPLOAD_RESPONSE"
    exit 1
  fi

# Step 7: Update Homebrew formula
if [ -z "$HOMEBREW_REPO_PATH" ]; then
  echo "Error: HOMEBREW_REPO_PATH environment variable not set."
  echo "Please set the environment variable and try again."
  exit 1
fi

if [ ! -d "$HOMEBREW_REPO_PATH" ]; then
  echo "Error: Homebrew repo path does not exist: $HOMEBREW_REPO_PATH"
  echo "Please check the HOMEBREW_REPO_PATH environment variable."
  exit 1
fi

echo "Updating Homebrew formula..."

  # Determine the download URL
  if [ -n "$DOWNLOAD_URL" ]; then
    FORMULA_URL="$DOWNLOAD_URL"
  else
    FORMULA_URL="https://codeberg.org/devcarlosmolero/vibewatch/releases/download/$VERSION/vibewatch-$VERSION.tar.gz"
  fi

  # Update the formula
  FORMULA_FILE="$HOMEBREW_REPO_PATH/Formula/vibewatch.rb"

  # Update URL and SHA in the formula
  sed -i '' "s|url '.*'|url '$FORMULA_URL'|" "$FORMULA_FILE"
  sed -i '' "s|sha256 '.*'|sha256 '$SHA256'|" "$FORMULA_FILE"

  echo "Formula updated successfully!"

  # Commit and push the formula changes
  echo "Committing and pushing Homebrew formula updates..."
  cd "$HOMEBREW_REPO_PATH"
  git add "$FORMULA_FILE"
  git commit -m "update vibewatch to version $VERSION"
  git push origin master

  echo "Homebrew formula updated and pushed!"

echo ""
echo "Release process complete!"
echo "Tag: $VERSION"
echo "Tarball: vibewatch-$VERSION.tar.gz"
echo "SHA-256: $SHA256"

if [ -n "$DOWNLOAD_URL" ]; then
  echo "Download URL: $DOWNLOAD_URL"
fi

echo ""
echo "Next steps (if not already automated):"
if [ -z "$CODEBERG_API_TOKEN" ]; then
  echo "Error: CODEBERG_API_TOKEN environment variable not set."
  echo "Please set the environment variable and try again."
  exit 1
fi
if [ "$UPDATE_HOMEBREW" = false ]; then
  echo "2. Update Homebrew formula with:"
  echo "   URL: https://codeberg.org/devcarlosmolero/vibewatch/releases/download/$VERSION/vibewatch-$VERSION.tar.gz"
  echo "   SHA-256: $SHA256"
fi

