package engine

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
	"github.com/tehgeii/wifi-speed-tester/internal/testutil"
)

func testEngine(srv *testutil.SpeedServer, p *testutil.Pinger) *Engine {
	cfg := config.Default()
	cfg.Servers = []config.Server{{Name: "Local", DownloadURL: srv.DownloadURL(), UploadURL: srv.UploadURL()}}
	cfg.ConnectivityHost = "localhost"
	cfg.PingCount, cfg.PingIntervalMs = 10, 10
	cfg.DownloadSeconds, cfg.UploadSeconds, cfg.WarmupSeconds = 2, 2, 0.5
	cfg.Streams = 2
	return &Engine{
		Config: cfg,
		Detect: func(context.Context) (*model.NetworkInfo, error) {
			return &model.NetworkInfo{Connection: model.ConnWiFi, AdapterName: "Wi-Fi", Gateway: "192.168.1.1", IPv4: []string{"192.168.1.23"}}, nil
		},
		NewPinger: func() (measure.Pinger, error) { return p, nil },
		Client:    measure.NewHTTPClient(2),
	}
}

type recorder struct {
	mu     sync.Mutex
	events []Event
}

func (r *recorder) emit(e Event) { r.mu.Lock(); r.events = append(r.events, e); r.mu.Unlock() }

func (r *recorder) stages() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []string
	for _, e := range r.events {
		if e.Type == "stage" {
			out = append(out, e.Stage)
		}
	}
	return out
}

func TestFullRun(t *testing.T) {
	srv := testutil.NewSpeedServer(1_000_000)
	defer srv.Close()
	p := &testutil.Pinger{RTT: 2 * time.Millisecond, Lose: map[string]bool{"8.8.8.8": true}}
	rec := &recorder{}
	res := testEngine(srv, p).Run(context.Background(), Options{Mode: "general"}, rec.emit)

	if res.Failure != nil || res.Cancelled {
		t.Fatalf("unexpected failure: %+v", res.Failure)
	}
	if got := strings.Join(rec.stages(), ","); got != "connection,ping,download,upload" {
		t.Fatalf("stages = %s", got)
	}
	for _, c := range res.Checks {
		if c.Status == model.CheckFail || c.Status == model.CheckPending {
			t.Errorf("check %s = %s", c.Name, c.Status)
		}
	}
	if res.PingMs <= 0 || res.PacketLoss != 0 {
		t.Fatalf("ping %v loss %v (a fully silent target must not count as loss)", res.PingMs, res.PacketLoss)
	}
	var primary, gw int
	for _, p := range res.Pings {
		if p.PrimaryTarget {
			primary++
			if p.Target != "1.1.1.1" {
				t.Errorf("primary target = %s", p.Target)
			}
		}
		if p.IsGateway {
			gw++
		}
	}
	if primary != 1 || gw != 1 {
		t.Fatalf("primary=%d gateway=%d", primary, gw)
	}
	if res.Download.Mbps < 8 || res.Upload.Mbps < 8 {
		t.Fatalf("down %.1f up %.1f", res.Download.Mbps, res.Upload.Mbps)
	}
	if res.Download.LoadedSamples == 0 {
		t.Fatal("latency under load was not sampled")
	}
	if res.Quality == nil || res.Quality.Level == model.QualityUnknown {
		t.Fatalf("quality = %+v", res.Quality)
	}
	last := rec.events[len(rec.events)-1]
	if last.Type != "result" || last.Result != res {
		t.Fatal("final event must carry the result")
	}
}

func TestCancelDuringDownload(t *testing.T) {
	srv := testutil.NewSpeedServer(1_000_000)
	defer srv.Close()
	e := testEngine(srv, &testutil.Pinger{RTT: time.Millisecond})
	e.Config.DownloadSeconds = 30
	ctx, cancel := context.WithCancel(context.Background())
	start := time.Now()
	res := e.Run(ctx, Options{}, func(ev Event) {
		if ev.Type == "stage" && ev.Stage == StageDownload {
			time.AfterFunc(200*time.Millisecond, cancel)
		}
	})
	if !res.Cancelled || res.Upload != nil || res.Quality != nil || res.Failure != nil {
		t.Fatalf("cancelled result wrong: cancelled=%v upload=%v failure=%v", res.Cancelled, res.Upload, res.Failure)
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("cancellation was not prompt")
	}
}

func TestInternetUnavailable(t *testing.T) {
	srv := testutil.NewSpeedServer(0)
	srv.Close() // nothing listening
	res := testEngine(srv, &testutil.Pinger{RTT: time.Millisecond}).Run(context.Background(), Options{}, nil)
	if res.Failure == nil || res.Failure.Title != "Internet Unavailable" {
		t.Fatalf("failure = %+v", res.Failure)
	}
	if !strings.Contains(res.Failure.Suggestion, "Speed test cannot continue") {
		t.Fatalf("suggestion = %q", res.Failure.Suggestion)
	}
	if res.Download != nil || len(res.Pings) != 0 {
		t.Fatal("tests must not run without internet")
	}
}

func TestNoAdapter(t *testing.T) {
	srv := testutil.NewSpeedServer(0)
	defer srv.Close()
	e := testEngine(srv, &testutil.Pinger{})
	e.Detect = func(context.Context) (*model.NetworkInfo, error) { return nil, errors.New("none") }
	res := e.Run(context.Background(), Options{}, nil)
	if res.Failure == nil || res.Failure.Title != "No Network Connection" {
		t.Fatalf("failure = %+v", res.Failure)
	}
}

func TestServerErrorIsExplained(t *testing.T) {
	srv := testutil.NewSpeedServer(0)
	defer srv.Close()
	e := testEngine(srv, &testutil.Pinger{RTT: time.Millisecond})
	srv.Status = 500 // connectivity check accepts any HTTP response
	res := e.Run(context.Background(), Options{}, nil)
	if res.Failure == nil || res.Failure.Title != "Speed Test Failed" {
		t.Fatalf("failure = %+v", res.Failure)
	}
	if !strings.Contains(res.Failure.Reason, "returned an error") || res.Failure.Suggestion == "" {
		t.Fatalf("reason = %q", res.Failure.Reason)
	}
	if res.Quality == nil {
		t.Fatal("ping results should still be rated")
	}
}

func TestFallbackToSecondServer(t *testing.T) {
	good := testutil.NewSpeedServer(1_000_000)
	defer good.Close()
	bad := testutil.NewSpeedServer(0)
	bad.Status = 503
	defer bad.Close()
	e := testEngine(good, &testutil.Pinger{RTT: time.Millisecond})
	e.Config.Servers = append([]config.Server{{Name: "Broken", DownloadURL: bad.DownloadURL(), UploadURL: bad.UploadURL()}}, e.Config.Servers...)
	res := e.Run(context.Background(), Options{}, nil)
	if res.Download.Server != "Local" || res.Download.Mbps <= 0 || res.Upload.Server != "Local" {
		t.Fatalf("download from %q (%.1f), upload from %q", res.Download.Server, res.Download.Mbps, res.Upload.Server)
	}
}

func TestExplainProxyBlock(t *testing.T) {
	f := Explain("x", errors.New(`Get "https://speed.cloudflare.com/__down?bytes=0": Forbidden`))
	if !strings.Contains(f.Reason, "proxy") {
		t.Fatalf("reason = %q", f.Reason)
	}
}

// Seen on a hosted Windows runner: every ICMP target times out. Latency must
// then come from TCP connects to the test server instead of being empty.
func TestTCPFallbackWhenICMPBlocked(t *testing.T) {
	srv := testutil.NewSpeedServer(1_000_000)
	defer srv.Close()
	p := &testutil.Pinger{RTT: time.Millisecond, Lose: map[string]bool{"192.168.1.1": true, "1.1.1.1": true, "8.8.8.8": true}}
	res := testEngine(srv, p).Run(context.Background(), Options{}, nil)
	if res.PingMs <= 0 {
		t.Fatal("no latency measured")
	}
	var tcp *model.PingStats
	for i := range res.Pings {
		if res.Pings[i].Method == "tcp" {
			tcp = &res.Pings[i]
		}
	}
	if tcp == nil || !tcp.PrimaryTarget || tcp.Received == 0 {
		t.Fatalf("tcp fallback missing: %+v", res.Pings)
	}
	if res.Download.LoadedSamples == 0 {
		t.Fatal("latency under load should use the TCP pinger too")
	}
	if res.Quality.Level == model.QualityUnknown {
		t.Fatal("rating should be available with TCP latency")
	}
	if !strings.Contains(strings.Join(res.Quality.Notes, " "), "blocks ICMP") {
		t.Fatalf("notes = %v", res.Quality.Notes)
	}
}
