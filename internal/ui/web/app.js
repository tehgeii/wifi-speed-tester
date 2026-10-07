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
    details: 'Details', copy: 'Copy result', export: 'Export:', 'export.png': 'PNG card',
    'export.sensitive': 'Include network identifiers',
    'export.sensitiveHint': 'Wi-Fi name, BSSID and IP addresses are hidden unless this is ticked.',
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
    'about.title': 'About and privacy', 'about.privacy': 'Privacy', 'about.accuracy': 'Accuracy', 'about.config': 'Configuration',
    'about.intro': 'WiFi Speed + Ping Tester — a portable utility that measures download, upload, ping, jitter and packet loss.',
    'about.traffic': 'Testing is local except for the traffic needed to measure your connection:',
    'about.local': 'The app does not read your files, browser history or passwords, and does not run in the background. History is stored only on this computer:',
    'about.accuracyText': 'Results depend on the test server, routing, Wi-Fi signal, other traffic, VPNs and this device. They show what was measured now, not a guaranteed ISP speed.',
    'about.configText': 'Place this file next to the .exe to change test servers, ping targets, durations and quality thresholds:',
    'about.trafficSpeed': 'Speed test data to: {0}', 'about.trafficPing': 'Ping (ICMP, or TCP if ICMP is blocked) to: {0}',
    'about.trafficDNS': 'A DNS lookup of the test server name', 'about.trafficUpdate': 'If enabled in Settings: an update check to api.github.com',
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
    'wifi.hotspot': 'Gateway address is typical of a phone hotspot.',
    'wifi.ssidHidden': 'Windows hides the Wi-Fi name. On Windows 11, turn on Settings → Privacy & security → Location → Let desktop apps access your location.',
    'wifi.unavailable': 'Wi-Fi details unavailable: {0}',
    'conn.Wi-Fi': 'Wi-Fi', 'conn.Ethernet': 'Ethernet', 'conn.Hotspot / Tethering': 'Hotspot / Tethering', 'conn.Cellular': 'Cellular',
    'conn.VPN / Virtual': 'VPN / Virtual', 'conn.Unknown': 'Unknown',
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
    details: 'Detail', copy: 'Salin hasil', export: 'Export:', 'export.png': 'Kartu PNG',
    'export.sensitive': 'Sertakan identitas jaringan',
    'export.sensitiveHint': 'Nama Wi-Fi, BSSID dan alamat IP disembunyikan kecuali dicentang.',
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
    'about.title': 'Tentang dan privasi', 'about.privacy': 'Privasi', 'about.accuracy': 'Akurasi', 'about.config': 'Konfigurasi',
    'about.intro': 'WiFi Speed + Ping Tester — aplikasi portable untuk mengukur download, upload, ping, jitter dan packet loss.',
    'about.traffic': 'Tes berjalan lokal, kecuali lalu lintas yang memang dibutuhkan untuk mengukur koneksi:',
    'about.local': 'Aplikasi ini tidak membaca file, riwayat browser atau password, dan tidak berjalan di latar belakang. Riwayat hanya disimpan di komputer ini:',
    'about.accuracyText': 'Hasil dipengaruhi server tes, rute jaringan, sinyal Wi-Fi, pemakaian lain, VPN dan perangkat ini. Angkanya menunjukkan hasil pengukuran saat ini, bukan jaminan kecepatan dari ISP.',
    'about.configText': 'Taruh file ini di samping .exe untuk mengubah server tes, target ping, durasi dan batas rating:',
    'about.trafficSpeed': 'Data speed test ke: {0}', 'about.trafficPing': 'Ping (ICMP, atau TCP kalau ICMP diblokir) ke: {0}',
    'about.trafficDNS': 'Pencarian DNS untuk nama server tes', 'about.trafficUpdate': 'Kalau diaktifkan di Pengaturan: cek update ke api.github.com',
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
    'wifi.hotspot': 'Alamat gateway ini biasanya milik hotspot HP.',
    'wifi.ssidHidden': 'Windows menyembunyikan nama Wi-Fi. Di Windows 11, nyalakan Settings → Privacy & security → Location → Let desktop apps access your location.',
    'wifi.unavailable': 'Detail Wi-Fi tidak tersedia: {0}',
    'conn.Wi-Fi': 'Wi-Fi', 'conn.Ethernet': 'Ethernet (kabel)', 'conn.Hotspot / Tethering': 'Hotspot / Tethering', 'conn.Cellular': 'Seluler',
    'conn.VPN / Virtual': 'VPN / Virtual', 'conn.Unknown': 'Tidak diketahui',
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
  ['#failureCard', '#summaryCard', '#detailsCard'].forEach(s => $(s).classList.add('hidden'));
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
  $('#historyBtn').disabled = on;
  $('#settingsBtn').disabled = on;
  if (!on) activeMetric('');
}

const stageLabel = s => ({ connection: t('gauge.checking'), ping: t('stage.ping'), download: 'Download', upload: 'Upload' }[s] || '');

function onTestEvent(ev) {
  switch (ev.type) {
    case 'stage':
      $$('.stage').forEach(s => s.classList.toggle('on', s.dataset.stage === ev.stage));
      activeMetric(ev.stage === 'ping' ? 'ping' : ev.stage);
      setGauge(stageLabel(ev.stage), '—', '', 0);
      break;
    case 'check':
      renderCheck(ev.check);
      break;
    case 'network':
      renderNetwork(ev.network);
      break;
    case 'progress':
      stageProgress(ev.stage, ev.progress);
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
    case 'result':
      setRunning(false);
      renderResult(ev.result);
      break;
  }
}

// ---- result ---------------------------------------------------------------
function speedSub(s) {
  if (!s) return '';
  if (s.error) return s.mbps > 0 ? t('incomplete') : t('failed');
  return s.loadedSamples ? t('underLoad', ms(s.loadedPingMs)) : '';
}

function renderResult(r, fromHistory) {
  state.result = r;
  if (r.network) renderNetwork(r.network);
  if (!fromHistory) (r.checks || []).forEach(renderCheck);
  const quick = r.mode === 'quick';

  const dl = r.download, ul = r.upload;
  const sp = s => (s && s.mbps > 0 ? speed(s.mbps) : '—');
  setMetric('download', sp(dl), quick ? t('notInQuick') : speedSub(dl));
  setMetric('upload', sp(ul), quick ? t('notInQuick') : speedSub(ul));
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
  }
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
    const val = stab ? `${t('level.' + g.grade)} (${g.value})` : g.value;
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
    g.fillText(`Jitter ${r.pingMs > 0 ? ms(r.jitterMs) + ' ms' : '—'}     Loss ${r.pingMs > 0 ? pct(r.packetLossPct) + '%' : '—'}`, 32, 210);
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
  resetRun(mode);
  setRunning(true);
  setGauge(t('gauge.starting'), '—', '', 0);
  try {
    await bridge.call('startTest', mode);
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

// ---- history --------------------------------------------------------------
async function showHistory() {
  $('#dashboard').classList.add('hidden');
  $('#historyView').classList.remove('hidden');
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
  $('#histEmpty').classList.toggle('hidden', !empty);
  ['#histTable', '#histFilters', '#trendBox', '#compareTable'].forEach(s => $(s).classList.toggle('hidden', empty));
  $$('#historyView .sub-head').forEach(el => el.classList.toggle('hidden', empty));

  $('#histTable tbody').innerHTML = list.map(h => `
    <tr data-id="${esc(h.id)}" tabindex="0">
      <td>${esc(new Date(h.date).toLocaleString(locale(), { dateStyle: 'medium', timeStyle: 'short' }))}${h.mode !== 'general' ? ` <span class="muted">(${esc(h.mode === 'quick' ? t('quick') : t('mode.' + h.mode))})</span>` : ''}</td>
      <td>${esc(connName(h.connection))}</td>
      <td>${h.download ? speed(h.download) : '—'} / ${h.upload ? speed(h.upload) : '—'} ${state.unit}</td>
      <td>${h.ping ? ms(h.ping) + ' ms' : '—'}</td>
      <td>${h.ping ? ms(h.jitter) + ' ms' : '—'}</td>
      <td>${h.ping ? pct(h.loss) + '%' : '—'}</td>
      <td class="q-${esc(h.quality || 'UNKNOWN')}">${esc(h.quality ? t('level.' + h.quality) : '—')}</td>
    </tr>`).join('');
  $$('#histTable tbody tr').forEach(tr => {
    const open = async () => {
      try {
        const r = await bridge.call('historyItem', tr.dataset.id);
        if (!r) return;
        hideHistory();
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
  const draw = hover => {
    const { ctx, W, H } = setupCanvas(canvas);
    const muted = cssVar('--muted');
    if (pts.length < 2) { ctx.fillStyle = muted; ctx.fillText(t('trend.noData'), 12, H / 2); return null; }
    const vals = pts.flatMap(h => series.map(s => s.get(h))).filter(v => v != null);
    const max = niceMax(Math.max(...vals));
    const L = 52, R = 56, T = 10, B = 22;
    const x = i => L + (i / (pts.length - 1)) * (W - L - R);
    const y = v => T + (1 - v / max) * (H - T - B);
    // Recessive grid with three ticks.
    ctx.strokeStyle = cssVar('--line'); ctx.lineWidth = 1; ctx.fillStyle = muted; ctx.textAlign = 'right';
    for (const f of [0, 0.5, 1]) {
      const yy = Math.round(y(max * f)) + 0.5;
      ctx.beginPath(); ctx.moveTo(L, yy); ctx.lineTo(W - R, yy); ctx.stroke();
      ctx.fillText(axisNum(max * f), L - 6, yy + 4);
    }
    ctx.textAlign = 'left';
    const d0 = new Date(pts[0].date).toLocaleDateString(locale(), { day: 'numeric', month: 'short' });
    const d1 = new Date(pts[pts.length - 1].date).toLocaleDateString(locale(), { day: 'numeric', month: 'short' });
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
    show(Math.max(0, Math.min(pts.length - 1, Math.round(f * (pts.length - 1)))));
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

function hideHistory() {
  $('#historyView').classList.add('hidden');
  $('#dashboard').classList.remove('hidden');
}
$('#historyBtn').onclick = showHistory;
$('#histBack').onclick = hideHistory;
$$('#histRange button').forEach(b => b.onclick = () => { state.histRange = +b.dataset.range; renderHistory(); });
$('#histClear').onclick = async () => {
  if (!confirm(t('history.confirmClear'))) return;
  try { await bridge.call('clearHistory'); showHistory(); } catch (e) { toast(e.message); }
};
$('#histCsv').onclick = () => bridge.call('export', { format: 'history-csv', unit: state.unit, includeSensitive: $('#includeSensitive').checked }).catch(e => toast(e.message));
window.addEventListener('resize', () => { if (!$('#historyView').classList.contains('hidden')) renderTrend(filteredHistory()); });

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
$('#settingsBtn').onclick = () => {
  $('#langSelect').value = lang;
  $('#updateToggle').checked = prefs.updateCheck !== 'off';
  $('#updateStatus').textContent = '';
  $('#settingsDlg').showModal();
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

// ---- about ------------------------------------------------------------------
$('#aboutBtn').onclick = () => {
  const i = state.info || {};
  $('#aboutVersion').textContent = i.version ? 'v' + i.version : '';
  const servers = [state.server ? state.server.name : null, ...(i.servers || [])].filter(Boolean);
  $('#aboutTraffic').innerHTML = [
    t('about.trafficSpeed', servers.join(', ')),
    t('about.trafficPing', (i.pingTargets || []).join(', ')),
    t('about.trafficDNS'),
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
});

// ---- init ---------------------------------------------------------------------
(async function init() {
  try { Object.assign(prefs, await bridge.call('getPrefs')); } catch { /* defaults */ }
  if (prefs.unit === 'MB/s' || prefs.unit === 'Mbps') state.unit = prefs.unit;
  if (prefs.mode === 'gaming' || prefs.mode === 'general') state.mode = prefs.mode;
  // First run: follow the Windows display language.
  lang = prefs.lang === 'id' || prefs.lang === 'en' ? prefs.lang : ((navigator.language || '').toLowerCase().startsWith('id') ? 'id' : 'en');
  if (prefs.lang !== lang) store.set('lang', lang);
  applyText();
  $$('#modeSeg button').forEach(x => x.classList.toggle('on', x.dataset.mode === state.mode));
  $$('#unitSeg button').forEach(x => x.classList.toggle('on', x.dataset.unit === state.unit));
  unitLabels();
  try {
    state.info = await bridge.call('info');
    if (state.info.configError) toast(t('toast.configError', state.info.configError), 8000);
  } catch (e) { toast(e.message); }
  try { state.server = await bridge.call('getServer'); } catch { state.server = null; }
  renderStatic();
  bridge.call('detectNetwork').catch(() => {});
  if (prefs.updateCheck !== 'off') bridge.call('checkUpdate').catch(() => {});
})();

window.addEventListener('keydown', e => {
  if (e.key === 'Escape' && state.running) $('#cancelBtn').click();
});
