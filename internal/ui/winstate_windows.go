//go:build windows

package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Window comfort: remember size and position, match the title bar to the
// page theme, and flash the taskbar button when something finishes while
// the window is in the background.

var (
	modDwmapi                 = windows.NewLazySystemDLL("dwmapi.dll")
	procDwmSetWindowAttribute = modDwmapi.NewProc("DwmSetWindowAttribute")
	procGetWindowPlacement    = modUser32.NewProc("GetWindowPlacement")
	procSetWindowPlacement    = modUser32.NewProc("SetWindowPlacement")
	procMonitorFromRect       = modUser32.NewProc("MonitorFromRect")
	procSetWindowLongPtrW     = modUser32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW       = modUser32.NewProc("CallWindowProcW")
	procSetWindowPos          = modUser32.NewProc("SetWindowPos")
	procFlashWindowEx         = modUser32.NewProc("FlashWindowEx")
	procGetForegroundWindow   = modUser32.NewProc("GetForegroundWindow")
)

type point struct{ X, Y int32 }

type rect struct{ Left, Top, Right, Bottom int32 }

// windowPlacement is WINDOWPLACEMENT.
type windowPlacement struct {
	Length         uint32
	Flags          uint32
	ShowCmd        uint32
	MinPosition    point
	MaxPosition    point
	NormalPosition rect
}

// savedWindow is what window.json holds.
type savedWindow struct {
	Left, Top, Right, Bottom int32
	Maximized                bool
}

const (
	swShowNormal    = 1
	swShowMaximized = 3
	wmClose         = 0x0010
	gwlpWndProc     = ^uintptr(3) // -4
)

func windowFile(dataDir string) string { return filepath.Join(dataDir, "window.json") }

// loadWindow returns the saved placement if it is sane and still on a
// connected monitor (screens change between runs).
func loadWindow(dataDir string) (savedWindow, bool) {
	var s savedWindow
	b, err := os.ReadFile(windowFile(dataDir))
	if err != nil || json.Unmarshal(b, &s) != nil {
		return s, false
	}
	if s.Right-s.Left < 400 || s.Bottom-s.Top < 300 {
		return s, false
	}
	r := rect{s.Left, s.Top, s.Right, s.Bottom}
	const monitorDefaultToNull = 0
	if m, _, _ := procMonitorFromRect.Call(uintptr(unsafe.Pointer(&r)), monitorDefaultToNull); m == 0 {
		return s, false
	}
	return s, true
}

func restoreWindow(hwnd uintptr, s savedWindow) {
	wp := windowPlacement{ShowCmd: swShowNormal, NormalPosition: rect{s.Left, s.Top, s.Right, s.Bottom}}
	if s.Maximized {
		wp.ShowCmd = swShowMaximized
	}
	wp.Length = uint32(unsafe.Sizeof(wp))
	procSetWindowPlacement.Call(hwnd, uintptr(unsafe.Pointer(&wp)))
}

func saveWindow(hwnd uintptr, dataDir string) {
	var wp windowPlacement
	wp.Length = uint32(unsafe.Sizeof(wp))
	if ok, _, _ := procGetWindowPlacement.Call(hwnd, uintptr(unsafe.Pointer(&wp))); ok == 0 {
		return
	}
	r := wp.NormalPosition
	b, err := json.Marshal(savedWindow{r.Left, r.Top, r.Right, r.Bottom, wp.ShowCmd == swShowMaximized})
	if err != nil {
		return
	}
	_ = os.MkdirAll(dataDir, 0o755)
	_ = os.WriteFile(windowFile(dataDir), b, 0o644)
}

// The window procedure belongs to go-webview2, which has no close hook, so
// it is subclassed once to save the placement before the window goes away.
var (
	origWndProc  uintptr
	closeHook    func(hwnd uintptr)
	subclassProc = windows.NewCallback(func(hwnd, msg, wp, lp uintptr) uintptr {
		if msg == wmClose && closeHook != nil {
			closeHook(hwnd)
			closeHook = nil
		}
		r, _, _ := procCallWindowProcW.Call(origWndProc, hwnd, msg, wp, lp)
		return r
	})
)

func onClose(hwnd uintptr, f func(hwnd uintptr)) {
	closeHook = f
	origWndProc, _, _ = procSetWindowLongPtrW.Call(hwnd, gwlpWndProc, subclassProc)
}

// setDarkTitleBar asks Windows 10 (20H1 and later; 19 is the attribute
// number before that) and 11 for a dark or light window frame.
func setDarkTitleBar(hwnd uintptr, dark bool) {
	var v uint32
	if dark {
		v = 1
	}
	for _, attr := range []uintptr{20, 19} { // DWMWA_USE_IMMERSIVE_DARK_MODE
		if r, _, _ := procDwmSetWindowAttribute.Call(hwnd, attr, uintptr(unsafe.Pointer(&v)), unsafe.Sizeof(v)); r == 0 {
			break
		}
	}
	// Repaint the frame now rather than at the next activation.
	const swpNoSize, swpNoMove, swpNoZOrder, swpNoActivate, swpFrameChanged = 0x1, 0x2, 0x4, 0x10, 0x20
	procSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0, swpNoSize|swpNoMove|swpNoZOrder|swpNoActivate|swpFrameChanged)
}

// flashWInfo is FLASHWINFO.
type flashWInfo struct {
	Size    uint32
	Hwnd    uintptr
	Flags   uint32
	Count   uint32
	Timeout uint32
}

// flashTaskbar flashes the taskbar button until the window is brought to
// the front; nothing happens when it already is.
func flashTaskbar(hwnd uintptr) {
	if fg, _, _ := procGetForegroundWindow.Call(); fg == hwnd {
		return
	}
	const flashwTray, flashwTimerNoFG = 0x2, 0xC
	fi := flashWInfo{Hwnd: hwnd, Flags: flashwTray | flashwTimerNoFG}
	fi.Size = uint32(unsafe.Sizeof(fi))
	procFlashWindowEx.Call(uintptr(unsafe.Pointer(&fi)))
}
