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
| Speed is well below your plan (test with a cable, then use the ISP report) | a plan is set in Settings and download or upload is below `planLowPct` (default 50 %) of it |

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
- **Share of your plan** ("Download is 62% of your 100 Mbps plan") when a
  plan is set in Settings, and a note when the Wi-Fi link speed itself is
  below the plan, so the plan cannot be reached over this Wi-Fi.

The plan only adds notes and a tip; it never changes the rating.

## What is it good for?

After each stage of a test (ping, download, upload) the app judges everyday
activities, so the list fills in while the test runs. The code is
`internal/analysis/suitability.go`. Each activity needs every listed value to
be at least as good as its threshold:

| Activity | Smooth when | Usable (may stutter) when |
|---|---|---|
| Browsing & chat | download ≥ 5 Mbps, ping ≤ 100 ms, loss ≤ 1 % | download ≥ 1.5, ping ≤ 300, loss ≤ 5 |
| YouTube 1080p | download ≥ 8 Mbps, loss ≤ 2 % | download ≥ 5, loss ≤ 5 |
| YouTube / Netflix 4K | download ≥ 25 Mbps, loss ≤ 2 % | download ≥ 15, loss ≤ 5 |
| HD video calls | download ≥ 4, upload ≥ 3 Mbps, ping ≤ 100 ms, jitter ≤ 30 ms, loss ≤ 1 % | download ≥ 1.5, upload ≥ 1, ping ≤ 200, jitter ≤ 50, loss ≤ 3 |
| Online games | ping ≤ 60 ms, jitter ≤ 15 ms, loss ≤ 0.5 %, latency increase under load ≤ 30 ms | ping ≤ 100, jitter ≤ 30, loss ≤ 2, increase ≤ 80 |
| Live streaming | upload ≥ 6 Mbps, loss ≤ 1 % | upload ≥ 3, loss ≤ 3 |
| Large downloads | download ≥ 56 Mbps (50 GB in about 2 hours) | download ≥ 14 (about 8 hours) |

Otherwise the activity is "not enough", with the value that falls short. An
activity is shown as soon as one of its values is already "not enough", even
if other measurements are still running. Quick Ping judges only browsing,
video calls and games, from latency alone. Latency under load is optional:
when it was not measured, games are judged without it.

## Bufferbloat grade

The grade shows how much latency rises while the connection is busy (the
larger of the download and upload increase), on the scale common bufferbloat
tests use:

| Grade | Increase under load |
|---|---|
| A+ | under 5 ms |
| A | under 30 ms |
| B | under 60 ms |
| C | under 200 ms |
| D | under 400 ms |
| F | 400 ms or more |

It is shown only when latency under load was measured. It does not change
the overall rating; the profile's `loadedLatency` band does that.

## Automatic diagnosis

The Diagnosis tab runs up to three normal tests and compares them. A test
counts as *healthy* when its rating is EXCELLENT or GOOD, the router answers
normally (no loss and no unusual delay to the gateway), and the download is
not below `planLowPct` of the plan when one is set.

1. **Where you are** (always).
2. **Near the router**, about 2 m away (only offered on a wireless link).
3. **On a LAN cable** (optional). If Windows still reports Wi-Fi, the step is
   ignored and the conclusion says so.

A step is skipped once the answer is already clear: a healthy or offline
first test, a test that is already on a cable, or a good result near the
router.

| What the tests show | Conclusion | Offered next |
|---|---|---|
| The router answers but the internet does not | No internet beyond the router | ISP report |
| The router does not answer | Not connected to the router | (advice in the text) |
| Every measured step is healthy | Healthy | Stability monitor |
| Cable healthy, where you were not | The problem is the Wi-Fi | Wi-Fi channel check |
| Near the router healthy, where you were not | Weak Wi-Fi coverage | Wi-Fi channel check |
| Delay or loss already to the router | The router or local Wi-Fi is struggling | Wi-Fi channel check (unless found on a cable) |
| Router normal, still slow or unstable (also on a cable or near the router) | Most likely the ISP or the route | ISP report, stability monitor |
| Wi-Fi only, one test, not healthy | Not clear yet: run it with the near-router step | Stability monitor |

When the test server's name took more than 150 ms to look up, the conclusion
also suggests the DNS test.

## Ping to regions

Each region is timed with 5 TCP connections (median), like the TCP fallback
for ping. Bars use the online-game limits from "What is it good for?": up to
60 ms is good, up to 100 ms is fine, above that is high. Singapore gets its
own verdict because most Southeast Asian game servers run there: up to 40 ms
great, up to 80 ms fine for most games, above that high.

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
