'use strict';
// Drives the real WiFiSpeedTester.exe on Windows through WebView2's DevTools
// port (start it with WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS set to
// --remote-debugging-port=9222). Unlike ui-tests/run.cjs this covers the
// actual window, the Go bridge and the Windows code paths. Nothing here opens
// a Save dialog, which would wait for a click. Used by CI (build.yml,
// ui-windows); hosted runners have no Wi-Fi, so the signal finder is
// expected to report that.
const fs = require('fs');
const path = require('path');
const { chromium } = require('playwright');

const SHOTS = process.env.UI_SHOTS || 'ui-windows-shots';
let failures = 0;
const check = (ok, what) => {
  console.log((ok ? '  ok   ' : '  FAIL ') + what);
  if (!ok) failures++;
};
const sleep = ms => new Promise(r => setTimeout(r, ms));
async function eventually(fn, timeout = 10000) {
  const end = Date.now() + timeout;
  for (;;) {
    try { if (await fn()) return true; } catch { /* not ready yet */ }
    if (Date.now() > end) return false;
    await sleep(200);
  }
}

async function main() {
  fs.mkdirSync(SHOTS, { recursive: true });
  let browser = null;
  let lastErr = null;
  for (let i = 0; i < 60 && !browser; i++) {
    // Chromium listens on 127.0.0.1; [::1] covers a resolver that prefers IPv6.
    for (const host of ['127.0.0.1', '[::1]']) {
      try { browser = await chromium.connectOverCDP(`http://${host}:9222`, { timeout: 5000 }); break; } catch (e) { lastErr = e; }
    }
    if (!browser) await sleep(1000);
  }
  if (!browser) throw new Error('the app\'s WebView2 did not open its DevTools port; last error: ' + (lastErr && lastErr.message.split('\n')[0]));
  let page = null;
  await eventually(async () => {
    for (const ctx of browser.contexts()) {
      for (const p of ctx.pages()) {
        if (await p.evaluate(() => typeof state !== 'undefined' && !!state.info)) { page = p; return true; }
      }
    }
    return false;
  }, 30000);
  if (!page) throw new Error('the app page did not load');
  const errors = [];
  page.on('pageerror', e => errors.push(e.message));
  page.on('console', m => { if (m.type() === 'error') errors.push(m.text()); });
  page.setDefaultTimeout(20000);
  const shot = name => page.screenshot({ path: path.join(SHOTS, name + '.png') });

  console.log('window and bridge');
  const info = await page.evaluate(() => state.info);
  console.log('  version', info.version, '| page', page.url());
  check(!!info.version && info.canOpenUrl === true, 'the Go bridge answers (version, browser host)');
  check(await page.evaluate(() => bridge.call('setTitleBar', true).then(() => true)), 'dark title bar call succeeds');
  check(await page.evaluate(() => bridge.call('setTitleBar', false).then(() => true)), 'light title bar call succeeds');
  check(await page.evaluate(() => bridge.call('attention').then(() => true)), 'taskbar flash call succeeds');
  check(await eventually(async () => !/Detecting/.test(await page.textContent('#netConn')), 20000), 'network detected: ' + await page.textContent('#netConn'));
  await shot('start');

  console.log('tabs and theme');
  for (const [key, tab, view] of [['2', '#tabDiag', '#diagView'], ['3', '#tabHistory', '#historyView'], ['4', '#tabTools', '#toolsView'], ['1', '#tabTest', '#dashboard']]) {
    await page.keyboard.press('Control+' + key);
    check(await eventually(async () => await page.getAttribute(tab, 'aria-selected') === 'true' && await page.isVisible(view)), `Ctrl+${key} opens ${tab}`);
  }
  await page.click('#settingsBtn');
  await page.selectOption('#themeSelect', 'dark');
  check(await page.evaluate(() => getComputedStyle(document.body).backgroundColor) === 'rgb(14, 17, 22)', 'Dark theme');
  await shot('dark');
  await page.selectOption('#themeSelect', 'system');
  await page.keyboard.press('Escape');

  console.log('quick ping (real network)');
  await page.click('#quickBtn');
  await page.waitForSelector('#startBtn:not(.hidden)', { timeout: 60000 });
  const quickOk = await page.isVisible('#summaryCard') || await page.isVisible('#failureCard');
  check(quickOk, 'Quick Ping finishes with a result or an explained error: ' + (await page.textContent('#qualityLevel')));
  await shot('quick');

  console.log('tools');
  await page.click('#tabTools');
  await page.click('[data-tool="signalCard"]');
  await page.click('#sigStart');
  check(await eventually(async () => (await page.textContent('#sigError')).length > 0 || await page.isVisible('#sigLive'), 8000),
    'signal finder reads Windows Wi-Fi: ' + ((await page.textContent('#sigError')) || 'live reading'));
  await page.click('[data-tool="regionCard"]');
  check(await eventually(() => page.evaluate(() => !sig.running)), 'leaving the finder stops it');
  await page.click('#regStart');
  await page.waitForFunction(() => !reg.running, null, { timeout: 60000 });
  const regs = await page.evaluate(() => reg.results.map(r => `${r.label} ${r.ok > 0 ? Math.round(r.medianMs) + ' ms' : 'no answer'}`));
  console.log('  ' + regs.join(' | '));
  check(regs.length === 9 && regs.some(r => /ms$/.test(r)), 'ping to regions answers from the cloud regions');
  await shot('regions');

  console.log('report a problem');
  await page.click('#aboutBtn');
  await page.click('#aboutProblem');
  await page.waitForSelector('#problemDlg[open]');
  const report = await page.textContent('#problemInfo');
  console.log(report.split('\n').map(l => '  | ' + l).join('\n'));
  check(/System: Windows 1[01] \(10\.0\.\d+\)/.test(report), 'the report names the Windows release');
  await page.keyboard.press('Escape');

  check(errors.length === 0, 'no page errors ' + errors.join(' | '));
  console.log(failures ? `\n${failures} check(s) failed` : '\nall Windows UI checks passed');
  await browser.close().catch(() => {});
  process.exit(failures ? 1 : 0);
}

main().catch(e => { console.error(e); process.exit(1); });
