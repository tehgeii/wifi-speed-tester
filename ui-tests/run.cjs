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

// Screenshots for the website (site/): the window, cropped below the last
// card so short views carry no empty space.
async function siteShot(page, name) {
  if (!SHOTS) return;
  const h = await page.evaluate(() => Math.ceil(Math.max(...[...document.querySelectorAll('main .card, section .card, .card')]
    .filter(el => el.offsetParent !== null).map(el => el.getBoundingClientRect().bottom)) + 16));
  const vp = page.viewportSize();
  await page.screenshot({ path: path.join(SHOTS, 'site-' + name + '.png'), clip: { x: 0, y: 0, width: vp.width, height: Math.min(vp.height, Math.max(300, h)) } });
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

// Polls fn until it returns true or the timeout passes. Some updates land
// after an awaited bridge call, so a single read can race them.
async function eventually(fn, timeout = 5000) {
  const end = Date.now() + timeout;
  for (;;) {
    if (await fn()) return true;
    if (Date.now() > end) return false;
    await new Promise(r => setTimeout(r, 100));
  }
}
const textMatches = (page, sel, re) => eventually(async () => re.test(await page.textContent(sel)));

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
  check(await eventually(async () => {
    const p = await page.evaluate(() => bridge.call('getPrefs'));
    return p.planDown === '50' && p.planUp === '10';
  }), 'plan saved in settings');

  console.log('quick ping');
  await page.click('#quickBtn');
  await page.waitForSelector('#summaryCard:not(.hidden), #failureCard:not(.hidden)', { timeout: 30000 });
  await page.waitForSelector('#startBtn:not(.hidden)');
  check(!(await page.textContent('#qualityLevel')).includes('—'), 'quick ping gives a rating');
  check((await page.textContent('#m-download .sub')) === 'not part of Quick Ping', 'quick ping skips download');
  check(await page.locator('#actList .act').count() === 3 && await page.isVisible('#actQuickNote'), 'quick ping judges only latency-based activities');
  check(!(await page.isVisible('#cmpLine')), 'the first test has nothing to compare with');

  console.log('full test');
  await page.click('#startBtn');
  check(await visible(page, '#cancelBtn'), 'cancel is shown while testing');
  check(await page.isDisabled('#tabTools') && await page.isDisabled('#tabHistory'), 'other tabs are locked while testing');
  await page.waitForSelector('#activityCard:not(.hidden) .act-pending', { timeout: 30000 });
  check(await page.isVisible('#cancelBtn'), 'activities appear while the test is still running');
  await page.waitForSelector('#summaryCard:not(.hidden)', { timeout: 60000 });
  await page.waitForSelector('#startBtn:not(.hidden)');
  check(await page.locator('#actList .act').count() === 7 && await page.locator('#actList .act-pending').count() === 0, 'all 7 activities are judged at the end');
  check(/50 GB/.test(await page.textContent('#actList [data-key="bigdl"] .ar')), 'large download shows an estimated time');
  // A maximised window fits all 7 cards in one row; each status must still
  // sit on its own line under the name, never squeezed beside it.
  for (const width of [1100, 1920]) {
    await page.setViewportSize({ width, height: 900 });
    const squeezed = await page.evaluate(() => [...document.querySelectorAll('#actList .act')].filter(c => {
      const n = c.querySelector('.an').getBoundingClientRect(), s = c.querySelector('.as').getBoundingClientRect();
      return s.top < n.bottom - 1;
    }).map(c => c.dataset.key));
    check(squeezed.length === 0, `activity status sits under the name at ${width} px ${squeezed.join(' ')}`);
    if (width === 1920) await shot(page, 'activities-wide');
  }
  await page.setViewportSize({ width: 1100, height: 900 });
  check(await page.isVisible('#gradeRow') && /^(A\+|A|B|C|D|F)$/.test(await page.textContent('#gradeBadge')), 'bufferbloat grade is shown');
  check(!(await page.isVisible('#cmpLine')), 'a full test is not compared with a Quick Ping');
  // The comparison rules, with fixed numbers.
  const parts = await page.evaluate(() => [
    compareParts({ mode: 'general', download: { mbps: 118 }, upload: { mbps: 9.5 }, pingMs: 30 }, { download: 100, upload: 10, ping: 18 }),
    compareParts({ mode: 'general', download: { mbps: 102 }, upload: { mbps: 7 }, pingMs: 20 }, { download: 100, upload: 10, ping: 30 }),
    compareParts({ mode: 'quick', pingMs: 41 }, { download: 0, upload: 0, ping: 40 }),
  ].map(ps => ps.map(p => p.cls + ':' + p.text).join(' | ')));
  check(parts[0] === 'better:Download +18% | worse:Upload −5% | worse:Ping +12 ms', 'faster, slower and higher ping are told apart: ' + parts[0]);
  check(parts[1] === 'same:Download about the same | worse:Upload −30% | better:Ping −10 ms', 'small changes count as the same, lower ping is better: ' + parts[1]);
  check(parts[2] === 'same:Ping about the same', 'Quick Ping compares only the ping: ' + parts[2]);
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

  console.log('network identifiers in exports');
  const savedText = async () => {
    const msg = await toastAfter(page, () => page.click('[data-export="txt"]'), /^Saved to /);
    const p = msg.replace(/^Saved to /, '').trim();
    return fs.existsSync(p) ? fs.readFileSync(p, 'utf8') : '';
  };
  const ip = await page.evaluate(() => ((state.result.network || {}).ipv4 || [])[0] || '');
  if (ip) {
    const masked = ip.split('.').slice(0, 3).join('.') + '.x';
    check((await page.textContent('#sensitiveHint')).includes(masked), 'the hint shows how the IP will appear');
    let txt = await savedText();
    check(txt.includes(masked) && !txt.includes(ip), 'TXT hides the IP address by default');
    await page.check('#includeSensitive');
    check(/only with people you trust/.test(await page.textContent('#sensitiveHint')), 'the hint warns when identifiers are included');
    txt = await savedText();
    check(txt.includes(ip), 'TXT contains the IP address when ticked');
    await page.uncheck('#includeSensitive');
    check((await page.textContent('#sensitiveHint')).includes(masked), 'unticking restores the masked hint');
  } else {
    console.log('  (no IPv4 address on this machine; masking checks skipped)');
  }

  console.log('history');
  await page.keyboard.press('Control+3');
  await page.waitForSelector('#histTable tbody tr');
  check(await page.getAttribute('#tabHistory', 'aria-selected') === 'true', 'Ctrl+3 opens the History tab');
  check(await page.locator('#histTable tbody tr').count() === 2, 'both tests are in the history');
  check(await page.locator('#histTable .grade.sm').count() >= 1, 'history shows the bufferbloat grade');
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
  const pdfMsg = await toastAfter(page, () => page.click('#histReportPdf'), /^Saved to .*\.pdf/);
  const pdfPath = pdfMsg.replace(/^Saved to /, '').trim();
  const pdf = fs.existsSync(pdfPath) ? fs.readFileSync(pdfPath) : Buffer.alloc(0);
  check(pdf.subarray(0, 8).toString() === '%PDF-1.4' && pdf.length > 2000 && pdf.subarray(-6).toString() === '%%EOF\n', 'ISP report as PDF');
  await checkNoRawKeys(page, 'history');
  await shot(page, 'history');
  await page.click('#tabTest');

  console.log('tools');
  await page.click('#tabTools');
  check(await page.isVisible('#monitorCard') && !(await page.isVisible('#dnsCard')), 'tools open one at a time');
  await page.click('[data-tool="wifiCard"]');
  await page.click('#scanStart');
  await page.waitForSelector('#scanResult:not(.hidden)');
  check(await page.locator('#netTable tbody tr').count() === 8, 'Wi-Fi scan lists the networks');
  check((await page.textContent('#netTable')).includes('(hidden network)'), 'hidden networks are labelled');
  check((await page.textContent('#chanTable')).includes('suggested'), 'a less crowded channel is suggested');
  check((await page.textContent('#scanAdvice')).includes('Channel 1'), 'scan advice names the channel');

  await page.click('[data-tool="dnsCard"]');
  await page.click('#dnsStart');
  await page.waitForFunction(() => document.querySelector('#dnsAdvice').children.length > 0, null, { timeout: 70000 });
  check(await page.locator('#dnsTable tbody tr').count() >= 1, 'DNS test lists servers');

  await page.click('[data-tool="monitorCard"]');
  const options = await page.locator('#monTarget option').allTextContents();
  check(options.some(o => o.includes('Game server')), 'the monitor can watch the custom target');
  await page.click('#monStart');
  await page.waitForFunction(() => mon.samples.length >= 3, null, { timeout: 15000 });
  check(await page.isDisabled('#scanStart'), 'Wi-Fi scan is locked during the monitor');
  check(/\d+:\d\d \/ 5:00/.test(await page.textContent('#monElapsed')), 'monitor shows elapsed time');
  await page.click('#tabTest');
  check(await page.isDisabled('#startBtn'), 'speed test is locked during the monitor');
  const busy = await page.evaluate(() => bridge.call('startTest', 'quick').then(() => 'started', e => e.message));
  check(busy !== 'started', 'the engine also refuses a test during the monitor');
  await page.click('#tabTools');
  check(await page.isVisible('#monitorCard'), 'the last tool stays selected');
  await shot(page, 'tools-live');
  await page.click('#monStop');
  await page.waitForSelector('#monResult:not(.hidden)', { timeout: 10000 });
  check((await page.textContent('#monVerdict')).length > 10, 'monitor gives a verdict');
  check(await page.isEnabled('#monStart') && await page.isEnabled('#scanStart'), 'tools unlock after the monitor');
  const monExp = await toastAfter(page, () => page.click('#monExport'), /^Saved to /);
  check(/^Saved to /.test(monExp), 'monitor report export');
  await checkNoRawKeys(page, 'tools');
  await shot(page, 'tools');
  await page.click('#tabTest');
  check(await page.isEnabled('#startBtn'), 'speed test unlocks after the monitor');

  console.log('signal finder');
  await page.click('#tabTools');
  await page.click('[data-tool="signalCard"]');
  check(!(await page.isVisible('#sigLive')) && await page.isDisabled('#spotSave'), 'nothing is read before Start');
  await page.click('#sigStart');
  await page.waitForFunction(() => sig.samples.length >= 3, null, { timeout: 10000 });
  check(/^\d+%$/.test(await page.textContent('#sigPct')) && /^(Strong|Fair|Weak)$/.test(await page.textContent('#sigWord')), 'live signal in % with a word');
  check((await page.textContent('#sigMeta')).includes('HomeNet') && /dBm/.test(await page.textContent('#sigDbm')), 'the finder shows the network and dBm');
  check(await page.isVisible('#sigStop') && !(await page.isVisible('#sigStart')), 'Stop replaces Start while reading');
  const t0 = await page.evaluate(() => sig.samples[0].t);
  const t2 = await page.evaluate(() => sig.samples[2].t);
  check(t2 - t0 > 1.5 && t2 - t0 < 3.5, 'one reading per second');
  await page.check('#sigBeep');
  await page.fill('#spotName', 'Bedroom');
  await page.click('#spotSave');
  await page.waitForSelector('#spotBox:not(.hidden)');
  await page.waitForTimeout(1200);
  await page.fill('#spotName', 'Kitchen');
  await page.press('#spotName', 'Enter');
  check(await eventually(async () => await page.locator('#spotTable tbody tr').count() === 2), 'spots are saved');
  const pcts = (await page.locator('#spotTable tbody td:nth-child(3) b').allTextContents()).map(s => parseInt(s, 10));
  check(pcts.length === 2 && pcts[0] >= pcts[1], 'spots are ranked best first');
  const noName = await toastAfter(page, () => page.click('#spotSave'), /name/);
  check(/Give the spot a name/.test(noName), 'a spot needs a name');
  await shot(page, 'signal');
  await page.click('[data-tool="regionCard"]');
  check(await eventually(() => page.evaluate(() => !sig.running)), 'leaving the finder stops it');
  await page.click('[data-tool="signalCard"]');
  check(await page.isVisible('#spotBox') && await page.isVisible('#sigStart'), 'saved spots stay; reading waits for Start');
  await page.click('#spotClear');
  check(await eventually(async () => !(await page.isVisible('#spotBox'))), 'spots can be cleared');

  console.log('ping to regions');
  await page.click('[data-tool="regionCard"]');
  await page.click('#regStart');
  await page.waitForFunction(() => !reg.running && reg.results.length === 3, null, { timeout: 30000 });
  const regRows = await page.locator('#regTable tbody tr').evaluateAll(rows => rows.map(r => r.dataset.label));
  check(regRows.length === 3 && regRows[2] === 'Closed port', 'every region is listed; one that did not answer goes last');
  check((await page.textContent('#regTable tbody tr:last-child')).includes('no answer'), 'a region without an answer says so');
  check(/Singapore .* great for online games/.test(await page.textContent('#regAdvice')), 'advice about Singapore for games');
  check(await page.isEnabled('#regStart'), 'Measure can be run again');
  await checkNoRawKeys(page, 'regions');
  await shot(page, 'regions');

  console.log('diagnosis');
  await page.click('#tabTest');
  const histN = (await page.evaluate(() => bridge.call('history'))).length;
  await page.keyboard.press('Control+2');
  check(await page.getAttribute('#tabDiag', 'aria-selected') === 'true', 'Ctrl+2 opens the Diagnosis tab');
  check(await page.getAttribute('#diagSteps [data-step="now"]', 'class') === 'run' && !(await page.isVisible('#diagSkip')), 'step 1 is active and cannot be skipped');
  await page.click('#diagGo');
  check(await page.isVisible('#diagCancel') && await page.isDisabled('#tabTest') && await page.isDisabled('#tabHistory') && await page.isDisabled('#settingsBtn'), 'other tabs are locked during a step');
  await page.click('#diagCancel');
  check(await textMatches(page, '#diagMsg', /cancelled/), 'a cancelled step can be run again');
  check(await page.isEnabled('#tabTest') && await page.isVisible('#diagGo'), 'tabs unlock after cancelling');
  await page.click('#diagGo');
  check(await eventually(async () => /\d/.test(await page.textContent('#diagValue')), 20000), 'live values while a step runs');
  await page.waitForSelector('#diagResult:not(.hidden), #diagSteps [data-step="near"].run', { timeout: 60000 });
  check(/↓ /.test(await page.textContent('#diagSteps [data-step="now"] .ds-state')), 'the finished step shows its download speed');
  // Steps 2 and 3, whatever this machine's own connection is.
  await page.evaluate(() => {
    const s = diag.steps.find(x => x.kind === 'now');
    diagReset();
    Object.assign(diag, { steps: [s], marks: { now: 'done' }, step: 'near' });
    renderDiag();
  });
  check(await page.isVisible('#diagSkip') && /close to the router/.test(await page.textContent('#diagMsg')), 'step 2 asks to move near the router');
  await page.click('#diagSkip');
  check(/LAN cable/.test(await page.textContent('#diagMsg')) && await page.textContent('#diagSkip') === 'Skip, show the conclusion', 'step 3 offers a cable test');
  check(await page.textContent('#diagSteps [data-step="near"] .ds-state') === 'skipped', 'a skipped step is marked');
  await page.click('#diagGo');
  await page.waitForSelector('#diagResult:not(.hidden)', { timeout: 60000 });
  const facts = await page.locator('#diagFacts li').allTextContents();
  check(facts.some(f => f.startsWith('Where you are:')) && facts.some(f => f.startsWith('With a cable:')), 'the conclusion lists each measured step');
  check((await page.textContent('#diagTitle')).length > 5 && (await page.textContent('#diagDetail')).length > 20, 'the conclusion explains itself');
  check(await page.locator('#diagActions button').count() >= 1, 'the conclusion offers a next step');
  check((await page.evaluate(() => bridge.call('history'))).length === histN, 'diagnosis tests are not saved in History');
  await checkNoRawKeys(page, 'diagnosis');
  await shot(page, 'diagnosis');
  const action = await page.getAttribute('#diagActions button', 'data-action');
  await page.click('#diagActions button');
  check(await page.getAttribute('#tabDiag', 'aria-selected') === 'false', `the "${action}" action opens its tool`);
  await page.click('#tabDiag');
  check(await page.isVisible('#diagResult'), 'the conclusion stays until Run again');
  await page.evaluate(() => { window.diagForSite = diag.steps; });
  await page.click('#diagAgain');
  check(await page.getAttribute('#diagSteps [data-step="now"]', 'class') === 'run' && await page.isVisible('#diagGo'), 'Run again starts over');

  // A finished step is coloured by its rating, not always green.
  await page.evaluate(() => {
    Object.assign(diag, { steps: [{ kind: 'now', result: { quality: { level: 'POOR' }, download: { mbps: 4.8 } } }], marks: { now: 'done' }, step: 'near' });
    renderDiag();
  });
  check(await page.getAttribute('#diagSteps [data-step="now"]', 'class') === 'done q-bad', 'a poor step result is shown in red');
  await page.evaluate(() => { diag.steps[0].result.quality.level = 'GOOD'; renderDiag(); });
  check(await page.getAttribute('#diagSteps [data-step="now"]', 'class') === 'done q-ok', 'a good step result is shown in green');

  // A strong signal where the user is: no walk to the router.
  await page.evaluate(() => {
    diagReset();
    return diagStepDone({ mode: 'general', pingMs: 15, jitterMs: 2, download: { mbps: 4.8 }, upload: { mbps: 4.9 },
      network: { connection: 'Wi-Fi', wifi: { signalPercent: 85 } }, quality: { level: 'POOR' } });
  });
  check(await textMatches(page, '#diagSteps [data-step="near"] .ds-state', /not needed \(85% signal\)/), 'a strong signal skips the near-router step');
  check(/already strong \(85%\)/.test(await page.textContent('#diagMsg')) && await page.textContent('#diagGo') === 'Test with the cable', 'the cable step explains why');
  await page.click('#diagSkip');
  await page.waitForSelector('#diagResult:not(.hidden)');
  check(/ISP/.test(await page.textContent('#diagTitle')) && /already strong \(85%\)/.test(await page.textContent('#diagFacts')), 'a slow test on a strong signal points to the ISP');
  await page.click('#diagAgain');

  // Over a phone hotspot (the user's realme: phone answers in 22 ms) there
  // is no router to walk to: the diagnosis ends after step 1.
  await page.evaluate(() => {
    diagReset();
    return diagStepDone({ mode: 'general', pingMs: 40, jitterMs: 6, download: { mbps: 8 }, upload: { mbps: 6 },
      network: { connection: 'Hotspot / Tethering', metered: true, wifi: { signalPercent: 95 } }, quality: { level: 'POOR' },
      pings: [{ isGateway: true, sent: 20, received: 20, avgMs: 22 }, { primaryTarget: true, sent: 20, received: 20, avgMs: 40 }] });
  });
  await page.waitForSelector('#diagResult:not(.hidden)');
  check(/phone hotspot/.test(await page.textContent('#diagTitle')) && !/router/i.test(await page.textContent('#diagTitle')), 'a slow hotspot test concludes "via a phone hotspot": ' + await page.textContent('#diagTitle'));
  check(await textMatches(page, '#diagSteps [data-step="near"] .ds-state', /skipped/), 'no walk to a router over a hotspot');
  await shot(page, 'diag-hotspot');
  await page.click('#diagAgain');
  await page.click('#tabTest');

  console.log('theme');
  const bg = () => page.evaluate(() => getComputedStyle(document.body).backgroundColor);
  const lightBg = await bg();
  await page.click('#settingsBtn');
  await page.selectOption('#themeSelect', 'dark');
  check(await bg() === 'rgb(14, 17, 22)', 'Dark theme applies at once');
  check(await eventually(async () => (await page.evaluate(() => bridge.call('getPrefs'))).theme === 'dark'), 'theme is saved');
  await page.selectOption('#themeSelect', 'system');
  check(await bg() === lightBg, 'Follow Windows goes back to the system theme');
  await page.keyboard.press('Escape');

  console.log('about');
  await page.click('#aboutBtn');
  await page.waitForSelector('#aboutDlg[open]');
  const about = await page.textContent('#aboutTraffic');
  check(about.includes('9.9.9.9') && about.includes('Game server'), 'About lists DNS-test servers and custom targets');
  check(about.includes('Ping to regions') && about.includes('127.0.0.1:18099'), 'About lists the region hosts');

  console.log('report a problem');
  await page.click('#aboutProblem');
  await page.waitForSelector('#problemDlg[open]');
  check(!(await page.evaluate(() => document.querySelector('#aboutDlg').open)), 'the problem dialog replaces About');
  const info = await page.textContent('#problemInfo');
  const myIp = await page.evaluate(() => ((state.network || {}).ipv4 || [])[0] || '');
  check(/WiFi Speed \+ Ping Tester/.test(info) && /Connection: /.test(info) && /Last test: /.test(info), 'the report shows app, connection and last test');
  check(!myIp || !info.includes(myIp), 'the report holds no IP address');
  const copiedInfo = await toastAfter(page, () => page.click('#problemCopy'), /copied/);
  check(/Info copied/.test(copiedInfo) && (await page.evaluate(() => navigator.clipboard.readText())) === info, 'Copy info puts the report on the clipboard');
  await ctx.route('https://github.com/**', r => r.fulfill({ status: 200, body: 'issue form' }));
  const [popup] = await Promise.all([page.waitForEvent('popup'), page.click('#problemIssue')]);
  const issueUrl = new URL(popup.url());
  check(issueUrl.href.startsWith('https://github.com/tehgeii/wifi-speed-tester/issues/new?') && issueUrl.searchParams.get('body').includes(info.split('\n')[0]), 'Open GitHub Issues prefills a new issue');
  await popup.close();
  await page.keyboard.press('Escape');

  console.log('data saver and repeat');
  const dataLine = () => page.textContent('#dataLine');
  check(/This test used ≈ \d/.test(await dataLine()), 'the last test shows the data it used');
  const single = await page.evaluate(() => state.history.find(h => h.mode !== 'quick').dataMB);
  await page.click('#repeatBtn');
  const est = (await dataLine()).match(/Estimated data: ≈ ([\d.]+) (MB|GB)/);
  check(est && Math.abs((est[2] === 'GB' ? est[1] * 1000 : +est[1]) - single * 3) <= Math.max(1, single * 0.2), 'estimate for 3 runs from the last test on this connection');
  await page.click('#saverBtn');
  check(await page.getAttribute('#saverBtn', 'aria-pressed') === 'true' && /Data saver on: about 135 MB per test/.test(await dataLine()), 'data saver shows its cap (3 runs)');
  await page.click('#saverBtn');
  // Home Wi-Fi that Windows reports as metered gets the hint too.
  await page.evaluate(() => renderNetwork({ ...state.network, connection: 'Wi-Fi', metered: true }));
  check(await page.isVisible('#hotspotHint') && /metered/.test(await page.textContent('#hotspotMsg')), 'a metered connection suggests data saver');
  await page.evaluate(() => renderNetwork({ ...state.network, metered: false }));
  check(!(await page.isVisible('#hotspotHint')), 'no hint on unmetered Wi-Fi');
  await page.evaluate(() => renderNetwork({ ...state.network, connection: 'Hotspot / Tethering' }));
  check(await page.isVisible('#hotspotHint') && /phone hotspot/.test(await page.textContent('#hotspotMsg')), 'a phone hotspot suggests data saver');
  await page.click('#hotspotOn');
  check(await page.getAttribute('#saverBtn', 'aria-pressed') === 'true' && !(await page.isVisible('#hotspotHint')), 'one click turns data saver on');
  check(await eventually(async () => {
    const p = await page.evaluate(() => bridge.call('getPrefs'));
    return p.dataSaver === 'on' && p.repeat === 'on';
  }), 'data saver and repeat are remembered');
  const histBefore = (await page.evaluate(() => bridge.call('history'))).length;
  await page.click('#startBtn');
  await page.waitForFunction(() => /Test 2 of 3/.test(document.querySelector('#gaugeLabel').textContent), null, { timeout: 60000 });
  check(/Data used: \d/.test(await dataLine()), 'live data counter while testing');
  await page.waitForSelector('#summaryCard:not(.hidden)', { timeout: 120000 });
  await page.waitForSelector('#startBtn:not(.hidden)');
  check(await page.isVisible('#runsBox') && await page.locator('#runsBox tbody tr').count() === 3, 'the 3 runs are listed');
  check(/This test used ≈ \d/.test(await dataLine()), 'data used by the test is shown afterwards');
  check(await page.isVisible('#cmpLine') && /^Compared with the previous test \(.+\): Download /.test(await page.textContent('#cmpLine')),
    'the second full test is compared with the first: ' + await page.textContent('#cmpLine'));
  await shot(page, 'repeat-result');
  const notes = await page.textContent('#notes');
  check(notes.includes('3 times') && notes.includes('Data saver'), 'notes explain the median and data saver');
  const hist = await page.evaluate(() => bridge.call('history'));
  check(hist.length === histBefore + 1 && hist[0].runs === 3 && hist[0].dataSaver && hist[0].dataMB > 0, 'a repeated test is saved once, with its data use');
  await page.click('#repeatBtn');
  await page.click('#saverBtn');
  await page.click('#tabHistory');
  await page.waitForSelector('#histTable tbody tr');
  check(/Data used by tests this month: .*\(\d+ tests\)/.test(await page.textContent('#histData')), 'history shows this month\'s data use');
  check(await textMatches(page, '#histTable tbody tr:first-child', /3×/), 'history marks the repeated test');
  await page.click('#tabTest');

  console.log('Indonesian');
  await page.click('#settingsBtn');
  await page.selectOption('#langSelect', 'id');
  await page.keyboard.press('Escape');
  // #netConn and #gaugeLabel start with a placeholder and then show live data.
  const wrong = await page.evaluate(() => [...document.querySelectorAll('[data-i18n]:not(#netConn):not(#gaugeLabel)')]
    .filter(el => el.textContent !== TEXT.id[el.dataset.i18n]).map(el => el.dataset.i18n));
  check(wrong.length === 0, 'every label switched to Indonesian ' + wrong.join(' '));
  // The result is re-rendered after it is reloaded in the new language.
  check(await textMatches(page, '#m-download .sub', /% dari paket/), 'plan share in Indonesian');
  // Engine texts follow the language of the run.
  await page.click('#quickBtn');
  await page.waitForSelector('#startBtn:not(.hidden)', { timeout: 30000 });
  await page.waitForSelector('#summaryCard:not(.hidden)');
  check((await page.textContent('#pingTable')).includes('Gateway Lokal'), 'ping target labels in Indonesian');
  check((await page.textContent('#actList')).includes('Game online'), 'activities in Indonesian');
  check(/^Dibanding tes sebelumnya \(.+\): Ping /.test(await page.textContent('#cmpLine')), 'the second Quick Ping is compared with the first, in Indonesian');
  if (SHOTS) {
    // The website shows the app in Indonesian.
    await page.click('#tabHistory');
    await page.waitForSelector('#histTable tbody tr');
    await siteShot(page, 'history');
    await page.click('#histTable tbody tr:nth-child(2)');
    await page.waitForSelector('#summaryCard:not(.hidden)');
    await siteShot(page, 'result');
    await page.click('#tabDiag');
    await page.evaluate(() => { diag.steps = window.diagForSite; diag.marks = { now: 'done', cable: 'done' }; return diagConclude(); });
    await page.waitForSelector('#diagResult:not(.hidden)');
    await siteShot(page, 'diagnosis');
    await page.click('#tabTools');
    await page.click('[data-tool="signalCard"]');
    await page.click('#sigStart');
    await page.waitForFunction(() => sig.samples.length >= 9, null, { timeout: 15000 });
    await siteShot(page, 'signal');
    await page.click('[data-tool="regionCard"]');
    await page.click('#regStart');
    await page.waitForFunction(() => !reg.running && reg.results.length === 3, null, { timeout: 30000 });
    await siteShot(page, 'regions');
  }
  await page.click('#tabTools');
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
  await dp.click('#tabHistory');
  await dp.waitForSelector('#histTable tbody tr');
  await noHorizontalScroll(dp, 'history');
  await shot(dp, 'history-dark');
  await dp.click('#histTable tbody tr:first-child');
  await dp.waitForSelector('#summaryCard:not(.hidden)');
  await noHorizontalScroll(dp, 'result and export actions');
  await shot(dp, 'result-dark');
  await dp.click('#tabTools');
  await dp.click('[data-tool="wifiCard"]');
  await dp.click('#scanStart');
  await dp.waitForSelector('#scanResult:not(.hidden)');
  await noHorizontalScroll(dp, 'tools');
  await shot(dp, 'tools-dark');
  await dp.click('[data-tool="regionCard"]');
  await dp.click('#regStart');
  await dp.waitForFunction(() => !reg.running && reg.results.length === 3, null, { timeout: 30000 });
  await noHorizontalScroll(dp, 'regions');
  await dp.click('[data-tool="signalCard"]');
  await dp.click('#sigStart');
  await dp.waitForFunction(() => sig.samples.length >= 2, null, { timeout: 10000 });
  await noHorizontalScroll(dp, 'signal finder');
  await shot(dp, 'signal-dark');
  await dp.click('#tabDiag');
  check(await dp.evaluate(() => !sig.running), 'leaving the Tools tab stops the signal finder');
  await noHorizontalScroll(dp, 'diagnosis');
  await shot(dp, 'diagnosis-dark');
  await dp.evaluate(() => { prefs.theme = 'light'; applyTheme(); });
  check(await dp.evaluate(() => getComputedStyle(document.body).backgroundColor) === 'rgb(244, 246, 249)', 'Light theme overrides a dark Windows');
  await dark.close();
  await browser.close();

  check(errors.length === 0, 'no page errors ' + errors.join(' | '));
  console.log(failures ? `\n${failures} check(s) failed` : '\nall UI checks passed');
  process.exit(failures ? 1 : 0);
}

main().catch(e => { console.error(e); process.exit(1); });
