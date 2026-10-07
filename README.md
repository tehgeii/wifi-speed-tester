# WiFi Speed + Ping Tester

> Check your speed. Measure your ping. Understand your connection.

A lightweight, portable Windows utility that measures **download, upload,
ping, jitter, packet loss** and **latency under load**. It shows network and
Wi-Fi details and gives a plain-language connection rating. There is no
installer, no background service and no startup entry: extract the zip and run
`WiFiSpeedTester.exe`.

```
WiFiSpeedTester-Portable-<version>-win-x64.zip
├── WiFiSpeedTester.exe
├── WiFiSpeedTester.config.example.json
└── README.txt
```

## Download

Get the latest version from
**[Releases](https://github.com/tehgeii/wifi-speed-tester/releases/latest)**:

- **`…-win-x64.zip`**: most Windows 10/11 PCs (Intel / AMD).
- **`…-win-arm64.zip`**: Windows on ARM (Snapdragon).

Extract the zip and double-click `WiFiSpeedTester.exe`. The app is not
code-signed yet, so Windows SmartScreen may show "Windows protected your PC";
click **More info → Run anyway**.

## Features

| Area | What you get |
|------|--------------|
| Network detection | Active adapter, type (Wi-Fi / Ethernet / hotspot-tethering / cellular / VPN), IPv4, IPv6, gateway, DNS |
| Wi-Fi details | SSID, BSSID, signal %, RSSI, band, channel, frequency, Wi-Fi standard, link speed (labelled as *not* internet speed) |
| Internet check | Adapter ✓ → Gateway ✓ → DNS ✓ → Internet ✓; stops with a clear message if the internet is unreachable |
| Ping | Router, Cloudflare DNS and Google DNS by default (configurable): sent/received, min/avg/max, jitter, loss |
| Download / Upload | Multi-connection HTTP test with warm-up excluded, live gauge, throughput chart, retry and fallback to backup servers |
| Latency under load | Ping sampled during download and upload, with a cautious bufferbloat hint |
| Rating | EXCELLENT / GOOD / FAIR / POOR / UNSTABLE from documented, configurable thresholds ([docs/QUALITY.md](docs/QUALITY.md)). **Why this rating?** shows every metric's grade and thresholds |
| Tips | Concrete next steps for what was measured (turn on SQM/QoS, use 5 GHz, try a cable, restart the router, VPN active, contact the ISP…) |
| Gaming mode | Rates ping, jitter, loss and latency stability instead of bandwidth |
| Quick Ping | About 5 seconds: connection check and ping only, for a quick check before gaming |
| Test servers | Cloudflare by default; **Find nearby servers** ranks public LibreSpeed servers by latency, and the chosen one is used first, with Cloudflare as backup |
| Language | English and Bahasa Indonesia (follows Windows on first run; switch in Settings ⚙) |
| Units | Mbps (default) or MB/s (80 Mbps = 10 MB/s) |
| Export | TXT, JSON, CSV and a PNG result card. SSID, BSSID and IPs are masked unless you choose to include them |
| History | Stored locally as JSON next to the exe; speed and ping trend charts, averages per connection (Wi-Fi vs Ethernet…), 7/30-day filters, reload, export to CSV, clear |
| Copy result | Short summary for WhatsApp, Discord or anywhere, without the Wi-Fi name or IPs |
| Updates | Optional check for a newer release on startup (Settings ⚙), with a link to the download |
| Errors | Reason and suggestion in plain language, with technical details in a collapsible section |
| Cancel | Stops all network activity immediately (Esc also works) |

## Requirements

- Windows 10 or 11, x64 or ARM64.
- The [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/),
  which ships with Windows 11 and most Windows 10 installations. If it is
  missing the app says so and offers the download link. `--cli` works without
  it.
- No administrator rights. Ping uses the Windows ICMP helper API, not raw
  sockets.

## Privacy

Testing is local except for the traffic needed to measure the connection:

- speed-test data to the configured test server (default: Cloudflare, `speed.cloudflare.com`);
- ICMP echo (ping) to your router and to the configured targets (default `1.1.1.1`, `8.8.8.8`);
- a DNS lookup of the test server name;
- only when you press **Find nearby servers**: the public server list from
  `librespeed.org` and one small request to each listed server;
- only if **Check for updates** is on (Settings ⚙, on by default): one
  request to `api.github.com` for the latest release.

The app does not read files, browser history or passwords, has no telemetry
and does not run in the background. The About (`?`) dialog lists exactly
where traffic goes for the current configuration.

Files it writes:

- `WiFiSpeedTester-data\` next to the exe holds `history.json`,
  `settings.json` (language, unit, mode, update check) and `server.json`
  (the server picked in the app, if any). If that folder is read-only it falls back to
  `%APPDATA%\WiFiSpeedTester`.
- `%LOCALAPPDATA%\WiFiSpeedTester\WebView2` is the browser-engine cache.

> On Windows 11, Windows only reveals the Wi-Fi network name (SSID) when
> *Location → Let desktop apps access your location* is on. Everything else
> works without it, and the app shows a note when the SSID is hidden.

## Accuracy

Results depend on the test server, routing, ISP, Wi-Fi signal, other traffic,
VPNs and the device. The app reports what it **measured** ("Measured download
speed: 48.6 Mbps") and never presents it as your ISP plan speed.

How it measures:

- **Download**: 4 parallel HTTP/1.1 connections (one TCP connection each)
  stream data for 10 s. Requests start at 10 MB and grow up to 100 MB on fast
  links, which keeps the request count low so servers don't rate-limit. If the
  server refuses a larger size, the test stays at the largest one that worked. The
  first 1.5 s of warm-up are excluded, and the speed comes from bytes
  received. If the server stops answering partway, the data already received
  is still reported, marked as incomplete.
- **Upload**: 4 parallel connections POST incompressible data. Only requests
  the server **acknowledged** are counted, each credited for the part that
  overlaps the measurement window, so data still sitting in local socket
  buffers cannot inflate the result.
- **Ping**: probes are sent every 200 ms without waiting for earlier replies,
  as the `ping` command does. Jitter is the mean absolute difference between
  consecutive samples.
- **ICMP blocked?** Some networks (corporate, cloud) drop ping entirely. If no
  internet target answers, latency is measured from TCP connection setup to
  the test server instead, and the result says so.

## Configuration

Create `WiFiSpeedTester.config.json` next to the exe. The easiest start is to
rename the bundled `WiFiSpeedTester.config.example.json`. Any field you leave
out keeps its default. An invalid file is reported in the UI, and the app
falls back to the defaults.

```jsonc
{
  "servers": [                       // tried in order; later ones are backups
    { "name": "Cloudflare",
      "downloadUrl": "https://speed.cloudflare.com/__down?bytes={bytes}",
      "uploadUrl":   "https://speed.cloudflare.com/__up" },
    { "name": "My LibreSpeed", "type": "librespeed",   // optional
      "url": "https://speedtest.example.com/backend/" }
  ],
  "pingTargets": [                   // "gateway" = your router
    { "label": "Local Gateway",  "host": "gateway" },
    { "label": "Cloudflare DNS", "host": "1.1.1.1" },
    { "label": "Google DNS",     "host": "8.8.8.8" }
  ],
  "pingCount": 20, "pingIntervalMs": 200, "pingTimeoutMs": 1000,
  "downloadSeconds": 10, "uploadSeconds": 10, "warmupSeconds": 1.5,
  "streams": 4,
  "measureLoadedPing": true,
  "connectivityHost": "speed.cloudflare.com",
  "serverListUrl": "https://librespeed.org/backend-servers/servers.php",
  "historyLimit": 200,
  "profiles": { "general": { ... }, "gaming": { ... }, "quick": { ... } }   // see docs/QUALITY.md
}
```

A URL-template server must accept `GET <downloadUrl>` with `{bytes}` replaced
by a byte count and return that many bytes. It must also accept a `POST` of
arbitrary size to `uploadUrl`. A `librespeed` server needs only the backend
`url`; the app uses `garbage.php` and `empty.php` there (override with
`dlPath`, `ulPath`, `pingPath`). Only use servers you are allowed to test
against.

## Command line

```
WiFiSpeedTester.exe                 open the window
WiFiSpeedTester.exe --cli           run a test in the terminal and print a report
WiFiSpeedTester.exe --cli --gaming  gaming profile
WiFiSpeedTester.exe --cli --quick   ping only (about 5 seconds)
WiFiSpeedTester.exe --cli --lang id report in Bahasa Indonesia
WiFiSpeedTester.exe --cli --json    full result as JSON (redirect with > file.json)
WiFiSpeedTester.exe --list-servers  rank public LibreSpeed servers by latency
WiFiSpeedTester.exe --print-config  print the default configuration
WiFiSpeedTester.exe --write-config  write WiFiSpeedTester.config.json next to the exe
```

## Building

Requires Go 1.24+. No C compiler is needed: the Windows build is pure Go
(`CGO_ENABLED=0`) and can be cross-compiled from Linux or macOS.

```sh
scripts/build.sh 1.0.0      # -> dist/WiFiSpeedTester-Portable-1.0.0-win-x64.zip (+ win-arm64)
go test ./...               # unit and end-to-end engine tests
```

CI (`.github/workflows/build.yml`) runs `go vet` and the tests on Linux and
Windows and builds both zips as artifacts. It also runs a real `--cli` smoke
test on a Windows runner.

To publish a release: **Actions → build → Run workflow**, choose `main`, and
enter a version such as `1.0.1`. The workflow builds both zips and
`SHA256SUMS.txt`, creates the tag `v1.0.1`, and publishes the release with
the notes from `packaging/RELEASE_NOTES.md` plus the changes since the last
release. Pushing a `v*` tag does the same.

The icon, DPI manifest and version info come from
`cmd/wifispeedtester/rsrc_windows_*.syso`. Regenerate them after changing
`winres/` with:

```sh
go run github.com/tc-hib/go-winres@v0.3.3 make --in winres/winres.json --out cmd/wifispeedtester/rsrc --arch amd64,arm64
```

### Working on the UI without Windows

The window is an embedded web page (`internal/ui/web`) hosted in WebView2. The
same page and API can be served locally:

```sh
go run ./cmd/fakespeedserver &          # local speed-test server, bandwidth-capped
# point a WiFiSpeedTester.config.json at http://127.0.0.1:8099/down?bytes={bytes} and /up
go run ./cmd/wifispeedtester --serve 8088   # then open http://127.0.0.1:8088
```

## Architecture

Measurement logic never lives in the UI:

```
Network test (engine, measure, network)
  → Test result model (model)
  → Analysis (analysis)
  → View-model / API (app)
  → UI (ui: WebView2 window or dev server → web/app.js)
```

```
cmd/wifispeedtester   entry point: window, --cli, --serve
internal/network      adapter detection (GetAdaptersAddresses), Wi-Fi info (wlanapi), DNS/internet checks
internal/measure      ping (IcmpSendEcho), download/upload testers, jitter / packet-loss / unit math
internal/engine       runs the full test, emits progress events, cancellation, user-facing errors
internal/analysis     ConnectionQualityEngine: deterministic rating, notes and tips
internal/i18n         English / Indonesian text for everything the Go side writes
internal/servers      LibreSpeed server list and latency ranking
internal/update       latest-release check on GitHub
internal/export       TXT / JSON / CSV with masking of identifying fields
internal/history      local JSON history
internal/config       configuration and defaults
internal/app          view-model: the API the UI calls; long work runs off the UI thread
internal/ui           WebView2 host, Save dialog, dev server, embedded web assets
```

## Scope

This is a simple connection-testing utility. It is not a monitoring
platform, packet analyzer, router admin tool, Wi-Fi hacking tool or
vulnerability scanner, and it is not meant to become one.

## Credits

- **[@tehgeii](https://github.com/tehgeii)**: product owner. Wrote the
  [product brief](docs/PRODUCT_BRIEF.md) and owns the project.
- **Claude Code**: implementation, tests and CI, built from that brief.
