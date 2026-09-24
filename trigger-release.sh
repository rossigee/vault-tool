#!/bin/bash
# Manual release trigger script for vault-tool
# Usage: ./trigger-release.sh v0.1.6

set -euo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 <tag>"
  echo "Example: $0 v0.1.6"
  exit 1
fi

TAG="$1"

# Ensure tag exists
if ! git rev-parse "$TAG" > /dev/null 2>&1; then
  echo "Error: Tag '$TAG' does not exist"
  exit 1
fi

echo "Triggering release build for tag: $TAG"
echo ""
echo "Option 1: Manually push the tag to trigger Gitea Actions"
echo "  git push origin $TAG"
echo ""
echo "Option 2: Use Gitea web UI to manually dispatch the workflow"
echo "  1. Go to https://git.golder.lan/rossgolderltd/vault-tool/actions"
echo "  2. Click 'Release' workflow"
echo "  3. Click 'Run workflow'"
echo "  4. Enter tag: $TAG"
echo "  5. Click 'Run workflow'"
echo ""
echo "Option 3: Use Gitea API (requires token in GITEA_TOKEN env var)"
echo "  curl -X POST -H 'Authorization: token \$GITEA_TOKEN' \\"
echo "    https://git.golder.lan/api/v1/repos/rossgolderltd/vault-tool/actions/workflows/release.yaml/dispatches \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"ref\":\"master\",\"inputs\":{\"tag\":\"$TAG\"}}'"
