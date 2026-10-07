package analysis

import (
	"strings"
	"testing"

	"github.com/tehgeii/wifi-speed-tester/internal/config"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
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
		q := Evaluate(c.r, "general", p, i18n.EN)
		if q.Level != c.want {
			t.Errorf("%s: level = %s, want %s (grades %+v)", c.name, q.Level, c.want, q.Grades)
		}
	}
}

func TestDeterministic(t *testing.T) {
	p := config.Default().Profile("general")
	r := result(48.6, 11.2, 18, 3, 0)
	a, b := Evaluate(r, "general", p, i18n.EN), Evaluate(r, "general", p, i18n.EN)
	if a.Level != b.Level || a.Summary != b.Summary {
		t.Fatal("same input produced different output")
	}
}

func TestGamingIgnoresBandwidthButUsesLoadedLatency(t *testing.T) {
	p := config.Default().Profile("gaming")
	r := result(3, 0.5, 15, 2, 0) // slow but responsive
	if q := Evaluate(r, "gaming", p, i18n.EN); q.Level != model.QualityExcellent {
		t.Fatalf("gaming level = %s, want EXCELLENT", q.Level)
	}
	r.Download.LoadedPingMs, r.Download.LoadedSamples = 140, 20 // +125 ms under load
	q := Evaluate(r, "gaming", p, i18n.EN)
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
	q := Evaluate(r, "general", p, i18n.EN)
	if q.Level != model.QualityGood {
		t.Fatalf("level = %s", q.Level)
	}
	if !strings.Contains(strings.Join(q.Notes, " "), "Upload was not measured") {
		t.Fatalf("notes = %v", q.Notes)
	}
	if q := Evaluate(&model.TestResult{}, "general", p, i18n.EN); q.Level != model.QualityUnknown {
		t.Fatalf("empty result level = %s", q.Level)
	}
}

func TestSummaryIsHonest(t *testing.T) {
	q := Evaluate(result(48.6, 11.2, 18, 3, 0), "general", config.Default().Profile("general"), i18n.EN)
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
	q := Evaluate(r, "general", config.Default().Profile("general"), i18n.EN)
	if !strings.Contains(strings.Join(q.Notes, " "), "beyond the router") {
		t.Fatalf("notes = %v", q.Notes)
	}
	r.Pings[0].AvgMs, r.Pings[0].PacketLossPct = 40, 5
	q = Evaluate(r, "general", config.Default().Profile("general"), i18n.EN)
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

func TestTooFewMetricsIsUnknown(t *testing.T) {
	// Seen on a CI runner: ICMP blocked, download rate-limited, only upload measured.
	r := &model.TestResult{Download: &model.SpeedResult{Error: "limited"}, Upload: &model.SpeedResult{Mbps: 2577}}
	q := Evaluate(r, "general", config.Default().Profile("general"), i18n.EN)
	if q.Level != model.QualityUnknown {
		t.Fatalf("level = %s, want UNKNOWN", q.Level)
	}
	if !strings.Contains(strings.Join(q.Notes, " "), "Too few results") {
		t.Fatalf("notes = %v", q.Notes)
	}
}

func TestIndonesian(t *testing.T) {
	r := result(48.6, 11.2, 18, 3, 0)
	r.Upload.Error = "x"
	q := Evaluate(r, "general", config.Default().Profile("general"), i18n.ID)
	if !strings.Contains(q.Summary, "Kualitas koneksi: BAGUS (profil Umum)") || !strings.Contains(q.Summary, "Download terukur 48.6 Mbps") {
		t.Fatalf("summary = %q", q.Summary)
	}
	if !strings.Contains(strings.Join(q.Notes, " "), "Upload tidak terukur") {
		t.Fatalf("notes = %v", q.Notes)
	}
	for _, g := range q.Grades {
		if g.Key == "" || g.Metric == "" || g.Metric == "metric."+g.Key {
			t.Fatalf("grade label missing: %+v", g)
		}
	}
}

func TestGradeDetails(t *testing.T) {
	q := Evaluate(result(48.6, 11.2, 18, 3, 0), "gaming", config.Default().Profile("gaming"), i18n.EN)
	for _, g := range q.Grades {
		switch g.Key {
		case "ping":
			if !g.Counted || g.Limits != [3]float64{20, 40, 70} || g.Unit != "ms" || g.HigherBetter {
				t.Fatalf("ping grade = %+v", g)
			}
		case "download":
			if g.Counted || !g.HigherBetter {
				t.Fatalf("download should be shown but not counted in gaming: %+v", g)
			}
		}
	}
}

func TestTips(t *testing.T) {
	has := func(q *model.Quality, sub string) bool { return strings.Contains(strings.Join(q.Tips, " "), sub) }
	p := config.Default().Profile("general")

	// Wi-Fi on 2.4 GHz, weak signal, jitter, big latency increase under load.
	r := result(60, 20, 30, 20, 0)
	r.Network = &model.NetworkInfo{Connection: model.ConnWiFi, WiFi: &model.WiFiInfo{SignalPercent: 45, Band: "2.4 GHz"}}
	r.Download.LoadedPingMs, r.Download.LoadedSamples = 200, 10
	q := Evaluate(r, "general", p, i18n.EN)
	for _, want := range []string{"SQM", "Move closer", "5 GHz", "LAN cable"} {
		if !has(q, want) {
			t.Errorf("missing tip %q in %v", want, q.Tips)
		}
	}

	// Healthy Ethernet: no tips at all.
	q = Evaluate(result(300, 50, 12, 2, 0), "general", p, i18n.EN)
	if len(q.Tips) != 0 {
		t.Errorf("healthy connection got tips: %v", q.Tips)
	}

	// VPN and a slow download.
	r = result(5, 20, 15, 2, 0)
	r.Network = &model.NetworkInfo{Connection: model.ConnVPN}
	q = Evaluate(r, "general", p, i18n.EN)
	if !has(q, "VPN") || !has(q, "Pause other downloads") {
		t.Errorf("tips = %v", q.Tips)
	}

	// High ping, healthy router → suggest a nearer server; loss beyond router → ISP.
	r = result(300, 50, 150, 2, 1)
	r.Pings = []model.PingStats{
		{IsGateway: true, Sent: 20, Received: 20, AvgMs: 2},
		{PrimaryTarget: true, Sent: 20, Received: 20, AvgMs: 150, PacketLossPct: 1},
	}
	q = Evaluate(r, "general", p, i18n.EN)
	if !has(q, "nearer server") || !has(q, "ISP") {
		t.Errorf("tips = %v", q.Tips)
	}
}

func TestQuickProfile(t *testing.T) {
	r := &model.TestResult{PingMs: 15, JitterMs: 2}
	q := Evaluate(r, "quick", config.Default().Profile("quick"), i18n.EN)
	if q.Level != model.QualityExcellent {
		t.Fatalf("quick level = %s (%v)", q.Level, q.Notes)
	}
	if strings.Contains(strings.Join(q.Notes, " "), "not measured") {
		t.Fatalf("quick ping must not complain about missing speed: %v", q.Notes)
	}
}
