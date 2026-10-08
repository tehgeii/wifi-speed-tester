#!/usr/bin/env bash
# End-to-end UI test: serves the embedded page with the real Go API
# (--serve) against the local fake speed-test server and drives it with
# Playwright (ui-tests/run.cjs).
# Usage: scripts/ui-test.sh   (needs Go, Node and the playwright package with Chromium)
# Screenshots go to ui-tests/screenshots (set UI_SHOTS= to skip them).
set -euo pipefail
cd "$(dirname "$0")/.."

WORK="$(mktemp -d)"
cleanup() {
  kill $(jobs -p) 2>/dev/null || true
  rm -rf "$WORK"
}
trap cleanup EXIT

go build -o "$WORK/wst" ./cmd/wifispeedtester
go build -o "$WORK/fakesrv" ./cmd/fakespeedserver
cp ui-tests/config.json "$WORK/WiFiSpeedTester.config.json"
# English, and no update check: the test must not depend on GitHub.
mkdir -p "$WORK/WiFiSpeedTester-data"
echo '{"lang": "en", "updateCheck": "off"}' > "$WORK/WiFiSpeedTester-data/settings.json"

"$WORK/fakesrv" -addr 127.0.0.1:18099 -rate 2000000 > "$WORK/fakesrv.log" 2>&1 &
WST_FAKE_WIFI_SCAN="$PWD/ui-tests/wifi-scan.json" "$WORK/wst" --serve 18088 > "$WORK/wst.log" 2>&1 &

for _ in $(seq 50); do
  curl -fs -o /dev/null http://127.0.0.1:18088/ && break
  sleep 0.2
done

status=0
UI_URL=http://127.0.0.1:18088/ UI_SHOTS="${UI_SHOTS-ui-tests/screenshots}" node ui-tests/run.cjs || status=$?
if [ "$status" -ne 0 ]; then
  echo "--- app log ---"; cat "$WORK/wst.log"
fi
exit "$status"
