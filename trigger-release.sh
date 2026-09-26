#!/bin/bash
# Release trigger script for vault-tool
# Usage: ./trigger-release.sh v0.2.2

set -euo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 <tag>"
  echo "Example: $0 v0.2.2"
  exit 1
fi

TAG="$1"

# Ensure tag exists
if ! git rev-parse "$TAG" > /dev/null 2>&1; then
  echo "Error: Tag '$TAG' does not exist"
  exit 1
fi

echo "Triggering release for tag: $TAG"
echo ""
echo "Option 1: Push the tag to GitHub to trigger the release workflow"
echo "  git push origin $TAG"
echo ""
echo "Option 2: Use GitHub CLI to manually dispatch the workflow"
echo "  gh workflow run release.yaml -f tag=$TAG"
echo ""
echo "The release workflow will:"
echo "  - Build the Debian package"
echo "  - Run tests and security scans"
echo "  - Create a GitHub Release with the .deb attached"
