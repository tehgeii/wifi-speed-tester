package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/update"
)

// installUpdate downloads and installs the release found by the last update
// check ("Update now"). It reports through "updateInstall" events: stage
// download (with done/total bytes), verify, install, then done or error. The
// new version starts after restartApp.
func (a *App) installUpdate() error {
	if err := a.busyErr(); err != nil {
		return err
	}
	exe := a.exe
	if exe == "" {
		return errors.New("the program's location is unknown; download the new version manually")
	}
	a.updMu.Lock()
	if a.update == nil || !a.update.CanInstall || a.installing {
		a.updMu.Unlock()
		return errors.New(i18n.T(a.lang(), "app.noUpdate"))
	}
	info := *a.update
	a.installing = true
	a.updMu.Unlock()

	go func() {
		err := update.Install(context.Background(), a.httpClient(), info, exe, func(stage string, done, total int64) {
			a.emit("updateInstall", map[string]any{"stage": stage, "done": done, "total": total})
		})
		a.updMu.Lock()
		a.installing = false
		if err == nil {
			a.installed = info.Latest
			if a.update != nil {
				a.update.CanInstall = false
			}
		}
		a.updMu.Unlock()
		if err != nil {
			a.logf("update: %v", err)
			a.emit("updateInstall", map[string]any{"stage": "error", "error": err.Error()})
			return
		}
		a.emit("updateInstall", map[string]any{"stage": "done", "version": info.Latest})
	}()
	return nil
}

// restartApp starts the updated program and closes this window.
func (a *App) restartApp() error {
	if a.host.Restart == nil {
		return errors.New("not supported")
	}
	if err := a.busyErr(); err != nil {
		return err
	}
	return a.host.Restart(a.exe)
}

// busyErr refuses while a test or the monitor runs: a download would skew
// the measurement, and a restart would end it.
func (a *App) busyErr() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch {
	case a.running:
		return errors.New(i18n.T(a.lang(), "app.busy"))
	case a.monCancel != nil:
		return errors.New(i18n.T(a.lang(), "app.monBusy"))
	}
	return nil
}

func exePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if p, err := filepath.EvalSymlinks(exe); err == nil {
		exe = p
	}
	return exe, nil
}
