package app

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func TestScrubAddresses(t *testing.T) {
	got := scrubAddresses("dial tcp 192.168.1.23:443 and [2001:db8::1]:80: refused; 10.0.0.1")
	if strings.Contains(got, "192.168.1.23") || strings.Contains(got, "2001:db8") || strings.Contains(got, "10.0.0.1") ||
		!strings.Contains(got, "192.168.x.x:443") || !strings.Contains(got, "refused") {
		t.Fatalf("scrubbed = %q", got)
	}
}

func TestProblemReport(t *testing.T) {
	a, _, _ := newTestApp(t)
	r := a.problemReport()
	text := r["text"].(string)
	if !strings.Contains(text, "WiFi Speed + Ping Tester") || !strings.Contains(text, "Connection: Ethernet") || !strings.Contains(text, "Last test: none") {
		t.Fatalf("report = %s", text)
	}

	a.last = &model.TestResult{
		Mode: "general", StartedAt: time.Now(),
		Network: &model.NetworkInfo{IPv4: []string{"192.168.1.23"}, WiFi: &model.WiFiInfo{SSID: "SecretHome", BSSID: "aa:bb:cc:dd:ee:ff"}},
		Failure: &model.Failure{Title: "Download failed", Technical: "dial tcp 203.0.113.9:443: i/o timeout"},
	}
	r = a.problemReport()
	text = r["text"].(string)
	for _, secret := range []string{"192.168.1.23", "SecretHome", "aa:bb:cc", "203.0.113.9"} {
		if strings.Contains(text, secret) || strings.Contains(r["url"].(string), url.QueryEscape(secret)) {
			t.Errorf("report leaks %q:\n%s", secret, text)
		}
	}
	if !strings.Contains(text, "Error: Download failed") || !strings.Contains(text, "203.0.x.x:443") {
		t.Fatalf("report = %s", text)
	}
	u, err := url.Parse(r["url"].(string))
	if err != nil || !strings.HasPrefix(r["url"].(string), "https://github.com/tehgeii/wifi-speed-tester/issues/new?") {
		t.Fatalf("url = %v (%v)", r["url"], err)
	}
	if q := u.Query(); q.Get("title") != "Problem report: Download failed" || !strings.Contains(q.Get("body"), text) {
		t.Fatalf("issue = %v", q)
	}
	// The page opens it with openUrl, which only allows this project.
	var opened string
	a.host.OpenURL = func(u string) error { opened = u; return nil }
	if _, err := call(t, a, "openUrl", r["url"]); err != nil || opened != r["url"] {
		t.Fatalf("openUrl: %v (opened %q)", err, opened)
	}
	if copied, err := a.copyProblem(text); copied || err != nil {
		t.Fatalf("copy without a host = %v, %v", copied, err)
	}
}
