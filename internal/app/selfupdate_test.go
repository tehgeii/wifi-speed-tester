package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/update"
)

func TestInstallUpdateFlow(t *testing.T) {
	a, ev, _ := newTestApp(t)
	exe := filepath.Join(t.TempDir(), update.ExeName)
	os.WriteFile(exe, []byte("program"), 0o755)
	a.exe = exe

	// Nothing to install before a check found a release.
	if _, err := call(t, a, "installUpdate"); err == nil || !strings.Contains(err.Error(), "check for updates first") {
		t.Fatalf("without a check: %v", err)
	}
	// Not while a test runs: the download would skew it.
	a.update = &update.Info{Latest: "9.0.0", Newer: true, CanInstall: true, ZipName: "x.zip",
		ZipURL: "https://example.com/x.zip", SumsURL: "https://example.com/SHA256SUMS.txt"}
	a.mu.Lock()
	a.running = true
	a.mu.Unlock()
	if _, err := call(t, a, "installUpdate"); err == nil || !strings.Contains(err.Error(), "test is already running") {
		t.Fatalf("during a test: %v", err)
	}
	a.mu.Lock()
	a.running = false
	a.mu.Unlock()

	// A failed install is reported as an event and leaves the program alone.
	if _, err := call(t, a, "installUpdate"); err != nil {
		t.Fatal(err)
	}
	e := ev.wait(t, "updateInstall", "error", 5*time.Second)
	if !strings.Contains(e["error"].(string), "refusing") {
		t.Fatalf("event = %v", e)
	}
	if b, _ := os.ReadFile(exe); string(b) != "program" {
		t.Fatal("program changed")
	}
	if a.installing || a.installed != "" {
		t.Fatalf("state after a failure: installing %v installed %q", a.installing, a.installed)
	}
}

func TestRestartApp(t *testing.T) {
	a, _, _ := newTestApp(t)
	if _, err := call(t, a, "restartApp"); err == nil {
		t.Fatal("restart without a window host must fail")
	}
	var got string
	a.host.Restart = func(p string) error { got = p; return nil }
	a.exe = `C:\Apps\WiFiSpeedTester.exe`
	if _, err := call(t, a, "restartApp"); err != nil || got != a.exe {
		t.Fatalf("restart: %v, path %q", err, got)
	}
	if info, _ := call(t, a, "info"); info.(map[string]any)["justUpdated"] != false {
		t.Fatal("info.justUpdated")
	}
}
