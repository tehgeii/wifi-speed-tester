// Package app is the view-model layer between the engine and the UI. The UI
// (an embedded web page) calls methods through Call and receives progress
// through the Emit callback; it never runs measurements itself.
package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/engine"
	"github.com/tehgeii/wifi-speed-tester/internal/export"
	"github.com/tehgeii/wifi-speed-tester/internal/history"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
	"github.com/tehgeii/wifi-speed-tester/internal/network"
	"github.com/tehgeii/wifi-speed-tester/internal/servers"
	"github.com/tehgeii/wifi-speed-tester/internal/update"
)

// Version is set at build time with -ldflags "-X .../internal/app.Version=...".
var Version = "dev"

// Host is what the window (WebView2 or the dev server) provides.
type Host struct {
	// Emit delivers an event to the page. Must be safe from any goroutine.
	Emit func(name string, payload any)
	// SaveFile asks where to save data and writes it. It returns the chosen
	// path, or "" when the user cancelled.
	SaveFile func(defaultName, filterName, ext string, data []byte) (string, error)
	// CopyText puts text on the clipboard. Optional: without it the page
	// copies the returned text itself.
	CopyText func(text string) error
	// OpenURL opens a web page in the default browser. Optional.
	OpenURL func(url string) error
}

// App holds the state of one window.
type App struct {
	host    Host
	cfg     config.Config
	cfgPath string
	cfgErr  error
	eng     *engine.Engine
	hist    *history.Store
	dataDir string

	mu      sync.Mutex
	cancel  context.CancelFunc
	running bool
	last    *model.TestResult
	logf    func(format string, args ...any)
}

// New loads configuration from exeDir and prepares the data directory.
func New(exeDir string, host Host, logf func(string, ...any)) *App {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	cfg, path, err := config.Load(exeDir)
	if err != nil {
		logf("config: %v (using defaults)", err)
	}
	a := &App{host: host, cfg: cfg, cfgPath: path, cfgErr: err, logf: logf}
	a.dataDir = DataDir(exeDir)
	a.hist = history.Open(filepath.Join(a.dataDir, "history.json"), cfg.HistoryLimit)
	a.eng = engine.New(cfg)
	return a
}

// DataDir picks a writable folder for history: next to the executable for a
// true portable install, otherwise the per-user config directory.
func DataDir(exeDir string) string {
	portable := filepath.Join(exeDir, "WiFiSpeedTester-data")
	if err := os.MkdirAll(portable, 0o755); err == nil {
		probe := filepath.Join(portable, ".write-test")
		if f, err := os.Create(probe); err == nil {
			f.Close()
			os.Remove(probe)
			return portable
		}
	}
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "WiFiSpeedTester")
	}
	return filepath.Join(os.TempDir(), "WiFiSpeedTester")
}

// Call dispatches a UI request. Every method returns quickly; long work runs
// on goroutines and reports back through events, so the window never freezes.
func (a *App) Call(method string, params []json.RawMessage) (any, error) {
	arg := func(i int, v any) error {
		if i >= len(params) {
			return nil
		}
		return json.Unmarshal(params[i], v)
	}
	switch method {
	case "info":
		return a.info(), nil
	case "detectNetwork":
		go a.detectNetwork()
		return nil, nil
	case "startTest":
		var mode string
		if err := arg(0, &mode); err != nil {
			return nil, err
		}
		return nil, a.StartTest(mode)
	case "cancelTest":
		a.CancelTest()
		return nil, nil
	case "history":
		return a.historyList()
	case "historyItem":
		var id string
		if err := arg(0, &id); err != nil {
			return nil, err
		}
		r, err := a.hist.Get(id)
		if r != nil && r.Quality != nil {
			// Notes and tips are text; rebuild them in the current language.
			// The rating itself is deterministic, so it does not change.
			r.Quality = analysis.Evaluate(r, r.Mode, a.cfg.Profile(r.Mode), a.lang())
		}
		return r, err
	case "copyResult":
		var req struct {
			ID   string `json:"id"`
			Unit string `json:"unit"`
		}
		if err := arg(0, &req); err != nil {
			return nil, err
		}
		return a.copyResult(req.ID, req.Unit)
	case "findServers":
		go a.findServers()
		return nil, nil
	case "getServer":
		return a.selectedServer(), nil
	case "setServer":
		var srv *config.Server
		if err := arg(0, &srv); err != nil {
			return nil, err
		}
		return nil, a.setServer(srv)
	case "checkUpdate":
		go a.checkUpdate()
		return nil, nil
	case "openUrl":
		var u string
		if err := arg(0, &u); err != nil {
			return nil, err
		}
		return nil, a.openURL(u)
	case "clearHistory":
		return nil, a.hist.Clear()
	case "export":
		var req ExportRequest
		if err := arg(0, &req); err != nil {
			return nil, err
		}
		go a.reportSave(func() (string, error) { return a.Export(req) })
		return nil, nil
	case "savePng":
		var dataURL string
		if err := arg(0, &dataURL); err != nil {
			return nil, err
		}
		go a.reportSave(func() (string, error) { return a.savePNG(dataURL) })
		return nil, nil
	case "getPrefs":
		return a.prefs(), nil
	case "setPrefs":
		var p map[string]string
		if err := arg(0, &p); err != nil {
			return nil, err
		}
		return nil, a.savePrefs(p)
	}
	return nil, fmt.Errorf("unknown method %q", method)
}

// reportSave runs a save off the UI thread (the host may show a modal
// dialog) and reports the outcome as a "saved" event.
func (a *App) reportSave(fn func() (string, error)) {
	path, err := fn()
	switch {
	case err != nil:
		a.emit("saved", map[string]any{"error": err.Error()})
	case path == "":
		a.emit("saved", map[string]any{"cancelled": true})
	default:
		a.emit("saved", map[string]any{"path": path})
	}
}

func (a *App) emit(name string, payload any) {
	if a.host.Emit != nil {
		a.host.Emit(name, payload)
	}
}

func (a *App) info() map[string]any {
	servers := make([]string, len(a.cfg.Servers))
	for i, s := range a.cfg.Servers {
		servers[i] = s.Name
	}
	targets := make([]string, len(a.cfg.PingTargets))
	for i, t := range a.cfg.PingTargets {
		targets[i] = t.Label + " (" + t.Host + ")"
	}
	m := map[string]any{
		"version":     Version,
		"configPath":  a.cfgPath,
		"configFile":  config.FileName,
		"servers":     servers,
		"pingTargets": targets,
		"historyPath": a.hist.Path(),
		"releasesUrl": update.ReleasesPage,
		"canOpenUrl":  a.host.OpenURL != nil,
		"downloadSec": a.cfg.DownloadSeconds,
		"uploadSec":   a.cfg.UploadSeconds,
		"profiles":    a.cfg.Profiles,
	}
	if a.cfgErr != nil {
		m["configError"] = a.cfgErr.Error()
	}
	return m
}

func (a *App) detectNetwork() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	info, err := network.Detect(ctx)
	if err != nil {
		a.emit("network", map[string]any{"error": err.Error()})
		return
	}
	a.emit("network", map[string]any{"network": info})
}

// StartTest begins a run in the background. Only one run at a time.
func (a *App) StartTest(mode string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return errors.New(i18n.T(a.lang(), "app.busy"))
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel, a.running = cancel, true
	go func() {
		defer cancel()
		opts := engine.Options{Mode: mode, Lang: a.lang(), Server: a.selectedServer()}
		res := a.eng.Run(ctx, opts, func(ev engine.Event) {
			if ev.Type == "result" {
				return // sent below, after it is stored
			}
			a.emit("test", ev)
		})
		a.mu.Lock()
		a.running, a.cancel = false, nil
		a.last = res
		a.mu.Unlock()
		if !res.Cancelled && (res.Quality != nil || res.Failure == nil) {
			if err := a.hist.Add(res); err != nil {
				a.logf("history: %v", err)
			}
		}
		a.emit("test", engine.Event{Type: "result", Stage: engine.StageDone, Progress: 1, Result: res})
	}()
	return nil
}

// CancelTest stops the running test, if any. Network operations are aborted
// through the context; the engine then reports a cancelled result.
func (a *App) CancelTest() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
}

// HistoryEntry is the compact form shown in the history list.
type HistoryEntry struct {
	ID         string    `json:"id"`
	Date       time.Time `json:"date"`
	Mode       string    `json:"mode"`
	Connection string    `json:"connection"`
	Download   float64   `json:"download"`
	Upload     float64   `json:"upload"`
	Ping       float64   `json:"ping"`
	Jitter     float64   `json:"jitter"`
	Loss       float64   `json:"loss"`
	Quality    string    `json:"quality"`
}

func (a *App) historyList() ([]HistoryEntry, error) {
	all, err := a.hist.List()
	if err != nil {
		return nil, fmt.Errorf("history file could not be read: %w", err)
	}
	out := make([]HistoryEntry, 0, len(all))
	for _, r := range all {
		e := HistoryEntry{ID: r.ID, Date: r.StartedAt, Mode: r.Mode, Ping: r.PingMs, Jitter: r.JitterMs, Loss: r.PacketLoss}
		if r.Network != nil {
			e.Connection = string(r.Network.Connection)
		}
		if r.Download != nil {
			e.Download = r.Download.Mbps
		}
		if r.Upload != nil {
			e.Upload = r.Upload.Mbps
		}
		if r.Quality != nil {
			e.Quality = string(r.Quality.Level)
		}
		out = append(out, e)
	}
	return out, nil
}

// ExportRequest is sent by the UI.
type ExportRequest struct {
	Format           string `json:"format"` // txt | json | csv | history-csv
	Unit             string `json:"unit"`
	IncludeSensitive bool   `json:"includeSensitive"`
	ID               string `json:"id"` // history entry; "" = latest result
}

// Export renders a result and asks the host where to save it.
func (a *App) Export(req ExportRequest) (string, error) {
	opts := export.Options{Unit: req.Unit, IncludeSensitive: req.IncludeSensitive, Lang: a.lang()}
	if req.Format == "history-csv" {
		all, err := a.hist.List()
		if err != nil {
			return "", err
		}
		if len(all) == 0 {
			return "", errors.New(i18n.T(a.lang(), "app.historyEmpty"))
		}
		return a.save("WiFiSpeedTest-history.csv", "CSV file", "csv", export.CSV(all, opts))
	}

	a.mu.Lock()
	r := a.last
	a.mu.Unlock()
	if req.ID != "" {
		var err error
		if r, err = a.hist.Get(req.ID); err != nil {
			return "", err
		}
	}
	if r == nil {
		return "", errors.New(i18n.T(a.lang(), "app.runFirst"))
	}
	name := "WiFiSpeedTest-" + r.StartedAt.Format("2006-01-02-150405")
	switch req.Format {
	case "txt":
		return a.save(name+".txt", "Text report", "txt", export.TXT(r, opts))
	case "json":
		b, err := export.JSON(r, opts)
		if err != nil {
			return "", err
		}
		return a.save(name+".json", "JSON file", "json", b)
	case "csv":
		return a.save(name+".csv", "CSV file", "csv", export.CSV([]*model.TestResult{r}, opts))
	}
	return "", fmt.Errorf("unknown format %q", req.Format)
}

func (a *App) savePNG(dataURL string) (string, error) {
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		return "", errors.New("invalid image data")
	}
	png, err := base64.StdEncoding.DecodeString(dataURL[len(prefix):])
	if err != nil {
		return "", err
	}
	name := "WiFiSpeedTest-card-" + time.Now().Format("2006-01-02-150405") + ".png"
	return a.save(name, "PNG image", "png", png)
}

func (a *App) save(name, filter, ext string, data []byte) (string, error) {
	if a.host.SaveFile != nil {
		return a.host.SaveFile(name, filter, ext, data)
	}
	dir := filepath.Join(a.dataDir, "exports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, name)
	return p, os.WriteFile(p, data, 0o644)
}

// prefs are small UI preferences (unit, mode) kept in the data folder, since
// the embedded page has no persistent browser storage.
func (a *App) prefs() map[string]string {
	out := map[string]string{}
	if b, err := os.ReadFile(filepath.Join(a.dataDir, "settings.json")); err == nil {
		_ = json.Unmarshal(b, &out)
	}
	return out
}

func (a *App) savePrefs(p map[string]string) error {
	allowed := map[string]bool{"unit": true, "mode": true, "lang": true, "updateCheck": true}
	clean := map[string]string{}
	for k, v := range p {
		if allowed[k] && len(v) < 32 {
			clean[k] = v
		}
	}
	b, _ := json.MarshalIndent(clean, "", "  ")
	if err := os.MkdirAll(a.dataDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(a.dataDir, "settings.json"), b, 0o644)
}

// lang is the UI language chosen in the page (saved in settings.json).
func (a *App) lang() i18n.Lang { return i18n.Parse(a.prefs()["lang"]) }

// result returns history entry id, or the latest result.
func (a *App) result(id string) (*model.TestResult, error) {
	if id != "" {
		if r, err := a.hist.Get(id); r != nil || err != nil {
			return r, err
		}
		// Not in history (e.g. a cancelled run): fall back to the latest.
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.last, nil
}

func (a *App) copyResult(id, unit string) (map[string]any, error) {
	r, err := a.result(id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errors.New(i18n.T(a.lang(), "app.runFirst"))
	}
	text := export.Share(r, unit, a.lang())
	copied := false
	if a.host.CopyText != nil {
		if err := a.host.CopyText(text); err != nil {
			return nil, err
		}
		copied = true
	}
	return map[string]any{"text": text, "copied": copied}, nil
}

// serverFile keeps the server picked in the UI (tried before the
// configured ones).
func (a *App) serverFile() string { return filepath.Join(a.dataDir, "server.json") }

func (a *App) selectedServer() *config.Server {
	b, err := os.ReadFile(a.serverFile())
	if err != nil {
		return nil
	}
	var s config.Server
	if json.Unmarshal(b, &s) != nil || s.Validate() != nil {
		return nil
	}
	return &s
}

// setServer saves the chosen server; nil goes back to the configured list.
func (a *App) setServer(s *config.Server) error {
	if s == nil {
		err := os.Remove(a.serverFile())
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := s.Validate(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	if err := os.MkdirAll(a.dataDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(a.serverFile(), b, 0o644)
}

// findServers fetches the public LibreSpeed list and ranks it by latency,
// reporting progress as "servers" events.
func (a *App) findServers() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	a.emit("servers", map[string]any{"stage": "fetch"})
	list, err := servers.Fetch(ctx, a.httpClient(), a.cfg.ServerListURL)
	if err != nil {
		a.emit("servers", map[string]any{"error": err.Error()})
		return
	}
	ranked := servers.Probe(ctx, list, 8, func(done, total int) {
		a.emit("servers", map[string]any{"stage": "probe", "done": done, "total": total})
	})
	a.emit("servers", map[string]any{"stage": "done", "list": ranked})
}

func (a *App) httpClient() *http.Client {
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
}

func (a *App) checkUpdate() {
	info, err := update.Check(context.Background(), a.httpClient(), "https://api.github.com", Version)
	if err != nil {
		a.emit("update", map[string]any{"error": err.Error()})
		return
	}
	a.emit("update", map[string]any{"info": info})
}

// openURL only opens this project's GitHub pages, so the page cannot be
// used to launch arbitrary programs or sites.
func (a *App) openURL(u string) error {
	if !strings.HasPrefix(u, "https://github.com/"+update.Repo+"/") {
		return errors.New("refusing to open " + u)
	}
	if a.host.OpenURL == nil {
		return errors.New("not supported")
	}
	return a.host.OpenURL(u)
}

// ExportsDir is the fallback folder used when no save dialog is available.
func (a *App) ExportsDir() string { return filepath.Join(a.dataDir, "exports") }
