## What's new in 1.2

- **Tools** (new button in the top bar):
  - **Stability monitor**: pings every second for 5–30 minutes to catch short drops and lag spikes, with a live chart, a verdict and a TXT report. It runs only while the window is open.
  - **DNS test**: compares your DNS with Cloudflare, Google and Quad9.
  - **Wi-Fi channel check**: shows how crowded each channel is and suggests a better one.
- **Your ISP plan** (Settings ⚙): results show "% of plan", with a tip when speed is far below it.
- **ISP report** (History): a text summary of your tests to send to your ISP.
- **Extra ping targets** (Settings ⚙): add up to 5, for example a game server.
- Fixes: very slow uploads are now estimated instead of failing; history charts use real time spacing; exporting an old or cancelled result works; opening the app twice brings the existing window to the front; the router and TCP ping labels are translated.

## Download

| Your PC | File |
|---|---|
| Most Windows 10/11 PCs (Intel / AMD) | **`WiFiSpeedTester-Portable-…-win-x64.zip`** |
| Windows on ARM (Snapdragon) | `WiFiSpeedTester-Portable-…-win-arm64.zip` |

1. Download the zip and extract it anywhere.
2. Double-click **`WiFiSpeedTester.exe`**. Nothing is installed.
3. If Windows shows **"Windows protected your PC"**, click **More info → Run anyway**. The app is not code-signed yet, which is why SmartScreen warns.

The app needs the Microsoft Edge WebView2 Runtime, which is built into Windows 11 and most Windows 10 installations. If it is missing, the app offers the download link.

`SHA256SUMS.txt` lists checksums so you can verify the download (`certutil -hashfile <zip> SHA256`).
