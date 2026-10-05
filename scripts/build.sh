#!/usr/bin/env bash
# Builds the portable Windows release: dist/WiFiSpeedTester-Portable-<version>-<arch>.zip
# Usage: scripts/build.sh [version]   (needs Go and python3; runs on Linux, macOS or Git Bash)
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="${1:-$(git describe --tags --always 2>/dev/null | sed 's/^v//' || echo dev)}"
PKG=github.com/tehgeii/wifi-speed-tester
mkdir -p dist

for ARCH in amd64 arm64; do
  case "$ARCH" in amd64) LABEL=win-x64 ;; arm64) LABEL=win-arm64 ;; esac
  STAGE="dist/stage-$LABEL"
  mkdir -p "$STAGE"
  echo "Building $LABEL ($VERSION)"
  GOOS=windows GOARCH=$ARCH CGO_ENABLED=0 go build -trimpath \
    -ldflags "-H windowsgui -s -w -X $PKG/internal/app.Version=$VERSION" \
    -o "$STAGE/WiFiSpeedTester.exe" ./cmd/wifispeedtester
  go run ./cmd/wifispeedtester --print-config > "$STAGE/WiFiSpeedTester.config.example.json"
  cp packaging/README.txt "$STAGE/README.txt"
  ZIP="dist/WiFiSpeedTester-Portable-$VERSION-$LABEL.zip"
  (cd "$STAGE" && python3 -m zipfile -c "../$(basename "$ZIP")" WiFiSpeedTester.exe WiFiSpeedTester.config.example.json README.txt)
  echo "  -> $ZIP"
done
