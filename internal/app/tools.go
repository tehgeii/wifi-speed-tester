package app

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/dnsbench"
	"github.com/tehgeii/wifi-speed-tester/internal/export"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
	"github.com/tehgeii/wifi-speed-tester/internal/monitor"
	"github.com/tehgeii/wifi-speed-tester/internal/network"
)

// MaxCustomTargets limits the ping targets a user can add in the app.
const MaxCustomTargets = 5

// readJSON and writeJSON keep small files in the data folder.
func (a *App) readJSON(name string, v any) bool {
	b, err := os.ReadFile(filepath.Join(a.dataDir, name))
	return err == nil && json.Unmarshal(b, v) == nil
}

func (a *App) writeJSON(name string, v any) error {
	if err := os.MkdirAll(a.dataDir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(a.dataDir, name+".tmp")
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(a.dataDir, name))
}

// ---- ISP plan --------------------------------------------------------------

// plan is the ISP plan from the settings (nil when not set).
func (a *App) plan() *model.Plan {
	p := a.prefs()
	num := func(k string) float64 {
		v, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(p[k]), ",", "."), 64)
		if err != nil || v <= 0 || v > 100000 {
			return 0
		}
		return v
	}
	pl := &model.Plan{DownMbps: num("planDown"), UpMbps: num("planUp"), LowPct: a.cfg.PlanLowPct}
	if pl.DownMbps == 0 && pl.UpMbps == 0 {
		return nil
	}
	return pl
}

// ---- custom ping targets ---------------------------------------------------

func (a *App) customTargets() []config.PingTarget {
	var ts []config.PingTarget
	if !a.readJSON("targets.json", &ts) {
		return nil
	}
	out := ts[:0]
	for _, t := range ts {
		if config.ValidTargetHost(t.Host) && t.Label != "" {
			out = append(out, t)
		}
	}
	return out
}

func (a *App) setTargets(ts []config.PingTarget) error {
	if len(ts) > MaxCustomTargets {
		return errors.New(i18n.T(a.lang(), "app.tooMany", MaxCustomTargets))
	}
	clean := make([]config.PingTarget, 0, len(ts))
	for _, t := range ts {
		t.Host = strings.TrimSpace(t.Host)
		t.Label = strings.TrimSpace(t.Label)
		if !config.ValidTargetHost(t.Host) {
			return errors.New(i18n.T(a.lang(), "app.badTarget"))
		}
		if t.Label == "" {
			t.Label = t.Host
		}
		if r := []rune(t.Label); len(r) > 40 {
			t.Label = string(r[:40])
		}
		clean = append(clean, t)
	}
	return a.writeJSON("targets.json", clean)
}

// monitorTargets are the internet targets the monitor can watch: the
// configured ones (except the router, which it always watches) and the
// user's own.
func (a *App) monitorTargets() []config.PingTarget {
	var out []config.PingTarget
	for _, t := range a.cfg.PingTargets {
		if !strings.EqualFold(t.Host, "gateway") {
			out = append(out, t)
		}
	}
	return append(out, a.customTargets()...)
}

// ---- busy state ------------------------------------------------------------

// claim reserves the network for one activity (speed test or monitor).
func (a *App) claim(kind string, cancel context.CancelFunc) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch {
	case a.running:
		return errors.New(i18n.T(a.lang(), "app.busy"))
	case a.monCancel != nil:
		return errors.New(i18n.T(a.lang(), "app.monBusy"))
	}
	if kind == "test" {
		a.running, a.cancel = true, cancel
	} else {
		a.monCancel = cancel
	}
	return nil
}

// ---- DNS test --------------------------------------------------------------

func (a *App) dnsTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	lang := a.lang()
	var system []string
	if info, err := a.eng.Detect(ctx); err == nil {
		system = info.DNS
	}
	list := dnsbench.Servers(system, a.cfg.DNSServers, lang)
	res := dnsbench.Run(ctx, list, dnsbench.Domains, dnsbench.Options{
		OnProgress: func(done, total int) { a.emit("dns", map[string]any{"done": done, "total": total}) },
	})
	a.emit("dns", map[string]any{"results": res, "advice": analysis.DNSAdvice(res, lang)})
}

// ---- Wi-Fi scan ------------------------------------------------------------

func (a *App) wifiScan(rescan bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	nets, err := network.ScanWiFi(ctx, rescan)
	if err != nil {
		a.emit("wifiscan", map[string]any{"error": err.Error()})
		return
	}
	a.emit("wifiscan", map[string]any{"report": analysis.AnalyzeScan(nets, a.lang())})
}

// ---- stability monitor -----------------------------------------------------

// MonitorRequest is sent by the UI.
type MonitorRequest struct {
	Minutes int    `json:"minutes"`
	Host    string `json:"host"` // "" = first configured internet target
}

func (a *App) startMonitor(req MonitorRequest) error {
	if req.Minutes < 1 || req.Minutes > int(monitor.MaxDuration/time.Minute) {
		req.Minutes = 5
	}
	targets := a.monitorTargets()
	target := config.PingTarget{Label: "1.1.1.1", Host: "1.1.1.1"}
	if len(targets) > 0 {
		target = targets[0]
	}
	for _, t := range targets {
		if t.Host == req.Host {
			target = t
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := a.claim("monitor", cancel); err != nil {
		cancel()
		return err
	}
	go func() {
		defer cancel()
		defer func() {
			a.mu.Lock()
			a.monCancel = nil
			a.mu.Unlock()
		}()
		sum, samples, err := a.runMonitor(ctx, target, time.Duration(req.Minutes)*a.monitorUnit)
		if err != nil {
			a.emit("monitor", map[string]any{"error": err.Error()})
			return
		}
		a.mu.Lock()
		a.lastMonitor = sum
		a.mu.Unlock()
		var all []*model.MonitorSummary
		a.readJSON("monitors.json", &all)
		all = append([]*model.MonitorSummary{sum}, all...)
		if len(all) > 20 {
			all = all[:20]
		}
		if err := a.writeJSON("monitors.json", all); err != nil {
			a.logf("monitors: %v", err)
		}
		a.emit("monitor", map[string]any{"summary": sum, "samples": samples})
	}()
	return nil
}

func (a *App) runMonitor(ctx context.Context, target config.PingTarget, dur time.Duration) (*model.MonitorSummary, []model.MonitorSample, error) {
	lang := a.lang()
	info, _ := a.eng.Detect(ctx)
	ip, err := measure.ResolveIP(ctx, target.Host)
	if err != nil {
		return nil, nil, err
	}
	sum := &model.MonitorSummary{StartedAt: time.Now(), Target: target.Host, TargetLabel: target.Label, Method: "icmp", Network: info}
	p, err := a.eng.NewPinger()
	if err != nil {
		p = nil
	}
	// Networks that block ICMP get the TCP fallback, as in the speed test.
	icmpOK := false
	for i := 0; p != nil && i < 3 && !icmpOK && ctx.Err() == nil; i++ {
		_, perr := p.Ping(ctx, ip, time.Second)
		icmpOK = perr == nil
	}
	var probe measure.Pinger = p
	if !icmpOK {
		srv := a.cfg.Servers[0]
		if s := a.selectedServer(); s != nil {
			srv = *s
		}
		u, perr := url.Parse(srv.Ping())
		if perr != nil {
			return nil, nil, perr
		}
		port := 443
		if n, e := strconv.Atoi(u.Port()); e == nil {
			port = n
		} else if u.Scheme == "http" {
			port = 80
		}
		if ip, err = measure.ResolveIP(ctx, u.Hostname()); err != nil {
			return nil, nil, err
		}
		probe = measure.TCPPinger{Port: port}
		sum.Method, sum.Target, sum.TargetLabel = "tcp", u.Hostname(), i18n.T(lang, "target.tcp")
	}
	var gwIP net.IP
	var gwPinger measure.Pinger
	if info != nil && info.Gateway != "" && p != nil {
		gwIP, gwPinger = net.ParseIP(info.Gateway), p
	}
	a.emit("monitor", map[string]any{"started": true, "target": sum.TargetLabel, "host": sum.Target, "method": sum.Method, "seconds": dur.Seconds()})
	start := time.Now()
	samples, cut := monitor.Run(ctx, probe, ip, gwPinger, gwIP, monitor.Options{
		Duration: dur,
		OnSample: func(s model.MonitorSample) { a.emit("monitor", map[string]any{"sample": s}) },
	})
	stats := monitor.Summarize(samples, time.Second)
	stats.StartedAt, stats.Target, stats.TargetLabel, stats.Method, stats.Network = sum.StartedAt, sum.Target, sum.TargetLabel, sum.Method, info
	stats.DurationSec = min(time.Since(start), dur).Seconds()
	stats.Cancelled = cut
	analysis.MonitorVerdict(&stats, lang)
	return &stats, samples, nil
}

func (a *App) stopMonitor() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.monCancel != nil {
		a.monCancel()
	}
}

func (a *App) monitorHistory() []*model.MonitorSummary {
	var all []*model.MonitorSummary
	a.readJSON("monitors.json", &all)
	return all
}

// ---- reports ---------------------------------------------------------------

// ReportRequest selects the tests for the ISP report.
type ReportRequest struct {
	Days             int    `json:"days"` // 0 = all
	Connection       string `json:"connection"`
	Unit             string `json:"unit"`
	IncludeSensitive bool   `json:"includeSensitive"`
}

func (a *App) ispReport(req ReportRequest) (string, error) {
	all, err := a.hist.List()
	if err != nil {
		return "", err
	}
	f := export.ReportFilter{Connection: req.Connection}
	if req.Days > 0 {
		f.Since = time.Now().AddDate(0, 0, -req.Days)
	}
	opts := export.Options{Unit: req.Unit, IncludeSensitive: req.IncludeSensitive, Lang: a.lang()}
	data := export.ISPReport(all, a.monitorHistory(), a.plan(), f, opts)
	return a.save("WiFiSpeedTest-ISP-report-"+time.Now().Format("2006-01-02")+".txt", "Text report", "txt", data)
}

func (a *App) exportMonitor(unitSensitive bool) (string, error) {
	a.mu.Lock()
	s := a.lastMonitor
	a.mu.Unlock()
	if s == nil {
		if all := a.monitorHistory(); len(all) > 0 {
			s = all[0]
		}
	}
	if s == nil {
		return "", errors.New(i18n.T(a.lang(), "app.runFirst"))
	}
	data := export.MonitorTXT(s, export.Options{IncludeSensitive: unitSensitive, Lang: a.lang()})
	return a.save("WiFiSpeedTest-monitor-"+s.StartedAt.Format("2006-01-02-150405")+".txt", "Text report", "txt", data)
}
