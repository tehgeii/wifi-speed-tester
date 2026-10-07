// Package network detects the active network adapter, Wi-Fi details and basic
// internet reachability.
package network

import (
	"context"
	"net"
	"strings"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Detect returns information about the adapter that carries the default
// route. Wi-Fi details are filled in when the OS exposes them.
func Detect(ctx context.Context) (*model.NetworkInfo, error) {
	info, err := detectAdapter()
	if err != nil {
		return nil, err
	}
	if info.Connection == model.ConnWiFi {
		if wifi, err := wifiInfo(info); err == nil && wifi != nil {
			info.WiFi = wifi
			if hotspotGateway(info.Gateway) {
				wifi.NoteKey = "wifi.hotspot"
			}
		} else if err != nil {
			info.WiFi = &model.WiFiInfo{NoteKey: "wifi.unavailable", Note: err.Error()}
		}
	}
	return info, nil
}

// classify infers the connection type from the adapter description.
func classify(base model.ConnectionType, description string) model.ConnectionType {
	d := strings.ToLower(description)
	for _, k := range []string{"remote ndis", "apple mobile device", "android", "iphone", "usb tethering", "rndis"} {
		if strings.Contains(d, k) {
			return model.ConnTethering
		}
	}
	for _, k := range []string{"vpn", "tap-windows", "wireguard", "wintun", "openvpn", "tailscale", "zerotier", "anyconnect", "fortinet"} {
		if strings.Contains(d, k) {
			return model.ConnVPN
		}
	}
	for _, k := range []string{"mobile broadband", "wwan", "lte", "5g modem"} {
		if strings.Contains(d, k) {
			return model.ConnCellular
		}
	}
	return base
}

// hotspotGateway reports whether gw is a default address used by common phone
// hotspots (iPhone 172.20.10.1, Android 192.168.43.1 and 192.168.49.1).
func hotspotGateway(gw string) bool {
	switch gw {
	case "172.20.10.1", "192.168.43.1", "192.168.49.1":
		return true
	}
	return false
}

func splitAddrs(ips []net.IP, info *model.NetworkInfo) {
	for _, ip := range ips {
		if ip == nil || ip.IsLoopback() {
			continue
		}
		if v4 := ip.To4(); v4 != nil {
			info.IPv4 = append(info.IPv4, v4.String())
		} else if !ip.IsLinkLocalUnicast() {
			info.IPv6 = append(info.IPv6, ip.String())
		}
	}
}

// BandFromFrequency maps a centre frequency in MHz to a band label.
func BandFromFrequency(mhz int) string {
	switch {
	case mhz >= 2400 && mhz < 2500:
		return "2.4 GHz"
	case mhz >= 4900 && mhz < 5900:
		return "5 GHz"
	case mhz >= 5925 && mhz < 7125:
		return "6 GHz"
	}
	return ""
}

// ChannelFromFrequency converts an IEEE 802.11 centre frequency to a channel.
func ChannelFromFrequency(mhz int) int {
	switch {
	case mhz == 2484:
		return 14
	case mhz >= 2412 && mhz < 2484:
		return (mhz - 2407) / 5
	case mhz >= 5925 && mhz < 7125:
		return (mhz - 5950) / 5
	case mhz >= 4900 && mhz < 5900:
		return (mhz - 5000) / 5
	}
	return 0
}

// BandFromChannel guesses the band when no frequency is known. 6 GHz
// channel numbers overlap with the others, so this is only a fallback.
func BandFromChannel(ch int) string {
	switch {
	case ch >= 1 && ch <= 14:
		return "2.4 GHz"
	case ch >= 32 && ch <= 177:
		return "5 GHz"
	}
	return ""
}
