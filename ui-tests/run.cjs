'use strict';
// End-to-end check of the embedded UI against the real Go API, served by
// `wifispeedtester --serve` with the local fake speed-test server. Run it
// through scripts/ui-test.sh. Any page error, missing translation or broken
// flow fails the run.
const fs = require('fs');
const path = require('path');
const { chromium } = require('playwright');

const BASE = process.env.UI_URL || 'http://127.0.0.1:18088/';
const SHOTS = process.env.UI_SHOTS || '';
let failures = 0;

function check(ok, what) {
  console.log((ok ? '  ok   ' : '  FAIL ') + what);
  if (!ok) failures++;
}

async function shot(page, name) {
  if (SHOTS) await page.screenshot({ path: path.join(SHOTS, name + '.png'), fullPage: true });
}

// Runs action and waits for a toast whose text matches re.
async function toastAfter(page, action, re, timeout = 15000) {
  await page.evaluate(() => { const t = document.querySelector('#toast'); t.textContent = ''; t.classList.add('hidden'); });
  await action();
  await page.waitForFunction(src => {
    const t = document.querySelector('#toast');
    return !t.classList.contains('hidden') && new RegExp(src).test(t.textContent);
  }, re.source, { timeout }).catch(() => {});
  return page.textContent('#toast');
}

// Every key the page uses exists in both languages, and the two
// dictionaries have the same keys.
async function checkDictionaries(page) {
  const r = await page.evaluate(() => {
    const used = new Set();
    for (const el of document.querySelectorAll('[data-i18n],[data-i18n-title],[data-i18n-aria]')) {
      for (const k of [el.dataset.i18n, el.dataset.i18nTitle, el.dataset.i18nAria]) if (k) used.add(k);
    }
    const en = Object.keys(TEXT.en), id = Object.keys(TEXT.id);
    return {
      missingEn: [...used].filter(k => !(k in TEXT.en)),
      missingId: [...used].filter(k => !(k in TEXT.id)),
      onlyEn: en.filter(k => !(k in TEXT.id)),
      onlyId: id.filter(k => !(k in TEXT.en)),
    };
  });
  check(r.missingEn.length === 0, 'every label has English text ' + r.missingEn.join(' '));
  check(r.missingId.length === 0, 'every label has Indonesian text ' + r.missingId.join(' '));
  check(r.onlyEn.length === 0 && r.onlyId.length === 0, 'EN and ID dictionaries match ' + [...r.onlyEn, ...r.onlyId].join(' '));
}

// No element shows a raw translation key such as "mon.title".
async function checkNoRawKeys(page, where) {
  const raw = await page.evaluate(() => [...document.querySelectorAll('body *')]
    .filter(el => el.offsetParent !== null && el.children.length === 0)
    .map(el => el.textContent.trim())
    .filter(s => /^[a-z]+(\.[A-Za-z0-9]+)+$/.test(s) && (s in TEXT.en)));
  check(raw.length === 0, `no untranslated keys on ${where} ${raw.join(' ')}`);
}

async function noHorizontalScroll(page, where) {
  const over = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
  check(over <= 1, `${where} fits the window width (overflow ${over}px)`);
}

const visible = (page, sel) => page.isVisible(sel);

async function main() {
  if (SHOTS) fs.mkdirSync(SHOTS, { recursive: true });
  const browser = await chromium.launch();
  const errors = [];
  const watch = page => {
    page.on('pageerror', e => errors.push(e.message));
    page.on('console', m => { if (m.type() === 'error') errors.push(m.text()); });
  };

  // ---- light theme, English: every flow --------------------------------------
  const ctx = await browser.newContext({ viewport: { width: 1100, height: 900 }, colorScheme: 'light', locale: 'en-US' });
  await ctx.grantPermissions(['clipboard-read', 'clipboard-write'], { origin: BASE });
  const page = await ctx.newPage();
  watch(page);
  page.setDefaultTimeout(20000);
  await page.goto(BASE);
  await page.waitForFunction(() => document.querySelector('#netConn').textContent !== TEXT.en['net.detecting']);
  console.log('dashboard');
  await checkDictionaries(page);
  await checkNoRawKeys(page, 'dashboard');

  console.log('settings: plan and custom targets');
  await page.click('#settingsBtn');
  await page.fill('#planDown', '50');
  await page.dispatchEvent('#planDown', 'change');
  await page.fill('#planUp', '10');
  await page.dispatchEvent('#planUp', 'change');
  await page.click('#targetAdd');
  await page.fill('#targetRows .target-row:last-child .t-label', 'Game server');
  await page.fill('#targetRows .target-row:last-child .t-host', 'not a host!');
  await page.click('#targetSave');
  await page.waitForFunction(() => document.querySelector('#targetStatus').textContent !== '');
  check(await page.textContent('#targetStatus') !== 'Saved.', 'an invalid host is rejected');
  await page.fill('#targetRows .target-row:last-child .t-host', 'localhost');
  await page.click('#targetSave');
  await page.waitForFunction(() => document.querySelector('#targetStatus').textContent === 'Saved.');
  check(await page.locator('#targetRows .target-row').count() === 1, 'the saved target is listed');
  for (let i = 0; i < 6; i++) if (await page.isEnabled('#targetAdd')) await page.click('#targetAdd');
  check(await page.locator('#targetRows .target-row').count() === 5, 'at most 5 targets can be added');
  await shot(page, 'settings');
  await page.keyboard.press('Escape');
  const prefs = await page.evaluate(() => bridge.call('getPrefs'));
  check(prefs.planDown === '50' && prefs.planUp === '10', 'plan saved in settings');

  console.log('quick ping');
  await page.click('#quickBtn');
  await page.waitForSelector('#summaryCard:not(.hidden), #failureCard:not(.hidden)', { timeout: 30000 });
  await page.waitForSelector('#startBtn:not(.hidden)');
  check(!(await page.textContent('#qualityLevel')).includes('—'), 'quick ping gives a rating');
  check((await page.textContent('#m-download .sub')) === 'not part of Quick Ping', 'quick ping skips download');

  console.log('full test');
  await page.click('#startBtn');
  check(await visible(page, '#cancelBtn'), 'cancel is shown while testing');
  check(await page.isDisabled('#toolsBtn'), 'tools are locked while testing');
  await page.waitForSelector('#summaryCard:not(.hidden)', { timeout: 60000 });
  await page.waitForSelector('#startBtn:not(.hidden)');
  check(/% of plan/.test(await page.textContent('#m-download .sub')), 'download shows the share of the plan');
  await page.click('#detailsBtn');
  check((await page.textContent('#pingTable')).includes('Game server'), 'the custom target was pinged');
  await shot(page, 'result');
  for (const fmt of ['txt', 'json', 'csv', 'png']) {
    const msg = await toastAfter(page, () => page.click(`[data-export="${fmt}"]`), /^Saved to /);
    check(/^Saved to /.test(msg), `export ${fmt}`);
  }
  const copied = await toastAfter(page, () => page.click('#copyBtn'), /copied/);
  check(/copied/.test(copied), 'copy result');

  console.log('history');
  await page.click('#historyBtn');
  await page.waitForSelector('#histTable tbody tr');
  check(await page.locator('#histTable tbody tr').count() === 2, 'both tests are in the history');
  check(await page.isVisible('#histTable thead .plan-col'), 'the plan column is shown');
  // Ping has two points (quick + full test); speed only one.
  const plot = await page.locator('#plotPing').boundingBox();
  await page.mouse.move(plot.x + plot.width - 70, plot.y + plot.height / 2);
  check(await page.isVisible('#plotPing .tip'), 'trend chart tooltip on hover');
  await page.click('#histRange [data-range="7"]');
  check(await page.locator('#histTable tbody tr').count() === 2, '7-day filter keeps today\'s tests');
  const rep = await toastAfter(page, () => page.click('#histReport'), /^Saved to /);
  const repPath = rep.replace(/^Saved to /, '').trim();
  const repText = fs.existsSync(repPath) ? fs.readFileSync(repPath, 'utf8') : '';
  check(repText.includes('INTERNET QUALITY REPORT') && /Plan:\s+50\.0 Mbps/.test(repText), 'ISP report written with the plan');
  check(/Tests:\s+1\n/.test(repText), 'ISP report counts the full test only, not Quick Ping');
  await checkNoRawKeys(page, 'history');
  await shot(page, 'history');
  await page.click('#histBack');

  console.log('tools');
  await page.click('#toolsBtn');
  await page.click('#scanStart');
  await page.waitForSelector('#scanResult:not(.hidden)');
  check(await page.locator('#netTable tbody tr').count() === 8, 'Wi-Fi scan lists the networks');
  check((await page.textContent('#netTable')).includes('(hidden network)'), 'hidden networks are labelled');
  check((await page.textContent('#chanTable')).includes('suggested'), 'a less crowded channel is suggested');
  check((await page.textContent('#scanAdvice')).includes('Channel 1'), 'scan advice names the channel');

  await page.click('#dnsStart');
  await page.waitForFunction(() => document.querySelector('#dnsAdvice').children.length > 0, null, { timeout: 70000 });
  check(await page.locator('#dnsTable tbody tr').count() >= 1, 'DNS test lists servers');

  const options = await page.locator('#monTarget option').allTextContents();
  check(options.some(o => o.includes('Game server')), 'the monitor can watch the custom target');
  await page.click('#monStart');
  await page.waitForFunction(() => mon.samples.length >= 3, null, { timeout: 15000 });
  check(await page.isDisabled('#scanStart'), 'Wi-Fi scan is locked during the monitor');
  check(/\d+:\d\d \/ 5:00/.test(await page.textContent('#monElapsed')), 'monitor shows elapsed time');
  await page.click('#toolsBack');
  check(await page.isDisabled('#startBtn'), 'speed test is locked during the monitor');
  const busy = await page.evaluate(() => bridge.call('startTest', 'quick').then(() => 'started', e => e.message));
  check(busy !== 'started', 'the engine also refuses a test during the monitor');
  await page.click('#toolsBtn');
  await shot(page, 'tools-live');
  await page.click('#monStop');
  await page.waitForSelector('#monResult:not(.hidden)', { timeout: 10000 });
  check((await page.textContent('#monVerdict')).length > 10, 'monitor gives a verdict');
  check(await page.isEnabled('#monStart') && await page.isEnabled('#scanStart'), 'tools unlock after the monitor');
  const monExp = await toastAfter(page, () => page.click('#monExport'), /^Saved to /);
  check(/^Saved to /.test(monExp), 'monitor report export');
  await checkNoRawKeys(page, 'tools');
  await shot(page, 'tools');
  await page.click('#toolsBack');
  check(await page.isEnabled('#startBtn'), 'speed test unlocks after the monitor');

  console.log('about');
  await page.click('#aboutBtn');
  await page.waitForSelector('#aboutDlg[open]');
  const about = await page.textContent('#aboutTraffic');
  check(about.includes('9.9.9.9') && about.includes('Game server'), 'About lists DNS-test servers and custom targets');
  await page.keyboard.press('Escape');

  console.log('Indonesian');
  await page.click('#settingsBtn');
  await page.selectOption('#langSelect', 'id');
  await page.keyboard.press('Escape');
  // #netConn and #gaugeLabel start with a placeholder and then show live data.
  const wrong = await page.evaluate(() => [...document.querySelectorAll('[data-i18n]:not(#netConn):not(#gaugeLabel)')]
    .filter(el => el.textContent !== TEXT.id[el.dataset.i18n]).map(el => el.dataset.i18n));
  check(wrong.length === 0, 'every label switched to Indonesian ' + wrong.join(' '));
  check(/% dari paket/.test(await page.textContent('#m-download .sub')), 'plan share in Indonesian');
  // Engine texts follow the language of the run.
  await page.click('#quickBtn');
  await page.waitForSelector('#startBtn:not(.hidden)', { timeout: 30000 });
  await page.waitForSelector('#summaryCard:not(.hidden)');
  check((await page.textContent('#pingTable')).includes('Gateway Lokal'), 'ping target labels in Indonesian');
  await page.click('#toolsBtn');
  await checkNoRawKeys(page, 'tools (id)');
  await shot(page, 'tools-id');
  await ctx.close();

  // ---- dark theme, narrow window: layout ---------------------------------------
  console.log('dark theme, narrow window');
  const dark = await browser.newContext({ viewport: { width: 720, height: 900 }, colorScheme: 'dark' });
  const dp = await dark.newPage();
  watch(dp);
  await dp.goto(BASE);
  await dp.waitForFunction(() => typeof state !== 'undefined' && state.info);
  check(await dp.evaluate(() => getComputedStyle(document.body).backgroundColor) === 'rgb(14, 17, 22)', 'dark theme applied');
  await noHorizontalScroll(dp, 'dashboard');
  await dp.click('#historyBtn');
  await dp.waitForSelector('#histTable tbody tr');
  await noHorizontalScroll(dp, 'history');
  await shot(dp, 'history-dark');
  await dp.click('#histBack');
  await dp.click('#toolsBtn');
  await dp.click('#scanStart');
  await dp.waitForSelector('#scanResult:not(.hidden)');
  await noHorizontalScroll(dp, 'tools');
  await shot(dp, 'tools-dark');
  await dark.close();
  await browser.close();

  check(errors.length === 0, 'no page errors ' + errors.join(' | '));
  console.log(failures ? `\n${failures} check(s) failed` : '\nall UI checks passed');
  process.exit(failures ? 1 : 0);
}

main().catch(e => { console.error(e); process.exit(1); });
