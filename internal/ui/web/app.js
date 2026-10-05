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

// ---- helpers ------------------------------------------------------------
const $ = s => document.querySelector(s);
const $$ = s => Array.from(document.querySelectorAll(s));
// Preferences live in the app's data folder (the embedded page has no
// persistent browser storage).
const prefs = {};
const store = {
  get(k, d) { return prefs[k] ?? d; },
  set(k, v) { prefs[k] = v; bridge.call('setPrefs', prefs).catch(() => {}); },
};
const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const ms = v => (v < 10 ? v.toFixed(1) : Math.round(v).toString());
const pct = v => (Number.isInteger(v) ? v.toString() : v.toFixed(1));

const state = {
  unit: 'Mbps',
  mode: 'general',
  running: false,
  result: null,
  network: null,
  info: null,
  pingLive: [],
};

// 80 Mbps = 10 MB/s (megabits vs megabytes).
function speed(mbps) {
  if (state.unit === 'MB/s') return (mbps / 8).toFixed(2);
  return mbps >= 100 ? Math.round(mbps).toString() : mbps.toFixed(1);
}

function toast(msg, ms_ = 4000) {
  const t = $('#toast');
  t.textContent = msg;
  t.classList.remove('hidden');
  clearTimeout(toast.timer);
  toast.timer = setTimeout(() => t.classList.add('hidden'), ms_);
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
function renderNetwork(n) {
  state.network = n;
  if (!n) return;
  $('#netConn').textContent = n.connection || 'Unknown';
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
  if (w && w.note) { note.textContent = w.note; note.classList.remove('hidden'); } else note.classList.add('hidden');
}

// ---- progress -----------------------------------------------------------
function resetRun() {
  state.result = null;
  state.pingLive = [];
  $('#checks').innerHTML = '';
  $$('.stage').forEach(s => { s.classList.remove('on', 'err'); s.querySelector('i').style.width = '0'; s.querySelector('em').textContent = '0%'; });
  for (const id of ['download', 'upload', 'ping', 'jitter', 'loss']) setMetric(id, '—', '');
  $('#qualityLevel').textContent = '—';
  $('#qualityLevel').className = '';
  $('#qualityProfile').textContent = '';
  ['#failureCard', '#summaryCard', '#detailsCard'].forEach(s => $(s).classList.add('hidden'));
  $('#progressCard').classList.remove('hidden');
}

const checkIcon = { ok: '✓', warn: '!', fail: '✗', pending: '…' };
function renderCheck(c) {
  let row = $(`#checks [data-name="${CSS.escape(c.name)}"]`);
  if (!row) {
    row = document.createElement('div');
    row.dataset.name = c.name;
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
  $('#cancelBtn').disabled = false;
  $$('#modeSeg button').forEach(b => { b.disabled = on; });
  $('#historyBtn').disabled = on;
  if (!on) activeMetric('');
}

const stageLabel = { connection: 'Checking connection…', ping: 'Ping test', download: 'Download', upload: 'Upload' };

function onTestEvent(ev) {
  switch (ev.type) {
    case 'stage':
      $$('.stage').forEach(s => s.classList.toggle('on', s.dataset.stage === ev.stage));
      activeMetric(ev.stage === 'ping' ? 'ping' : ev.stage);
      setGauge(stageLabel[ev.stage] || '', '—', '', 0);
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
        state.pingLive.push(ev.rttMs);
        setGauge('Ping', ms(ev.rttMs), 'ms', 0);
        setMetric('ping', ms(ev.rttMs), 'live');
      } else if (ev.stage === 'download' || ev.stage === 'upload') {
        setGauge(stageLabel[ev.stage], speed(ev.mbps || 0), state.unit, ev.mbps || 0);
        setMetric(ev.stage, speed(ev.mbps || 0), 'measuring…');
      }
      break;
    case 'ping':
      break;
    case 'speed': {
      const s = ev.speed;
      stageProgress(ev.stage, 1);
      if (s.error && !(s.mbps > 0)) {
        $(`.stage[data-stage="${ev.stage}"]`).classList.add('err');
        setMetric(ev.stage, '—', 'failed');
      } else {
        setMetric(ev.stage, speed(s.mbps), s.loadedSamples ? `ping under load ${ms(s.loadedPingMs)} ms` : '');
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
function renderResult(r) {
  state.result = r;
  if (r.network) renderNetwork(r.network);
  (r.checks || []).forEach(renderCheck);

  const dl = r.download, ul = r.upload;
  const sp = s => (s && s.mbps > 0 ? speed(s.mbps) : '—');
  setMetric('download', sp(dl), dl && dl.error ? (dl.mbps > 0 ? 'incomplete' : 'failed') : (dl && dl.loadedSamples ? `ping under load ${ms(dl.loadedPingMs)} ms` : ''));
  setMetric('upload', sp(ul), ul && ul.error ? (ul.mbps > 0 ? 'incomplete' : 'failed') : (ul && ul.loadedSamples ? `ping under load ${ms(ul.loadedPingMs)} ms` : ''));
  if (r.pingMs > 0) {
    const primary = (r.pings || []).find(p => p.primaryTarget);
    setMetric('ping', ms(r.pingMs), primary ? `min ${ms(primary.minMs)} · max ${ms(primary.maxMs)}` : '');
    setMetric('jitter', ms(r.jitterMs), 'latency stability');
    const sent = (r.pings || []).filter(p => !p.isGateway && p.received > 0).reduce((a, p) => a + p.sent, 0);
    const recv = (r.pings || []).filter(p => !p.isGateway && p.received > 0).reduce((a, p) => a + p.received, 0);
    setMetric('loss', pct(r.packetLossPct), sent ? `${recv} of ${sent} received` : '');
  } else {
    setMetric('ping', '—', r.cancelled || r.failure ? '' : 'not available');
    setMetric('jitter', '—', '');
    setMetric('loss', '—', '');
  }

  if (r.cancelled) {
    setGauge('Cancelled', '—', '', 0);
    toast('Test cancelled.');
  } else if (r.failure && !r.quality) {
    setGauge('Failed', '—', '', 0);
  } else {
    setGauge('Download', sp(dl), dl && dl.mbps > 0 ? state.unit : '', dl ? dl.mbps : 0);
  }

  const q = r.quality;
  const ql = $('#qualityLevel');
  ql.textContent = q ? q.level : '—';
  ql.className = q ? 'q-' + q.level : '';
  $('#qualityTitle').textContent = r.mode === 'gaming' ? 'Gaming Connection' : 'Connection Quality';
  $('#qualityProfile').textContent = q ? (q.profile === 'gaming' ? 'Gaming profile' : 'General profile') : '';

  const f = r.failure;
  $('#failureCard').classList.toggle('hidden', !f);
  if (f) {
    $('#failTitle').textContent = f.title;
    $('#failReason').textContent = f.reason;
    $('#failSuggestion').textContent = f.suggestion;
    $('#failTech').textContent = f.technical || '—';
  }

  const showSummary = !!q;
  $('#summaryCard').classList.toggle('hidden', !showSummary);
  if (showSummary) {
    $('#summaryText').textContent = q.summary;
    $('#notes').innerHTML = (q.notes || []).map(n => `<li>${esc(n)}</li>`).join('');
    renderGaming(r);
  }
  renderDetails(r);
}

function renderGaming(r) {
  const panel = $('#gamingPanel');
  const q = r.quality;
  if (r.mode !== 'gaming' || !q) { panel.classList.add('hidden'); return; }
  panel.classList.remove('hidden');
  const good = g => g === 'EXCELLENT' || g === 'GOOD';
  const items = q.grades.filter(g => ['Ping', 'Jitter', 'Packet Loss', 'Latency Under Load'].includes(g.metric));
  $('#gameList').innerHTML = items.map(g => {
    const label = g.metric === 'Latency Under Load' ? 'Latency Stability' : g.metric;
    const val = g.metric === 'Latency Under Load' ? `${g.grade.charAt(0) + g.grade.slice(1).toLowerCase()} (${g.value})` : g.value;
    return `<li><span class="k">${esc(label)}</span><span class="v">${esc(val)} <span class="${good(g.grade) ? 'ok' : 'bad'}">${good(g.grade) ? '✓' : '✗'}</span></span></li>`;
  }).join('');
}

function renderDetails(r) {
  $('#pingTable tbody').innerHTML = (r.pings || []).map(p => {
    if (!p.sent) return `<tr><td>${esc(p.label)} <span class="muted">${esc(p.target)}</span></td><td colspan="7" class="muted">${esc(p.error || 'not measured')}</td></tr>`;
    const v = x => (p.received ? ms(x) + ' ms' : '—');
    if (!p.received) return `<tr><td>${esc(p.label)} <span class="muted">${esc(p.target)}</span></td><td>${p.sent}</td><td>0</td><td>${pct(p.packetLossPct)}%</td><td colspan="4" class="muted">No replies — the target may block ping. Not counted in the headline loss.</td></tr>`;
    return `<tr><td>${esc(p.label)} <span class="muted">${esc(p.target)}</span>${p.primaryTarget ? ' ★' : ''}</td><td>${p.sent}</td><td>${p.received}</td><td>${pct(p.packetLossPct)}%</td><td>${v(p.minMs)}</td><td>${v(p.avgMs)}</td><td>${v(p.maxMs)}</td><td>${v(p.jitterMs)}</td></tr>`;
  }).join('') || '<tr><td colspan="8" class="muted">Ping was not measured.</td></tr>';

  const dl = r.download, ul = r.upload;
  const hasLoaded = (dl && dl.loadedSamples) || (ul && ul.loadedSamples);
  $('#loadedBox').classList.toggle('hidden', !hasLoaded || !(r.pingMs > 0));
  if (hasLoaded) {
    $('#ldIdle').textContent = ms(r.pingMs) + ' ms';
    $('#ldDown').textContent = dl && dl.loadedSamples ? ms(dl.loadedPingMs) + ' ms' : '—';
    $('#ldUp').textContent = ul && ul.loadedSamples ? ms(ul.loadedPingMs) + ' ms' : '—';
  }
  drawChart($('#chartDown'), dl);
  drawChart($('#chartUp'), ul);

  const n = r.network || {};
  const w = n.wifi || {};
  const rows = [
    ['Connection', n.connection], ['Adapter', n.adapterName], ['Description', n.description],
    ['IPv4', (n.ipv4 || []).join(', ')], ['IPv6', (n.ipv6 || []).join(', ')], ['Gateway', n.gateway], ['DNS', (n.dns || []).join(', ')],
    ['SSID', w.ssid], ['BSSID', w.bssid], ['Signal', w.signalPercent ? w.signalPercent + '%' + (w.rssi ? ` (${w.rssi} dBm)` : '') : ''],
    ['Band', w.band], ['Channel', w.channel || ''], ['Frequency', w.frequencyMHz ? w.frequencyMHz + ' MHz' : ''], ['Wi-Fi standard', w.phyType],
    ['Wi-Fi link speed', w.rxRateMbps ? `${Math.round(w.rxRateMbps)} Mbps receive / ${Math.round(w.txRateMbps)} Mbps transmit (radio rate, not internet speed)` : ''],
  ].filter(([, v]) => v);
  $('#netDetails').innerHTML = rows.map(([k, v]) => `<dt>${esc(k)}</dt><dd>${esc(v)}</dd>`).join('');
  const server = (dl && dl.server) || (ul && ul.server);
  const streams = (dl && dl.streams) || (ul && ul.streams);
  $('#serverLine').textContent = server ? `Test server: ${server} · ${streams} parallel connections · started ${new Date(r.startedAt).toLocaleString()}` : '';
}

function drawChart(canvas, s) {
  const ctx = canvas.getContext('2d');
  const dpr = window.devicePixelRatio || 1;
  const W = canvas.clientWidth || 420, H = canvas.clientHeight || 110;
  canvas.width = W * dpr; canvas.height = H * dpr;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.clearRect(0, 0, W, H);
  const css = getComputedStyle(document.documentElement);
  const data = (s && s.samplesMbps) || [];
  ctx.fillStyle = css.getPropertyValue('--muted');
  ctx.font = '11px Segoe UI, system-ui, sans-serif';
  if (data.length < 2) { ctx.fillText('No data', 10, H / 2); return; }
  const max = Math.max(...data) * 1.1 || 1;
  const pad = 6;
  const x = i => pad + (i / (data.length - 1)) * (W - 2 * pad);
  const y = v => H - pad - (v / max) * (H - 2 * pad - 12);
  const accent = css.getPropertyValue('--accent').trim();
  ctx.beginPath();
  data.forEach((v, i) => (i ? ctx.lineTo(x(i), y(v)) : ctx.moveTo(x(i), y(v))));
  ctx.strokeStyle = accent; ctx.lineWidth = 2; ctx.stroke();
  ctx.lineTo(x(data.length - 1), H - pad); ctx.lineTo(x(0), H - pad); ctx.closePath();
  ctx.globalAlpha = 0.12; ctx.fillStyle = accent; ctx.fill(); ctx.globalAlpha = 1;
  ctx.fillStyle = css.getPropertyValue('--muted');
  ctx.fillText(`peak ${speed(Math.max(...data))} ${state.unit}`, 10, 14);
  if (s.mbps > 0) {
    const ry = y(s.mbps);
    ctx.setLineDash([4, 4]); ctx.strokeStyle = css.getPropertyValue('--muted');
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
  g.fillText('INTERNET TEST', 32, 44);
  g.fillStyle = '#9aa7b8'; g.font = '13px Segoe UI, system-ui, sans-serif';
  const conn = r.network ? r.network.connection : '';
  g.fillText(`${new Date(r.startedAt).toLocaleString()}${conn ? '  ·  ' + conn : ''}`, 32, 66);
  const big = (label, val, unit, x, yy) => {
    g.fillStyle = '#9aa7b8'; g.font = '600 12px Segoe UI, system-ui, sans-serif'; g.fillText(label, x, yy);
    g.fillStyle = '#ffffff'; g.font = '700 40px Segoe UI, system-ui, sans-serif'; g.fillText(val, x, yy + 44);
    const w = g.measureText(val).width;
    g.fillStyle = '#9aa7b8'; g.font = '600 15px Segoe UI, system-ui, sans-serif'; g.fillText(unit, x + w + 6, yy + 44);
  };
  const sp = s => (s && s.mbps > 0 ? speed(s.mbps) : '—');
  big('↓ DOWNLOAD', sp(r.download), state.unit, 32, 110);
  big('↑ UPLOAD', sp(r.upload), state.unit, 230, 110);
  big('● PING', r.pingMs > 0 ? ms(r.pingMs) : '—', 'ms', 428, 110);
  g.fillStyle = '#c9d4e2'; g.font = '15px Segoe UI, system-ui, sans-serif';
  g.fillText(`Jitter ${r.pingMs > 0 ? ms(r.jitterMs) + ' ms' : '—'}     Loss ${r.pingMs > 0 ? pct(r.packetLossPct) + '%' : '—'}`, 32, 210);
  const q = r.quality ? r.quality.level : 'UNKNOWN';
  const colors = { EXCELLENT: '#3fb950', GOOD: '#58a6ff', FAIR: '#d29922', POOR: '#f85149', UNSTABLE: '#a371f7', UNKNOWN: '#8d96a0' };
  g.fillStyle = 'rgba(255,255,255,.08)'; g.beginPath(); g.roundRect(32, 236, W - 64, 56, 12); g.fill();
  g.fillStyle = '#9aa7b8'; g.font = '600 13px Segoe UI, system-ui, sans-serif'; g.fillText(r.mode === 'gaming' ? 'GAMING CONNECTION' : 'CONNECTION', 52, 270);
  g.fillStyle = colors[q] || '#fff'; g.font = '800 26px Segoe UI, system-ui, sans-serif';
  g.fillText(q, W - 52 - g.measureText(q).width, 274);
  g.fillStyle = '#6f7d8f'; g.font = '11px Segoe UI, system-ui, sans-serif';
  g.fillText('WiFi Speed + Ping Tester · measured values, not a guaranteed ISP speed', 32, 318);
  return c.toDataURL('image/png');
}

// ---- actions --------------------------------------------------------------
async function start() {
  resetRun();
  setRunning(true);
  setGauge('Starting…', '—', '', 0);
  try {
    await bridge.call('startTest', state.mode);
  } catch (e) {
    setRunning(false);
    toast(e.message);
  }
}

$('#startBtn').onclick = start;
$('#cancelBtn').onclick = () => {
  $('#cancelBtn').disabled = true;
  setGauge('Cancelling…', '—', '', 0);
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
  if (state.result) renderResult(state.result);
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
  let list = [];
  try { list = await bridge.call('history') || []; } catch (e) { toast(e.message); }
  $('#histEmpty').classList.toggle('hidden', list.length > 0);
  $('#histTable').classList.toggle('hidden', list.length === 0);
  $('#histTable tbody').innerHTML = list.map(h => `
    <tr data-id="${esc(h.id)}">
      <td>${esc(new Date(h.date).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' }))}${h.mode === 'gaming' ? ' <span class="muted">(gaming)</span>' : ''}</td>
      <td>${esc(h.connection)}</td>
      <td>${h.download ? speed(h.download) : '—'} / ${h.upload ? speed(h.upload) : '—'} ${state.unit}</td>
      <td>${h.ping ? ms(h.ping) + ' ms' : '—'}</td>
      <td>${h.ping ? ms(h.jitter) + ' ms' : '—'}</td>
      <td>${h.ping ? pct(h.loss) + '%' : '—'}</td>
      <td class="q-${esc(h.quality || 'UNKNOWN')}">${esc(h.quality || '—')}</td>
    </tr>`).join('');
  $$('#histTable tbody tr').forEach(tr => tr.onclick = async () => {
    try {
      const r = await bridge.call('historyItem', tr.dataset.id);
      if (!r) return;
      hideHistory();
      resetRun();
      $('#progressCard').classList.add('hidden');
      renderResult(r);
    } catch (e) { toast(e.message); }
  });
  if (state.info) $('#histPath').textContent = 'Stored locally at ' + state.info.historyPath;
}
function hideHistory() {
  $('#historyView').classList.add('hidden');
  $('#dashboard').classList.remove('hidden');
}
$('#historyBtn').onclick = showHistory;
$('#histBack').onclick = hideHistory;
$('#histClear').onclick = async () => {
  if (!confirm('Delete all saved test results?')) return;
  try { await bridge.call('clearHistory'); showHistory(); } catch (e) { toast(e.message); }
};
$('#histCsv').onclick = () => bridge.call('export', { format: 'history-csv', unit: state.unit, includeSensitive: $('#includeSensitive').checked }).catch(e => toast(e.message));

// ---- about ------------------------------------------------------------------
$('#aboutBtn').onclick = () => {
  const i = state.info || {};
  $('#aboutVersion').textContent = i.version ? 'v' + i.version : '';
  $('#aboutTraffic').innerHTML = [
    `Speed test data to: ${esc((i.servers || []).join(', '))}`,
    `Ping (ICMP echo) to: ${esc((i.pingTargets || []).join(', '))}`,
    'A DNS lookup of the test server name',
  ].map(s => `<li>${s}</li>`).join('');
  $('#aboutHistory').textContent = i.historyPath || '';
  $('#aboutCfgName').textContent = i.configFile || '';
  $('#aboutCfg').textContent = i.configError ? 'Config error (defaults in use): ' + i.configError
    : i.configPath ? 'Loaded: ' + i.configPath : 'No config file found — using built-in defaults.';
  $('#aboutDlg').showModal();
};

// ---- events -----------------------------------------------------------------
bridge.listen((name, payload) => {
  if (name === 'test') onTestEvent(payload);
  else if (name === 'network') {
    if (payload.network) renderNetwork(payload.network);
    else if (!state.network) { $('#netConn').textContent = 'Not connected'; $('#netName').textContent = payload.error || '—'; }
  } else if (name === 'saved') {
    if (payload.error) toast('Export failed: ' + payload.error, 6000);
    else if (payload.path) toast('Saved to ' + payload.path, 6000);
  }
});

// ---- init ---------------------------------------------------------------------
(async function init() {
  try { Object.assign(prefs, await bridge.call('getPrefs')); } catch { /* defaults */ }
  if (prefs.unit === 'MB/s' || prefs.unit === 'Mbps') state.unit = prefs.unit;
  if (prefs.mode === 'gaming' || prefs.mode === 'general') state.mode = prefs.mode;
  $$('#modeSeg button').forEach(x => x.classList.toggle('on', x.dataset.mode === state.mode));
  $$('#unitSeg button').forEach(x => x.classList.toggle('on', x.dataset.unit === state.unit));
  unitLabels();
  try {
    state.info = await bridge.call('info');
    if (state.info.configError) toast('Config file problem — using defaults: ' + state.info.configError, 8000);
    const servers = (state.info.servers || []).join(', ');
    $('#privacyLine').textContent = `Tests send traffic only to the speed-test server (${servers}) and the ping targets. Nothing else leaves this computer.`;
  } catch (e) { toast(e.message); }
  bridge.call('detectNetwork').catch(() => {});
})();

window.addEventListener('keydown', e => {
  if (e.key === 'Escape' && state.running) $('#cancelBtn').click();
});
