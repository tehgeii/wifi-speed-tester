package monitor

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
	"github.com/tehgeii/wifi-speed-tester/internal/testutil"
)

func ok(rtt float64) model.MonitorSample { return model.MonitorSample{OK: true, RttMs: rtt} }
func lost() model.MonitorSample          { return model.MonitorSample{} }

func TestSummarize(t *testing.T) {
	var xs []model.MonitorSample
	for i := 0; i < 20; i++ {
		xs = append(xs, ok(20))
	}
	xs = append(xs, ok(200), ok(20), lost(), lost(), lost(), lost(), ok(20), lost(), ok(22))
	gw := true
	for i := range xs {
		xs[i].T = float64(i)
		xs[i].GwOK = &gw
	}
	s := Summarize(xs, time.Second)
	if s.Sent != 29 || s.Received != 24 || s.Outages != 1 || s.LongestOutage != 4 || s.Spikes != 1 {
		t.Fatalf("summary = %+v", s)
	}
	if s.MaxMs != 200 || s.MinMs != 20 || s.P95Ms != 22 || s.GatewaySent != 29 || s.GatewayLost != 0 {
		t.Fatalf("summary = %+v", s)
	}
	if e := Summarize(nil, time.Second); e.Sent != 0 || e.LossPct != 0 {
		t.Fatalf("empty = %+v", e)
	}
}

func TestRunDurationAndCancel(t *testing.T) {
	p := &testutil.Pinger{RTT: time.Millisecond, LoseEvery: 3}
	calls := 0
	start := time.Now()
	samples, cut := Run(context.Background(), p, net.IPv4(1, 1, 1, 1), p, net.IPv4(192, 168, 1, 1),
		Options{Duration: 500 * time.Millisecond, Interval: 50 * time.Millisecond, OnSample: func(model.MonitorSample) { calls++ }})
	if cut || len(samples) < 8 || len(samples) > 11 || calls != len(samples) {
		t.Fatalf("samples=%d calls=%d cut=%v", len(samples), calls, cut)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("took %v", d)
	}
	for i := 1; i < len(samples); i++ {
		if samples[i].T <= samples[i-1].T || samples[i].GwOK == nil {
			t.Fatalf("samples not ordered or missing gateway: %+v", samples)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(120*time.Millisecond, cancel)
	samples, cut = Run(ctx, p, net.IPv4(1, 1, 1, 1), nil, nil, Options{Duration: 10 * time.Second, Interval: 50 * time.Millisecond})
	if !cut || len(samples) == 0 || len(samples) > 5 || samples[0].GwOK != nil {
		t.Fatalf("cancel: samples=%d cut=%v", len(samples), cut)
	}
}
