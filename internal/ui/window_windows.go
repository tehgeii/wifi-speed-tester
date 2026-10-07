//go:build windows

package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"

	"github.com/tehgeii/wifi-speed-tester/internal/app"
)

// ErrNoWebView2 means the Microsoft Edge WebView2 Runtime is not installed.
var ErrNoWebView2 = errors.New("Microsoft Edge WebView2 Runtime not found")

// RunWindow opens the desktop window and blocks until it is closed.
func RunWindow(exeDir string, logf func(string, ...any)) error {
	// WebView2 keeps a browser profile; put it in the per-user cache rather
	// than next to the portable executable.
	dataPath := filepath.Join(os.TempDir(), "WiFiSpeedTester-WebView2")
	if d, err := os.UserCacheDir(); err == nil {
		dataPath = filepath.Join(d, "WiFiSpeedTester", "WebView2")
	}

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     os.Getenv("WST_DEBUG") == "1",
		AutoFocus: true,
		DataPath:  dataPath,
		WindowOptions: webview2.WindowOptions{
			Title:  "WiFi Speed + Ping Tester",
			Width:  1060,
			Height: 820,
			IconId: 1, // from the embedded .syso resource
			Center: true,
		},
	})
	if w == nil {
		return ErrNoWebView2
	}
	defer w.Destroy()
	w.SetSize(760, 600, webview2.HintMin)

	host := app.Host{
		Emit: func(name string, payload any) {
			n, _ := json.Marshal(name)
			p, err := json.Marshal(payload)
			if err != nil {
				return
			}
			js := fmt.Sprintf("window.__wstEvent && window.__wstEvent(%s, %s)", n, p)
			w.Dispatch(func() { w.Eval(js) })
		},
		SaveFile: func(name, filter, ext string, data []byte) (string, error) {
			// The dialog must run on the UI thread; this is called from a
			// worker goroutine, so hop over and wait for the answer.
			type res struct {
				path string
				err  error
			}
			ch := make(chan res, 1)
			w.Dispatch(func() {
				// A panic here would close the whole window; report it as
				// an export error instead.
				defer func() {
					if v := recover(); v != nil {
						ch <- res{"", fmt.Errorf("save dialog failed: %v", v)}
					}
				}()
				p, err := saveDialog(uintptr(w.Window()), name, filter, ext)
				ch <- res{p, err}
			})
			r := <-ch
			if r.err != nil || r.path == "" {
				return "", r.err
			}
			return r.path, os.WriteFile(r.path, data, 0o644)
		},
	}
	host.CopyText = func(text string) error { return copyText(uintptr(w.Window()), text) }
	host.OpenURL = func(url string) error { OpenURL(url); return nil }
	a := app.New(exeDir, host, logf)
	if err := w.Bind("wstCall", func(method string, params []json.RawMessage) (any, error) {
		return a.Call(method, params)
	}); err != nil {
		return err
	}
	w.SetHtml(InlinePage())
	w.Run()
	a.CancelTest()
	return nil
}

var (
	modComdlg32          = windows.NewLazySystemDLL("comdlg32.dll")
	procGetSaveFileNameW = modComdlg32.NewProc("GetSaveFileNameW")
	procCommDlgExtError  = modComdlg32.NewProc("CommDlgExtendedError")
	modUser32            = windows.NewLazySystemDLL("user32.dll")
	procMessageBoxW      = modUser32.NewProc("MessageBoxW")
)

type openFileName struct {
	StructSize    uint32
	Owner         uintptr
	Instance      uintptr
	Filter        *uint16
	CustomFilter  *uint16
	MaxCustFilter uint32
	FilterIndex   uint32
	File          *uint16
	MaxFile       uint32
	FileTitle     *uint16
	MaxFileTitle  uint32
	InitialDir    *uint16
	Title         *uint16
	Flags         uint32
	FileOffset    uint16
	FileExtension uint16
	DefExt        *uint16
	CustData      uintptr
	FnHook        uintptr
	TemplateName  *uint16
	Reserved      uintptr
	Reserved2     uint32
	FlagsEx       uint32
}

const (
	ofnOverwritePrompt = 0x2
	ofnNoChangeDir     = 0x8
	ofnPathMustExist   = 0x800
)

// saveDialog shows the standard Save As dialog. Returns "" on cancel.
func saveDialog(owner uintptr, defaultName, filterName, ext string) (string, error) {
	filter := dialogFilter(filterName, ext)
	file := make([]uint16, windows.MAX_LONG_PATH)
	name, err := windows.UTF16FromString(defaultName)
	if err != nil {
		return "", err
	}
	copy(file, name)
	defExt, _ := windows.UTF16PtrFromString(ext)
	title, _ := windows.UTF16PtrFromString("Save result")
	var initial *uint16
	if home, err := os.UserHomeDir(); err == nil {
		initial, _ = windows.UTF16PtrFromString(filepath.Join(home, "Documents"))
	}
	ofn := openFileName{
		Owner:       owner,
		Filter:      &filter[0],
		FilterIndex: 1,
		File:        &file[0],
		MaxFile:     uint32(len(file)),
		InitialDir:  initial,
		Title:       title,
		Flags:       ofnOverwritePrompt | ofnNoChangeDir | ofnPathMustExist,
		DefExt:      defExt,
	}
	ofn.StructSize = uint32(unsafe.Sizeof(ofn))
	ok, _, _ := procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ok == 0 {
		if code, _, _ := procCommDlgExtError.Call(); code != 0 {
			return "", fmt.Errorf("save dialog failed (code %#x)", code)
		}
		return "", nil
	}
	return windows.UTF16ToString(file), nil
}

// MessageBox shows a native message box and returns the button pressed.
func MessageBox(title, text string, flags uint32) int {
	t, _ := windows.UTF16PtrFromString(title)
	m, _ := windows.UTF16PtrFromString(text)
	r, _, _ := procMessageBoxW.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), uintptr(flags))
	return int(r)
}

// OpenURL opens a link in the default browser.
func OpenURL(url string) {
	verb, _ := windows.UTF16PtrFromString("open")
	u, _ := windows.UTF16PtrFromString(url)
	_ = windows.ShellExecute(0, verb, u, nil, nil, windows.SW_SHOWNORMAL)
}
