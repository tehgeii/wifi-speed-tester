package analysis

import (
	"strings"
	"testing"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func byKey(list []model.Activity) map[string]model.Activity {
	m := map[string]model.Activity{}
	for _, a := range list {
		m[a.Key] = a
	}
	return m
}

func goodResult() *model.TestResult {
	return &model.TestResult{
		Mode: "general", PingMs: 18, JitterMs: 2, PacketLoss: 0,
		Pings:    []model.PingStats{{Received: 20}},
		Download: &model.SpeedResult{Mbps: 95, LoadedSamples: 10, LoadedPingMs: 26},
		Upload:   &model.SpeedResult{Mbps: 40, LoadedSamples: 10, LoadedPingMs: 30},
	}
}

func TestSuitabilityAllGood(t *testing.T) {
	acts := byKey(Suitability(goodResult(), true, i18n.EN))
	if len(acts) != 7 {
		t.Fatalf("got %d activities", len(acts))
	}
	for k, a := range acts {
		if a.Status != ActOK {
			t.Errorf("%s = %s (%s), want ok", k, a.Status, a.Reason)
		}
	}
	if !strings.Contains(acts["bigdl"].Reason, "50 GB") || !strings.Contains(acts["bigdl"].Reason, "h") {
		t.Errorf("big download reason = %q", acts["bigdl"].Reason)
	}
	if acts["game"].Name != "Online games" {
		t.Errorf("name = %q", acts["game"].Name)
	}
}

func TestSuitabilityWhileRunning(t *testing.T) {
	r := goodResult()
	r.Download, r.Upload = nil, nil
	acts := byKey(Suitability(r, false, i18n.EN))
	for _, k := range []string{"browse", "hd", "uhd", "call", "live", "bigdl"} {
		if acts[k].Status != ActPending {
			t.Errorf("%s = %s, want pending before the speed test", k, acts[k].Status)
		}
	}
	// Latency under load is still to come, so even games wait.
	if acts["game"].Status != ActPending {
		t.Errorf("game = %s", acts["game"].Status)
	}

	// A clearly bad value is reported at once, without waiting.
	r.PingMs = 250
	acts = byKey(Suitability(r, false, i18n.ID))
	if acts["game"].Status != ActBad || acts["game"].Reason != "Ping 250 ms (idealnya ≤ 60 ms)" {
		t.Errorf("game = %+v", acts["game"])
	}
	if acts["call"].Status != ActBad {
		t.Errorf("call = %+v", acts["call"])
	}
	if acts["hd"].Status != ActPending {
		t.Errorf("hd = %+v", acts["hd"])
	}
}

func TestSuitabilityThresholds(t *testing.T) {
	r := goodResult()
	r.Download.Mbps = 10 // enough for 1080p, not for 4K
	acts := byKey(Suitability(r, true, i18n.EN))
	if acts["hd"].Status != ActOK || acts["uhd"].Status != ActBad || acts["uhd"].Reason != "Download 10.0 Mbps (ideal ≥ 25 Mbps)" {
		t.Fatalf("hd=%+v uhd=%+v", acts["hd"], acts["uhd"])
	}
	r.Download.Mbps = 20
	if a := byKey(Suitability(r, true, i18n.EN))["uhd"]; a.Status != ActWarn {
		t.Fatalf("uhd at 20 Mbps = %+v, want warn", a)
	}
	r.Download.Mbps = 95
	r.Download.LoadedPingMs = 120 // +102 ms under load: bad for games
	if a := byKey(Suitability(r, true, i18n.EN))["game"]; a.Status != ActBad || a.Reason != "Latency Under Load +102 ms (ideal ≤ +30 ms)" {
		t.Fatalf("game = %+v", a)
	}
}

func TestSuitabilityMissingAndQuick(t *testing.T) {
	r := goodResult()
	r.Upload = &model.SpeedResult{Error: "failed"}
	acts := byKey(Suitability(r, true, i18n.EN))
	if acts["live"].Status != ActNA || acts["call"].Status != ActNA || acts["hd"].Status != ActOK {
		t.Fatalf("live=%+v call=%+v hd=%+v", acts["live"], acts["call"], acts["hd"])
	}

	q := &model.TestResult{Mode: "quick", PingMs: 30, JitterMs: 3, Pings: []model.PingStats{{Received: 10}}}
	acts = byKey(Suitability(q, true, i18n.EN))
	if len(acts) != 3 || acts["browse"].Status != ActOK || acts["call"].Status != ActOK || acts["game"].Status != ActOK {
		t.Fatalf("quick = %+v", acts)
	}
}

func TestBloatGrade(t *testing.T) {
	cases := map[float64]string{0: "A+", 4.9: "A+", 5: "A", 29: "A", 30: "B", 59: "B", 60: "C", 199: "C", 200: "D", 399: "D", 400: "F", 900: "F"}
	for ms, want := range cases {
		if got := BloatGrade(ms); got != want {
			t.Errorf("BloatGrade(%v) = %s, want %s", ms, got, want)
		}
	}
	if g := GradeOf(goodResult()); g != "A" { // +12 ms
		t.Errorf("GradeOf = %s", g)
	}
	if g := GradeOf(&model.TestResult{PingMs: 20}); g != "" {
		t.Errorf("GradeOf without load = %q", g)
	}
}

func TestFormatDurationHours(t *testing.T) {
	cases := map[float64]string{3600: "1 h", 5400: "1 h 30 min", 7142: "1 h 59 min", 3599: "59 min 59 s", 7199: "2 h"}
	for sec, want := range cases {
		if got := FormatDuration(sec, i18n.EN); got != want {
			t.Errorf("FormatDuration(%v) = %q, want %q", sec, got, want)
		}
	}
	if got := FormatDuration(5400, i18n.ID); got != "1 jam 30 menit" {
		t.Errorf("ID = %q", got)
	}
}
