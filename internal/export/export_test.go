package export

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func sample() *model.TestResult {
	return &model.TestResult{
		ID: "x", StartedAt: time.Date(2026, 10, 4, 9, 30, 0, 0, time.UTC), Mode: "general",
		Network: &model.NetworkInfo{
			Connection: model.ConnWiFi, AdapterName: "Wi-Fi", Description: "Intel Wi-Fi 6 AX201",
			IPv4: []string{"192.168.1.23"}, IPv6: []string{"2001:db8:1:2::5"}, Gateway: "192.168.1.1", DNS: []string{"192.168.1.1"},
			WiFi: &model.WiFiInfo{SSID: "Home_5G", BSSID: "AA:BB:CC:DD:EE:FF", SignalPercent: 88, Band: "5 GHz", RxRateMbps: 866},
		},
		Checks:   []model.CheckItem{{Name: "Gateway", Status: model.CheckOK, Detail: "192.168.1.1 reachable"}},
		Pings:    []model.PingStats{{Label: "Local Gateway", Target: "192.168.1.1", IsGateway: true, Sent: 20, Received: 20, AvgMs: 2}},
		Download: &model.SpeedResult{Mbps: 48.6, Server: "Cloudflare"},
		Upload:   &model.SpeedResult{Mbps: 11.2},
		PingMs:   18, JitterMs: 3, PacketLoss: 0,
		Quality: &model.Quality{Level: model.QualityGood},
	}
}

func TestMasking(t *testing.T) {
	if MaskIP("192.168.1.23") != "192.168.1.x" || MaskText("Home_5G") != "Ho*****" || MaskText("ab") != "**" {
		t.Fatal("mask helpers")
	}
	r := sample()
	b, err := JSON(r, Options{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, secret := range []string{"192.168.1.23", "Home_5G", "AA:BB:CC", "2001:db8:1:2::5", "192.168.1.1\""} {
		if strings.Contains(s, secret) {
			t.Errorf("masked JSON still contains %q", secret)
		}
	}
	if r.Network.WiFi.SSID != "Home_5G" {
		t.Fatal("Sanitize must not modify the original")
	}
	full, _ := JSON(r, Options{IncludeSensitive: true})
	if !strings.Contains(string(full), "Home_5G") {
		t.Fatal("IncludeSensitive should keep the SSID")
	}
	var back model.TestResult
	if err := json.Unmarshal(full, &back); err != nil || back.Download.Mbps != 48.6 {
		t.Fatalf("JSON round trip: %v", err)
	}
}

func TestTXT(t *testing.T) {
	txt := string(TXT(sample(), Options{Unit: "Mbps"}))
	for _, want := range []string{"WIFI SPEED + PING TEST", "04 October 2026", "Wi-Fi", "48.6 Mbps", "11.2 Mbps", "18 ms", "Packet Loss:", "GOOD", "Ho*****", "not a guarantee"} {
		if !strings.Contains(txt, want) {
			t.Errorf("TXT missing %q\n%s", want, txt)
		}
	}
	if mb := string(TXT(sample(), Options{Unit: "MB/s"})); !strings.Contains(mb, "6.08 MB/s") {
		t.Errorf("MB/s report wrong:\n%s", mb)
	}
}

func TestCSV(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(string(CSV([]*model.TestResult{sample()}, Options{}))), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "date,") {
		t.Fatalf("csv = %q", lines)
	}
	if !strings.Contains(lines[1], "48.60") || !strings.Contains(lines[1], "GOOD") || strings.Contains(lines[1], "Home_5G") {
		t.Fatalf("row = %q", lines[1])
	}
}

func TestShare(t *testing.T) {
	s := Share(sample(), "Mbps", i18n.EN)
	for _, want := range []string{"WiFi Speed + Ping Test (General)", "↓ 48.6 Mbps   ↑ 11.2 Mbps", "Ping 18 ms · Jitter 3.0 ms · Loss 0%", "Quality: GOOD · Wi-Fi 5 GHz"} {
		if !strings.Contains(s, want) {
			t.Errorf("share missing %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "Home_5G") || strings.Contains(s, "192.168") {
		t.Fatalf("share leaks identifiers:\n%s", s)
	}
	if id := Share(sample(), "MB/s", i18n.ID); !strings.Contains(id, "Kualitas: BAGUS") || !strings.Contains(id, "6.08 MB/s") {
		t.Fatalf("id share:\n%s", id)
	}
}

func TestTXTIndonesian(t *testing.T) {
	txt := string(TXT(sample(), Options{Unit: "Mbps", Lang: i18n.ID}))
	for _, want := range []string{"Tanggal:", "04-10-2026", "Koneksi:", "Kualitas:", "Bukan jaminan"} {
		if !strings.Contains(txt, want) {
			t.Errorf("TXT (id) missing %q\n%s", want, txt)
		}
	}
}
