WiFi Speed + Ping Tester (portable)
===================================

Check your speed. Measure your ping. Understand your connection.

RUN
  Extract the zip anywhere and double-click WiFiSpeedTester.exe.
  No installer, no registry entries, no background service, no startup entry.

REQUIREMENTS
  Windows 10 or 11 (64-bit) with the Microsoft Edge WebView2 Runtime.
  WebView2 is built into Windows 11 and most Windows 10 installations.
  If it is missing the app offers to open the download page.

WHAT IT MEASURES
  Download, upload, ping (min/avg/max), jitter, packet loss, latency under
  load, plus network and Wi-Fi details (SSID, signal, band, channel, link
  speed). Wi-Fi link speed is the radio rate to your router, NOT your
  internet speed.

PRIVACY
  Testing is local except for the traffic needed to measure your connection:
  speed-test data to the test server (default: Cloudflare,
  speed.cloudflare.com) and ping to your router, 1.1.1.1 and 8.8.8.8.
  Only when you ask for it: the public server list from librespeed.org
  (Server -> Find nearby servers). If "Check for updates" is on (Settings,
  on by default): one request to api.github.com for the latest version.
  The app does not read your files, browser history or passwords.

FILES IT CREATES
  WiFiSpeedTester-data\   next to the exe: test history and UI settings
                          (falls back to %APPDATA%\WiFiSpeedTester if the
                          folder is read-only)
  %LOCALAPPDATA%\WiFiSpeedTester\WebView2   browser engine cache
  Delete these folders to remove every trace of the app.

CONFIGURATION (optional)
  Rename WiFiSpeedTester.config.example.json to WiFiSpeedTester.config.json
  and edit it to change test servers, ping targets, test durations and the
  connection-quality thresholds. Any field you remove keeps its default.

COMMAND LINE
  WiFiSpeedTester.exe --cli            run a test in the terminal
  WiFiSpeedTester.exe --cli --gaming   use the gaming profile
  WiFiSpeedTester.exe --cli --quick    ping only (about 5 seconds)
  WiFiSpeedTester.exe --cli --lang id  report in Bahasa Indonesia
  WiFiSpeedTester.exe --cli --json     print the full result as JSON
  WiFiSpeedTester.exe --list-servers   rank public LibreSpeed servers

ACCURACY
  Results depend on the test server, routing, your ISP, Wi-Fi signal, other
  traffic, VPNs and this device. They show what was measured at that moment,
  not a guaranteed ISP speed.

Windows 11: to see the Wi-Fi network name (SSID), Windows may require
Settings > Privacy & security > Location > "Let desktop apps access your
location" to be on. Everything else works without it.
