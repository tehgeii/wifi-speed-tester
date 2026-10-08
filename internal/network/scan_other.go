//go:build !windows

package network

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// scanWiFi on Linux uses NetworkManager's nmcli when available. For UI
// development without Wi-Fi, WST_FAKE_WIFI_SCAN may name a JSON file with
// a []model.WiFiNetwork to return instead.
func scanWiFi(ctx context.Context, rescan bool) ([]model.WiFiNetwork, error) {
	if p := os.Getenv("WST_FAKE_WIFI_SCAN"); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var out []model.WiFiNetwork
		return out, json.Unmarshal(b, &out)
	}
	args := []string{"-t", "-e", "yes", "-f", "IN-USE,SSID,BSSID,FREQ,SIGNAL", "device", "wifi", "list"}
	if rescan {
		args = append(args, "--rescan", "yes")
	}
	raw, err := exec.CommandContext(ctx, "nmcli", args...).Output()
	if err != nil {
		return nil, errors.New("Wi-Fi scan needs NetworkManager (nmcli)")
	}
	return parseNmcli(string(raw)), nil
}

// parseNmcli reads `nmcli -t -e yes -f IN-USE,SSID,BSSID,FREQ,SIGNAL`.
// Fields are ':'-separated with ':' inside values escaped as "\:".
func parseNmcli(out string) []model.WiFiNetwork {
	var nets []model.WiFiNetwork
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		var fields []string
		var cur strings.Builder
		for i := 0; i < len(line); i++ {
			switch {
			case line[i] == '\\' && i+1 < len(line):
				i++
				cur.WriteByte(line[i])
			case line[i] == ':':
				fields = append(fields, cur.String())
				cur.Reset()
			default:
				cur.WriteByte(line[i])
			}
		}
		fields = append(fields, cur.String())
		if len(fields) < 5 {
			continue
		}
		mhz, _ := strconv.Atoi(strings.Fields(fields[3] + " 0")[0])
		sig, _ := strconv.Atoi(fields[4])
		nets = append(nets, model.WiFiNetwork{
			SSID: fields[1], BSSID: strings.ToUpper(fields[2]), SignalPercent: sig, RSSI: sig/2 - 100,
			Channel: ChannelFromFrequency(mhz), Band: BandFromFrequency(mhz), FrequencyMHz: mhz,
			Connected: strings.TrimSpace(fields[0]) == "*",
		})
	}
	return nets
}
