# Connection Quality Rating

The rating is deterministic: the same measurements and thresholds always give
the same result. The code is `internal/analysis/quality.go`, and the
thresholds live in the `profiles` section of the config
(`WiFiSpeedTester.config.json`; run `WiFiSpeedTester.exe --print-config` to see
the defaults).

## How it works

1. Each metric gets a grade from its band in the active profile:
   - Lower is better (ping, jitter, packet loss, latency increase under
     load): `value <= excellent` → EXCELLENT, `<= good` → GOOD,
     `<= fair` → FAIR, otherwise POOR.
   - Higher is better (download, upload): `value >= excellent` → EXCELLENT,
     `>= good` → GOOD, `>= fair` → FAIR, otherwise POOR.
2. The overall level is the **worst grade** among the metrics listed in the
   profile's `use` array. Other metrics are still shown but do not lower the
   rating.
3. **UNSTABLE** overrides the result when packet loss `>= unstableLossPct`
   or jitter `>= unstableJitterMs`.
4. A metric that could not be measured, such as a failed upload, is skipped
   and a note says so. If fewer than half of the profile's metrics were
   measured, the level is UNKNOWN, so the app never shows EXCELLENT from
   upload alone, for example.

"Latency under load" is the median ping while the download or upload is
running, minus the idle ping. The larger of the two is used.

## Default thresholds

### General profile

Uses ping, jitter, packet loss, download and upload.

| Metric                  | Excellent | Good    | Fair    |
|-------------------------|-----------|---------|---------|
| Ping (ms)               | ≤ 20      | ≤ 50    | ≤ 100   |
| Jitter (ms)             | ≤ 5       | ≤ 15    | ≤ 30    |
| Packet loss (%)         | 0         | ≤ 0.5   | ≤ 2     |
| Download (Mbps)         | ≥ 100     | ≥ 25    | ≥ 10    |
| Upload (Mbps)           | ≥ 20      | ≥ 5     | ≥ 2     |
| Unstable when           | loss ≥ 5 % or jitter ≥ 50 ms |||

### Gaming profile

Uses ping, jitter, packet loss and latency under load. Bandwidth is shown but
not graded, because responsiveness matters more than raw speed for games.
Gaming mode also sends twice as many pings for a steadier jitter figure.

| Metric                          | Excellent | Good | Fair |
|---------------------------------|-----------|------|------|
| Ping (ms)                       | ≤ 20      | ≤ 40 | ≤ 70 |
| Jitter (ms)                     | ≤ 3       | ≤ 8  | ≤ 20 |
| Packet loss (%)                 | 0         | 0    | ≤ 1  |
| Latency increase under load (ms)| ≤ 10      | ≤ 30 | ≤ 80 |
| Unstable when                   | loss ≥ 2 % or jitter ≥ 30 ms |||

Ping is network latency, not frame rate. The gaming rating says nothing about
FPS.

### Quick Ping profile

Uses ping, jitter and packet loss only, with the gaming thresholds (no
download or upload is run).

## Why this rating?

The result card has a **Why this rating?** section. It lists every measured
metric with its value, its grade, and the thresholds of the active profile.
Metrics the profile does not count are shown greyed out and marked "not
counted".

## Tips the app adds

Tips are concrete actions. Each rule fires only on a measured problem, so a
healthy connection gets no tips.

| Tip | When |
|---|---|
| Turn on SQM / Smart Queue / QoS | latency under load is above the profile's `good` value, or the bufferbloat note fired |
| Move closer / reduce obstacles | Wi-Fi signal below 60 % |
| Use the 5 GHz network | connected on 2.4 GHz |
| Test once with a LAN cable | Wi-Fi with local delay/loss, jitter graded FAIR/POOR, or any packet loss |
| Restart the router / change Wi-Fi channel | delay or loss already appears at the router |
| VPN is active | the active adapter is a VPN |
| Try a nearer server | ping graded FAIR/POOR while the router looks fine |
| Pause other downloads/streaming | download or upload graded FAIR/POOR |
| Contact your ISP with a report | problems appear beyond the router, or loss on a wired link |

## Notes the app adds

- **Latency increases significantly under load** when the increase is above
  the profile's `fair` value, or above 30 ms and more than twice the idle
  ping. This can be an early sign of bufferbloat, but the app does not claim a
  diagnosis.
- **Where the problem is.** The app compares the router (gateway) ping with
  the internet ping:
  - loss, more than 20 ms average or more than 10 ms jitter to the router →
    likely a local network or Wi-Fi problem;
  - a healthy router but more than 80 ms of extra latency, or loss, on the internet
    target → added beyond the router (ISP or route);
  - no reply from the router → many routers block ping, so this alone is not
    a fault.
- **Weak Wi-Fi signal** when signal quality is below 40 %.
- A reminder that **Wi-Fi link speed is not internet speed**.

## Ping, jitter and packet loss definitions

- **Ping**: average round-trip time to the first internet ping target that
  answered (default 1.1.1.1). Min and max are shown in Details. If no ICMP
  target answers, the time to open a TCP connection to the test server is
  used instead, and the app adds a note.
- **Jitter**: mean absolute difference between consecutive ping samples. It
  measures how stable latency is, not how fast the connection is.
- **Packet loss**: lost probes ÷ sent probes over the internet targets.
  A target that never answers at all is listed in Details but left out of
  the headline loss, because that usually means the target blocks ICMP, not
  that packets are lost.
