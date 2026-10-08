package export

import (
	"strings"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func test(at time.Time, conn model.ConnectionType, down, up, ping float64) *model.TestResult {
	return &model.TestResult{
		StartedAt: at, Mode: "general", Network: &model.NetworkInfo{Connection: conn},
		Download: &model.SpeedResult{Mbps: down}, Upload: &model.SpeedResult{Mbps: up}, PingMs: ping, JitterMs: 2,
	}
}

func TestISPReport(t *testing.T) {
	day := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	all := []*model.TestResult{
		test(day.Add(9*time.Hour), model.ConnEthernet, 48, 10, 15),
		test(day.Add(10*time.Hour), model.ConnEthernet, 46, 10, 16),
		test(day.Add(20*time.Hour), model.ConnEthernet, 12, 9, 60), // busy evening
		test(day.Add(21*time.Hour), model.ConnWiFi, 20, 8, 40),
		{StartedAt: day.Add(22 * time.Hour), Mode: "quick", PingMs: 15, Network: &model.NetworkInfo{Connection: model.ConnWiFi}},
		test(day.Add(-48*time.Hour), model.ConnEthernet, 5, 1, 99), // before the period
	}
	plan := &model.Plan{DownMbps: 50, UpMbps: 10, LowPct: 50}
	rep := string(ISPReport(all, []*model.MonitorSummary{{StartedAt: day.Add(20 * time.Hour), DurationSec: 600, Verdict: "dropped 3 times"}},
		plan, ReportFilter{Since: day.Add(-time.Hour)}, Options{Unit: "Mbps", Lang: i18n.EN}))
	for _, want := range []string{
		"INTERNET QUALITY REPORT", "Tests:                     4",
		"Download:                  avg 31.5 Mbps · lowest 12.0 Mbps · highest 48.0 Mbps",
		"Average of plan:           download 63% · upload 92.5%",
		"Below 50% of plan:         2 of 4 tests",
		"06:00–12:00  tests: 2 ·", "18:00–24:00  tests: 2 ·",
		"which points to congestion at busy hours",
		"SLOWEST TESTS", "dropped 3 times", "1 of these tests were over Wi-Fi",
	} {
		if !strings.Contains(rep, want) {
			t.Errorf("report missing %q\n%s", want, rep)
		}
	}
	if strings.Contains(rep, "99 ms") {
		t.Error("test outside the period was included")
	}

	eth := string(ISPReport(all, nil, nil, ReportFilter{Since: day.Add(-time.Hour), Connection: "Ethernet"}, Options{Unit: "Mbps", Lang: i18n.ID}))
	if !strings.Contains(eth, "Jumlah tes:                3") || !strings.Contains(eth, "belum diisi") || strings.Contains(eth, "memakai Wi-Fi") {
		t.Errorf("filtered report:\n%s", eth)
	}
	if none := string(ISPReport(nil, nil, nil, ReportFilter{}, Options{Lang: i18n.EN})); !strings.Contains(none, "No speed tests") {
		t.Errorf("empty report:\n%s", none)
	}
}

func TestMonitorTXT(t *testing.T) {
	s := &model.MonitorSummary{
		StartedAt: time.Date(2026, 10, 1, 20, 0, 0, 0, time.UTC), DurationSec: 300, Target: "192.168.1.1", TargetLabel: "Router",
		Sent: 300, Received: 297, LossPct: 1, AvgMs: 20, MinMs: 10, MaxMs: 90, P95Ms: 40, Outages: 1, LongestOutage: 3,
		GatewaySent: 300, Verdict: "The connection dropped 1 times", Notes: []string{"beyond it"},
	}
	txt := string(MonitorTXT(s, Options{Lang: i18n.EN}))
	for _, want := range []string{"STABILITY MONITOR", "5 min", "192.168.1.x", "300 sent, 297 received", "1 (longest 3 s)", "Conclusion: The connection dropped", "- beyond it"} {
		if !strings.Contains(txt, want) {
			t.Errorf("monitor txt missing %q\n%s", want, txt)
		}
	}
	if strings.Contains(string(MonitorTXT(s, Options{Lang: i18n.EN, IncludeSensitive: true})), "192.168.1.x") {
		t.Error("IncludeSensitive should keep the address")
	}
}
