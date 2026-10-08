//go:build !windows

package network

import "testing"

func TestParseNmcli(t *testing.T) {
	out := "*:Home\\:5G:AA\\:BB\\:CC\\:DD\\:EE\\:FF:5180 MHz:82\n :Cafe:11\\:22\\:33\\:44\\:55\\:66:2437 MHz:40\n"
	nets := parseNmcli(out)
	if len(nets) != 2 {
		t.Fatalf("nets = %+v", nets)
	}
	a := nets[0]
	if a.SSID != "Home:5G" || a.BSSID != "AA:BB:CC:DD:EE:FF" || a.Channel != 36 || a.Band != "5 GHz" || !a.Connected || a.SignalPercent != 82 {
		t.Fatalf("first = %+v", a)
	}
	if nets[1].Channel != 6 || nets[1].Connected {
		t.Fatalf("second = %+v", nets[1])
	}
}
