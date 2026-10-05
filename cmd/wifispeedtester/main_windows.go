//go:build windows

package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"golang.org/x/sys/windows"

	"github.com/tehgeii/wifi-speed-tester/internal/app"
	"github.com/tehgeii/wifi-speed-tester/internal/ui"
)

const webView2URL = "https://developer.microsoft.com/microsoft-edge/webview2/"

// prepareConsole attaches to the parent console when started from a terminal
// with arguments. The executable is built as a GUI program, so without this
// --cli output would go nowhere. Streams that were redirected (for example
// "--cli --json > result.json") are left alone.
func prepareConsole() {
	if len(os.Args) < 2 {
		return
	}
	valid := func(h uint32) bool {
		fd, err := windows.GetStdHandle(h)
		return err == nil && fd != 0 && fd != windows.InvalidHandle
	}
	outOK, errOK := valid(windows.STD_OUTPUT_HANDLE), valid(windows.STD_ERROR_HANDLE)
	if outOK && errOK {
		return
	}
	const attachParentProcess = ^uint32(0) // ATTACH_PARENT_PROCESS = (DWORD)-1
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("AttachConsole")
	if r, _, _ := proc.Call(uintptr(attachParentProcess)); r == 0 {
		return
	}
	f, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		return
	}
	if !outOK {
		os.Stdout = f
	}
	if !errOK {
		os.Stderr = f
		log.SetOutput(f)
	}
}

func runGUI(exeDir string) {
	// The GUI has no console, so a crash would vanish without a trace.
	// Record it next to the history instead.
	if f, err := os.OpenFile(filepath.Join(app.DataDir(exeDir), "crash.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
		_ = debug.SetCrashOutput(f, debug.CrashOptions{})
		f.Close()
	}
	logf := func(string, ...any) {}
	if os.Getenv("WST_DEBUG") == "1" {
		if f, err := os.Create(filepath.Join(exeDir, "WiFiSpeedTester-debug.log")); err == nil {
			l := log.New(f, "", log.LstdFlags)
			logf = l.Printf
		}
	}
	err := ui.RunWindow(exeDir, logf)
	if errors.Is(err, ui.ErrNoWebView2) {
		const mbYesNo, mbIconWarning, idYes = 0x4, 0x30, 6
		if ui.MessageBox("WiFi Speed + Ping Tester",
			"This app needs the Microsoft Edge WebView2 Runtime, which is included with Windows 11 and most Windows 10 installations but was not found.\n\n"+
				"Open the download page now?\n\n(You can also run the test in a terminal: WiFiSpeedTester.exe --cli)",
			mbYesNo|mbIconWarning) == idYes {
			ui.OpenURL(webView2URL)
		}
		os.Exit(1)
	}
	if err != nil {
		const mbIconError = 0x10
		ui.MessageBox("WiFi Speed + Ping Tester", "The window could not be opened:\n\n"+err.Error(), mbIconError)
		os.Exit(1)
	}
}
