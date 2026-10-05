package analysis

import (
	"strings"
	"testing"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func result(down, up, ping, jitter, loss float64) *model.TestResult {
	return &model.TestResult{
		Download: &model.SpeedResult{Mbps: down},
		Upload:   &model.SpeedResult{Mbps: up},
		PingMs:   ping, JitterMs: jitter, PacketLoss: loss,
	}
}

func TestGeneralLevels(t *testing.T) {
	p := config.Default().Profile("general")
	cases := []struct {
		name string
		r    *model.TestResult
		want model.QualityLevel
	}{
		{"excellent", result(300, 50, 12, 2, 0), model.QualityExcellent},
		{"good (brief example)", result(48.6, 11.2, 18, 3, 0), model.QualityGood},
		{"fair from upload", result(48.6, 3, 18, 3, 0), model.QualityFair},
		{"poor from ping", result(48.6, 11.2, 180, 3, 0), model.QualityPoor},
		{"poor from slow download", result(4, 11.2, 18, 3, 0), model.QualityPoor},
		{"unstable from loss", result(300, 50, 12, 2, 6), model.QualityUnstable},
		{"unstable from jitter", result(300, 50, 12, 60, 0), model.QualityUnstable},
	}
	for _, c := range cases {
		q := Evaluate(c.r, "general", p)
		if q.Level != c.want {
			t.Errorf("%s: level = %s, want %s (grades %+v)", c.name, q.Level, c.want, q.Grades)
		}
	}
}

func TestDeterministic(t *testing.T) {
	p := config.Default().Profile("general")
	r := result(48.6, 11.2, 18, 3, 0)
	a, b := Evaluate(r, "general", p), Evaluate(r, "general", p)
	if a.Level != b.Level || a.Summary != b.Summary {
		t.Fatal("same input produced different output")
	}
}

func TestGamingIgnoresBandwidthButUsesLoadedLatency(t *testing.T) {
	p := config.Default().Profile("gaming")
	r := result(3, 0.5, 15, 2, 0) // slow but responsive
	if q := Evaluate(r, "gaming", p); q.Level != model.QualityExcellent {
		t.Fatalf("gaming level = %s, want EXCELLENT", q.Level)
	}
	r.Download.LoadedPingMs, r.Download.LoadedSamples = 140, 20 // +125 ms under load
	q := Evaluate(r, "gaming", p)
	if q.Level != model.QualityPoor {
		t.Fatalf("gaming level with bufferbloat = %s, want POOR", q.Level)
	}
	if !strings.Contains(strings.Join(q.Notes, " "), "under load") {
		t.Fatalf("missing under-load note: %v", q.Notes)
	}
}

func TestMissingMetrics(t *testing.T) {
	p := config.Default().Profile("general")
	r := result(50, 0, 18, 3, 0)
	r.Upload.Error = "failed"
	q := Evaluate(r, "general", p)
	if q.Level != model.QualityGood {
		t.Fatalf("level = %s", q.Level)
	}
	if !strings.Contains(strings.Join(q.Notes, " "), "Upload was not measured") {
		t.Fatalf("notes = %v", q.Notes)
	}
	if q := Evaluate(&model.TestResult{}, "general", p); q.Level != model.QualityUnknown {
		t.Fatalf("empty result level = %s", q.Level)
	}
}

func TestSummaryIsHonest(t *testing.T) {
	q := Evaluate(result(48.6, 11.2, 18, 3, 0), "general", config.Default().Profile("general"))
	if !strings.Contains(q.Summary, "Measured download 48.6 Mbps") || !strings.Contains(q.Summary, "not a guaranteed ISP speed") {
		t.Fatalf("summary = %q", q.Summary)
	}
}

func TestLocationNotes(t *testing.T) {
	r := result(50, 10, 90, 3, 0)
	r.Pings = []model.PingStats{
		{IsGateway: true, Sent: 20, Received: 20, AvgMs: 2},
		{PrimaryTarget: true, Sent: 20, Received: 20, AvgMs: 90},
	}
	q := Evaluate(r, "general", config.Default().Profile("general"))
	if !strings.Contains(strings.Join(q.Notes, " "), "beyond the router") {
		t.Fatalf("notes = %v", q.Notes)
	}
	r.Pings[0].AvgMs, r.Pings[0].PacketLossPct = 40, 5
	q = Evaluate(r, "general", config.Default().Profile("general"))
	if !strings.Contains(strings.Join(q.Notes, " "), "local network or Wi-Fi") {
		t.Fatalf("notes = %v", q.Notes)
	}
}

func TestFormat(t *testing.T) {
	if FormatSpeed(80, "MB/s") != "10.00 MB/s" || FormatSpeed(48.64, "Mbps") != "48.6 Mbps" || FormatSpeed(940.2, "Mbps") != "940 Mbps" {
		t.Fatal("speed format")
	}
	if FormatPct(0) != "0%" || FormatPct(4) != "4%" || FormatPct(2.5) != "2.5%" {
		t.Fatal("pct format")
	}
}
