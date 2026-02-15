#!/usr/bin/env bash
# Extracts the changelog section for a given version from CHANGELOG.md
# Usage: ./scripts/extract-changelog.sh v0.1.2
# Output: Section from ## [X.Y.Z] until the next ## or end of file

set -e

VERSION="${1:?Usage: $0 <version>}"
# Strip leading 'v' if present (v0.1.2 -> 0.1.2)
VERSION="${VERSION#v}"

CHANGELOG="${2:-CHANGELOG.md}"
if [[ ! -f "$CHANGELOG" ]]; then
  echo "Changelog file not found: $CHANGELOG" >&2
  exit 1
fi

# Escape dots for regex (0.1.2 -> 0\.1\.2)
VERSION_ESC="${VERSION//./\\.}"

# Extract from ## [X.Y.Z] until next ## or end
output=$(awk -v ver="$VERSION_ESC" '
  $0 ~ "^## \\[" ver "\\]" { found=1; print; next }
  found && /^## \[/ { exit }
  found { print }
' "$CHANGELOG")

echo "$output"

if [[ -z "$output" ]]; then
  echo "No changelog entry found for version $VERSION. Add a section to CHANGELOG.md before releasing." >&2
  exit 1
fi
