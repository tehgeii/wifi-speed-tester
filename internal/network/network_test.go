package network

import (
	"testing"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func TestFrequencyMapping(t *testing.T) {
	cases := []struct {
		mhz, ch int
		band    string
	}{{2412, 1, "2.4 GHz"}, {2437, 6, "2.4 GHz"}, {2484, 14, "2.4 GHz"}, {5180, 36, "5 GHz"}, {5745, 149, "5 GHz"}, {5955, 1, "6 GHz"}, {6115, 33, "6 GHz"}}
	for _, c := range cases {
		if got := ChannelFromFrequency(c.mhz); got != c.ch {
			t.Errorf("channel(%d) = %d, want %d", c.mhz, got, c.ch)
		}
		if got := BandFromFrequency(c.mhz); got != c.band {
			t.Errorf("band(%d) = %q, want %q", c.mhz, got, c.band)
		}
	}
	if BandFromChannel(6) != "2.4 GHz" || BandFromChannel(149) != "5 GHz" || BandFromChannel(0) != "" {
		t.Error("BandFromChannel")
	}
}

func TestClassify(t *testing.T) {
	cases := map[string]model.ConnectionType{
		"Intel(R) Ethernet Connection I219-V":       model.ConnEthernet,
		"Remote NDIS based Internet Sharing Device": model.ConnTethering,
		"Apple Mobile Device Ethernet":              model.ConnTethering,
		"TAP-Windows Adapter V9":                    model.ConnVPN,
		"WireGuard Tunnel":                          model.ConnVPN,
		"Generic Mobile Broadband Adapter":          model.ConnCellular,
	}
	for desc, want := range cases {
		if got := classify(model.ConnEthernet, desc); got != want {
			t.Errorf("classify(%q) = %s, want %s", desc, got, want)
		}
	}
	if !hotspotGateway("172.20.10.1") || hotspotGateway("192.168.1.1") {
		t.Error("hotspotGateway")
	}
}

func TestMarkHotspot(t *testing.T) {
	wifi := func(gw string, metered bool) *model.NetworkInfo {
		return &model.NetworkInfo{Connection: model.ConnWiFi, Gateway: gw, Metered: metered, WiFi: &model.WiFiInfo{SSID: "realme 16 Pro 5G", SignalPercent: 90}}
	}
	// A modern Android hotspot: random range, but metered.
	n := wifi("10.54.188.181", true)
	markHotspot(n)
	if n.Connection != model.ConnTethering || n.WiFi.SSID != "realme 16 Pro 5G" || n.WiFi.SignalPercent != 90 || n.WiFi.NoteKey != "wifi.hotspot" {
		t.Errorf("metered Wi-Fi: %+v %+v", n, n.WiFi)
	}
	// An iPhone: not reported as metered, but its fixed gateway.
	n = wifi("172.20.10.1", false)
	markHotspot(n)
	if n.Connection != model.ConnTethering {
		t.Errorf("iPhone gateway: %s", n.Connection)
	}
	// A home router stays Wi-Fi.
	n = wifi("192.168.1.1", false)
	markHotspot(n)
	if n.Connection != model.ConnWiFi || n.WiFi.NoteKey != "" {
		t.Errorf("home Wi-Fi: %s %q", n.Connection, n.WiFi.NoteKey)
	}
	// A note about missing Wi-Fi details is kept.
	n = wifi("10.0.0.1", true)
	n.WiFi.NoteKey = "wifi.unavailable"
	markHotspot(n)
	if n.Connection != model.ConnTethering || n.WiFi.NoteKey != "wifi.unavailable" {
		t.Errorf("unavailable note: %s %q", n.Connection, n.WiFi.NoteKey)
	}
	// A metered cable stays a cable (the flag still tells the app).
	n = &model.NetworkInfo{Connection: model.ConnEthernet, Gateway: "192.168.1.1", Metered: true}
	markHotspot(n)
	if n.Connection != model.ConnEthernet || !n.Metered || n.WiFi != nil {
		t.Errorf("metered Ethernet: %+v", n)
	}
}
