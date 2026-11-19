'use strict';
// Points the download buttons at the latest release's zips. Without
// JavaScript, or if GitHub cannot be reached, they open the release page.
(async () => {
  try {
    const r = await fetch('https://api.github.com/repos/tehgeii/wifi-speed-tester/releases/latest', { headers: { Accept: 'application/vnd.github+json' } });
    if (!r.ok) return;
    const rel = await r.json();
    const asset = arch => (rel.assets || []).find(a => a.name.endsWith(`-${arch}.zip`));
    const x64 = asset('win-x64'), arm = asset('win-arm64');
    const v = (rel.tag_name || '').replace(/^v/, '');
    if (x64) {
      const main = document.getElementById('dlMain');
      main.href = x64.browser_download_url;
      main.textContent = `⬇ Download v${v} untuk Windows`;
      document.getElementById('dlMeta').textContent = `Versi ${v} · ${(x64.size / 1e6).toFixed(1)} MB · gratis, portable, tanpa install · Windows 10 dan 11`;
    }
    if (arm) document.getElementById('dlArm').href = arm.browser_download_url;
  } catch { /* keep the release-page links */ }
})();
