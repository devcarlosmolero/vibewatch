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

if [ -z "$VERSION" ]; then
  echo "Error: Version not specified. Use -v flag."
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

# Step 4: Calculate checksum
echo "Calculating SHA-256 checksum..."
SHA256=$(shasum -a 256 vibewatch-$VERSION.tar.gz | awk '{print $1}')
echo "SHA-256 checksum: $SHA256"

# Step 5: Push tag
echo "Pushing tag to remote..."
git push origin "$VERSION"
echo "Tag pushed: $VERSION"

# Step 6: Create Codeberg release
if [ -z "$CODEBERG_API_TOKEN" ]; then
  echo "Error: CODEBERG_API_TOKEN environment variable not set."
  exit 1
fi

echo "Creating Codeberg release..."

RELEASE_RESPONSE_FILE=$(mktemp)

curl -v -sL --fail --max-time 30 -X POST \
  -H "Authorization: Bearer $CODEBERG_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tag_name":"'"$VERSION"'","name":"vibewatch '"$VERSION"'","body":"Automated release '"$VERSION"'"}' \
  "https://codeberg.org/api/v1/repos/devcarlosmolero/vibewatch/releases" \
  >"$RELEASE_RESPONSE_FILE"

RELEASE_RESPONSE=$(cat "$RELEASE_RESPONSE_FILE")
rm -f "$RELEASE_RESPONSE_FILE"

echo "$RELEASE_RESPONSE" | jq .

RELEASE_ID=$(echo "$RELEASE_RESPONSE" | jq -r '.id // empty')

if [ -z "$RELEASE_ID" ]; then
  echo "Error: Failed to create release via Codeberg API"
  echo "Response: $RELEASE_RESPONSE"

  ERROR_MSG=$(echo "$RELEASE_RESPONSE" | jq -r '.message // empty')
  if [ -n "$ERROR_MSG" ]; then
    echo "API Error: $ERROR_MSG"
  fi

  exit 1
fi

echo "Release created with ID: $RELEASE_ID"

# Upload tarball
echo "Uploading release asset..."

UPLOAD_RESPONSE=$(curl -s --max-time 60 -X POST \
  -H "Authorization: token $CODEBERG_API_TOKEN" \
  -F "attachment=@vibewatch-$VERSION.tar.gz" \
  -F "name=vibewatch-$VERSION.tar.gz" \
  "https://codeberg.org/api/v1/repos/devcarlosmolero/vibewatch/releases/$RELEASE_ID/assets")

if echo "$UPLOAD_RESPONSE" | jq -e '.name == "vibewatch-'$VERSION'.tar.gz"' >/dev/null; then
  echo "Asset uploaded successfully!"

  DOWNLOAD_URL=$(echo "$UPLOAD_RESPONSE" | jq -r '.browser_download_url')
  echo "Download URL: $DOWNLOAD_URL"

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
  exit 1
fi

if [ ! -d "$HOMEBREW_REPO_PATH" ]; then
  echo "Error: Homebrew repo path does not exist: $HOMEBREW_REPO_PATH"
  exit 1
fi

echo "Updating Homebrew formula..."

if [ -n "$DOWNLOAD_URL" ]; then
  FORMULA_URL="$DOWNLOAD_URL"
else
  FORMULA_URL="https://codeberg.org/devcarlosmolero/vibewatch/releases/download/$VERSION/vibewatch-$VERSION.tar.gz"
fi

FORMULA_FILE="$HOMEBREW_REPO_PATH/Formula/vibewatch.rb"

sed -i '' "s|url '.*'|url '$FORMULA_URL'|" "$FORMULA_FILE"
sed -i '' "s|sha256 '.*'|sha256 '$SHA256'|" "$FORMULA_FILE"

echo "Formula updated successfully!"

echo "Committing and pushing Homebrew formula updates..."

cd "$HOMEBREW_REPO_PATH"
git add "$FORMULA_FILE"
git commit -m "update vibewatch to version $VERSION"
git push origin master

echo "Homebrew formula updated and pushed!"

echo ""
echo "Release process complete!"
echo "Tag: $VERSION"
echo "SHA-256: $SHA256"

if [ -n "$DOWNLOAD_URL" ]; then
  echo "Download URL: $DOWNLOAD_URL"
fi

echo ""
echo "Release process complete!"
