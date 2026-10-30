
#!/usr/bin/env bash
set -euo pipefail

REPO="ak-repo/tcopy"
VERSION="${VERSION:-v0.1.0}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "Error: Only Linux is supported" >&2
  exit 1
fi

case "$(uname -m)" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture" >&2; exit 1 ;;
esac

ASSET="tcopy_${VERSION#v}_linux_${ARCH}.tar.gz"
BASE="https://github.com/$REPO/releases/download/$VERSION"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading tcopy $VERSION ($ARCH)..."

curl -fsSL "$BASE/$ASSET" -o "$TMP/$ASSET"
curl -fsSL "$BASE/checksums.txt" -o "$TMP/checksums.txt"

# Verify that the downloaded archive matches
# the checksum published with the release.
(
  cd "$TMP"
  grep -F "  $ASSET" checksums.txt > selected-checksum.txt
  sha256sum -c selected-checksum.txt
)

mkdir -p "$INSTALL_DIR"
tar -xzf "$TMP/$ASSET" -C "$TMP"

install -m 755 "$TMP/tcopy" "$INSTALL_DIR/tcopy"

echo "tcopy installed to $INSTALL_DIR/tcopy"
echo "Ensure $INSTALL_DIR is in your PATH."
