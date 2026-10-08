// Package engine runs a complete test: connectivity checks, ping, download,
// upload and analysis. It reports progress through events and never touches
// the UI directly.
package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
	"github.com/tehgeii/wifi-speed-tester/internal/network"
)

// Stage names used in events.
const (
	StageConnection = "connection"
	StagePing       = "ping"
	StageDownload   = "download"
	StageUpload     = "upload"
	StageDone       = "done"
)

// Event is a progress notification. Only the fields relevant to Type are set.
type Event struct {
	Type     string             `json:"type"` // stage | check | network | progress | ping | speed | result
	Stage    string             `json:"stage,omitempty"`
	Progress float64            `json:"progress"`
	Mbps     float64            `json:"mbps,omitempty"`
	RttMs    float64            `json:"rttMs,omitempty"`
	Check    *model.CheckItem   `json:"check,omitempty"`
	Network  *model.NetworkInfo `json:"network,omitempty"`
	Ping     *model.PingStats   `json:"ping,omitempty"`
	Speed    *model.SpeedResult `json:"speed,omitempty"`
	Result   *model.TestResult  `json:"result,omitempty"`
}

// Engine runs tests with a fixed configuration.
type Engine struct {
	Config config.Config
	// Detect, NewPinger and Client are replaceable for tests.
	Detect    func(context.Context) (*model.NetworkInfo, error)
	NewPinger func() (measure.Pinger, error)
	Client    *http.Client
}

// New returns an engine using the real network.
func New(cfg config.Config) *Engine {
	return &Engine{
		Config:    cfg,
		Detect:    network.Detect,
		NewPinger: measure.NewPinger,
		Client:    measure.NewHTTPClient(cfg.Streams),
	}
}

// Options selects what to run.
type Options struct {
	Mode string    // "general", "gaming", or "quick" (connectivity + ping only)
	Lang i18n.Lang // language of notes, tips, checks and errors
	// Server, when set, is tried first; the configured servers remain as
	// fallbacks.
	Server *config.Server
	// ExtraTargets are ping targets the user added in the app; they are
	// measured after the configured ones.
	ExtraTargets []config.PingTarget
	// Plan is the user's ISP plan; results are compared with it.
	Plan *model.Plan
}

type runner struct {
	e       *Engine
	cfg     config.Config
	emit    func(Event)
	res     *model.TestResult
	pinger  measure.Pinger
	primary net.IP // internet ping target used for latency under load
	server  int    // index of the server that worked
	errs    map[string]error
	lang    i18n.Lang
	// loadPinger creates the pinger used for latency under load: ICMP, or
	// TCP when ICMP is blocked.
	loadPinger func() (measure.Pinger, error)
}

// Run executes a full test. It always returns a result; when ctx is
// cancelled the result has Cancelled set and contains whatever finished.
// emit is called from the engine's goroutines and must not block for long.
func (e *Engine) Run(ctx context.Context, opts Options, emit func(Event)) *model.TestResult {
	if emit == nil {
		emit = func(Event) {}
	}
	mode := opts.Mode
	if mode != "gaming" && mode != "quick" {
		mode = "general"
	}
	cfg := e.Config
	if opts.Server != nil {
		cfg.Servers = append([]config.Server{*opts.Server}, cfg.Servers...)
	}
	if len(opts.ExtraTargets) > 0 {
		cfg.PingTargets = append(append([]config.PingTarget(nil), cfg.PingTargets...), opts.ExtraTargets...)
	}
	r := &runner{e: e, cfg: cfg, emit: emit, errs: map[string]error{}, lang: opts.Lang, loadPinger: e.NewPinger, res: &model.TestResult{
		ID:        newID(),
		StartedAt: time.Now(),
		Mode:      mode,
	}}
	if opts.Plan != nil && (opts.Plan.DownMbps > 0 || opts.Plan.UpMbps > 0) {
		pl := *opts.Plan
		pl.LowPct = cfg.PlanLowPct
		r.res.Plan = &pl
	}
	defer func() {
		r.res.FinishedAt = time.Now()
		if errors.Is(ctx.Err(), context.Canceled) {
			r.res.Cancelled = true
			r.res.Failure = nil
		}
		measured := r.res.PingMs > 0 || r.res.Download != nil && r.res.Download.Mbps > 0 || r.res.Upload != nil && r.res.Upload.Mbps > 0
		if !r.res.Cancelled && measured {
			r.res.Quality = analysis.Evaluate(r.res, mode, r.cfg.Profile(mode), r.lang)
		}
		emit(Event{Type: "result", Stage: StageDone, Progress: 1, Result: r.res})
	}()

	p, err := e.NewPinger()
	if err == nil {
		r.pinger = p
		defer p.Close()
	}

	if !r.connectivity(ctx) || ctx.Err() != nil {
		return r.res
	}
	r.ping(ctx, mode)
	if ctx.Err() != nil || mode == "quick" {
		return r.res
	}
	r.res.Download = r.speed(ctx, StageDownload)
	if ctx.Err() != nil {
		return r.res
	}
	r.res.Upload = r.speed(ctx, StageUpload)
	if ctx.Err() != nil {
		return r.res
	}
	if dErr, uErr := r.errs[StageDownload], r.errs[StageUpload]; dErr != nil && uErr != nil {
		r.res.Failure = Explain(r.lang, r.t("fail.speed.title"), dErr)
		r.res.Failure.Technical = "download: " + dErr.Error() + "; upload: " + uErr.Error()
	}
	return r.res
}

func newID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return time.Now().Format("20060102-150405") + "-" + hex.EncodeToString(b)
}

func (r *runner) t(key string, args ...any) string { return i18n.T(r.lang, key, args...) }

// check records the state of one pre-flight step (key: adapter, gateway,
// dns, internet).
func (r *runner) check(key string, st model.CheckStatus, detail string) {
	item := model.CheckItem{Key: key, Name: r.t("check." + key), Status: st, Detail: detail}
	for i := range r.res.Checks {
		if r.res.Checks[i].Key == key {
			r.res.Checks[i] = item
			r.emit(Event{Type: "check", Stage: StageConnection, Check: &item})
			return
		}
	}
	r.res.Checks = append(r.res.Checks, item)
	r.emit(Event{Type: "check", Stage: StageConnection, Check: &item})
}

func (r *runner) fail(title, reason, suggestion string, err error) bool {
	r.res.Failure = &model.Failure{Title: title, Reason: reason, Suggestion: suggestion}
	if err != nil {
		r.res.Failure.Technical = err.Error()
	}
	return false
}

// connectivity runs the pre-flight checks. It returns false when testing
// cannot continue.
func (r *runner) connectivity(ctx context.Context) bool {
	r.emit(Event{Type: "stage", Stage: StageConnection})
	for _, n := range []string{"adapter", "gateway", "dns", "internet"} {
		r.check(n, model.CheckPending, "")
	}

	info, err := r.e.Detect(ctx)
	if err != nil {
		r.check("adapter", model.CheckFail, err.Error())
		return r.fail(r.t("fail.noNetwork.title"), r.t("fail.noNetwork.reason"), r.t("fail.noNetwork.suggestion"), err)
	}
	r.res.Network = info
	r.emit(Event{Type: "network", Stage: StageConnection, Network: info})
	r.check("adapter", model.CheckOK, string(info.Connection)+" — "+info.AdapterName)

	if info.Gateway == "" {
		r.check("gateway", model.CheckWarn, r.t("check.noGateway"))
	} else if r.pinger == nil {
		r.check("gateway", model.CheckWarn, r.t("check.noPing"))
	} else {
		ok := false
		for i := 0; i < 3 && !ok && ctx.Err() == nil; i++ {
			_, err := r.pinger.Ping(ctx, net.ParseIP(info.Gateway), time.Second)
			ok = err == nil
		}
		if ok {
			r.check("gateway", model.CheckOK, r.t("check.gatewayOK", info.Gateway))
		} else {
			r.check("gateway", model.CheckWarn, r.t("check.gatewayNoReply", info.Gateway))
		}
	}
	if ctx.Err() != nil {
		return false
	}

	if d, err := network.CheckDNS(ctx, r.cfg.ConnectivityHost); err != nil {
		if ctx.Err() != nil {
			return false
		}
		r.check("dns", model.CheckFail, err.Error())
		r.check("internet", model.CheckFail, r.t("check.unavailable"))
		return r.fail(r.t("fail.offline.title"), r.t("fail.dns.reason"), r.t("fail.dns.suggestion"), err)
	} else {
		r.check("dns", model.CheckOK, r.t("check.dnsOK", analysis.FormatMs(float64(d)/float64(time.Millisecond))))
	}

	var lastErr error
	for i, s := range r.cfg.Servers {
		if lastErr = network.CheckInternet(ctx, r.e.Client, s.Ping()); lastErr == nil {
			r.server = i
			r.check("internet", model.CheckOK, r.t("check.internetOK", s.Name))
			return true
		}
		if ctx.Err() != nil {
			return false
		}
	}
	r.check("internet", model.CheckFail, r.t("check.unavailable"))
	f := Explain(r.lang, r.t("fail.offline.title"), lastErr)
	f.Reason = r.t("fail.offline.reasonPrefix") + f.Reason
	f.Suggestion = r.t("fail.offline.suggPrefix") + f.Suggestion
	r.res.Failure = f
	return false
}

// ping measures every configured target concurrently.
func (r *runner) ping(ctx context.Context, mode string) {
	r.emit(Event{Type: "stage", Stage: StagePing})
	defer func() {
		if r.primary == nil && ctx.Err() == nil {
			r.tcpFallback(ctx, mode)
		}
	}()
	if r.pinger == nil {
		return
	}
	type job struct {
		target config.PingTarget
		host   string
		ip     net.IP
		gw     bool
	}
	var jobs []job
	for _, t := range r.cfg.PingTargets {
		host, gw := t.Host, false
		if strings.EqualFold(host, "gateway") {
			if r.res.Network == nil || r.res.Network.Gateway == "" {
				continue
			}
			host, gw = r.res.Network.Gateway, true
		}
		ip, err := measure.ResolveIP(ctx, host)
		if err != nil {
			r.res.Pings = append(r.res.Pings, model.PingStats{Label: r.targetLabel(t), Target: host, IsGateway: gw, Error: err.Error()})
			continue
		}
		jobs = append(jobs, job{t, host, ip, gw})
	}
	if len(jobs) == 0 {
		return
	}

	count, interval := r.cfg.PingCount, r.cfg.PingInterval()
	if mode == "gaming" {
		count *= 2 // more samples for a steadier jitter figure
	}
	var mu sync.Mutex
	completed := 0
	results := make([]model.PingStats, len(jobs))
	var wg sync.WaitGroup
	for i, j := range jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			st, _ := measure.RunPing(ctx, r.pinger, r.targetLabel(j.target), j.host, j.ip, measure.PingOptions{
				Count: count, Interval: interval, Timeout: r.cfg.PingTimeout(),
				OnReply: func(_ int, rtt float64, ok bool) {
					mu.Lock()
					completed++
					ev := Event{Type: "progress", Stage: StagePing, Progress: float64(completed) / float64(count*len(jobs))}
					mu.Unlock()
					if ok && !j.gw {
						ev.RttMs = rtt
					}
					r.emit(ev)
				},
			})
			st.IsGateway, st.Method = j.gw, "icmp"
			results[i] = st
		}()
	}
	wg.Wait()

	// Headline figures come from the first internet target that answered;
	// loss is pooled over internet targets that answered at all (a target
	// with zero replies is more likely blocking ICMP than losing packets).
	sent, recv := 0, 0
	for i := range results {
		st := &results[i]
		if st.IsGateway || st.Received == 0 {
			continue
		}
		if r.primary == nil {
			st.PrimaryTarget = true
			r.primary = jobs[i].ip
			r.res.PingMs = st.AvgMs
			r.res.JitterMs = st.JitterMs
		}
		sent += st.Sent
		recv += st.Received
	}
	r.res.PacketLoss = measure.PacketLoss(sent, recv)
	r.res.Pings = append(r.res.Pings, results...)
	for i := range results {
		r.emit(Event{Type: "ping", Stage: StagePing, Ping: &results[i]})
	}
}

// targetLabel translates the built-in router label; labels the user typed
// are shown as written.
func (r *runner) targetLabel(t config.PingTarget) string {
	if strings.EqualFold(t.Host, "gateway") && (t.Label == "" || t.Label == "Local Gateway") {
		return r.t("target.gateway")
	}
	return t.Label
}

// tcpFallback measures latency with TCP connections to the test server
// when no ICMP target answered (common on corporate and cloud networks).
func (r *runner) tcpFallback(ctx context.Context, mode string) {
	u, err := url.Parse(r.cfg.Servers[r.server].Ping())
	if err != nil || u.Hostname() == "" {
		return
	}
	port := 443
	if p, err := strconv.Atoi(u.Port()); err == nil {
		port = p
	} else if u.Scheme == "http" {
		port = 80
	}
	ip, err := measure.ResolveIP(ctx, u.Hostname())
	if err != nil {
		return
	}
	count := r.cfg.PingCount
	if mode == "gaming" {
		count *= 2
	}
	tp := measure.TCPPinger{Port: port}
	st, _ := measure.RunPing(ctx, tp, r.t("target.tcp"), u.Hostname(), ip, measure.PingOptions{
		Count: count, Interval: r.cfg.PingInterval(), Timeout: r.cfg.PingTimeout(),
		OnReply: func(n int, rtt float64, ok bool) {
			ev := Event{Type: "progress", Stage: StagePing, Progress: float64(n+1) / float64(count)}
			if ok {
				ev.RttMs = rtt
			}
			r.emit(ev)
		},
	})
	st.Method = "tcp"
	if st.Received > 0 {
		st.PrimaryTarget = true
		r.primary = ip
		r.res.PingMs, r.res.JitterMs, r.res.PacketLoss = st.AvgMs, st.JitterMs, st.PacketLossPct
		r.loadPinger = func() (measure.Pinger, error) { return tp, nil }
	}
	r.res.Pings = append(r.res.Pings, st)
	r.emit(Event{Type: "ping", Stage: StagePing, Ping: &st})
}

// speed runs a download or upload test, falling back to the next server
// when one fails before transferring data.
func (r *runner) speed(ctx context.Context, stage string) *model.SpeedResult {
	r.emit(Event{Type: "stage", Stage: stage})
	dur := r.cfg.DownloadDuration()
	run := func(ctx context.Context, s config.Server, o measure.SpeedOptions) (model.SpeedResult, error) {
		return measure.Download(ctx, r.e.Client, s.DownloadFor, o)
	}
	if stage == StageUpload {
		dur = r.cfg.UploadDuration()
		run = func(ctx context.Context, s config.Server, o measure.SpeedOptions) (model.SpeedResult, error) {
			return measure.Upload(ctx, r.e.Client, s.Upload(), o)
		}
	}
	opts := measure.SpeedOptions{
		Duration: dur, Warmup: r.cfg.Warmup(), Streams: r.cfg.Streams,
		OnProgress: func(f, mbps float64) {
			r.emit(Event{Type: "progress", Stage: stage, Progress: f, Mbps: mbps})
		},
	}

	var res model.SpeedResult
	var err error
	n := len(r.cfg.Servers)
	for k := 0; k < n; k++ {
		i := (r.server + k) % n
		s := r.cfg.Servers[i]

		var loaded []float64
		loadCtx, stopLoad := context.WithCancel(ctx)
		var lw sync.WaitGroup
		if r.cfg.MeasureLoadedPing && r.primary != nil {
			if lp, perr := r.loadPinger(); perr == nil {
				lw.Add(1)
				go func() {
					defer lw.Done()
					defer lp.Close()
					select { // let the connection ramp up first
					case <-loadCtx.Done():
						return
					case <-time.After(opts.Warmup):
					}
					loaded = measure.SampleLatency(loadCtx, lp, r.primary, 250*time.Millisecond, r.cfg.PingTimeout())
				}()
			}
		}
		res, err = run(ctx, s, opts)
		stopLoad()
		lw.Wait()

		res.Server = s.Name
		if len(loaded) > 0 {
			res.LoadedPingMs = measure.Median(loaded)
			res.LoadedSamples = len(loaded)
		}
		if err == nil || ctx.Err() != nil {
			r.server = i
			break
		}
		if res.Bytes > 0 {
			break // partial result from this server; keep it
		}
	}
	if err != nil && ctx.Err() == nil {
		r.errs[stage] = err
		res.Error = Explain(r.lang, "", err).Reason
	}
	out := res
	r.emit(Event{Type: "speed", Stage: stage, Progress: 1, Speed: &out})
	return &out
}
