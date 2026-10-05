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
