package analysis

import (
	"strings"
	"testing"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func TestDNSAdvice(t *testing.T) {
	join := func(s []string) string { return strings.Join(s, " ") }
	slowSys := []model.DNSResult{
		{Label: "Cloudflare", MedianMs: 12, OK: 20},
		{Label: "Your DNS", System: true, MedianMs: 80, OK: 20},
	}
	if a := join(DNSAdvice(slowSys, i18n.EN)); !strings.Contains(a, "Cloudflare answered fastest") || !strings.Contains(a, "Hardware properties") {
		t.Fatalf("advice = %s", a)
	}
	fastSys := []model.DNSResult{
		{Label: "Your DNS", System: true, MedianMs: 10, OK: 20},
		{Label: "Cloudflare", MedianMs: 12, OK: 20},
	}
	if a := join(DNSAdvice(fastSys, i18n.EN)); !strings.Contains(a, "already among the fastest") {
		t.Fatalf("advice = %s", a)
	}
	// Close enough (under 15 ms difference): keep.
	close := []model.DNSResult{{Label: "Google", MedianMs: 20, OK: 20}, {Label: "Your DNS", System: true, MedianMs: 30, OK: 20}}
	if a := join(DNSAdvice(close, i18n.EN)); !strings.Contains(a, "no change needed") {
		t.Fatalf("advice = %s", a)
	}
	failing := []model.DNSResult{{Label: "Quad9", MedianMs: 20, OK: 20}, {Label: "Your DNS", System: true, MedianMs: 15, OK: 12, Failed: 8}}
	if a := join(DNSAdvice(failing, i18n.ID)); !strings.Contains(a, "gagal 8 dari 20") {
		t.Fatalf("advice = %s", a)
	}
	dead := []model.DNSResult{{Label: "Quad9", MedianMs: 20, OK: 20}, {Label: "Your DNS", System: true, Failed: 20}}
	if a := join(DNSAdvice(dead, i18n.EN)); !strings.Contains(a, "failed 20 of 20") {
		t.Fatalf("advice = %s", a)
	}
	if a := join(DNSAdvice([]model.DNSResult{{Label: "x", Failed: 3}}, i18n.EN)); !strings.Contains(a, "No DNS server answered") {
		t.Fatalf("advice = %s", a)
	}
}

func TestMonitorVerdict(t *testing.T) {
	s := &model.MonitorSummary{Sent: 300, Received: 300, DurationSec: 300}
	MonitorVerdict(s, i18n.ID)
	if s.Verdict != "Stabil: tidak ada putus dan lonjakan besar selama 5 menit." {
		t.Fatalf("verdict = %q", s.Verdict)
	}
	s = &model.MonitorSummary{Sent: 300, Received: 290, LossPct: 3.3, Outages: 2, LongestOutage: 7, DurationSec: 300, GatewaySent: 300, GatewayLost: 0}
	MonitorVerdict(s, i18n.EN)
	if s.Verdict != "The connection dropped 2 times (longest 7 s). This is what causes sudden lag or disconnects." || len(s.Notes) != 1 || !strings.Contains(s.Notes[0], "beyond it") {
		t.Fatalf("verdict = %q notes %v", s.Verdict, s.Notes)
	}
	s.GatewayLost = 9
	MonitorVerdict(s, i18n.EN)
	if !strings.Contains(strings.Join(s.Notes, " "), "router also missed replies (3%)") {
		t.Fatalf("notes = %v", s.Notes)
	}
	s = &model.MonitorSummary{Sent: 600, Received: 600, Spikes: 9, DurationSec: 600, Method: "tcp", Cancelled: true}
	MonitorVerdict(s, i18n.EN)
	if !strings.Contains(s.Verdict, "jumped 9 times") || len(s.Notes) != 2 {
		t.Fatalf("verdict = %q notes %v", s.Verdict, s.Notes)
	}
	if FormatDuration(90, i18n.EN) != "1 min 30 s" || FormatDuration(45, i18n.ID) != "45 detik" {
		t.Fatal("FormatDuration")
	}
}
