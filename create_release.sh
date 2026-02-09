#!/bin/bash
# Script to create a GitHub release for syncer v0.1.0
# This script should be run by the repository maintainer with appropriate permissions

set -e

VERSION="v0.1.0"
REPO="nir414/pc-setup"

echo "Creating release $VERSION for $REPO"

# Create and push git tag
git tag -a "$VERSION" -m "Release $VERSION - Initial syncer release"
git push origin "$VERSION"

echo "Tag created and pushed successfully!"
echo ""
echo "Next steps:"
echo "1. Go to https://github.com/$REPO/releases/new?tag=$VERSION"
echo "2. Upload the following binaries from the 'releases/' directory:"
echo "   - syncer-windows-amd64.exe"
echo "   - syncer-linux-amd64"
echo "   - syncer-darwin-amd64"
echo "   - syncer-darwin-arm64"
echo "3. Copy the content from RELEASE_NOTES.md as the release description"
echo "4. Mark as 'Latest release'"
echo "5. Click 'Publish release'"
echo ""
echo "Alternatively, use the GitHub CLI if available:"
echo "gh release create $VERSION ./releases/* --title \"Syncer $VERSION\" --notes-file RELEASE_NOTES.md"
