package analysis

import (
	"strings"
	"testing"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// res builds a result rated level on the given connection, with the router
// answering in gwMs.
func res(level model.QualityLevel, conn model.ConnectionType, down, gwMs float64) *model.TestResult {
	return &model.TestResult{
		Mode: "general", PingMs: 20, JitterMs: 2,
		Network:  &model.NetworkInfo{Connection: conn, WiFi: &model.WiFiInfo{SignalPercent: 60}},
		Download: &model.SpeedResult{Mbps: down}, Upload: &model.SpeedResult{Mbps: down / 4},
		Pings: []model.PingStats{
			{IsGateway: true, Sent: 10, Received: 10, AvgMs: gwMs},
			{PrimaryTarget: true, Sent: 10, Received: 10, AvgMs: 20},
		},
		Quality: &model.Quality{Level: level},
	}
}

func steps(kv ...any) []model.DiagStep {
	var out []model.DiagStep
	for i := 0; i < len(kv); i += 2 {
		out = append(out, model.DiagStep{Kind: kv[i].(string), Result: kv[i+1].(*model.TestResult)})
	}
	return out
}

func TestDiagnose(t *testing.T) {
	good := model.QualityGood
	poor := model.QualityPoor
	wifi, cable := model.ConnWiFi, model.ConnEthernet
	cases := []struct {
		name  string
		steps []model.DiagStep
		kind  string
	}{
		{"healthy", steps("now", res(good, wifi, 80, 3)), "healthy"},
		{"coverage", steps("now", res(poor, wifi, 5, 4), "near", res(good, wifi, 80, 3)), "coverage"},
		{"router: local delay near the router too", steps("now", res(poor, wifi, 5, 45), "near", res(poor, wifi, 8, 40)), "router"},
		{"isp: router fine, still poor near it", steps("now", res(poor, wifi, 5, 3), "near", res(poor, wifi, 6, 2)), "isp"},
		{"wifi: good on a cable", steps("now", res(poor, wifi, 5, 3), "near", res(poor, wifi, 6, 2), "cable", res(good, cable, 90, 1)), "wifi"},
		{"isp: poor on a cable", steps("now", res(poor, wifi, 5, 3), "cable", res(poor, cable, 6, 1)), "isp"},
		{"unclear: one Wi-Fi test", steps("now", res(poor, wifi, 5, 3)), "unclear"},
		{"isp: one cable test", steps("now", res(poor, cable, 5, 1)), "isp"},
	}
	for _, c := range cases {
		d := Diagnose(c.steps, i18n.EN)
		if d.Kind != c.kind {
			t.Errorf("%s: kind = %s, want %s", c.name, d.Kind, c.kind)
		}
		if d.Title == "" || d.Detail == "" || strings.HasPrefix(d.Title, "diag.") || len(d.Facts) != len(c.steps) {
			t.Errorf("%s: %+v", c.name, d)
		}
	}
}

func TestDiagnoseOfflineAndExtras(t *testing.T) {
	off := &model.TestResult{Mode: "general", Failure: &model.Failure{Title: "x"},
		Checks: []model.CheckItem{{Key: "adapter", Status: model.CheckOK}, {Key: "gateway", Status: model.CheckFail}}}
	if d := Diagnose(steps("now", off), i18n.ID); d.Kind != "offlineLocal" || !strings.Contains(d.Title, "router") {
		t.Fatalf("offline local = %+v", d)
	}
	off.Checks[1].Status = model.CheckOK
	if d := Diagnose(steps("now", off), i18n.EN); d.Kind != "offline" {
		t.Fatalf("offline = %+v", d)
	}

	// A "cable" step still on Wi-Fi is ignored, and slow DNS is pointed out.
	now := res(model.QualityPoor, model.ConnWiFi, 5, 3)
	now.DNSMs = 400
	d := Diagnose(steps("now", now, "cable", res(model.QualityGood, model.ConnWiFi, 90, 1)), i18n.EN)
	if d.Kind != "unclear" || !strings.Contains(strings.Join(d.Facts, " "), "still on Wi-Fi") {
		t.Fatalf("cable on Wi-Fi = %+v", d)
	}
	if d.Actions[len(d.Actions)-1] != "dns" {
		t.Fatalf("actions = %v", d.Actions)
	}
	if f := d.Facts[0]; !strings.Contains(f, "Where you are") || !strings.Contains(f, "Wi-Fi signal 60%") {
		t.Fatalf("fact = %q", f)
	}
}

func TestRegionAdvice(t *testing.T) {
	res := []model.RegionResult{
		{Label: "Jakarta", MedianMs: 12, OK: 5},
		{Label: "Singapore", MedianMs: 28, OK: 5},
		{Label: "Tokyo", OK: 0, Error: "timeout"},
	}
	adv := RegionAdvice(res, i18n.ID)
	if len(adv) != 2 || !strings.Contains(adv[0], "Jakarta (12 ms)") || !strings.Contains(adv[1], "Singapura 28 ms: sangat bagus") {
		t.Fatalf("advice = %v", adv)
	}
	res[1].MedianMs = 120
	if adv := RegionAdvice(res, i18n.EN); !strings.Contains(adv[1], "high for online games") {
		t.Fatalf("advice = %v", adv)
	}
	if adv := RegionAdvice([]model.RegionResult{{Label: "X"}}, i18n.EN); len(adv) != 1 || !strings.Contains(adv[0], "No region answered") {
		t.Fatalf("advice = %v", adv)
	}
}
