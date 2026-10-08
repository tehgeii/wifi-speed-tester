package analysis

import (
	"strings"
	"testing"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func ap(ssid string, ch, rssi int, connected bool) model.WiFiNetwork {
	band := "2.4 GHz"
	if ch > 14 {
		band = "5 GHz"
	}
	return model.WiFiNetwork{SSID: ssid, BSSID: ssid + "-bssid", Channel: ch, Band: band, RSSI: rssi, Connected: connected}
}

func TestAnalyzeScanRecommendsQuieterChannel(t *testing.T) {
	nets := []model.WiFiNetwork{
		ap("Home", 6, -45, true),
		ap("A", 6, -50, false), ap("B", 6, -55, false), ap("C", 7, -60, false), ap("D", 5, -58, false),
		ap("E", 1, -85, false),
		ap("Home", 44, -60, false), // own 5 GHz radio, same name: not interference
	}
	rep := AnalyzeScan(nets, i18n.EN)
	if rep.Current == nil || rep.Current.Channel != 6 {
		t.Fatalf("current = %+v", rep.Current)
	}
	if rep.Recommended != 11 {
		t.Fatalf("recommended = %d (channels %+v)", rep.Recommended, rep.Channels)
	}
	adv := strings.Join(rep.Advice, " ")
	if !strings.Contains(adv, "Channel 11 looks less crowded") || !strings.Contains(adv, "about 2 nearby networks") {
		t.Fatalf("advice = %s", adv)
	}
	var cur, rec bool
	for _, c := range rep.Channels {
		if c.Channel == 6 && c.Current {
			cur = true
		}
		if c.Channel == 11 && c.Recommended {
			rec = true
		}
		if c.Band == "5 GHz" && c.Channel == 44 && c.Networks != 0 {
			t.Fatal("own network counted as a neighbour")
		}
	}
	if !cur || !rec {
		t.Fatalf("channels = %+v", rep.Channels)
	}
	if rep.Networks[0].RSSI != -45 {
		t.Fatal("networks must be sorted strongest first")
	}
}

func TestAnalyzeScanKeepsGoodChannel(t *testing.T) {
	nets := []model.WiFiNetwork{ap("Home", 11, -50, true), ap("A", 1, -60, false), ap("B", 6, -62, false)}
	rep := AnalyzeScan(nets, i18n.ID)
	if rep.Recommended != 0 || !strings.Contains(strings.Join(rep.Advice, " "), "Tidak perlu diganti") {
		t.Fatalf("report = %+v", rep)
	}
}

func TestAnalyzeScanCrowded24AndNotConnected(t *testing.T) {
	var nets []model.WiFiNetwork
	for i := 0; i < 9; i++ {
		nets = append(nets, ap(string(rune('A'+i)), []int{1, 6, 11}[i%3], -70, false))
	}
	nets = append(nets, ap("Home", 6, -50, true))
	if adv := strings.Join(AnalyzeScan(nets, i18n.EN).Advice, " "); !strings.Contains(adv, "2.4 GHz is crowded here (10 networks)") {
		t.Fatalf("advice = %s", adv)
	}
	nets[len(nets)-1].Connected = false
	nets = append(nets, ap("Far5", 36, -80, false))
	adv := strings.Join(AnalyzeScan(nets, i18n.EN).Advice, " ")
	if !strings.Contains(adv, "On 2.4 GHz, channel") || !strings.Contains(adv, "On 5 GHz, channel 40") {
		t.Fatalf("advice = %s", adv)
	}
	if adv := AnalyzeScan(nil, i18n.EN).Advice; len(adv) != 1 || !strings.Contains(adv[0], "No Wi-Fi networks") {
		t.Fatalf("advice = %v", adv)
	}
}
