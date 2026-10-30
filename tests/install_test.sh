#!/usr/bin/env bash
set -euo pipefail
# Local release fixtures; no network or desktop required.
ROOT=$(cd "$(dirname "$0")/.." && pwd)
FIXTURE=$(mktemp -d)
trap 'rm -rf "$FIXTURE"' EXIT
mkdir -p "$FIXTURE/bin" "$FIXTURE/release" "$FIXTURE/payload"
cat > "$FIXTURE/bin/uname" <<'SCRIPT'
#!/usr/bin/env bash
case "$1" in -s) echo Linux;; -m) echo x86_64;; esac
SCRIPT
cat > "$FIXTURE/bin/curl" <<'SCRIPT'
#!/usr/bin/env bash
set -euo pipefail
cp "$RELEASE_FIXTURE/${2##*/}" "$4"
SCRIPT
printf '#!/bin/sh\necho fixture\n' > "$FIXTURE/payload/tcopy"
chmod +x "$FIXTURE/bin/uname" "$FIXTURE/bin/curl" "$FIXTURE/payload/tcopy"
ASSET=tcopy_0.1.0_linux_amd64.tar.gz
tar -czf "$FIXTURE/release/$ASSET" -C "$FIXTURE/payload" tcopy
(cd "$FIXTURE/release" && sha256sum "$ASSET" > checksums.txt)
export RELEASE_FIXTURE="$FIXTURE/release"
export PATH="$FIXTURE/bin:$PATH"
VERSION=v0.1.0 INSTALL_DIR="$FIXTURE/install" bash "$ROOT/install.sh"
test "$("$FIXTURE/install/tcopy")" = fixture
printf corrupt >> "$FIXTURE/release/$ASSET"
if VERSION=v0.1.0 INSTALL_DIR="$FIXTURE/rejected" bash "$ROOT/install.sh"; then
  echo 'Installer accepted a corrupted archive' >&2
  exit 1
fi
test ! -e "$FIXTURE/rejected/tcopy"
echo 'Installer fixture checks passed'
