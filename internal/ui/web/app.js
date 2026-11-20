'use strict';
// UI only: renders state and forwards user actions. All measurements run in
// the Go engine, which reports progress through events.

// ---- bridge -------------------------------------------------------------
// In the desktop app WebView2 exposes window.wstCall; in the development
// server the same methods are reached over HTTP and events over SSE.
const bridge = {
  call(method, ...args) {
    if (typeof window.wstCall === 'function') return window.wstCall(method, args);
    return fetch('/api/' + method, { method: 'POST', body: JSON.stringify(args) })
      .then(r => r.json())
      .then(j => { if (j.error) throw new Error(j.error); return j.result; });
  },
  listen(cb) {
    window.__wstEvent = cb;
    if (typeof window.wstCall !== 'function') {
      const es = new EventSource('/events');
      es.onmessage = e => { const m = JSON.parse(e.data); cb(m.name, m.payload); };
    }
  },
};

// ---- text ---------------------------------------------------------------
// Static labels and UI messages. Texts produced by the engine (notes, tips,
// errors, check details) come already translated from Go.
const TEXT = {
  en: {
    tagline: 'Check your speed. Measure your ping. Understand your connection.',
    'aria.mode': 'Test mode', 'aria.unit': 'Speed unit',
    'mode.general': 'General', 'mode.gaming': 'Gaming',
    history: 'History', settings: 'Settings', close: 'Close', back: 'Back',
    'update.download': 'Download',
    'update.available': 'Version {0} is available (you have {1}).',
    'update.latest': 'You have the latest version ({0}).',
    'update.error': 'Could not check for updates: {0}',
    'update.checking': 'Checking…',
    'net.connection': 'Connection', 'net.network': 'Network', 'net.detecting': 'Detecting…',
    'net.signal': 'Signal', 'net.link': 'Wi-Fi Link Speed',
    'net.linkHint': 'Radio rate between this device and the router. Not your internet speed.',
    'net.notConnected': 'Not connected',
    'gauge.ready': 'Ready', 'gauge.starting': 'Starting…', 'gauge.cancelling': 'Cancelling…',
    'gauge.cancelled': 'Cancelled', 'gauge.failed': 'Failed', 'gauge.checking': 'Checking connection…',
    start: 'Start Test', cancel: 'Cancel',
    quick: 'Quick Ping', 'quick.hint': 'About 5 seconds: connection check and ping only, no download or upload.',
    server: 'Server', 'server.hint': 'Choose the test server',
    'quality.title': 'Connection Quality', 'quality.gaming': 'Gaming Connection', 'quality.quick': 'Quick Ping',
    'stage.ping': 'Ping Test',
    'fail.reason': 'Reason:', 'fail.suggestion': 'Suggestion:', 'fail.tech': 'Technical details',
    'gaming.title': 'Gaming Connection', 'gaming.stability': 'Latency Stability',
    'tips.title': 'What you can do',
    'why.title': 'Why this rating?',
    'why.intro': 'Each measured value gets a grade. The overall rating is the worst grade among the metrics this profile counts.',
    'why.notCounted': 'shown only, not counted in this profile',
    'why.lower': 'Excellent ≤ {0} · Good ≤ {1} · Fair ≤ {2} {3}',
    'why.higher': 'Excellent ≥ {0} · Good ≥ {1} · Fair ≥ {2} {3}',
    details: 'Details', copy: 'Copy result', 'export.png': 'PNG card',
    'out.share': 'Share', 'out.save': 'Save as file',
    'export.sensitive': 'Include network identifiers in TXT, JSON and CSV',
    'privacy.hidden': 'Hidden in the files: {0}. Copy result and the PNG card never contain them.',
    'privacy.shown': 'The files will contain your full {0}. Share them only with people you trust, such as your ISP. Copy result and the PNG card still never contain them.',
    'privacy.ssid': 'Wi-Fi name', 'privacy.ip': 'IP addresses', 'privacy.bssid': 'the router\'s hardware address (BSSID)',
    and: 'and',
    'history.csvNames': 'with Wi-Fi names', 'history.csvNamesHint': 'Adds each test\'s Wi-Fi name to the CSV. The CSV never contains IP addresses.',
    'details.ping': 'Ping Details', 'col.target': 'Target', 'col.sent': 'Sent', 'col.received': 'Received', 'col.avg': 'Average',
    'details.loaded': 'Latency Under Load', 'details.idle': 'Idle Ping', 'details.duringDown': 'During Download', 'details.duringUp': 'During Upload',
    'details.throughput': 'Throughput', 'details.network': 'Network',
    'history.title': 'Test History', 'history.csv': 'Export CSV', 'history.clear': 'Clear', 'history.all': 'All tests',
    'history.empty': 'No tests yet.', 'history.storedAt': 'Stored only on this computer: {0}',
    'history.confirmClear': 'Delete all saved test results?',
    'range.7': '7 days', 'range.30': '30 days', 'range.all': 'All', 'conn.all': 'All',
    'trend.speed': 'Speed', 'trend.ping': 'Ping (ms)', 'trend.noData': 'Not enough tests in this selection yet.',
    'trend.unit': 'Mbps',
    'compare.title': 'Average by connection', 'compare.tests': 'Tests',
    'col.date': 'Date', 'col.quality': 'Quality',
    'server.title': 'Test server', 'server.current': 'Current',
    'server.intro': 'The default is Cloudflare. You can pick a nearby public LibreSpeed server instead; Cloudflare stays as a backup if it fails.',
    'server.find': 'Find nearby servers', 'server.useDefault': 'Use default ({0})',
    'server.name': 'Server', 'server.sponsor': 'Provider', 'server.latency': 'Latency',
    'server.privacy': 'Finding servers downloads the public list from librespeed.org and contacts each server once to measure latency.',
    'server.fetching': 'Downloading the server list…', 'server.probing': 'Measuring latency… {0}/{1}',
    'server.found': '{0} reachable servers, fastest first. Click one to use it.',
    'server.none': 'No server answered. Check your connection and try again.',
    'server.error': 'Could not get the server list: {0}', 'server.saved': 'Test server set to {0}.',
    'server.default': '{0} (default)',
    'settings.lang': 'Language', 'settings.update': 'Check for updates when the app starts',
    'settings.updateHint': 'Asks GitHub for the latest release. Nothing else is sent.', 'settings.checkNow': 'Check now',
    'about.copyright': '© 2026–present tehgeii. All rights reserved. See LICENSE.txt for what you may do with the app.',
    'about.title': 'About and privacy', 'about.privacy': 'Privacy', 'about.accuracy': 'Accuracy', 'about.config': 'Configuration',
    'about.intro': 'WiFi Speed + Ping Tester — a portable utility that measures download, upload, ping, jitter and packet loss.',
    'about.traffic': 'Testing is local except for the traffic needed to measure your connection:',
    'about.local': 'The app does not read your files, browser history or passwords, and does not run in the background. History is stored only on this computer:',
    'about.accuracyText': 'Results depend on the test server, routing, Wi-Fi signal, other traffic, VPNs and this device. They show what was measured now, not a guaranteed ISP speed.',
    'about.configText': 'Place this file next to the .exe to change test servers, ping targets, durations and quality thresholds:',
    'about.trafficSpeed': 'Speed test data to: {0}', 'about.trafficPing': 'Ping (ICMP, or TCP if ICMP is blocked) to: {0}',
    'about.trafficDNS': 'A DNS lookup of the test server name', 'about.trafficUpdate': 'If enabled in Settings: an update check to api.github.com',
    'about.trafficDNSTest': 'Only when you run the DNS test: lookups of popular site names to your DNS and {0}',
    'about.trafficMonitor': 'Only while the stability monitor runs: one ping per second to the chosen target and your router',
    'about.trafficScan': 'The Wi-Fi channel check sends nothing; it reads the networks Windows already sees',
    'about.trafficSignal': 'The Wi-Fi signal finder sends nothing; it reads the signal from Windows',
    'about.trafficRegions': 'Only when you run Ping to regions: a few TCP connections, without data, to {0}',
    'about.cfgError': 'Config error (defaults in use): {0}', 'about.cfgLoaded': 'Loaded: {0}', 'about.cfgNone': 'No config file found — using built-in defaults.',
    privacy: 'Tests send traffic only to the speed-test server ({0}) and the ping targets. Nothing else leaves this computer.',
    live: 'live', measuring: 'measuring…', failed: 'failed', incomplete: 'incomplete', notAvailable: 'not available',
    notInQuick: 'not part of Quick Ping',
    underLoad: 'ping under load {0} ms', minmax: 'min {0} · max {1}', stability: 'latency stability',
    received: '{0} of {1} received', noReplies: 'No replies — the target may block ping. Not counted in the headline loss.',
    pingNotMeasured: 'Ping was not measured.', notMeasured: 'not measured', noData: 'No data', peak: 'peak {0}',
    'toast.cancelled': 'Test cancelled.', 'toast.saved': 'Saved to {0}', 'toast.exportFailed': 'Export failed: {0}',
    'toast.copied': 'Result copied — paste it into WhatsApp, Discord or anywhere.', 'toast.copyFailed': 'Could not copy: {0}',
    'toast.configError': 'Config file problem — using defaults: {0}',
    'row.adapter': 'Adapter', 'row.description': 'Description', 'row.frequency': 'Frequency', 'row.standard': 'Wi-Fi standard',
    'row.link': 'Wi-Fi link speed', 'row.linkVal': '{0} Mbps receive / {1} Mbps transmit (radio rate, not internet speed)',
    serverLine: 'Test server: {0} · {1} parallel connections · started {2}',
    'card.title': 'INTERNET TEST', 'card.footer': 'WiFi Speed + Ping Tester · measured values, not a guaranteed ISP speed',
    'card.general': 'CONNECTION', 'card.gaming': 'GAMING CONNECTION', 'card.quick': 'QUICK PING',
    'level.EXCELLENT': 'EXCELLENT', 'level.GOOD': 'GOOD', 'level.FAIR': 'FAIR', 'level.POOR': 'POOR', 'level.UNSTABLE': 'UNSTABLE', 'level.UNKNOWN': 'UNKNOWN',
    'profile.general': 'General profile', 'profile.gaming': 'Gaming profile', 'profile.quick': 'Quick Ping',
    'wifi.hotspot': 'Looks like a phone hotspot (a metered connection or a typical hotspot address).',
    'wifi.ssidHidden': 'Windows hides the Wi-Fi name. On Windows 11, turn on Settings → Privacy & security → Location → Let desktop apps access your location.',
    'wifi.unavailable': 'Wi-Fi details unavailable: {0}',
    'conn.Wi-Fi': 'Wi-Fi', 'conn.Ethernet': 'Ethernet', 'conn.Hotspot / Tethering': 'Hotspot / Tethering', 'conn.Cellular': 'Cellular',
    'conn.VPN / Virtual': 'VPN / Virtual', 'conn.Unknown': 'Unknown',
    tools: 'Tools', 'tools.title': 'Tools', 'tab.test': 'Test', 'aria.tabs': 'Sections',
    'mon.short': 'Catch short drops and lag spikes', 'dns.short': 'Is your DNS slowing websites down?', 'scan.short': 'Find a less crowded channel',
    'act.title': 'What is it good for?', 'act.quickNote': 'Quick Ping only checks latency, so activities that need speed are left out.',
    'act.s.ok': 'Smooth', 'act.s.warn': 'Usable, may stutter', 'act.s.bad': 'Not enough', 'act.s.pending': 'Measuring', 'act.s.na': 'Not measured',
    'grade.title': 'Bufferbloat grade', 'grade.text': 'ping rises {0} ms while the connection is busy',
    'grade.hint': 'How much ping rises while the connection is busy downloading or uploading. A+ and A are great for games and calls.',
    'settings.theme': 'Theme', 'theme.system': 'Follow Windows', 'theme.light': 'Light', 'theme.dark': 'Dark',
    'card.grade': 'Bufferbloat {0}',
    saver: 'Data saver', 'saver.hint': 'Shorter test that stops early to save mobile data: about {0} MB per test.',
    repeat: 'Repeat 3×', 'repeat.hint': 'Runs the test 3 times and shows the middle values: steadier results and better evidence for your ISP. Uses 3× the data.',
    'data.live': 'Data used: {0}', 'data.used': 'This test used ≈ {0} of data', 'data.estimate': 'Estimated data: ≈ {0}',
    'data.saverOn': 'Data saver on: about {0} per test', 'data.turnOn': 'Turn on',
    'data.hotspot': "You're on a phone hotspot or mobile data. Turn on Data saver?",
    'data.metered': 'Windows marks this connection as metered (it has a data limit). Turn on Data saver?',
    'data.month': 'Data used by tests this month: {0} ({1} tests)', 'data.perTest': 'per test',
    'run.label': 'Test {0} of {1}', 'runs.title': '{0} runs (the figures above are the middle values)', 'runs.col': 'Run',
    'hist.saver': 'data saver',
    'history.reportPdf': 'ISP report (PDF)', 'history.reportTxt': 'TXT', 'history.reportHint': 'A report of the selected tests, with charts, that you can send to your ISP. TXT is the same as plain text',
    'col.plan': '% of plan', 'plan.of': '{0}% of plan',
    'plan.title': 'Your internet plan', 'plan.intro': 'Optional. Results then show what share of your plan you actually get.',
    'plan.saved': 'Plan saved.',
    'targets.title': 'Extra ping targets', 'targets.intro': 'For example a game server. Each test also pings these (up to 5), and the stability monitor can watch them.',
    'targets.add': 'Add target', 'targets.save': 'Save targets', 'targets.saved': 'Saved.', 'targets.label': 'Name', 'targets.host': 'IP address or host name',
    'targets.remove': 'Remove', 'targets.max': 'Up to {0} targets.',
    'settings.other': 'Other',
    'mon.title': 'Stability monitor',
    'mon.intro': 'Pings every second while this window is open, to catch short drops and lag spikes that a single test misses. It stops when you press Stop or close the app; nothing runs in the background.',
    'mon.duration': 'Duration', 'mon.target': 'Target', 'mon.start': 'Start monitor', 'mon.stop': 'Stop', 'mon.stopping': 'Stopping…',
    'mon.elapsed': 'Time', 'mon.routerLost': 'Router lost', 'mon.lostKey': 'no reply', 'mon.export': 'Export report (TXT)',
    'mon.lost': 'lost', 'mon.tcpNote': 'ICMP ping is blocked here, so the monitor times TCP connections to the test server instead.',
    'mon.error': 'The monitor could not start: {0}', 'mon.busy': 'The stability monitor is running. Stop it first to run a speed test.',
    'mon.noRouter': 'not measured', 'min': 'min',
    'dns.title': 'DNS test',
    'dns.intro': 'DNS turns website names into addresses. A slow DNS makes every site start loading late. This compares your current DNS with public ones.',
    'dns.start': 'Test DNS', 'dns.server': 'DNS server', 'dns.median': 'Typical answer time', 'dns.ok': 'Answered',
    'dns.running': 'Testing… {0}/{1}', 'dns.noAnswer': 'no answer',
    'scan.title': 'Wi-Fi channel check',
    'scan.intro': 'Lists the Wi-Fi networks Windows can see around you (read-only) and checks whether your channel is crowded.',
    'scan.start': 'Scan Wi-Fi', 'scan.scanning': 'Scanning… (about 5 seconds)', 'scan.found': '{0} networks found.',
    'scan.error': 'Scan failed: {0}', 'scan.channels': 'Channels', 'scan.networks': 'Networks', 'scan.busy': 'Crowding',
    'scan.nearby': 'Nearby networks', 'scan.name': 'Name', 'scan.hidden': '(hidden network)',
    'scan.yours': 'your channel', 'scan.suggested': 'suggested', 'scan.connected': 'connected',
    'busy.low': 'low', 'busy.medium': 'medium', 'busy.high': 'high',
    'prob.btn': 'Report a problem', 'prob.head': 'Something not working?',
    'prob.intro': 'Tell the developer what went wrong. This information goes with it; it holds no IP address, Wi-Fi name or device name.',
    'prob.issue': 'Open GitHub Issues', 'prob.copy': 'Copy info', 'prob.copied': 'Info copied. Paste it into your message.',
    'prob.copyHint': 'No GitHub account? Copy the info and send it with WhatsApp or email.',
    'tab.diag': 'Diagnosis',
    'diag.title': 'Automatic diagnosis',
    'diag.intro': 'Finds out whether a slow connection comes from the Wi-Fi, the router or the ISP, in up to three short tests. Just follow the steps; nothing on your router is changed. These tests are not saved in History.',
    'diag.s.now': 'Test where you are', 'diag.s.near': 'Test near the router', 'diag.s.cable': 'Test with a LAN cable (optional)', 'diag.s.result': 'Conclusion',
    'diag.start': 'Start diagnosis', 'diag.skip': 'Skip this step', 'diag.again': 'Run again', 'diag.skipped': 'skipped',
    'diag.skippedStrong': 'not needed ({0}% signal)',
    'diag.cableMsgStrong': 'The Wi-Fi signal where you are is already strong ({0}%), so a test near the router would not change anything. Optional: plug a LAN cable from this laptop into the router and wait until Windows shows the wired connection. No cable? Skip this step.',
    'diag.ready': 'Step 1 runs a full test right where you normally use the internet (less than a minute).',
    'diag.nearMsg': 'Now take this laptop close to the router: about 2 meters away, in the same room, with no wall in between. Then start the next test.',
    'diag.goNear': 'Test near the router',
    'diag.cableMsg': 'Optional: plug a LAN cable from this laptop into the router and wait until Windows shows the wired connection. No cable? Skip this step.',
    'diag.goCable': 'Test with the cable', 'diag.skipToResult': 'Skip, show the conclusion',
    'diag.cancelled': 'The test was cancelled. You can run this step again.',
    'diag.act.channel': 'Check the Wi-Fi channel', 'diag.act.dns': 'Test DNS', 'diag.act.ispReport': 'Go to the ISP report', 'diag.act.monitor': 'Open the stability monitor',
    'sig.title': 'Wi-Fi signal finder', 'sig.short': 'Find the best spot, live',
    'sig.intro': 'Walk around with this laptop: the signal updates every second, so you can find the best spot for the laptop or for the router. It uses no internet data and stops when you leave this tool.',
    'sig.start': 'Start', 'sig.beep': 'Beep (faster = stronger)', 'sig.chart': 'Signal (%), last 60 seconds', 'sig.now': 'now',
    'sig.note': 'Windows refreshes the signal value every 1–5 seconds, depending on the Wi-Fi card.',
    'sig.spotName': 'Spot name, e.g. Bedroom', 'sig.save': 'Save this spot', 'sig.spots': 'Saved spots, best first', 'sig.spot': 'Spot', 'sig.clear': 'Clear spots',
    'sig.good': 'Strong', 'sig.fair': 'Fair', 'sig.weak': 'Weak',
    'sig.error': 'Cannot read the Wi-Fi signal: {0}', 'sig.saved': 'Saved "{0}".',
    'reg.title': 'Ping to regions', 'reg.short': 'Jakarta, Singapore, Tokyo…',
    'reg.intro': 'Latency to the cloud regions where game and streaming servers run. Most Southeast Asian game servers are in Singapore. This is the latency to the region, not to one particular game server.',
    'reg.start': 'Measure', 'reg.region': 'Region', 'reg.running': 'Measuring… {0}/{1}', 'reg.noAnswer': 'no answer',
  },
  id: {
    tagline: 'Cek kecepatan. Ukur ping. Pahami koneksimu.',
    'aria.mode': 'Mode tes', 'aria.unit': 'Satuan kecepatan',
    'mode.general': 'Umum', 'mode.gaming': 'Gaming',
    history: 'Riwayat', settings: 'Pengaturan', close: 'Tutup', back: 'Kembali',
    'update.download': 'Download',
    'update.available': 'Versi {0} sudah tersedia (kamu memakai {1}).',
    'update.latest': 'Kamu sudah memakai versi terbaru ({0}).',
    'update.error': 'Gagal mengecek update: {0}',
    'update.checking': 'Mengecek…',
    'net.connection': 'Koneksi', 'net.network': 'Jaringan', 'net.detecting': 'Mendeteksi…',
    'net.signal': 'Sinyal', 'net.link': 'Wi-Fi Link Speed',
    'net.linkHint': 'Kecepatan radio antara perangkat ini dan router. Bukan kecepatan internet kamu.',
    'net.notConnected': 'Tidak terhubung',
    'gauge.ready': 'Siap', 'gauge.starting': 'Memulai…', 'gauge.cancelling': 'Membatalkan…',
    'gauge.cancelled': 'Dibatalkan', 'gauge.failed': 'Gagal', 'gauge.checking': 'Mengecek koneksi…',
    start: 'Mulai Tes', cancel: 'Batal',
    quick: 'Ping Cepat', 'quick.hint': 'Sekitar 5 detik: cek koneksi dan ping saja, tanpa download atau upload.',
    server: 'Server', 'server.hint': 'Pilih server tes',
    'quality.title': 'Kualitas Koneksi', 'quality.gaming': 'Koneksi Gaming', 'quality.quick': 'Ping Cepat',
    'stage.ping': 'Tes Ping',
    'fail.reason': 'Penyebab:', 'fail.suggestion': 'Saran:', 'fail.tech': 'Detail teknis',
    'gaming.title': 'Koneksi Gaming', 'gaming.stability': 'Stabilitas Ping',
    'tips.title': 'Yang bisa kamu lakukan',
    'why.title': 'Kenapa rating-nya segini?',
    'why.intro': 'Setiap nilai yang terukur diberi nilai. Rating akhir adalah nilai terburuk dari metrik yang dihitung oleh profil ini.',
    'why.notCounted': 'hanya ditampilkan, tidak dihitung di profil ini',
    'why.lower': 'Sangat bagus ≤ {0} · Bagus ≤ {1} · Cukup ≤ {2} {3}',
    'why.higher': 'Sangat bagus ≥ {0} · Bagus ≥ {1} · Cukup ≥ {2} {3}',
    details: 'Detail', copy: 'Salin hasil', 'export.png': 'Kartu PNG',
    'out.share': 'Bagikan', 'out.save': 'Simpan sebagai file',
    'export.sensitive': 'Sertakan identitas jaringan di TXT, JSON dan CSV',
    'privacy.hidden': 'Disembunyikan di file: {0}. Salin hasil dan kartu PNG tidak pernah memuat data ini.',
    'privacy.shown': 'File akan memuat {0} secara lengkap. Bagikan hanya ke pihak yang kamu percaya, misalnya ISP. Salin hasil dan kartu PNG tetap tidak memuat data ini.',
    'privacy.ssid': 'nama Wi-Fi', 'privacy.ip': 'alamat IP', 'privacy.bssid': 'alamat hardware router (BSSID)',
    and: 'dan',
    'history.csvNames': 'dengan nama Wi-Fi', 'history.csvNamesHint': 'Menambahkan nama Wi-Fi tiap tes ke CSV. CSV tidak pernah memuat alamat IP.',
    'details.ping': 'Detail Ping', 'col.target': 'Target', 'col.sent': 'Terkirim', 'col.received': 'Diterima', 'col.avg': 'Rata-rata',
    'details.loaded': 'Ping Saat Sibuk', 'details.idle': 'Ping Normal', 'details.duringDown': 'Saat Download', 'details.duringUp': 'Saat Upload',
    'details.throughput': 'Throughput', 'details.network': 'Jaringan',
    'history.title': 'Riwayat Tes', 'history.csv': 'Export CSV', 'history.clear': 'Hapus', 'history.all': 'Semua tes',
    'history.empty': 'Belum ada tes.', 'history.storedAt': 'Disimpan hanya di komputer ini: {0}',
    'history.confirmClear': 'Hapus semua hasil tes yang tersimpan?',
    'range.7': '7 hari', 'range.30': '30 hari', 'range.all': 'Semua', 'conn.all': 'Semua',
    'trend.speed': 'Kecepatan', 'trend.ping': 'Ping (ms)', 'trend.noData': 'Belum cukup tes untuk pilihan ini.',
    'trend.unit': 'Mbps',
    'compare.title': 'Rata-rata per koneksi', 'compare.tests': 'Tes',
    'col.date': 'Tanggal', 'col.quality': 'Kualitas',
    'server.title': 'Server tes', 'server.current': 'Saat ini',
    'server.intro': 'Bawaannya Cloudflare. Kamu bisa memilih server LibreSpeed publik yang lebih dekat; Cloudflare tetap jadi cadangan kalau server itu gagal.',
    'server.find': 'Cari server terdekat', 'server.useDefault': 'Pakai bawaan ({0})',
    'server.name': 'Server', 'server.sponsor': 'Penyedia', 'server.latency': 'Latensi',
    'server.privacy': 'Pencarian server mengunduh daftar publik dari librespeed.org dan menghubungi tiap server sekali untuk mengukur latensi.',
    'server.fetching': 'Mengunduh daftar server…', 'server.probing': 'Mengukur latensi… {0}/{1}',
    'server.found': '{0} server bisa dijangkau, tercepat di atas. Klik salah satu untuk memakainya.',
    'server.none': 'Tidak ada server yang merespons. Cek koneksi lalu coba lagi.',
    'server.error': 'Gagal mengambil daftar server: {0}', 'server.saved': 'Server tes diganti ke {0}.',
    'server.default': '{0} (bawaan)',
    'settings.lang': 'Bahasa', 'settings.update': 'Cek update saat aplikasi dibuka',
    'settings.updateHint': 'Menanyakan rilis terbaru ke GitHub. Tidak ada data lain yang dikirim.', 'settings.checkNow': 'Cek sekarang',
    'about.copyright': '© 2026–sekarang tehgeii. Seluruh hak dilindungi. Lihat LICENSE.txt untuk apa saja yang boleh dilakukan dengan aplikasi ini.',
    'about.title': 'Tentang dan privasi', 'about.privacy': 'Privasi', 'about.accuracy': 'Akurasi', 'about.config': 'Konfigurasi',
    'about.intro': 'WiFi Speed + Ping Tester — aplikasi portable untuk mengukur download, upload, ping, jitter dan packet loss.',
    'about.traffic': 'Tes berjalan lokal, kecuali lalu lintas yang memang dibutuhkan untuk mengukur koneksi:',
    'about.local': 'Aplikasi ini tidak membaca file, riwayat browser atau password, dan tidak berjalan di latar belakang. Riwayat hanya disimpan di komputer ini:',
    'about.accuracyText': 'Hasil dipengaruhi server tes, rute jaringan, sinyal Wi-Fi, pemakaian lain, VPN dan perangkat ini. Angkanya menunjukkan hasil pengukuran saat ini, bukan jaminan kecepatan dari ISP.',
    'about.configText': 'Taruh file ini di samping .exe untuk mengubah server tes, target ping, durasi dan batas rating:',
    'about.trafficSpeed': 'Data speed test ke: {0}', 'about.trafficPing': 'Ping (ICMP, atau TCP kalau ICMP diblokir) ke: {0}',
    'about.trafficDNS': 'Pencarian DNS untuk nama server tes', 'about.trafficUpdate': 'Kalau diaktifkan di Pengaturan: cek update ke api.github.com',
    'about.trafficDNSTest': 'Hanya saat kamu menjalankan tes DNS: pencarian nama situs populer ke DNS kamu dan {0}',
    'about.trafficMonitor': 'Hanya selama monitor stabilitas berjalan: satu ping per detik ke target yang dipilih dan router kamu',
    'about.trafficScan': 'Cek channel Wi-Fi tidak mengirim apa pun; hanya membaca jaringan yang sudah terlihat oleh Windows',
    'about.trafficSignal': 'Pencari sinyal Wi-Fi tidak mengirim apa pun; hanya membaca sinyal dari Windows',
    'about.trafficRegions': 'Hanya saat kamu menjalankan Ping ke region: beberapa koneksi TCP tanpa data ke {0}',
    'about.cfgError': 'Konfigurasi bermasalah (memakai bawaan): {0}', 'about.cfgLoaded': 'Dimuat: {0}', 'about.cfgNone': 'File konfigurasi tidak ditemukan — memakai pengaturan bawaan.',
    privacy: 'Tes hanya mengirim data ke server speed test ({0}) dan target ping. Tidak ada data lain yang keluar dari komputer ini.',
    live: 'langsung', measuring: 'mengukur…', failed: 'gagal', incomplete: 'tidak lengkap', notAvailable: 'tidak tersedia',
    notInQuick: 'tidak termasuk Ping Cepat',
    underLoad: 'ping saat sibuk {0} ms', minmax: 'min {0} · maks {1}', stability: 'stabilitas ping',
    received: '{0} dari {1} diterima', noReplies: 'Tidak ada balasan — target mungkin memblokir ping. Tidak dihitung di packet loss utama.',
    pingNotMeasured: 'Ping tidak diukur.', notMeasured: 'tidak terukur', noData: 'Tidak ada data', peak: 'puncak {0}',
    'toast.cancelled': 'Tes dibatalkan.', 'toast.saved': 'Disimpan ke {0}', 'toast.exportFailed': 'Export gagal: {0}',
    'toast.copied': 'Hasil disalin — tinggal tempel di WhatsApp, Discord atau di mana saja.', 'toast.copyFailed': 'Gagal menyalin: {0}',
    'toast.configError': 'File konfigurasi bermasalah — memakai bawaan: {0}',
    'row.adapter': 'Adapter', 'row.description': 'Deskripsi', 'row.frequency': 'Frekuensi', 'row.standard': 'Standar Wi-Fi',
    'row.link': 'Wi-Fi link speed', 'row.linkVal': '{0} Mbps terima / {1} Mbps kirim (kecepatan radio, bukan kecepatan internet)',
    serverLine: 'Server tes: {0} · {1} koneksi paralel · dimulai {2}',
    'card.title': 'TES INTERNET', 'card.footer': 'WiFi Speed + Ping Tester · nilai terukur, bukan jaminan kecepatan ISP',
    'card.general': 'KONEKSI', 'card.gaming': 'KONEKSI GAMING', 'card.quick': 'PING CEPAT',
    'level.EXCELLENT': 'SANGAT BAGUS', 'level.GOOD': 'BAGUS', 'level.FAIR': 'CUKUP', 'level.POOR': 'BURUK', 'level.UNSTABLE': 'TIDAK STABIL', 'level.UNKNOWN': 'TIDAK DIKETAHUI',
    'profile.general': 'Profil Umum', 'profile.gaming': 'Profil Gaming', 'profile.quick': 'Ping Cepat',
    'wifi.hotspot': 'Sepertinya hotspot HP (koneksi berkuota atau alamat khas hotspot).',
    'wifi.ssidHidden': 'Windows menyembunyikan nama Wi-Fi. Di Windows 11, nyalakan Settings → Privacy & security → Location → Let desktop apps access your location.',
    'wifi.unavailable': 'Detail Wi-Fi tidak tersedia: {0}',
    'conn.Wi-Fi': 'Wi-Fi', 'conn.Ethernet': 'Ethernet (kabel)', 'conn.Hotspot / Tethering': 'Hotspot / Tethering', 'conn.Cellular': 'Seluler',
    'conn.VPN / Virtual': 'VPN / Virtual', 'conn.Unknown': 'Tidak diketahui',
    tools: 'Alat', 'tools.title': 'Alat', 'tab.test': 'Tes', 'aria.tabs': 'Bagian',
    'mon.short': 'Tangkap putus sesaat dan lonjakan lag', 'dns.short': 'Apakah DNS bikin website lambat?', 'scan.short': 'Cari channel yang lebih sepi',
    'act.title': 'Cocok untuk apa?', 'act.quickNote': 'Ping Cepat hanya mengecek ping, jadi aktivitas yang butuh kecepatan tidak ditampilkan.',
    'act.s.ok': 'Lancar', 'act.s.warn': 'Bisa, kadang tersendat', 'act.s.bad': 'Kurang', 'act.s.pending': 'Mengukur', 'act.s.na': 'Tidak terukur',
    'grade.title': 'Nilai bufferbloat', 'grade.text': 'ping naik {0} ms saat koneksi sedang sibuk',
    'grade.hint': 'Seberapa besar ping naik saat koneksi sibuk download atau upload. A+ dan A sangat bagus untuk game dan video call.',
    'settings.theme': 'Tema', 'theme.system': 'Ikuti Windows', 'theme.light': 'Terang', 'theme.dark': 'Gelap',
    'card.grade': 'Bufferbloat {0}',
    saver: 'Hemat kuota', 'saver.hint': 'Tes lebih singkat yang berhenti lebih awal supaya hemat kuota: sekitar {0} MB per tes.',
    repeat: 'Ulangi 3×', 'repeat.hint': 'Tes dijalankan 3 kali lalu diambil nilai tengahnya: hasil lebih stabil dan bukti lebih kuat untuk ISP. Memakai kuota 3×.',
    'data.live': 'Kuota terpakai: {0}', 'data.used': 'Tes ini memakai kuota ≈ {0}', 'data.estimate': 'Perkiraan kuota: ≈ {0}',
    'data.saverOn': 'Hemat kuota aktif: sekitar {0} per tes', 'data.turnOn': 'Nyalakan',
    'data.hotspot': 'Kamu pakai hotspot HP atau data seluler. Nyalakan Hemat kuota?',
    'data.metered': 'Windows menandai koneksi ini berkuota (ada batas data). Nyalakan Hemat kuota?',
    'data.month': 'Kuota terpakai untuk tes bulan ini: {0} ({1} tes)', 'data.perTest': 'per tes',
    'run.label': 'Tes {0} dari {1}', 'runs.title': '{0} putaran (angka di atas adalah nilai tengahnya)', 'runs.col': 'Putaran',
    'hist.saver': 'hemat kuota',
    'history.reportPdf': 'Laporan ISP (PDF)', 'history.reportTxt': 'TXT', 'history.reportHint': 'Laporan dari tes yang dipilih, lengkap dengan grafik, untuk dikirim ke ISP kamu. TXT berisi hal yang sama dalam teks biasa',
    'col.plan': '% paket', 'plan.of': '{0}% dari paket',
    'plan.title': 'Paket internet kamu', 'plan.intro': 'Opsional. Hasil tes lalu menunjukkan berapa persen dari paket yang benar-benar kamu dapat.',
    'plan.saved': 'Paket disimpan.',
    'targets.title': 'Target ping tambahan', 'targets.intro': 'Misalnya server game. Setiap tes juga ikut nge-ping target ini (maks. 5), dan monitor stabilitas bisa memantaunya.',
    'targets.add': 'Tambah target', 'targets.save': 'Simpan target', 'targets.saved': 'Tersimpan.', 'targets.label': 'Nama', 'targets.host': 'Alamat IP atau nama host',
    'targets.remove': 'Hapus', 'targets.max': 'Maksimal {0} target.',
    'settings.other': 'Lainnya',
    'mon.title': 'Monitor stabilitas',
    'mon.intro': 'Nge-ping setiap detik selama jendela ini terbuka, untuk menangkap putus sesaat dan lonjakan lag yang tidak terlihat di satu kali tes. Berhenti saat kamu menekan Stop atau menutup aplikasi; tidak ada yang berjalan di latar belakang.',
    'mon.duration': 'Durasi', 'mon.target': 'Target', 'mon.start': 'Mulai monitor', 'mon.stop': 'Stop', 'mon.stopping': 'Menghentikan…',
    'mon.elapsed': 'Waktu', 'mon.routerLost': 'Router hilang', 'mon.lostKey': 'tidak dibalas', 'mon.export': 'Export laporan (TXT)',
    'mon.lost': 'hilang', 'mon.tcpNote': 'Ping ICMP diblokir di sini, jadi monitor mengukur waktu koneksi TCP ke server tes sebagai gantinya.',
    'mon.error': 'Monitor tidak bisa dimulai: {0}', 'mon.busy': 'Monitor stabilitas sedang berjalan. Hentikan dulu untuk menjalankan speed test.',
    'mon.noRouter': 'tidak diukur', 'min': 'mnt',
    'dns.title': 'Tes DNS',
    'dns.intro': 'DNS mengubah nama website menjadi alamat. DNS yang lambat membuat setiap website telat mulai terbuka. Tes ini membandingkan DNS kamu sekarang dengan DNS publik.',
    'dns.start': 'Tes DNS', 'dns.server': 'Server DNS', 'dns.median': 'Waktu jawab biasanya', 'dns.ok': 'Dijawab',
    'dns.running': 'Mengetes… {0}/{1}', 'dns.noAnswer': 'tidak menjawab',
    'scan.title': 'Cek channel Wi-Fi',
    'scan.intro': 'Menampilkan jaringan Wi-Fi di sekitarmu yang terlihat oleh Windows (hanya membaca) dan mengecek apakah channel kamu ramai.',
    'scan.start': 'Scan Wi-Fi', 'scan.scanning': 'Memindai… (sekitar 5 detik)', 'scan.found': '{0} jaringan ditemukan.',
    'scan.error': 'Scan gagal: {0}', 'scan.channels': 'Channel', 'scan.networks': 'Jaringan', 'scan.busy': 'Keramaian',
    'scan.nearby': 'Jaringan di sekitar', 'scan.name': 'Nama', 'scan.hidden': '(jaringan tersembunyi)',
    'scan.yours': 'channel kamu', 'scan.suggested': 'disarankan', 'scan.connected': 'terhubung',
    'busy.low': 'sepi', 'busy.medium': 'sedang', 'busy.high': 'ramai',
    'prob.btn': 'Laporkan masalah', 'prob.head': 'Ada yang tidak jalan?',
    'prob.intro': 'Ceritakan ke pembuat aplikasi apa yang salah. Info di bawah ikut terkirim; isinya tidak memuat alamat IP, nama Wi-Fi, atau nama perangkat.',
    'prob.issue': 'Buka GitHub Issues', 'prob.copy': 'Salin info', 'prob.copied': 'Info disalin. Tempel di pesanmu.',
    'prob.copyHint': 'Tidak punya akun GitHub? Salin info ini lalu kirim lewat WhatsApp atau email.',
    'tab.diag': 'Diagnosa',
    'diag.title': 'Diagnosa otomatis',
    'diag.intro': 'Mencari tahu apakah koneksi lambat karena Wi-Fi, router, atau ISP, lewat maksimal tiga tes singkat. Ikuti saja langkahnya; tidak ada pengaturan router yang diubah. Tes di sini tidak disimpan ke Riwayat.',
    'diag.s.now': 'Tes di posisi sekarang', 'diag.s.near': 'Tes dekat router', 'diag.s.cable': 'Tes pakai kabel LAN (opsional)', 'diag.s.result': 'Kesimpulan',
    'diag.start': 'Mulai diagnosa', 'diag.skip': 'Lewati langkah ini', 'diag.again': 'Ulangi diagnosa', 'diag.skipped': 'dilewati',
    'diag.skippedStrong': 'tidak perlu (sinyal {0}%)',
    'diag.cableMsgStrong': 'Sinyal Wi-Fi di posisimu sudah kuat ({0}%), jadi tes dekat router tidak akan mengubah apa pun. Opsional: colok kabel LAN dari laptop ini ke router, lalu tunggu sampai Windows menampilkan koneksi kabel. Tidak punya kabel? Lewati saja langkah ini.',
    'diag.ready': 'Langkah 1 menjalankan tes lengkap di tempat kamu biasa memakai internet (kurang dari 1 menit).',
    'diag.nearMsg': 'Sekarang bawa laptop ini ke dekat router: sekitar 2 meter, di ruangan yang sama, tanpa tembok di antaranya. Lalu mulai tes berikutnya.',
    'diag.goNear': 'Tes dekat router',
    'diag.cableMsg': 'Opsional: colok kabel LAN dari laptop ini ke router, lalu tunggu sampai Windows menampilkan koneksi kabel. Tidak punya kabel? Lewati saja langkah ini.',
    'diag.goCable': 'Tes pakai kabel', 'diag.skipToResult': 'Lewati, lihat kesimpulan',
    'diag.cancelled': 'Tes dibatalkan. Kamu bisa mengulang langkah ini.',
    'diag.act.channel': 'Cek channel Wi-Fi', 'diag.act.dns': 'Tes DNS', 'diag.act.ispReport': 'Buka Laporan ISP', 'diag.act.monitor': 'Buka monitor stabilitas',
    'sig.title': 'Pencari sinyal Wi-Fi', 'sig.short': 'Cari posisi terbaik, langsung',
    'sig.intro': 'Bawa laptop ini berkeliling: sinyal diperbarui tiap detik, jadi kamu bisa menemukan posisi terbaik untuk laptop atau untuk router. Tidak memakai kuota internet dan berhenti saat kamu keluar dari alat ini.',
    'sig.start': 'Mulai', 'sig.beep': 'Bunyi (makin cepat = makin kuat)', 'sig.chart': 'Sinyal (%), 60 detik terakhir', 'sig.now': 'sekarang',
    'sig.note': 'Windows memperbarui nilai sinyal tiap 1–5 detik, tergantung kartu Wi-Fi.',
    'sig.spotName': 'Nama titik, misalnya Kamar', 'sig.save': 'Simpan titik ini', 'sig.spots': 'Titik tersimpan, terbaik di atas', 'sig.spot': 'Titik', 'sig.clear': 'Hapus semua titik',
    'sig.good': 'Kuat', 'sig.fair': 'Cukup', 'sig.weak': 'Lemah',
    'sig.error': 'Sinyal Wi-Fi tidak bisa dibaca: {0}', 'sig.saved': '"{0}" disimpan.',
    'reg.title': 'Ping ke region', 'reg.short': 'Jakarta, Singapura, Tokyo…',
    'reg.intro': 'Latensi ke region cloud tempat server game dan streaming berjalan. Sebagian besar server game Asia Tenggara ada di Singapura. Ini latensi ke region-nya, bukan ke server game tertentu.',
    'reg.start': 'Ukur', 'reg.region': 'Region', 'reg.running': 'Mengukur… {0}/{1}', 'reg.noAnswer': 'tidak menjawab',
  },
};

let lang = 'en';
function t(key, ...args) {
  const s = (TEXT[lang] && TEXT[lang][key]) ?? TEXT.en[key] ?? key;
  return s.replace(/\{(\d)\}/g, (_, i) => String(args[+i] ?? ''));
}
const connName = c => (c ? t('conn.' + c) : '');
const locale = () => (lang === 'id' ? 'id-ID' : undefined);

function applyText() {
  document.documentElement.lang = lang;
  $$('[data-i18n]').forEach(el => { el.textContent = t(el.dataset.i18n); });
  $$('[data-i18n-title]').forEach(el => { el.title = t(el.dataset.i18nTitle); });
  $$('[data-i18n-aria]').forEach(el => { el.setAttribute('aria-label', t(el.dataset.i18nAria)); });
}

// ---- helpers ------------------------------------------------------------
const $ = s => document.querySelector(s);
const $$ = s => Array.from(document.querySelectorAll(s));
// Preferences live in the app's data folder (the embedded page has no
// persistent browser storage).
const prefs = {};
const store = {
  set(k, v) { prefs[k] = v; bridge.call('setPrefs', prefs).catch(() => {}); },
};
const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const ms = v => (v < 10 ? v.toFixed(1) : Math.round(v).toString());
const pct = v => (Number.isInteger(v) ? v.toString() : v.toFixed(1));
const num = v => (Number.isInteger(v) ? String(v) : String(+v.toFixed(2)));

const state = {
  unit: 'Mbps',
  mode: 'general',
  running: false,
  runMode: 'general',
  result: null,
  network: null,
  info: null,
  server: null,
  update: null,
  history: [],
  histConn: 'all',
  histRange: 0,
  saver: false,
  repeat: false,
  run: 0,
  runs: 0,
  dataLive: 0,
};

// 80 Mbps = 10 MB/s (megabits vs megabytes).
function speed(mbps) {
  if (state.unit === 'MB/s') return (mbps / 8).toFixed(2);
  return mbps >= 100 ? Math.round(mbps).toString() : mbps.toFixed(1);
}

function toast(msg, dur = 4000) {
  const el = $('#toast');
  el.textContent = msg;
  el.classList.remove('hidden');
  clearTimeout(toast.timer);
  toast.timer = setTimeout(() => el.classList.add('hidden'), dur);
}
$('#toast').onclick = () => $('#toast').classList.add('hidden');

// ---- gauge & metrics ----------------------------------------------------
// Log scale so 1 Mbps and 1 Gbps both read well.
function setGauge(label, value, unit, mbpsForArc) {
  $('#gaugeLabel').textContent = label;
  $('#gaugeValue').textContent = value;
  $('#gaugeUnit').textContent = unit || '';
  let f = 0;
  if (mbpsForArc > 0) f = Math.min(1, Math.log10(1 + mbpsForArc) / Math.log10(1001));
  $('#gaugeFill').style.strokeDashoffset = (251.3 * (1 - f)).toFixed(1);
}

function setMetric(id, value, sub) {
  const el = $('#m-' + id);
  el.querySelector('b').textContent = value;
  if (sub !== undefined) el.querySelector('.sub').textContent = sub;
}

function unitLabels() {
  for (const id of ['download', 'upload']) $('#m-' + id + ' small').textContent = state.unit;
}

function activeMetric(id) {
  $$('.metric').forEach(m => m.classList.toggle('active', m.id === 'm-' + id));
}

// ---- network card -------------------------------------------------------
function wifiNote(w) {
  if (!w) return '';
  if (w.noteKey) return t(w.noteKey, w.note || '');
  return w.note || '';
}

function renderNetwork(n) {
  state.network = n;
  if (!n) return;
  $('#netConn').textContent = connName(n.connection) || t('conn.Unknown');
  const w = n.wifi;
  $('#netName').textContent = (w && w.ssid) || n.adapterName || '—';
  $('#netIP').textContent = (n.ipv4 && n.ipv4[0]) || (n.ipv6 && n.ipv6[0]) || '—';
  $('#netGW').textContent = n.gateway || '—';
  const wifiRow = $('#wifiRow');
  const note = $('#netNote');
  if (w && (w.signalPercent || w.band || w.rxRateMbps)) {
    wifiRow.classList.remove('hidden');
    $('#wifiSignal').textContent = w.signalPercent ? w.signalPercent + '%' + (w.rssi ? ` (${w.rssi} dBm)` : '') : '—';
    $('#wifiBand').textContent = w.band || '—';
    $('#wifiChannel').textContent = w.channel || '—';
    const link = Math.max(w.rxRateMbps || 0, w.txRateMbps || 0);
    $('#wifiLink').textContent = link ? Math.round(link) + ' Mbps' : '—';
  } else {
    wifiRow.classList.add('hidden');
  }
  const nt = wifiNote(w);
  note.textContent = nt;
  note.classList.toggle('hidden', !nt);
  renderTestOpts(); // the hotspot hint and data estimate depend on the connection
}

// ---- progress -----------------------------------------------------------
function resetRun(mode) {
  state.result = null;
  $('#checks').innerHTML = '';
  $$('.stage').forEach(s => {
    s.classList.remove('on', 'err');
    s.querySelector('i').style.width = '0';
    s.querySelector('em').textContent = '0%';
    s.classList.toggle('hidden', mode === 'quick' && s.dataset.stage !== 'ping');
  });
  for (const id of ['download', 'upload', 'ping', 'jitter', 'loss']) setMetric(id, '—', '');
  if (mode === 'quick') { setMetric('download', '—', t('notInQuick')); setMetric('upload', '—', t('notInQuick')); }
  $('#qualityLevel').textContent = '—';
  $('#qualityLevel').className = '';
  $('#qualityProfile').textContent = '';
  ['#failureCard', '#summaryCard', '#detailsCard', '#activityCard'].forEach(s => $(s).classList.add('hidden'));
  $('#progressCard').classList.remove('hidden');
}

const checkIcon = { ok: '✓', warn: '!', fail: '✗', pending: '…' };
function renderCheck(c) {
  const key = c.key || c.name;
  let row = $(`#checks [data-key="${CSS.escape(key)}"]`);
  if (!row) {
    row = document.createElement('div');
    row.dataset.key = key;
    $('#checks').appendChild(row);
  }
  row.className = c.status;
  row.innerHTML = `<span class="ic">${checkIcon[c.status] || ''}</span><span>${esc(c.name)}</span><span class="d">${esc(c.detail)}</span>`;
}

function stageProgress(stage, f) {
  const el = $(`.stage[data-stage="${stage}"]`);
  if (!el) return;
  const p = Math.round(Math.max(0, Math.min(1, f)) * 100);
  el.querySelector('i').style.width = p + '%';
  el.querySelector('em').textContent = p + '%';
}

function setRunning(on) {
  state.running = on;
  $('#startBtn').classList.toggle('hidden', on);
  $('#cancelBtn').classList.toggle('hidden', !on);
  $('#subActions').classList.toggle('hidden', on);
  $('#cancelBtn').disabled = false;
  $$('#modeSeg button').forEach(b => { b.disabled = on; });
  lockTabs(on);
  if (!on) activeMetric('');
  renderTestOpts();
}

const stageLabel = s => (state.runs > 1 && state.running ? t('run.label', state.run, state.runs) + ' · ' : '')
  + ({ connection: t('gauge.checking'), ping: t('stage.ping'), download: 'Download', upload: 'Upload' }[s] || '');

function onTestEvent(ev) {
  if (diag.running) { onDiagTestEvent(ev); return; }
  switch (ev.type) {
    case 'stage':
      $$('.stage').forEach(s => s.classList.toggle('on', s.dataset.stage === ev.stage));
      activeMetric(ev.stage === 'ping' ? 'ping' : ev.stage);
      setGauge(stageLabel(ev.stage), '—', '', 0);
      break;
    case 'run':
      state.run = ev.run;
      state.runs = ev.runs;
      if (ev.run > 1) resetStages();
      break;
    case 'check':
      renderCheck(ev.check);
      break;
    case 'network':
      renderNetwork(ev.network);
      break;
    case 'progress':
      stageProgress(ev.stage, ev.progress);
      if (ev.dataBytes) { state.dataLive = ev.dataBytes; renderDataLine(); }
      if (ev.stage === 'ping' && ev.rttMs) {
        setGauge('Ping', ms(ev.rttMs), 'ms', 0);
        setMetric('ping', ms(ev.rttMs), t('live'));
      } else if (ev.stage === 'download' || ev.stage === 'upload') {
        setGauge(stageLabel(ev.stage), speed(ev.mbps || 0), state.unit, ev.mbps || 0);
        setMetric(ev.stage, speed(ev.mbps || 0), t('measuring'));
      }
      break;
    case 'speed': {
      const s = ev.speed;
      stageProgress(ev.stage, 1);
      if (s.error && !(s.mbps > 0)) {
        $(`.stage[data-stage="${ev.stage}"]`).classList.add('err');
        setMetric(ev.stage, '—', t('failed'));
      } else {
        setMetric(ev.stage, speed(s.mbps), s.loadedSamples ? t('underLoad', ms(s.loadedPingMs)) : '');
      }
      break;
    }
    case 'suitability':
      renderActivities(ev.activities, state.runMode);
      break;
    case 'result':
      setRunning(false);
      state.showUsed = true;
      renderResult(ev.result);
      refreshHistoryList();
      bridge.call('attention').catch(() => {});
      break;
  }
}

// ---- result ---------------------------------------------------------------
function speedSub(s) {
  if (!s) return '';
  if (s.error) return s.mbps > 0 ? t('incomplete') : t('failed');
  return s.loadedSamples ? t('underLoad', ms(s.loadedPingMs)) : '';
}

// Same masking as the Go exporter (internal/export): Home_5G -> Ho*****,
// 192.168.1.23 -> 192.168.1.x.
const maskText = s => { const r = [...s]; return r.length <= 2 ? '*'.repeat(r.length) : r.slice(0, 2).join('') + '*'.repeat(r.length - 2); };
const maskIPv4 = s => s.split('.').slice(0, 3).join('.') + '.x';
const maskIPv6 = s => s.split(':').slice(0, 2).join(':') + ':x:x:x:x:x:x';

// Explains under the export buttons exactly what the checkbox changes, with
// this result's own values, so the effect is visible before saving.
function renderPrivacyHint() {
  const shown = $('#includeSensitive').checked;
  const n = (state.result && state.result.network) || state.network;
  const w = (n && n.wifi) || {};
  const ip = n && ((n.ipv4 || [])[0] ? maskIPv4(n.ipv4[0]) : (n.ipv6 || [])[0] ? maskIPv6(n.ipv6[0]) : '');
  // Only what this result really has; with no network info, everything.
  const parts = [];
  if (!n || w.ssid) parts.push(t('privacy.ssid') + (shown || !w.ssid ? '' : ` (${maskText(w.ssid)})`));
  if (!n || ip) parts.push(t('privacy.ip') + (shown || !ip ? '' : ` (${ip})`));
  if (!n || w.bssid) parts.push(t('privacy.bssid'));
  const list = parts.length > 1 ? parts.slice(0, -1).join(', ') + ' ' + t('and') + ' ' + parts[parts.length - 1] : parts[0] || '';
  $('#sensitiveHint').textContent = list ? t(shown ? 'privacy.shown' : 'privacy.hidden', list) : '';
}
$('#includeSensitive').onchange = renderPrivacyHint;

// "62% of plan · ping under load 31 ms" when the user entered their plan.
function withPlan(sub, s, plan) {
  if (!(plan > 0) || !s || !(s.mbps > 0)) return sub;
  const p = t('plan.of', Math.round(s.mbps / plan * 100));
  return sub ? p + ' · ' + sub : p;
}

function renderResult(r, fromHistory) {
  state.result = r;
  if (r.network) renderNetwork(r.network);
  if (!fromHistory) (r.checks || []).forEach(renderCheck);
  const quick = r.mode === 'quick';

  const dl = r.download, ul = r.upload;
  const sp = s => (s && s.mbps > 0 ? speed(s.mbps) : '—');
  const pl = r.plan || {};
  setMetric('download', sp(dl), quick ? t('notInQuick') : withPlan(speedSub(dl), dl, pl.downMbps));
  setMetric('upload', sp(ul), quick ? t('notInQuick') : withPlan(speedSub(ul), ul, pl.upMbps));
  if (r.pingMs > 0) {
    const primary = (r.pings || []).find(p => p.primaryTarget);
    setMetric('ping', ms(r.pingMs), primary ? t('minmax', ms(primary.minMs), ms(primary.maxMs)) : '');
    setMetric('jitter', ms(r.jitterMs), t('stability'));
    const counted = (r.pings || []).filter(p => !p.isGateway && p.received > 0);
    const sent = counted.reduce((a, p) => a + p.sent, 0);
    const recv = counted.reduce((a, p) => a + p.received, 0);
    setMetric('loss', pct(r.packetLossPct), sent ? t('received', recv, sent) : '');
  } else {
    setMetric('ping', '—', r.cancelled || r.failure ? '' : t('notAvailable'));
    setMetric('jitter', '—', '');
    setMetric('loss', '—', '');
  }

  if (r.cancelled) {
    setGauge(t('gauge.cancelled'), '—', '', 0);
    if (!fromHistory) toast(t('toast.cancelled'));
  } else if (r.failure && !r.quality) {
    setGauge(t('gauge.failed'), '—', '', 0);
  } else if (quick) {
    setGauge('Ping', r.pingMs > 0 ? ms(r.pingMs) : '—', r.pingMs > 0 ? 'ms' : '', 0);
  } else {
    setGauge('Download', sp(dl), dl && dl.mbps > 0 ? state.unit : '', dl ? dl.mbps : 0);
  }

  const q = r.quality;
  const ql = $('#qualityLevel');
  ql.textContent = q ? t('level.' + q.level) : '—';
  ql.className = q ? 'q-' + q.level : '';
  $('#qualityTitle').textContent = t(r.mode === 'gaming' ? 'quality.gaming' : quick ? 'quality.quick' : 'quality.title');
  $('#qualityProfile').textContent = q ? t('profile.' + (q.profile || 'general')) : '';

  const f = r.failure;
  $('#failureCard').classList.toggle('hidden', !f);
  if (f) {
    $('#failTitle').textContent = f.title;
    $('#failReason').textContent = f.reason;
    $('#failSuggestion').textContent = f.suggestion;
    $('#failTech').textContent = f.technical || '—';
  }

  $('#summaryCard').classList.toggle('hidden', !q);
  if (q) {
    $('#summaryText').textContent = q.summary;
    $('#notes').innerHTML = (q.notes || []).map(n => `<li>${esc(n)}</li>`).join('');
    const tips = q.tips || [];
    $('#tipsBox').classList.toggle('hidden', tips.length === 0);
    $('#tips').innerHTML = tips.map(n => `<li>${esc(n)}</li>`).join('');
    renderWhy(q);
    renderGaming(r);
    renderGrade(r);
    renderRuns(r);
    renderPrivacyHint();
  }
  renderActivities(r.activities, r.mode);
  renderDetails(r);
}

function limitsText(g) {
  const u = g.unit === '%' ? '%' : ' ' + g.unit;
  const [a, b, c] = g.limits.map(num);
  return t(g.higherBetter ? 'why.higher' : 'why.lower', a, b, c, u.trim());
}

function renderWhy(q) {
  $('#whyRows').innerHTML = (q.grades || []).map(g => `
    <tr class="${g.counted ? '' : 'muted'}">
      <td>${esc(g.metric)}</td>
      <td><b>${esc(g.value)}</b></td>
      <td><span class="pill q-${esc(g.grade)}">${esc(t('level.' + g.grade))}</span></td>
      <td class="small">${esc(limitsText(g))}${g.counted ? '' : ' — ' + esc(t('why.notCounted'))}</td>
    </tr>`).join('');
}

function renderGaming(r) {
  const panel = $('#gamingPanel');
  const q = r.quality;
  if (r.mode !== 'gaming' || !q) { panel.classList.add('hidden'); return; }
  panel.classList.remove('hidden');
  const good = g => g === 'EXCELLENT' || g === 'GOOD';
  const items = q.grades.filter(g => ['ping', 'jitter', 'packetLoss', 'loadedLatency'].includes(g.key));
  $('#gameList').innerHTML = items.map(g => {
    const stab = g.key === 'loadedLatency';
    const label = stab ? t('gaming.stability') : g.metric;
    const val = stab ? `${r.bloatGrade || t('level.' + g.grade)} (${g.value})` : g.value;
    return `<li><span class="k">${esc(label)}</span><span class="v">${esc(val)} <span class="${good(g.grade) ? 'ok' : 'bad'}">${good(g.grade) ? '✓' : '✗'}</span></span></li>`;
  }).join('');
}

function renderDetails(r) {
  $('#pingTable tbody').innerHTML = (r.pings || []).map(p => {
    const name = `${esc(p.label)} <span class="muted">${esc(p.target)}</span>${p.method === 'tcp' ? ' <span class="muted">(TCP)</span>' : ''}`;
    if (!p.sent) return `<tr><td>${name}</td><td colspan="7" class="muted">${esc(p.error || t('notMeasured'))}</td></tr>`;
    if (!p.received) return `<tr><td>${name}</td><td>${p.sent}</td><td>0</td><td>${pct(p.packetLossPct)}%</td><td colspan="4" class="muted">${esc(t('noReplies'))}</td></tr>`;
    const v = x => ms(x) + ' ms';
    return `<tr><td>${name}${p.primaryTarget ? ' ★' : ''}</td><td>${p.sent}</td><td>${p.received}</td><td>${pct(p.packetLossPct)}%</td><td>${v(p.minMs)}</td><td>${v(p.avgMs)}</td><td>${v(p.maxMs)}</td><td>${v(p.jitterMs)}</td></tr>`;
  }).join('') || `<tr><td colspan="8" class="muted">${esc(t('pingNotMeasured'))}</td></tr>`;

  const dl = r.download, ul = r.upload;
  const hasLoaded = (dl && dl.loadedSamples) || (ul && ul.loadedSamples);
  $('#loadedBox').classList.toggle('hidden', !hasLoaded || !(r.pingMs > 0));
  if (hasLoaded) {
    $('#ldIdle').textContent = ms(r.pingMs) + ' ms';
    $('#ldDown').textContent = dl && dl.loadedSamples ? ms(dl.loadedPingMs) + ' ms' : '—';
    $('#ldUp').textContent = ul && ul.loadedSamples ? ms(ul.loadedPingMs) + ' ms' : '—';
  }
  $('#throughputBox').classList.toggle('hidden', r.mode === 'quick');
  if (r.mode !== 'quick') {
    drawSamples($('#chartDown'), dl);
    drawSamples($('#chartUp'), ul);
  }

  const n = r.network || {};
  const w = n.wifi || {};
  const rows = [
    [t('net.connection'), connName(n.connection)], [t('row.adapter'), n.adapterName], [t('row.description'), n.description],
    ['IPv4', (n.ipv4 || []).join(', ')], ['IPv6', (n.ipv6 || []).join(', ')], ['Gateway', n.gateway], ['DNS', (n.dns || []).join(', ')],
    ['SSID', w.ssid], ['BSSID', w.bssid], [t('net.signal'), w.signalPercent ? w.signalPercent + '%' + (w.rssi ? ` (${w.rssi} dBm)` : '') : ''],
    ['Band', w.band], ['Channel', w.channel || ''], [t('row.frequency'), w.frequencyMHz ? w.frequencyMHz + ' MHz' : ''], [t('row.standard'), w.phyType],
    [t('row.link'), w.rxRateMbps ? t('row.linkVal', Math.round(w.rxRateMbps), Math.round(w.txRateMbps)) : ''],
  ].filter(([, v]) => v);
  $('#netDetails').innerHTML = rows.map(([k, v]) => `<dt>${esc(k)}</dt><dd>${esc(v)}</dd>`).join('');
  const server = (dl && dl.server) || (ul && ul.server);
  const streams = (dl && dl.streams) || (ul && ul.streams);
  $('#serverLine').textContent = server ? t('serverLine', server, streams, new Date(r.startedAt).toLocaleString(locale())) : '';
}

function cssVar(name) { return getComputedStyle(document.documentElement).getPropertyValue(name).trim(); }

function setupCanvas(canvas) {
  const ctx = canvas.getContext('2d');
  const dpr = window.devicePixelRatio || 1;
  const W = canvas.clientWidth || 420, H = canvas.clientHeight || 110;
  canvas.width = W * dpr; canvas.height = H * dpr;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.clearRect(0, 0, W, H);
  ctx.font = '11px Segoe UI, system-ui, sans-serif';
  return { ctx, W, H };
}

// Throughput samples of one run (single series, so no legend).
function drawSamples(canvas, s) {
  const { ctx, W, H } = setupCanvas(canvas);
  const data = (s && s.samplesMbps) || [];
  ctx.fillStyle = cssVar('--muted');
  if (data.length < 2) { ctx.fillText(t('noData'), 10, H / 2); return; }
  const max = Math.max(...data) * 1.1 || 1;
  const pad = 6;
  const x = i => pad + (i / (data.length - 1)) * (W - 2 * pad);
  const y = v => H - pad - (v / max) * (H - 2 * pad - 12);
  const accent = cssVar('--series-1');
  ctx.beginPath();
  data.forEach((v, i) => (i ? ctx.lineTo(x(i), y(v)) : ctx.moveTo(x(i), y(v))));
  ctx.strokeStyle = accent; ctx.lineWidth = 2; ctx.lineJoin = 'round'; ctx.stroke();
  ctx.lineTo(x(data.length - 1), H - pad); ctx.lineTo(x(0), H - pad); ctx.closePath();
  ctx.globalAlpha = 0.12; ctx.fillStyle = accent; ctx.fill(); ctx.globalAlpha = 1;
  ctx.fillStyle = cssVar('--muted');
  ctx.fillText(t('peak', `${speed(Math.max(...data))} ${state.unit}`), 10, 14);
  if (s.mbps > 0) {
    const ry = y(s.mbps);
    ctx.setLineDash([4, 4]); ctx.strokeStyle = cssVar('--muted');
    ctx.lineWidth = 1; ctx.beginPath(); ctx.moveTo(pad, ry); ctx.lineTo(W - pad, ry); ctx.stroke(); ctx.setLineDash([]);
  }
}

// ---- PNG result card ------------------------------------------------------
function makeCard(r) {
  const c = document.createElement('canvas');
  const W = 600, H = 340, S = 2;
  c.width = W * S; c.height = H * S;
  const g = c.getContext('2d');
  g.scale(S, S);
  const grad = g.createLinearGradient(0, 0, W, H);
  grad.addColorStop(0, '#0f2a4a'); grad.addColorStop(1, '#16213a');
  g.fillStyle = grad;
  g.beginPath(); g.roundRect(0, 0, W, H, 20); g.fill();
  g.fillStyle = '#8fb5ff'; g.font = '600 14px Segoe UI, system-ui, sans-serif';
  g.fillText(t('card.title'), 32, 44);
  g.fillStyle = '#9aa7b8'; g.font = '13px Segoe UI, system-ui, sans-serif';
  const conn = r.network ? connName(r.network.connection) : '';
  g.fillText(`${new Date(r.startedAt).toLocaleString(locale())}${conn ? '  ·  ' + conn : ''}`, 32, 66);
  const big = (label, val, unit, x, yy) => {
    g.fillStyle = '#9aa7b8'; g.font = '600 12px Segoe UI, system-ui, sans-serif'; g.fillText(label, x, yy);
    g.fillStyle = '#ffffff'; g.font = '700 40px Segoe UI, system-ui, sans-serif'; g.fillText(val, x, yy + 44);
    const w = g.measureText(val).width;
    g.fillStyle = '#9aa7b8'; g.font = '600 15px Segoe UI, system-ui, sans-serif'; g.fillText(unit, x + w + 6, yy + 44);
  };
  const sp = s => (s && s.mbps > 0 ? speed(s.mbps) : '—');
  if (r.mode === 'quick') {
    big('● PING', r.pingMs > 0 ? ms(r.pingMs) : '—', 'ms', 32, 110);
    big('JITTER', r.pingMs > 0 ? ms(r.jitterMs) : '—', 'ms', 230, 110);
    big('LOSS', r.pingMs > 0 ? pct(r.packetLossPct) : '—', '%', 428, 110);
  } else {
    big('↓ DOWNLOAD', sp(r.download), state.unit, 32, 110);
    big('↑ UPLOAD', sp(r.upload), state.unit, 230, 110);
    big('● PING', r.pingMs > 0 ? ms(r.pingMs) : '—', 'ms', 428, 110);
    g.fillStyle = '#c9d4e2'; g.font = '15px Segoe UI, system-ui, sans-serif';
    g.fillText(`Jitter ${r.pingMs > 0 ? ms(r.jitterMs) + ' ms' : '—'}     Loss ${r.pingMs > 0 ? pct(r.packetLossPct) + '%' : '—'}${r.bloatGrade ? '     ' + t('card.grade', r.bloatGrade) : ''}`, 32, 210);
  }
  const q = r.quality ? r.quality.level : 'UNKNOWN';
  const colors = { EXCELLENT: '#3fb950', GOOD: '#58a6ff', FAIR: '#d29922', POOR: '#f85149', UNSTABLE: '#a371f7', UNKNOWN: '#8d96a0' };
  g.fillStyle = 'rgba(255,255,255,.08)'; g.beginPath(); g.roundRect(32, 236, W - 64, 56, 12); g.fill();
  g.fillStyle = '#9aa7b8'; g.font = '600 13px Segoe UI, system-ui, sans-serif';
  g.fillText(t('card.' + (r.mode === 'gaming' || r.mode === 'quick' ? r.mode : 'general')), 52, 270);
  const qt = t('level.' + q);
  g.fillStyle = colors[q] || '#fff'; g.font = '800 26px Segoe UI, system-ui, sans-serif';
  g.fillText(qt, W - 52 - g.measureText(qt).width, 274);
  g.fillStyle = '#6f7d8f'; g.font = '11px Segoe UI, system-ui, sans-serif';
  g.fillText(t('card.footer'), 32, 318);
  return c.toDataURL('image/png');
}

// ---- actions --------------------------------------------------------------
async function start(mode) {
  state.runMode = mode;
  state.run = state.runs = state.dataLive = 0;
  resetRun(mode);
  setRunning(true);
  setGauge(t('gauge.starting'), '—', '', 0);
  renderDataLine();
  try {
    await bridge.call('startTest', mode, { repeat: state.repeat && mode !== 'quick', dataSaver: state.saver });
  } catch (e) {
    setRunning(false);
    toast(e.message);
  }
}

$('#startBtn').onclick = () => start(state.mode);
$('#quickBtn').onclick = () => start('quick');
$('#cancelBtn').onclick = () => {
  $('#cancelBtn').disabled = true;
  setGauge(t('gauge.cancelling'), '—', '', 0);
  bridge.call('cancelTest');
};

$$('#modeSeg button').forEach(b => b.onclick = () => {
  state.mode = b.dataset.mode;
  store.set('mode', state.mode);
  $$('#modeSeg button').forEach(x => x.classList.toggle('on', x === b));
});
$$('#unitSeg button').forEach(b => b.onclick = () => {
  state.unit = b.dataset.unit;
  store.set('unit', state.unit);
  $$('#unitSeg button').forEach(x => x.classList.toggle('on', x === b));
  unitLabels();
  if (state.result) renderResult(state.result, true);
  if (!$('#historyView').classList.contains('hidden')) renderHistory();
});

$$('[data-export]').forEach(b => b.onclick = () => {
  if (!state.result) return;
  const fmt = b.dataset.export;
  if (fmt === 'png') {
    bridge.call('savePng', makeCard(state.result)).catch(e => toast(e.message));
    return;
  }
  bridge.call('export', { format: fmt, unit: state.unit, includeSensitive: $('#includeSensitive').checked, id: state.result.id })
    .catch(e => toast(e.message));
});

$('#copyBtn').onclick = async () => {
  if (!state.result) return;
  try {
    const res = await bridge.call('copyResult', { id: state.result.id, unit: state.unit });
    if (!res.copied) await navigator.clipboard.writeText(res.text);
    toast(t('toast.copied'));
  } catch (e) {
    toast(t('toast.copyFailed', e.message || e));
  }
};

$('#detailsBtn').onclick = () => {
  const d = $('#detailsCard');
  d.classList.toggle('hidden');
  if (!d.classList.contains('hidden')) {
    if (state.result) renderDetails(state.result);
    d.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
};

// ---- activities & bufferbloat grade ------------------------------------------
// The engine sends the activity list after each stage, so it fills in while
// the test runs; the final list comes with the result.
const actIcon = { browse: '🌐', hd: '▶️', uhd: '🎬', call: '📹', game: '🎮', live: '📡', bigdl: '⬇️' };
function renderActivities(list, mode) {
  const card = $('#activityCard');
  if (!list || !list.length) { card.classList.add('hidden'); return; }
  card.classList.remove('hidden');
  $('#actQuickNote').classList.toggle('hidden', mode !== 'quick');
  $('#actList').innerHTML = list.map(a => `
    <li class="act act-${esc(a.status)}" data-key="${esc(a.key)}">
      <span class="ai" aria-hidden="true">${actIcon[a.key] || '•'}</span>
      <span class="an">${esc(a.name)}</span>
      <span class="as">${esc(t('act.s.' + a.status))}</span>
      <span class="ar">${esc(a.reason || '')}</span>
    </li>`).join('');
}

const gradeClass = g => 'g-' + (g === 'A+' ? 'Ap' : g);
function gradeBadge(g, small) {
  return ` <span class="grade${small ? ' sm' : ''} ${gradeClass(g)}" title="${esc(t('grade.title'))}">${esc(g)}</span>`;
}
function renderGrade(r) {
  const g = r.bloatGrade;
  $('#gradeRow').classList.toggle('hidden', !g);
  if (!g) return;
  const badge = $('#gradeBadge');
  badge.textContent = g;
  badge.className = 'grade ' + gradeClass(g);
  const loaded = [r.download, r.upload].filter(s => s && s.loadedSamples).map(s => s.loadedPingMs);
  $('#gradeText').textContent = t('grade.text', ms(Math.max(0, Math.max(...loaded) - r.pingMs)));
}

// ---- data saver, repeat & data usage ---------------------------------------
// Quota is counted the way phone plans do (decimal MB/GB).
function fmtBytes(n) {
  if (n >= 1e9) return (n / 1e9).toFixed(2) + ' GB';
  if (n >= 10e6) return Math.round(n / 1e6) + ' MB';
  return (n / 1e6).toFixed(1) + ' MB';
}

// Under the Start button: live usage while testing, the usage of the test
// just finished, or an estimate for the next one.
function renderDataLine() {
  const el = $('#dataLine');
  const r = state.result;
  if (state.running) el.textContent = state.dataLive ? t('data.live', fmtBytes(state.dataLive)) : '';
  else if (state.showUsed && r && r.mode !== 'quick' && r.dataBytes) el.textContent = t('data.used', fmtBytes(r.dataBytes));
  else el.textContent = dataEstimate();
}

// The estimate comes from the last comparable test on this kind of
// connection; data saver has a known cap.
function dataEstimate() {
  const runs = state.repeat ? 3 : 1;
  if (state.saver) return t('data.saverOn', fmtBytes(((state.info && state.info.dataSaverMB) || 45) * 1e6 * runs));
  const conn = state.network && state.network.connection;
  const h = state.history.find(x => x.connection === conn && x.mode !== 'quick' && !x.dataSaver && !(x.runs > 1) && x.dataMB > 0);
  return h ? t('data.estimate', fmtBytes(h.dataMB * 1e6 * runs)) : '';
}

const viaPhone = () => ['Hotspot / Tethering', 'Cellular'].includes(state.network && state.network.connection);
// A metered link (as Windows reports it) has a data limit too.
const onMobileData = () => viaPhone() || !!(state.network && state.network.metered);
function renderTestOpts() {
  $('#saverBtn').setAttribute('aria-pressed', state.saver);
  $('#repeatBtn').setAttribute('aria-pressed', state.repeat);
  $('#saverBtn').title = t('saver.hint', Math.round((state.info && state.info.dataSaverMB) || 45));
  $('#repeatBtn').title = t('repeat.hint');
  $('#testOpts').classList.toggle('hidden', state.running);
  $('#hotspotHint').classList.toggle('hidden', state.running || state.saver || !onMobileData());
  $('#hotspotMsg').textContent = t(viaPhone() ? 'data.hotspot' : 'data.metered');
  renderDataLine();
}
function setSaver(on) {
  state.saver = on;
  state.showUsed = false;
  store.set('dataSaver', on ? 'on' : 'off');
  renderTestOpts();
}
$('#saverBtn').onclick = () => setSaver(!state.saver);
$('#hotspotOn').onclick = () => setSaver(true);
$('#repeatBtn').onclick = () => {
  state.repeat = !state.repeat;
  state.showUsed = false;
  store.set('repeat', state.repeat ? 'on' : 'off');
  renderTestOpts();
};

// Between rounds of a repeated test the stage bars start over.
function resetStages() {
  $$('.stage').forEach(s => {
    s.classList.remove('on', 'err');
    s.querySelector('i').style.width = '0';
    s.querySelector('em').textContent = '0%';
  });
  $('#checks').innerHTML = '';
}

function renderRuns(r) {
  const runs = r.runs || [];
  $('#runsBox').classList.toggle('hidden', runs.length < 2);
  if (runs.length < 2) return;
  $('#runsTitle').textContent = t('runs.title', runs.length);
  const sp = v => (v > 0 ? speed(v) + ' ' + state.unit : '—');
  $('#runsBox tbody').innerHTML = runs.map((x, i) => `<tr><td>${i + 1}</td><td>${sp(x.downMbps)}</td><td>${sp(x.upMbps)}</td>
    <td>${x.pingMs > 0 ? ms(x.pingMs) + ' ms' : '—'}</td><td>${x.pingMs > 0 ? ms(x.jitterMs) + ' ms' : '—'}</td>
    <td>${x.pingMs > 0 ? pct(+x.lossPct.toFixed(1)) + '%' : '—'}</td></tr>`).join('');
}

async function refreshHistoryList() {
  try { state.history = await bridge.call('history') || []; } catch { /* keep */ }
  renderDataLine();
}

// ---- views & theme -------------------------------------------------------------
const viewSel = { test: '#dashboard', diag: '#diagView', history: '#historyView', tools: '#toolsView' };
function showView(v) {
  if (v !== 'tools') sigStop();
  for (const [k, sel] of Object.entries(viewSel)) $(sel).classList.toggle('hidden', k !== v);
  $$('.tabs [role="tab"]').forEach(b => {
    const on = b.dataset.view === v;
    b.setAttribute('aria-selected', on);
    b.tabIndex = on ? 0 : -1;
  });
  if (v === 'history') loadHistory();
  if (v === 'tools') openTools();
  if (v === 'diag') renderDiag();
  window.scrollTo(0, 0);
}
// While a test or a diagnosis runs, only its own tab stays usable.
function lockTabs(on) {
  $$('.tabs [role="tab"]').forEach(b => { b.disabled = on && b.getAttribute('aria-selected') !== 'true'; });
  $('#settingsBtn').disabled = on;
}
$$('.tabs [role="tab"]').forEach(b => { b.onclick = () => showView(b.dataset.view); });
// Arrow keys move between tabs, as in any tab strip.
$('.tabs').addEventListener('keydown', e => {
  if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return;
  const tabs = $$('.tabs [role="tab"]').filter(b => !b.disabled);
  const i = tabs.indexOf(document.activeElement);
  if (i < 0) return;
  const next = tabs[(i + (e.key === 'ArrowRight' ? 1 : tabs.length - 1)) % tabs.length];
  next.focus();
  next.click();
  e.preventDefault();
});

const darkMQ = window.matchMedia('(prefers-color-scheme: dark)');
const themePref = () => (prefs.theme === 'light' || prefs.theme === 'dark' ? prefs.theme : 'system');
// "system" follows Windows; the window frame follows the page.
function applyTheme() {
  const p = themePref();
  if (p === 'system') delete document.documentElement.dataset.theme;
  else document.documentElement.dataset.theme = p;
  bridge.call('setTitleBar', p === 'dark' || (p === 'system' && darkMQ.matches)).catch(() => {});
  redrawCharts();
}
darkMQ.addEventListener('change', () => { if (themePref() === 'system') applyTheme(); });
$('#themeSelect').onchange = () => { store.set('theme', $('#themeSelect').value); applyTheme(); };

// Canvas charts read the theme colors when drawn, so redraw visible ones.
function redrawCharts() {
  if (state.result && !$('#detailsCard').classList.contains('hidden')) renderDetails(state.result);
  if (!$('#historyView').classList.contains('hidden')) renderTrend(filteredHistory());
  if (!$('#toolsView').classList.contains('hidden') && mon.samples.length) drawMonitor();
  if (!$('#signalCard').classList.contains('hidden') && sig.samples.length) drawSignal();
}

// ---- history --------------------------------------------------------------
async function loadHistory() {
  try { state.history = await bridge.call('history') || []; } catch (e) { toast(e.message); state.history = []; }
  renderHistory();
}

function filteredHistory() {
  const since = state.histRange ? Date.now() - state.histRange * 86400e3 : 0;
  return state.history.filter(h => (state.histConn === 'all' || h.connection === state.histConn) && new Date(h.date).getTime() >= since);
}

function renderHistory() {
  const all = state.history;
  // Connection filter: one chip per connection type seen.
  const conns = [...new Set(all.map(h => h.connection).filter(Boolean))];
  if (state.histConn !== 'all' && !conns.includes(state.histConn)) state.histConn = 'all';
  const connSeg = $('#histConn');
  connSeg.innerHTML = '';
  for (const c of ['all', ...conns]) {
    const b = document.createElement('button');
    b.textContent = c === 'all' ? t('conn.all') : connName(c);
    b.classList.toggle('on', c === state.histConn);
    b.onclick = () => { state.histConn = c; renderHistory(); };
    connSeg.appendChild(b);
  }
  $$('#histRange button').forEach(b => b.classList.toggle('on', +b.dataset.range === state.histRange));

  const list = filteredHistory();
  const empty = all.length === 0;
  const month = all.filter(h => { const d = new Date(h.date), n = new Date(); return d.getMonth() === n.getMonth() && d.getFullYear() === n.getFullYear() && h.dataMB > 0; });
  $('#histData').textContent = month.length ? t('data.month', fmtBytes(month.reduce((a, h) => a + h.dataMB, 0) * 1e6), month.length) : '';
  $('#histEmpty').classList.toggle('hidden', !empty);
  ['#histTable', '#histFilters', '#trendBox', '#compareTable'].forEach(s => $(s).classList.toggle('hidden', empty));
  // The plan column only appears once a test was run with a plan set.
  $('#histTable').classList.toggle('no-plan', !all.some(h => h.planPct > 0));
  $$('#historyView .sub-head').forEach(el => el.classList.toggle('hidden', empty));

  $('#histTable tbody').innerHTML = list.map(h => `
    <tr data-id="${esc(h.id)}" tabindex="0">
      <td>${esc(new Date(h.date).toLocaleString(locale(), { dateStyle: 'medium', timeStyle: 'short' }))}${h.mode !== 'general' ? ` <span class="muted">(${esc(h.mode === 'quick' ? t('quick') : t('mode.' + h.mode))})</span>` : ''}</td>
      <td>${esc(connName(h.connection))}</td>
      <td>${h.download ? speed(h.download) : '—'} / ${h.upload ? speed(h.upload) : '—'} ${state.unit}${h.dataMB >= 0.1 ? ` <span class="muted small">· ${esc(fmtBytes(h.dataMB * 1e6))}${h.dataSaver ? ' · ' + esc(t('hist.saver')) : ''}${h.runs > 1 ? ' · ' + h.runs + '×' : ''}</span>` : ''}</td>
      <td>${h.ping ? ms(h.ping) + ' ms' : '—'}${h.grade ? gradeBadge(h.grade, true) : ''}</td>
      <td>${h.ping ? ms(h.jitter) + ' ms' : '—'}</td>
      <td>${h.ping ? pct(h.loss) + '%' : '—'}</td>
      <td class="plan-col">${h.planPct > 0 ? Math.round(h.planPct) + '%' : '—'}</td>
      <td class="q-${esc(h.quality || 'UNKNOWN')}">${esc(h.quality ? t('level.' + h.quality) : '—')}</td>
    </tr>`).join('');
  $$('#histTable tbody tr').forEach(tr => {
    const open = async () => {
      try {
        const r = await bridge.call('historyItem', tr.dataset.id);
        if (!r) return;
        showView('test');
        resetRun(r.mode);
        $('#progressCard').classList.add('hidden');
        renderResult(r, true);
      } catch (e) { toast(e.message); }
    };
    tr.onclick = open;
    tr.onkeydown = e => { if (e.key === 'Enter') open(); };
  });

  renderCompare(list);
  renderTrend(list);
  if (state.info) $('#histPath').textContent = t('history.storedAt', state.info.historyPath);
}

// Wi-Fi vs Ethernet (and other types): averages per connection.
function renderCompare(list) {
  const groups = {};
  for (const h of list) (groups[h.connection || 'Unknown'] ||= []).push(h);
  const avg = (xs, f) => { const v = xs.map(f).filter(x => x > 0); return v.length ? v.reduce((a, b) => a + b, 0) / v.length : 0; };
  const avg0 = (xs, f) => { const v = xs.filter(h => h.ping > 0).map(f); return v.length ? v.reduce((a, b) => a + b, 0) / v.length : -1; };
  $('#compareTable tbody').innerHTML = Object.entries(groups).map(([c, xs]) => {
    const d = avg(xs, h => h.download), u = avg(xs, h => h.upload), p = avg(xs, h => h.ping), j = avg0(xs, h => h.jitter), l = avg0(xs, h => h.loss);
    return `<tr><td>${esc(connName(c))}</td><td>${xs.length}</td><td>${d ? speed(d) + ' ' + state.unit : '—'}</td><td>${u ? speed(u) + ' ' + state.unit : '—'}</td><td>${p ? ms(p) + ' ms' : '—'}</td><td>${j >= 0 ? ms(j) + ' ms' : '—'}</td><td>${l >= 0 ? pct(+l.toFixed(1)) + '%' : '—'}</td></tr>`;
  }).join('');
}

// Trend charts: speed (download + upload, one axis) and ping on its own
// chart, oldest to newest. Hover or arrow keys show every value at a test.
function renderTrend(list) {
  const pts = [...list].reverse(); // oldest first
  const speedPts = pts.filter(h => h.download > 0 || h.upload > 0);
  const pingPts = pts.filter(h => h.ping > 0);
  const conv = v => (state.unit === 'MB/s' ? v / 8 : v);
  plot($('#plotSpeed'), speedPts, [
    { name: 'Download', color: '--series-1', get: h => (h.download > 0 ? conv(h.download) : null) },
    { name: 'Upload', color: '--series-2', get: h => (h.upload > 0 ? conv(h.upload) : null) },
  ], v => (state.unit === 'MB/s' ? v.toFixed(2) : v >= 100 ? Math.round(v) : v.toFixed(1)) + ' ' + state.unit);
  plot($('#plotPing'), pingPts, [
    { name: 'Ping', color: '--series-3', get: h => h.ping },
  ], v => ms(v) + ' ms');
}

function plot(box, pts, series, fmt) {
  const canvas = box.querySelector('canvas');
  const tip = box.querySelector('.tip');
  tip.classList.add('hidden');
  // x follows the test time, so a gap of days looks like a gap.
  const times = pts.map(h => new Date(h.date).getTime());
  const span = times.length > 1 ? times[times.length - 1] - times[0] : 0;
  const frac = times.map((v, i) => (span > 0 ? (v - times[0]) / span : i / Math.max(1, times.length - 1)));
  const draw = hover => {
    const { ctx, W, H } = setupCanvas(canvas);
    const muted = cssVar('--muted');
    if (pts.length < 2) { ctx.fillStyle = muted; ctx.fillText(t('trend.noData'), 12, H / 2); return null; }
    const vals = pts.flatMap(h => series.map(s => s.get(h))).filter(v => v != null);
    const max = niceMax(Math.max(...vals));
    const L = 52, R = 56, T = 10, B = 22;
    const x = i => L + frac[i] * (W - L - R);
    const y = v => T + (1 - v / max) * (H - T - B);
    // Recessive grid with three ticks.
    ctx.strokeStyle = cssVar('--line'); ctx.lineWidth = 1; ctx.fillStyle = muted; ctx.textAlign = 'right';
    for (const f of [0, 0.5, 1]) {
      const yy = Math.round(y(max * f)) + 0.5;
      ctx.beginPath(); ctx.moveTo(L, yy); ctx.lineTo(W - R, yy); ctx.stroke();
      ctx.fillText(axisNum(max * f), L - 6, yy + 4);
    }
    ctx.textAlign = 'left';
    // Same day at both ends: show the times instead of two equal dates.
    const sameDay = new Date(times[0]).toDateString() === new Date(times[times.length - 1]).toDateString();
    const fmtX = v => (sameDay ? new Date(v).toLocaleTimeString(locale(), { hour: '2-digit', minute: '2-digit' })
      : new Date(v).toLocaleDateString(locale(), { day: 'numeric', month: 'short' }));
    const d0 = fmtX(times[0]), d1 = fmtX(times[times.length - 1]);
    ctx.fillText(d0, L, H - 6);
    ctx.textAlign = 'right'; ctx.fillText(d1, W - R, H - 6); ctx.textAlign = 'left';
    if (hover != null) {
      ctx.strokeStyle = muted; ctx.beginPath(); ctx.moveTo(Math.round(x(hover)) + 0.5, T); ctx.lineTo(Math.round(x(hover)) + 0.5, H - B); ctx.stroke();
    }
    const surface = cssVar('--card');
    const labels = [];
    for (const s of series) {
      const col = cssVar(s.color);
      ctx.strokeStyle = col; ctx.lineWidth = 2; ctx.lineJoin = 'round'; ctx.beginPath();
      let started = false;
      pts.forEach((h, i) => {
        const v = s.get(h);
        if (v == null) { started = false; return; }
        if (!started) { ctx.moveTo(x(i), y(v)); started = true; } else ctx.lineTo(x(i), y(v));
      });
      ctx.stroke();
      // Direct label at the newest point (value in text ink, not series color).
      const lastI = pts.map(s.get).map((v, i) => (v != null ? i : -1)).filter(i => i >= 0).pop();
      if (lastI != null) {
        const v = s.get(pts[lastI]);
        dot(ctx, x(lastI), y(v), col, surface);
        labels.push({ x: x(lastI) + 8, y: y(v) + 4, text: fmt(v).replace(/ .*/, '') });
      }
      if (hover != null && s.get(pts[hover]) != null) dot(ctx, x(hover), y(s.get(pts[hover])), col, surface);
    }
    // Direct labels in text ink, pushed apart so close values don't collide.
    labels.sort((a, b) => a.y - b.y);
    for (let i = 1; i < labels.length; i++) labels[i].y = Math.max(labels[i].y, labels[i - 1].y + 13);
    ctx.fillStyle = cssVar('--ink');
    for (const l of labels) ctx.fillText(l.text, l.x, l.y);
    return { x, L, R, W };
  };
  const geo = draw(null);
  const show = i => {
    const g = draw(i);
    if (!g) return;
    const h = pts[i];
    tip.innerHTML = '';
    const head = document.createElement('div');
    head.className = 'tip-date';
    head.textContent = new Date(h.date).toLocaleString(locale(), { dateStyle: 'medium', timeStyle: 'short' }) + ' · ' + connName(h.connection);
    tip.appendChild(head);
    for (const s of series) {
      const v = s.get(h);
      const row = document.createElement('div');
      row.className = 'tip-row';
      const key = document.createElement('i'); key.className = 'tip-key'; key.style.background = cssVar(s.color);
      const val = document.createElement('b'); val.textContent = v == null ? '—' : fmt(v);
      const name = document.createElement('span'); name.textContent = s.name;
      row.append(key, val, name);
      tip.appendChild(row);
    }
    tip.classList.remove('hidden');
    // Sit beside the crosshair, on the side away from the newest points.
    const px = g.x(i);
    const left = px > canvas.clientWidth / 2 ? px - tip.offsetWidth - 12 : px + 12;
    tip.style.left = Math.min(Math.max(left, 0), canvas.clientWidth - tip.offsetWidth) + 'px';
    box.dataset.index = i;
  };
  const hide = () => { tip.classList.add('hidden'); draw(null); delete box.dataset.index; };
  if (!geo) { box.onpointermove = box.onpointerleave = box.onkeydown = null; box.removeAttribute('tabindex'); return; }
  box.tabIndex = 0;
  box.onpointermove = e => {
    const r = canvas.getBoundingClientRect();
    const f = (e.clientX - r.left - geo.L) / (geo.W - geo.L - geo.R);
    let best = 0;
    frac.forEach((v, i) => { if (Math.abs(v - f) < Math.abs(frac[best] - f)) best = i; });
    show(best);
  };
  box.onpointerleave = hide;
  box.onblur = hide;
  box.onkeydown = e => {
    const cur = box.dataset.index != null ? +box.dataset.index : pts.length;
    if (e.key === 'ArrowLeft') { show(Math.max(0, cur - 1)); e.preventDefault(); }
    if (e.key === 'ArrowRight') { show(Math.min(pts.length - 1, cur + 1 > pts.length - 1 ? pts.length - 1 : cur + 1)); e.preventDefault(); }
  };
}

function dot(ctx, x, y, col, ring) {
  ctx.beginPath(); ctx.arc(x, y, 5, 0, Math.PI * 2);
  ctx.fillStyle = col; ctx.fill();
  ctx.lineWidth = 2; ctx.strokeStyle = ring; ctx.stroke();
}

// Axis ticks: at most three significant digits, no trailing zeros (50, 0.05).
function axisNum(v) { return String(Number(v.toPrecision(3))); }

function niceMax(v) {
  if (!(v > 0)) return 1;
  const p = Math.pow(10, Math.floor(Math.log10(v)));
  for (const m of [1, 2, 2.5, 5, 10]) if (v <= m * p) return m * p;
  return 10 * p;
}

$$('#histRange button').forEach(b => b.onclick = () => { state.histRange = +b.dataset.range; renderHistory(); });
$('#histClear').onclick = async () => {
  if (!confirm(t('history.confirmClear'))) return;
  try { await bridge.call('clearHistory'); loadHistory(); } catch (e) { toast(e.message); }
};
// The ISP report holds no network identifiers, so it needs no option.
const ispReport = format => bridge.call('ispReport', {
  days: state.histRange, connection: state.histConn === 'all' ? '' : state.histConn, unit: state.unit, format,
}).catch(e => toast(e.message));
$('#histReportPdf').onclick = () => ispReport('pdf');
$('#histReport').onclick = () => ispReport('txt');
$('#histCsv').onclick = () => bridge.call('export', { format: 'history-csv', unit: state.unit, includeSensitive: $('#histSensitive').checked }).catch(e => toast(e.message));
window.addEventListener('resize', () => {
  if (!$('#historyView').classList.contains('hidden')) renderTrend(filteredHistory());
  if (!$('#toolsView').classList.contains('hidden') && mon.samples.length) drawMonitor();
  if (!$('#signalCard').classList.contains('hidden') && sig.samples.length) drawSignal();
});

// ---- server picker ----------------------------------------------------------
function defaultServerName() { return (state.info && state.info.servers && state.info.servers[0]) || 'Cloudflare'; }
function renderServerName() {
  const name = state.server ? state.server.name : t('server.default', defaultServerName());
  $('#serverName').textContent = state.server ? state.server.name : defaultServerName();
  $('#serverBtn').title = t('server.hint') + ': ' + name;
  $('#useDefaultServer').textContent = t('server.useDefault', defaultServerName());
  $('#serverCurrent').textContent = name;
}
$('#serverBtn').onclick = () => {
  renderServerName();
  $('#serverDlg').showModal();
};
$('#findServers').onclick = () => {
  $('#findServers').disabled = true;
  $('#serverStatus').textContent = t('server.fetching');
  $('#serverTable').classList.add('hidden');
  bridge.call('findServers').catch(e => { $('#serverStatus').textContent = t('server.error', e.message); $('#findServers').disabled = false; });
};
$('#useDefaultServer').onclick = async () => {
  try {
    await bridge.call('setServer', null);
    state.server = null;
    renderServerName();
    toast(t('server.saved', defaultServerName()));
  } catch (e) { toast(e.message); }
};
function onServers(p) {
  if (p.error) { $('#serverStatus').textContent = t('server.error', p.error); $('#findServers').disabled = false; return; }
  if (p.stage === 'fetch') { $('#serverStatus').textContent = t('server.fetching'); return; }
  if (p.stage === 'probe') { $('#serverStatus').textContent = t('server.probing', p.done, p.total); return; }
  $('#findServers').disabled = false;
  const ok = (p.list || []).filter(c => !c.error).slice(0, 15);
  $('#serverStatus').textContent = ok.length ? t('server.found', ok.length) : t('server.none');
  const tbody = $('#serverTable tbody');
  tbody.innerHTML = '';
  for (const c of ok) {
    const tr = document.createElement('tr');
    tr.tabIndex = 0;
    for (const v of [c.server.name, c.server.sponsor || '—', ms(c.latencyMs) + ' ms']) {
      const td = document.createElement('td'); td.textContent = v; tr.appendChild(td);
    }
    const pick = async () => {
      try {
        await bridge.call('setServer', c.server);
        state.server = c.server;
        renderServerName();
        $('#serverDlg').close();
        toast(t('server.saved', c.server.name));
      } catch (e) { toast(e.message); }
    };
    tr.onclick = pick;
    tr.onkeydown = e => { if (e.key === 'Enter') pick(); };
    tbody.appendChild(tr);
  }
  $('#serverTable').classList.toggle('hidden', ok.length === 0);
}

// ---- settings & updates -----------------------------------------------------
$('#settingsBtn').onclick = async () => {
  $('#langSelect').value = lang;
  $('#themeSelect').value = themePref();
  $('#updateToggle').checked = prefs.updateCheck !== 'off';
  $('#updateStatus').textContent = '';
  $('#targetStatus').textContent = '';
  $('#planDown').value = prefs.planDown || '';
  $('#planUp').value = prefs.planUp || '';
  let ts = [];
  try { ts = await bridge.call('getTargets') || []; } catch { /* none */ }
  renderTargetRows(ts);
  $('#settingsDlg').showModal();
};
for (const id of ['planDown', 'planUp']) {
  $('#' + id).onchange = () => {
    const v = $('#' + id).value.trim();
    store.set(id, v && +v > 0 ? v : '');
  };
}
const MAX_TARGETS = 5;
function renderTargetRows(list) {
  const box = $('#targetRows');
  box.innerHTML = '';
  for (const tg of list) addTargetRow(tg);
  $('#targetAdd').disabled = list.length >= MAX_TARGETS;
}
function addTargetRow(tg = { label: '', host: '' }) {
  const box = $('#targetRows');
  if (box.children.length >= MAX_TARGETS) return;
  const row = document.createElement('div');
  row.className = 'target-row';
  const label = document.createElement('input');
  label.maxLength = 40; label.value = tg.label; label.placeholder = t('targets.label'); label.className = 't-label';
  label.setAttribute('aria-label', t('targets.label'));
  const host = document.createElement('input');
  host.value = tg.host; host.placeholder = t('targets.host'); host.className = 't-host mono'; host.spellcheck = false;
  host.setAttribute('aria-label', t('targets.host'));
  const del = document.createElement('button');
  del.className = 'ghost small'; del.textContent = '✕'; del.title = t('targets.remove'); del.setAttribute('aria-label', t('targets.remove'));
  del.onclick = () => { row.remove(); $('#targetAdd').disabled = false; };
  row.append(label, host, del);
  box.appendChild(row);
  $('#targetAdd').disabled = box.children.length >= MAX_TARGETS;
  return host;
}
$('#targetAdd').onclick = () => { const h = addTargetRow(); if (h) h.previousSibling.focus(); };
$('#targetSave').onclick = async () => {
  const list = $$('#targetRows .target-row')
    .map(r => ({ label: r.querySelector('.t-label').value.trim(), host: r.querySelector('.t-host').value.trim() }))
    .filter(x => x.host);
  try {
    await bridge.call('setTargets', list);
    renderTargetRows(await bridge.call('getTargets') || []);
    $('#targetStatus').textContent = t('targets.saved');
    loadMonitorTargets();
  } catch (e) { $('#targetStatus').textContent = e.message; }
};
$('#langSelect').onchange = async () => {
  lang = $('#langSelect').value;
  store.set('lang', lang);
  applyText();
  renderStatic();
  // Engine texts are stored with the result; reload them in the new language.
  if (state.result) {
    let r = state.result;
    if (r.id) { try { r = (await bridge.call('historyItem', r.id)) || r; } catch { /* keep */ } }
    renderResult(r, true);
  }
  if (!$('#historyView').classList.contains('hidden')) renderHistory();
  renderTools();
};
$('#updateToggle').onchange = () => store.set('updateCheck', $('#updateToggle').checked ? 'on' : 'off');
$('#checkNow').onclick = () => {
  state.manualUpdate = true;
  $('#updateStatus').textContent = t('update.checking');
  bridge.call('checkUpdate').catch(() => {});
};
function onUpdate(p) {
  const manual = state.manualUpdate;
  state.manualUpdate = false;
  if (p.error) { if (manual) $('#updateStatus').textContent = t('update.error', p.error); return; }
  const i = p.info;
  state.update = i;
  if (i.newer) {
    $('#updateText').textContent = t('update.available', i.latest, i.current);
    $('#updateBanner').classList.remove('hidden');
  }
  if (manual) $('#updateStatus').textContent = i.newer ? t('update.available', i.latest, i.current) : t('update.latest', i.current);
}
$('#updateOpen').onclick = () => {
  const url = (state.update && state.update.url) || (state.info && state.info.releasesUrl);
  if (!url) return;
  if (state.info && state.info.canOpenUrl) bridge.call('openUrl', url).catch(e => toast(e.message));
  else window.open(url, '_blank', 'noopener');
};
$('#updateClose').onclick = () => $('#updateBanner').classList.add('hidden');

// ---- tools: stability monitor, DNS test, Wi-Fi scan ---------------------------
// Each tool runs only while the user waits for it. The monitor in particular
// stops when the window closes; there is no background mode.
const mon = { running: false, samples: [], seconds: 0, label: '', method: '', summary: null };
const tools = { dns: null, scan: null };

function openTools() {
  loadMonitorTargets();
  showTool(prefs.tool && $('#' + prefs.tool) ? prefs.tool : 'monitorCard');
}

// One tool at a time; the picker remembers the last one.
function showTool(id) {
  $$('.tool-picker [role="tab"]').forEach(b => {
    const on = b.dataset.tool === id;
    b.setAttribute('aria-selected', on);
    b.tabIndex = on ? 0 : -1;
  });
  $$('#toolsView > section.tool').forEach(c => c.classList.toggle('hidden', c.id !== id));
  if (prefs.tool !== id) store.set('tool', id);
  if (id === 'monitorCard' && mon.samples.length) drawMonitor();
  // The signal finder runs only while it is on screen.
  if (id === 'signalCard') { loadSpots(); if (sig.samples.length) drawSignal(); } else sigStop();
}
$$('.tool-picker [role="tab"]').forEach(b => { b.onclick = () => showTool(b.dataset.tool); });

async function loadMonitorTargets() {
  let list = [];
  try { list = await bridge.call('monitorTargets') || []; } catch (e) { toast(e.message); }
  const sel = $('#monTarget');
  const cur = sel.value;
  sel.innerHTML = '';
  for (const tg of list) {
    const o = document.createElement('option');
    o.value = tg.host;
    o.textContent = tg.label && tg.label !== tg.host ? `${tg.label} (${tg.host})` : tg.host;
    sel.appendChild(o);
  }
  if (list.some(x => x.host === cur)) sel.value = cur;
}

function setMonRunning(on) {
  mon.running = on;
  $('#monStart').classList.toggle('hidden', on);
  $('#monStop').classList.toggle('hidden', !on);
  $('#monStop').disabled = false;
  $('#monStop').textContent = t('mon.stop');
  $('#monMinutes').disabled = on;
  $('#monTarget').disabled = on;
  // A Wi-Fi scan takes the radio off-channel and a speed test saturates the
  // link; either would distort what the monitor records.
  $('#scanStart').disabled = on || tools.scanning;
  for (const id of ['startBtn', 'quickBtn']) {
    $('#' + id).disabled = on;
    $('#' + id).title = on ? t('mon.busy') : (id === 'quickBtn' ? t('quick.hint') : '');
  }
}

$('#monStart').onclick = async () => {
  const minutes = +$('#monMinutes').value;
  Object.assign(mon, { samples: [], seconds: minutes * 60, label: '', method: '', summary: null });
  $('#monResult').classList.add('hidden');
  $('#monLive').classList.remove('hidden');
  $('#monTargetName').textContent = $('#monTarget').selectedOptions[0]?.textContent || '';
  renderMonStats();
  drawMonitor();
  setMonRunning(true);
  try {
    await bridge.call('startMonitor', { minutes, host: $('#monTarget').value });
  } catch (e) {
    setMonRunning(false);
    $('#monLive').classList.add('hidden');
    toast(e.message);
  }
};
$('#monStop').onclick = () => {
  $('#monStop').disabled = true;
  $('#monStop').textContent = t('mon.stopping');
  bridge.call('stopMonitor').catch(() => {});
};
// The report names the target; a LAN address in it is masked (192.168.1.x).
$('#monExport').onclick = () => bridge.call('exportMonitor', false).catch(e => toast(e.message));

const clock = sec => { sec = Math.max(0, Math.round(sec)); return Math.floor(sec / 60) + ':' + String(sec % 60).padStart(2, '0'); };

function onMonitor(p) {
  if (p.error) {
    setMonRunning(false);
    if (!mon.samples.length) $('#monLive').classList.add('hidden');
    toast(t('mon.error', p.error), 6000);
    return;
  }
  if (p.started) {
    Object.assign(mon, { seconds: p.seconds, label: p.target, method: p.method });
    $('#monTargetName').textContent = p.method === 'tcp' ? `${p.target} (TCP)` : p.target;
    if (p.method === 'tcp') toast(t('mon.tcpNote'), 6000);
    renderMonStats();
    return;
  }
  if (p.sample) {
    mon.samples.push(p.sample);
    renderMonStats();
    if (!$('#toolsView').classList.contains('hidden')) drawMonitor();
    return;
  }
  if (p.summary) {
    setMonRunning(false);
    if (p.samples) mon.samples = p.samples;
    mon.summary = p.summary;
    renderMonStats();
    drawMonitor();
    renderMonSummary();
    // Finished while the user looked at another view.
    if ($('#toolsView').classList.contains('hidden')) toast(p.summary.verdict, 8000);
    bridge.call('attention').catch(() => {});
  }
}

function renderMonStats() {
  const xs = mon.samples;
  const last = xs[xs.length - 1];
  const ok = xs.filter(s => s.ok);
  $('#monElapsed').textContent = `${clock(last ? last.t + 1 : 0)} / ${clock(mon.seconds)}`;
  $('#monNow').textContent = !last ? '—' : last.ok ? ms(last.rttMs) + ' ms' : t('mon.lost');
  $('#monAvg').textContent = ok.length ? ms(ok.reduce((a, s) => a + s.rttMs, 0) / ok.length) + ' ms' : '—';
  $('#monLoss').textContent = xs.length ? pct(+((xs.length - ok.length) / xs.length * 100).toFixed(1)) + '%' : '—';
  const gw = xs.filter(s => s.gwOk != null);
  $('#monGw').textContent = gw.length ? `${gw.filter(s => !s.gwOk).length} / ${gw.length}` : t('mon.noRouter');
}

function renderMonSummary() {
  const s = mon.summary;
  if (!s) return;
  $('#monResult').classList.remove('hidden');
  $('#monVerdict').textContent = s.verdict || '';
  $('#monNotes').innerHTML = (s.notes || []).map(n => `<li>${esc(n)}</li>`).join('');
}

// Live ping chart: one series; lost probes are marked in the status color at
// the bottom, so a drop reads as a red stripe rather than a missing line.
function drawMonitor() {
  const canvas = $('#plotMon canvas');
  const { ctx, W, H } = setupCanvas(canvas);
  const xs = mon.samples;
  const muted = cssVar('--muted');
  const L = 44, R = 12, T = 10, B = 22;
  const total = Math.max(mon.seconds, xs.length ? xs[xs.length - 1].t + 1 : 1);
  const x = sec => L + (sec / total) * (W - L - R);
  const okVals = xs.filter(s => s.ok).map(s => s.rttMs);
  const max = niceMax(okVals.length ? Math.max(...okVals) : 50);
  const y = v => T + (1 - v / max) * (H - T - B);
  ctx.strokeStyle = cssVar('--line'); ctx.lineWidth = 1; ctx.fillStyle = muted; ctx.textAlign = 'right';
  for (const f of [0, 0.5, 1]) {
    const yy = Math.round(y(max * f)) + 0.5;
    ctx.beginPath(); ctx.moveTo(L, yy); ctx.lineTo(W - R, yy); ctx.stroke();
    ctx.fillText(axisNum(max * f), L - 6, yy + 4);
  }
  ctx.textAlign = 'left'; ctx.fillText('0:00', L, H - 6);
  ctx.textAlign = 'right'; ctx.fillText(clock(total), W - R, H - 6); ctx.textAlign = 'left';
  // Lost probes: a stripe across the plot so drops are seen at a glance.
  ctx.fillStyle = cssVar('--bad');
  const bw = Math.max(2, (W - L - R) / total);
  for (const s of xs) if (!s.ok) { ctx.globalAlpha = 0.18; ctx.fillRect(x(s.t), T, bw, H - T - B); ctx.globalAlpha = 1; ctx.fillRect(x(s.t), H - B - 4, bw, 4); }
  ctx.strokeStyle = cssVar('--series-1'); ctx.lineWidth = 1.5; ctx.lineJoin = 'round'; ctx.beginPath();
  let started = false;
  for (const s of xs) {
    if (!s.ok) { started = false; continue; }
    if (!started) { ctx.moveTo(x(s.t), y(s.rttMs)); started = true; } else ctx.lineTo(x(s.t), y(s.rttMs));
  }
  ctx.stroke();
}

$('#dnsStart').onclick = () => {
  $('#dnsStart').disabled = true;
  $('#dnsStatus').textContent = t('dns.running', 0, '…');
  bridge.call('dnsTest').catch(e => { $('#dnsStart').disabled = false; $('#dnsStatus').textContent = e.message; });
};
function onDNS(p) {
  if (!p.results) { $('#dnsStatus').textContent = t('dns.running', p.done, p.total); return; }
  $('#dnsStart').disabled = false;
  $('#dnsStatus').textContent = '';
  tools.dns = p;
  renderDNS();
}
function renderDNS() {
  const p = tools.dns;
  if (!p) return;
  const rows = p.results || [];
  $('#dnsTable').classList.toggle('hidden', rows.length === 0);
  $('#dnsTable tbody').innerHTML = rows.map((r, i) => {
    const name = r.label.includes(r.server) ? esc(r.label) : `${esc(r.label)} <span class="muted">${esc(r.server)}</span>`;
    const time = r.ok ? `${ms(r.medianMs)} ms${i === 0 ? ' ★' : ''}` : `<span class="muted">${esc(t('dns.noAnswer'))}</span>`;
    return `<tr class="${r.system ? 'mine' : ''}"><td>${name}</td><td>${time}</td><td>${r.ok} / ${r.ok + r.failed}</td></tr>`;
  }).join('');
  $('#dnsAdvice').innerHTML = (p.advice || []).map(a => `<li>${esc(a)}</li>`).join('');
}

$('#scanStart').onclick = () => {
  tools.scanning = true;
  $('#scanStart').disabled = true;
  $('#scanStatus').textContent = t('scan.scanning');
  bridge.call('wifiScan', true).catch(e => onScan({ error: e.message }));
};
function onScan(p) {
  tools.scanning = false;
  $('#scanStart').disabled = mon.running;
  if (p.error) {
    $('#scanStatus').textContent = t('scan.error', p.error);
    return;
  }
  tools.scan = p.report;
  renderScan();
}
const busyLevel = score => t(score < 1 ? 'busy.low' : score < 3 ? 'busy.medium' : 'busy.high');
function renderScan() {
  const r = tools.scan;
  if (!r) return;
  const nets = r.networks || [];
  $('#scanStatus').textContent = t('scan.found', nets.length);
  $('#scanAdvice').innerHTML = (r.advice || []).map(a => `<li>${esc(a)}</li>`).join('');
  $('#scanResult').classList.toggle('hidden', nets.length === 0);
  // Empty 5 GHz channels add nothing; 1/6/11 stay so a free one shows.
  const chans = (r.channels || []).filter(c => c.networks > 0 || c.current || c.recommended || (c.band === '2.4 GHz' && [1, 6, 11].includes(c.channel)));
  const maxScore = Math.max(1, ...chans.map(c => c.score));
  $('#chanTable tbody').innerHTML = chans.map(c => {
    const tags = [c.current ? t('scan.yours') : '', c.recommended ? t('scan.suggested') : ''].filter(Boolean);
    return `<tr class="${c.current ? 'mine' : ''}${c.recommended ? ' suggested' : ''}">
      <td>${esc(c.band)}</td><td><b>${c.channel}</b>${tags.length ? ` <span class="tag">${esc(tags.join(' · '))}</span>` : ''}</td>
      <td>${c.networks}</td>
      <td><span class="busy"><i style="width:${Math.round(c.score / maxScore * 100)}%"></i></span> ${esc(busyLevel(c.score))}</td></tr>`;
  }).join('');
  $('#netTable tbody').innerHTML = nets.slice(0, 30).map(n => `
    <tr class="${n.connected ? 'mine' : ''}">
      <td>${n.ssid ? esc(n.ssid) : `<span class="muted">${esc(t('scan.hidden'))}</span>`}${n.connected ? ` <span class="tag">${esc(t('scan.connected'))}</span>` : ''}</td>
      <td>${n.channel || '—'} <span class="muted">${esc(n.band)}</span></td>
      <td>${n.signalPercent ? n.signalPercent + '%' : '—'}${n.rssi ? ` <span class="muted">${n.rssi} dBm</span>` : ''}</td>
    </tr>`).join('');
}

// Texts from Go stay in the language they were produced in; the next run
// uses the new one. Static labels and table headers switch right away.
function renderTools() {
  renderMonStats();
  if (mon.running) $('#monStop').textContent = t('mon.stop');
  for (const o of $$('#monMinutes option')) o.textContent = `${o.value} ${t('min')}`;
  if (tools.scan) renderScan();
  if (tools.dns) renderDNS();
  $('#spotName').placeholder = t('sig.spotName');
  renderSig();
  renderSpots();
  renderRegions();
  renderDiag();
}

// ---- Wi-Fi signal finder --------------------------------------------------------
// Reads the signal once a second while this tool is open; no internet data.
const sig = { running: false, token: 0, samples: [], wifi: null, spots: [], beepTimer: 0, audio: null };
const sigLevel = p => (p >= 60 ? 'good' : p >= 40 ? 'fair' : 'weak');

function sigStart() {
  if (sig.running) return;
  Object.assign(sig, { running: true, token: sig.token + 1, samples: [], wifi: null });
  $('#sigError').textContent = '';
  renderSig();
  bridge.call('signalStart', sig.token).catch(e => {
    sig.running = false;
    renderSig();
    $('#sigError').textContent = t('sig.error', e.message);
  });
}
function sigStop() {
  if (!sig.running) return;
  sig.running = false;
  renderSig();
  bridge.call('signalStop').catch(() => {});
}
$('#sigStart').onclick = () => { audioReady(); sigStart(); };
$('#sigStop').onclick = sigStop;
$('#sigBeep').onchange = () => { audioReady(); beepUpdate(); };

function onSignal(p) {
  if (p.token !== sig.token || !sig.running) return;
  if (p.stopped) { sig.running = false; renderSig(); return; }
  if (p.error) { $('#sigError').textContent = t('sig.error', p.error); return; }
  $('#sigError').textContent = '';
  sig.wifi = p.wifi;
  sig.samples.push({ t: p.t, pct: p.wifi.signalPercent, rssi: p.wifi.rssi });
  sig.samples = sig.samples.filter(s => s.t > p.t - 60);
  renderSig();
  drawSignal();
}

function renderSig() {
  $('#sigStart').classList.toggle('hidden', sig.running);
  $('#sigStop').classList.toggle('hidden', !sig.running);
  $('#spotSave').disabled = !sig.running || !sig.samples.length;
  const w = sig.wifi;
  $('#sigLive').classList.toggle('hidden', !w);
  beepUpdate();
  if (!w) return;
  const lvl = sigLevel(w.signalPercent);
  $('#sigPct').textContent = w.signalPercent + '%';
  $('#sigPct').className = 'sig-' + lvl;
  $('#sigWord').textContent = t('sig.' + lvl);
  $('#sigWord').className = 'sig-word sig-' + lvl;
  $('#sigDbm').textContent = w.rssi ? w.rssi + ' dBm' : '';
  $('#sigMeta').textContent = [w.ssid, w.band, w.channel ? 'ch ' + w.channel : '',
    w.rxRateMbps ? t('net.link') + ' ' + num(w.rxRateMbps) + ' Mbps' : ''].filter(Boolean).join(' · ');
}

function drawSignal() {
  const { ctx, W, H } = setupCanvas($('#plotSig canvas'));
  const xs = sig.samples;
  const L = 40, R = 12, T = 10, B = 22;
  // The newest reading sits at "now" on the right; older ones scroll left.
  const end = xs.length ? xs[xs.length - 1].t : 0;
  const x = sec => L + ((sec - (end - 60)) / 60) * (W - L - R);
  const y = v => T + (1 - v / 100) * (H - T - B);
  ctx.strokeStyle = cssVar('--line'); ctx.lineWidth = 1; ctx.fillStyle = cssVar('--muted'); ctx.textAlign = 'right';
  // 40% and 60% are where the reading turns weak / strong.
  for (const v of [0, 40, 60, 100]) {
    const yy = Math.round(y(v)) + 0.5;
    ctx.beginPath(); ctx.moveTo(L, yy); ctx.lineTo(W - R, yy); ctx.stroke();
    ctx.fillText(v + '%', L - 6, yy + 4);
  }
  ctx.textAlign = 'left'; ctx.fillText('−60 s', L, H - 6);
  ctx.textAlign = 'right'; ctx.fillText(t('sig.now'), W - R, H - 6); ctx.textAlign = 'left';
  if (!xs.length) return;
  const col = cssVar('--series-1');
  ctx.strokeStyle = col; ctx.lineWidth = 2; ctx.lineJoin = 'round'; ctx.beginPath();
  xs.forEach((s, i) => (i ? ctx.lineTo(x(s.t), y(s.pct)) : ctx.moveTo(x(s.t), y(s.pct))));
  ctx.stroke();
  const last = xs[xs.length - 1];
  dot(ctx, x(last.t), y(last.pct), col);
}

// Metal-detector beep: the stronger the signal, the faster it ticks. The
// audio is unlocked by a click, as browsers require.
function audioReady() {
  if (!$('#sigBeep').checked) return;
  try { sig.audio = sig.audio || new AudioContext(); sig.audio.resume(); } catch { /* no audio */ }
}
function beepUpdate() {
  const want = sig.running && !!sig.wifi && $('#sigBeep').checked;
  if (want && !sig.beepTimer) beepTick();
  if (!want && sig.beepTimer) { clearTimeout(sig.beepTimer); sig.beepTimer = 0; }
}
function beepTick() {
  const p = sig.wifi ? sig.wifi.signalPercent : 0;
  try {
    const a = sig.audio, o = a.createOscillator(), g = a.createGain();
    o.frequency.value = 520 + p * 6;
    g.gain.setValueAtTime(0.08, a.currentTime);
    g.gain.exponentialRampToValueAtTime(0.001, a.currentTime + 0.08);
    o.connect(g).connect(a.destination);
    o.start();
    o.stop(a.currentTime + 0.09);
  } catch { /* no audio */ }
  sig.beepTimer = setTimeout(() => { sig.beepTimer = 0; beepUpdate(); }, 1400 - p * 12);
}

async function loadSpots() {
  try { sig.spots = await bridge.call('getSpots') || []; } catch { sig.spots = []; }
  renderSpots();
}
function renderSpots() {
  const list = [...sig.spots].sort((a, b) => b.signalPercent - a.signalPercent);
  $('#spotBox').classList.toggle('hidden', list.length === 0);
  $('#spotTable tbody').innerHTML = list.map((s, i) => {
    const best = i === 0 && list.length > 1;
    return `<tr class="${best ? 'suggested' : ''}"><td>${i + 1}</td><td>${esc(s.name)}${best ? ' ★' : ''}</td>
      <td><b class="sig-${sigLevel(s.signalPercent)}">${s.signalPercent}%</b>${s.rssi ? ` <span class="muted">${s.rssi} dBm</span>` : ''}</td></tr>`;
  }).join('');
}
// A spot keeps the average of the last 5 readings, so one odd reading
// does not decide it.
$('#spotSave').onclick = async () => {
  const recent = sig.samples.slice(-5);
  if (!recent.length) return;
  const avg = list => (list.length ? Math.round(list.reduce((a, v) => a + v, 0) / list.length) : 0);
  const name = $('#spotName').value.trim();
  try {
    sig.spots = await bridge.call('saveSpot', { name, signalPercent: avg(recent.map(s => s.pct)), rssi: avg(recent.map(s => s.rssi).filter(Boolean)) }) || [];
    $('#spotName').value = '';
    renderSpots();
    toast(t('sig.saved', name));
  } catch (e) { toast(e.message); }
};
$('#spotName').onkeydown = e => { if (e.key === 'Enter') $('#spotSave').click(); };
$('#spotClear').onclick = async () => {
  try { await bridge.call('clearSpots'); sig.spots = []; renderSpots(); } catch (e) { toast(e.message); }
};

// ---- ping to regions ---------------------------------------------------------------
const reg = { running: false, total: 0, results: [], advice: [] };
$('#regStart').onclick = () => {
  Object.assign(reg, { running: true, total: 0, results: [], advice: [] });
  renderRegions();
  bridge.call('regionTest').catch(e => { reg.running = false; renderRegions(); toast(e.message); });
};
function onRegion(p) {
  if (p.results) Object.assign(reg, { running: false, results: p.results, advice: p.advice || [] });
  else if (p.result) reg.results.push(p.result);
  else if (p.total !== undefined) reg.total = p.total;
  renderRegions();
}
function renderRegions() {
  $('#regStart').disabled = reg.running;
  $('#regStatus').textContent = reg.running ? t('reg.running', reg.results.length, reg.total || '…') : '';
  // Fastest first, as they arrive; regions that did not answer go last.
  const rows = [...reg.results].sort((a, b) => (b.ok > 0) - (a.ok > 0) || a.medianMs - b.medianMs);
  $('#regTable').classList.toggle('hidden', rows.length === 0);
  const max = Math.max(1, ...rows.filter(r => r.ok > 0).map(r => r.medianMs));
  $('#regTable tbody').innerHTML = rows.map(r => {
    if (!(r.ok > 0)) return `<tr data-label="${esc(r.label)}"><td>${esc(r.label)}</td><td class="muted">${esc(t('reg.noAnswer'))}</td><td></td></tr>`;
    // Same limits as online games in "What is it good for?".
    const cls = r.medianMs <= 60 ? 'reg-good' : r.medianMs <= 100 ? 'reg-ok' : 'reg-high';
    return `<tr data-label="${esc(r.label)}"><td>${esc(r.label)}</td><td>${ms(r.medianMs)} ms</td>
      <td><span class="reg-bar ${cls}"><i style="width:${Math.max(4, Math.round(r.medianMs / max * 100))}%"></i></span></td></tr>`;
  }).join('');
  $('#regAdvice').innerHTML = reg.advice.map(a => `<li>${esc(a)}</li>`).join('');
}

// ---- guided diagnosis ------------------------------------------------------------
// Up to three tests (where you are, near the router, on a cable); Go draws
// the conclusion. A step is skipped once the answer is already clear. The
// tests are not stored in History.
const diag = { step: 'now', steps: [], marks: {}, running: false, cancelled: false, result: null, strongPct: 0 };

function diagReset() {
  Object.assign(diag, { step: 'now', steps: [], marks: {}, running: false, cancelled: false, result: null, strongPct: 0 });
  renderDiag();
}

function diagMarkText(k) {
  const m = diag.marks[k];
  if (m === 'skip') return t('diag.skipped');
  if (m === 'strong') return t('diag.skippedStrong', diag.strongPct);
  const st = diag.steps.find(x => x.kind === k);
  const r = st && st.result;
  if (!r) return '';
  if (r.download && r.download.mbps > 0) return '↓ ' + speed(r.download.mbps) + ' ' + state.unit;
  return r.failure ? '✗' : '✓';
}

// The colour of a finished step follows that test's rating, so a slow
// result is not shown in green.
function diagQualityClass(k) {
  const st = diag.steps.find(x => x.kind === k);
  const r = st && st.result;
  if (!r) return '';
  const level = r.quality && r.quality.level;
  if (level === 'EXCELLENT' || level === 'GOOD') return 'q-ok';
  if (level === 'FAIR') return 'q-warn';
  return level || r.failure ? 'q-bad' : '';
}

function renderDiag() {
  for (const k of ['now', 'near', 'cable', 'result']) {
    const li = $(`#diagSteps [data-step="${k}"]`);
    const m = diag.marks[k];
    li.className = k === diag.step && !diag.result ? 'run'
      : m === 'skip' || m === 'strong' ? 'skip'
        : m ? ('done ' + diagQualityClass(k)).trim() : '';
    li.querySelector('.ds-state').textContent = diagMarkText(k);
  }
  const done = !!diag.result;
  $('#diagPanel').classList.toggle('hidden', done);
  $('#diagResult').classList.toggle('hidden', !done);
  if (done) { renderDiagResult(); return; }
  const s = diag.step;
  const cableMsg = diag.marks.near === 'strong' ? t('diag.cableMsgStrong', diag.strongPct) : t('diag.cableMsg');
  $('#diagMsg').textContent = diag.cancelled ? t('diag.cancelled') : { now: t('diag.ready'), near: t('diag.nearMsg'), cable: cableMsg }[s];
  $('#diagGo').textContent = t({ now: 'diag.start', near: 'diag.goNear', cable: 'diag.goCable' }[s]);
  $('#diagGo').classList.toggle('hidden', diag.running);
  $('#diagCancel').classList.toggle('hidden', !diag.running);
  $('#diagSkip').classList.toggle('hidden', diag.running || s === 'now');
  $('#diagSkip').textContent = t(s === 'cable' ? 'diag.skipToResult' : 'diag.skip');
  $('#diagLive').classList.toggle('hidden', !diag.running);
}

function setDiagRunning(on) {
  diag.running = on;
  $('#diagCancel').disabled = false;
  lockTabs(on);
  renderDiag();
}

async function diagRun() {
  diag.cancelled = false;
  $('#diagStage').textContent = t('gauge.starting');
  $('#diagValue').textContent = '';
  setDiagRunning(true);
  try {
    await bridge.call('startTest', 'general', { diagnosis: true, dataSaver: state.saver });
  } catch (e) {
    setDiagRunning(false);
    toast(e.message);
  }
}
$('#diagGo').onclick = diagRun;
$('#diagCancel').onclick = () => { $('#diagCancel').disabled = true; bridge.call('cancelTest'); };
$('#diagSkip').onclick = () => {
  diag.marks[diag.step] = 'skip';
  diag.cancelled = false;
  if (diag.step === 'near') { diag.step = 'cable'; renderDiag(); } else diagConclude();
};
$('#diagAgain').onclick = diagReset;

// While a diagnosis step runs, test events update this panel only; the
// Test tab keeps the last normal result.
function onDiagTestEvent(ev) {
  switch (ev.type) {
    case 'stage':
      $('#diagStage').textContent = stageLabel(ev.stage);
      $('#diagValue').textContent = '';
      break;
    case 'network':
      renderNetwork(ev.network);
      break;
    case 'progress':
      if (ev.stage === 'ping' && ev.rttMs) $('#diagValue').textContent = ms(ev.rttMs) + ' ms';
      else if (ev.stage === 'download' || ev.stage === 'upload') $('#diagValue').textContent = speed(ev.mbps || 0) + ' ' + state.unit;
      break;
    case 'result':
      diagStepDone(ev.result);
      break;
  }
}

// Same limit as analysis.StrongSignal in Go.
const STRONG_SIGNAL = 70;
const wifiSignal = r => (r.network && r.network.wifi && r.network.wifi.signalPercent) || 0;

// A cable or a walk to the router only helps on a wireless link.
const wireless = r => !(r.network && ['Ethernet', 'Cellular'].includes(r.network.connection));

async function diagStepDone(r) {
  setDiagRunning(false);
  if (r.cancelled) { diag.cancelled = true; renderDiag(); return; }
  const kind = diag.step;
  diag.steps = diag.steps.filter(s => s.kind !== kind).concat([{ kind, result: r }]);
  diag.marks[kind] = 'done';
  let d = null;
  try { d = await bridge.call('diagnose', diag.steps); } catch (e) { toast(e.message); }
  // Over a phone hotspot there is no router to walk to or cable to plug.
  const clear = !d || kind === 'cable' || ['healthy', 'offline', 'offlineLocal', 'coverage', 'hotspot'].includes(d.kind)
    || (kind === 'now' && !wireless(r));
  if (clear) { diagConclude(d); return; }
  if (kind === 'now' && wifiSignal(r) >= STRONG_SIGNAL) {
    // Moving closer cannot help a strong signal; only a cable test can
    // still tell the Wi-Fi from the ISP.
    diag.marks.near = 'strong';
    diag.strongPct = wifiSignal(r);
    diag.step = 'cable';
  } else {
    diag.step = kind === 'now' ? 'near' : 'cable';
  }
  renderDiag();
}

async function diagConclude(d) {
  if (!d) {
    try { d = await bridge.call('diagnose', diag.steps); } catch (e) { toast(e.message); renderDiag(); return; }
  }
  for (const k of ['near', 'cable']) if (!diag.marks[k]) diag.marks[k] = 'skip';
  diag.marks.result = 'done';
  diag.step = 'result';
  diag.result = d;
  renderDiag();
  bridge.call('attention').catch(() => {});
}

function openTool(id) { showView('tools'); showTool(id); }
const diagActions = {
  channel: ['diag.act.channel', () => openTool('wifiCard')],
  dns: ['diag.act.dns', () => openTool('dnsCard')],
  ispReport: ['diag.act.ispReport', () => showView('history')],
  monitor: ['diag.act.monitor', () => openTool('monitorCard')],
};
function renderDiagResult() {
  const d = diag.result;
  $('#diagTitle').textContent = (d.kind === 'healthy' ? '✓ ' : '⚠ ') + d.title;
  $('#diagTitle').className = 'dk-' + d.kind;
  $('#diagDetail').textContent = d.detail;
  $('#diagFacts').innerHTML = (d.facts || []).map(f => `<li>${esc(f)}</li>`).join('');
  const box = $('#diagActions');
  box.innerHTML = '';
  for (const a of d.actions || []) {
    if (!diagActions[a]) continue; // "router" is advice in the text, not a tool
    const b = document.createElement('button');
    b.className = 'ghost small';
    b.dataset.action = a;
    b.textContent = t(diagActions[a][0]);
    b.onclick = diagActions[a][1];
    box.appendChild(b);
  }
  box.classList.toggle('hidden', box.children.length === 0);
}

// ---- report a problem ------------------------------------------------------------
// Go builds the report (no addresses or network names); the page shows it
// before anything leaves the computer.
let problem = null;
async function openProblem() {
  try { problem = await bridge.call('problemReport'); } catch (e) { toast(e.message); return; }
  $('#problemInfo').textContent = problem.text;
  if ($('#aboutDlg').open) $('#aboutDlg').close();
  $('#problemDlg').showModal();
}
$('#aboutProblem').onclick = openProblem;
$('#failProblem').onclick = openProblem;
$('#problemIssue').onclick = () => {
  if (!problem) return;
  if (state.info && state.info.canOpenUrl) bridge.call('openUrl', problem.url).catch(e => toast(e.message));
  else window.open(problem.url, '_blank', 'noopener');
};
$('#problemCopy').onclick = async () => {
  if (!problem) return;
  try {
    const copied = await bridge.call('copyProblem', problem.text);
    if (!copied) await navigator.clipboard.writeText(problem.text);
    toast(t('prob.copied'));
  } catch (e) { toast(t('toast.copyFailed', e.message || e)); }
};

// ---- about ------------------------------------------------------------------
$('#aboutBtn').onclick = async () => {
  // Refresh: ping targets may have changed in Settings since startup.
  try { state.info = await bridge.call('info'); } catch { /* keep the old one */ }
  const i = state.info || {};
  $('#aboutVersion').textContent = i.version ? 'v' + i.version : '';
  const servers = [state.server ? state.server.name : null, ...(i.servers || [])].filter(Boolean);
  $('#aboutTraffic').innerHTML = [
    t('about.trafficSpeed', servers.join(', ')),
    t('about.trafficPing', (i.pingTargets || []).join(', ')),
    t('about.trafficDNS'),
    t('about.trafficDNSTest', (i.dnsServers || []).join(', ')),
    t('about.trafficMonitor'),
    t('about.trafficRegions', (i.regions || []).join(', ')),
    t('about.trafficScan'),
    t('about.trafficSignal'),
    t('about.trafficUpdate'),
  ].map(s => `<li>${esc(s)}</li>`).join('');
  $('#aboutHistory').textContent = i.historyPath || '';
  $('#aboutCfgName').textContent = i.configFile || '';
  $('#aboutCfg').textContent = i.configError ? t('about.cfgError', i.configError)
    : i.configPath ? t('about.cfgLoaded', i.configPath) : t('about.cfgNone');
  $('#aboutDlg').showModal();
};

function renderStatic() {
  const servers = [state.server ? state.server.name : null, ...((state.info && state.info.servers) || [])].filter(Boolean);
  $('#privacyLine').textContent = t('privacy', [...new Set(servers)].join(', '));
  renderServerName();
  if (!state.result && !state.running) setGauge(t('gauge.ready'), '—', '', 0);
  if (state.network) renderNetwork(state.network);
  renderPrivacyHint();
  renderTestOpts();
  if (state.update && state.update.newer) $('#updateText').textContent = t('update.available', state.update.latest, state.update.current);
}

// ---- events -----------------------------------------------------------------
bridge.listen((name, payload) => {
  if (name === 'test') onTestEvent(payload);
  else if (name === 'network') {
    if (payload.network) renderNetwork(payload.network);
    else if (!state.network) { $('#netConn').textContent = t('net.notConnected'); $('#netName').textContent = payload.error || '—'; }
  } else if (name === 'saved') {
    if (payload.error) toast(t('toast.exportFailed', payload.error), 6000);
    else if (payload.path) toast(t('toast.saved', payload.path), 6000);
  } else if (name === 'servers') onServers(payload);
  else if (name === 'update') onUpdate(payload);
  else if (name === 'monitor') onMonitor(payload);
  else if (name === 'dns') onDNS(payload);
  else if (name === 'wifiscan') onScan(payload);
  else if (name === 'signal') onSignal(payload);
  else if (name === 'region') onRegion(payload);
});

// ---- init ---------------------------------------------------------------------
(async function init() {
  try { Object.assign(prefs, await bridge.call('getPrefs')); } catch { /* defaults */ }
  if (prefs.unit === 'MB/s' || prefs.unit === 'Mbps') state.unit = prefs.unit;
  if (prefs.mode === 'gaming' || prefs.mode === 'general') state.mode = prefs.mode;
  state.saver = prefs.dataSaver === 'on';
  state.repeat = prefs.repeat === 'on';
  // First run: follow the Windows display language.
  lang = prefs.lang === 'id' || prefs.lang === 'en' ? prefs.lang : ((navigator.language || '').toLowerCase().startsWith('id') ? 'id' : 'en');
  if (prefs.lang !== lang) store.set('lang', lang);
  applyText();
  applyTheme();
  renderTools();
  $$('#modeSeg button').forEach(x => x.classList.toggle('on', x.dataset.mode === state.mode));
  $$('#unitSeg button').forEach(x => x.classList.toggle('on', x.dataset.unit === state.unit));
  unitLabels();
  try {
    state.info = await bridge.call('info');
    if (state.info.configError) toast(t('toast.configError', state.info.configError), 8000);
  } catch (e) { toast(e.message); }
  try { state.server = await bridge.call('getServer'); } catch { state.server = null; }
  renderStatic();
  refreshHistoryList();
  bridge.call('detectNetwork').catch(() => {});
  if (prefs.updateCheck !== 'off') bridge.call('checkUpdate').catch(() => {});
})();

window.addEventListener('keydown', e => {
  if (e.key === 'Escape' && state.running) $('#cancelBtn').click();
  if (e.key === 'Escape' && diag.running) $('#diagCancel').click();
  // Ctrl+1/2/3 switch tabs, like browser tabs.
  if (e.ctrlKey && !e.altKey && !e.shiftKey && /^[1-9]$/.test(e.key)) {
    const tab = $$('.tabs [role="tab"]')[+e.key - 1];
    if (tab && !tab.disabled) { tab.click(); e.preventDefault(); }
  }
});
