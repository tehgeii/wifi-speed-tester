//go:build windows

package ui

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procFindWindowW         = modUser32.NewProc("FindWindowW")
	procShowWindow          = modUser32.NewProc("ShowWindow")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
)

// windowTitle is shared by the window and the single-instance lookup.
const windowTitle = "WiFi Speed + Ping Tester"

// SingleInstance claims a named mutex per data folder, so two windows never
// write the same history file. When another instance already holds it, that
// window is brought to the front and ok is false.
func SingleInstance(dataDir string) (release func(), ok bool) {
	sum := sha256.Sum256([]byte(strings.ToLower(dataDir)))
	name, _ := windows.UTF16PtrFromString(`Local\WiFiSpeedTester-` + hex.EncodeToString(sum[:8]))
	h, err := windows.CreateMutex(nil, false, name)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		if h != 0 {
			windows.CloseHandle(h)
		}
		activateExisting()
		return func() {}, false
	}
	if err != nil || h == 0 {
		return func() {}, true // cannot tell; do not block the app
	}
	return func() { windows.CloseHandle(h) }, true
}

func activateExisting() {
	class, _ := windows.UTF16PtrFromString("webview") // go-webview2's window class
	title, _ := windows.UTF16PtrFromString(windowTitle)
	hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(title)))
	if hwnd == 0 {
		return
	}
	const swRestore = 9
	procShowWindow.Call(hwnd, swRestore)
	procSetForegroundWindow.Call(hwnd)
}
