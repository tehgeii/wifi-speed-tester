//go:build !windows

package network

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Linux implementation, used for development and the CLI. It reads the
// default route from /proc/net/route.
func detectAdapter() (*model.NetworkInfo, error) {
	ifName, gw, err := defaultRoute()
	if err != nil {
		return nil, err
	}
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return nil, err
	}
	info := &model.NetworkInfo{AdapterName: ifName, Description: ifName, Gateway: gw}
	base := model.ConnEthernet
	if _, err := os.Stat("/sys/class/net/" + ifName + "/wireless"); err == nil {
		base = model.ConnWiFi
	} else if strings.HasPrefix(ifName, "tun") || strings.HasPrefix(ifName, "wg") || strings.HasPrefix(ifName, "tap") {
		base = model.ConnVPN
	} else if strings.HasPrefix(ifName, "usb") || strings.HasPrefix(ifName, "rndis") {
		base = model.ConnTethering
	} else if strings.HasPrefix(ifName, "ww") {
		base = model.ConnCellular
	}
	info.Connection = classify(base, info.Description)

	addrs, _ := iface.Addrs()
	var ips []net.IP
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok {
			ips = append(ips, n.IP)
		}
	}
	splitAddrs(ips, info)
	info.DNS = resolvConf()
	return info, nil
}

func defaultRoute() (string, string, error) {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Scan() // header
	best, bestGW, bestMetric := "", "", -1
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 7 || fields[1] != "00000000" {
			continue
		}
		metric, _ := strconv.Atoi(fields[6])
		if bestMetric >= 0 && metric >= bestMetric {
			continue
		}
		gw := ""
		if b, err := hex.DecodeString(fields[2]); err == nil && len(b) == 4 {
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, binary.LittleEndian.Uint32(b))
			gw = ip.String()
		}
		best, bestGW, bestMetric = fields[0], gw, metric
	}
	if best == "" {
		return "", "", errors.New("no default route found")
	}
	return best, bestGW, nil
}

func resolvConf() []string {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) >= 2 && fields[0] == "nameserver" {
			out = append(out, fields[1])
		}
	}
	return out
}

func wifiInfo(info *model.NetworkInfo) (*model.WiFiInfo, error) {
	w := &model.WiFiInfo{}
	if out, err := exec.Command("iw", "dev", info.AdapterName, "link").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			k, v, ok := strings.Cut(strings.TrimSpace(line), ":")
			if !ok {
				if strings.HasPrefix(line, "Connected to ") {
					w.BSSID = strings.ToUpper(strings.Fields(line)[2])
				}
				continue
			}
			v = strings.TrimSpace(v)
			switch k {
			case "SSID":
				w.SSID = v
			case "freq":
				f, _ := strconv.ParseFloat(strings.Fields(v)[0], 64)
				w.FrequencyMHz = int(f)
			case "signal":
				w.RSSI, _ = strconv.Atoi(strings.Fields(v)[0])
			case "rx bitrate":
				w.RxRateMbps, _ = strconv.ParseFloat(strings.Fields(v)[0], 64)
			case "tx bitrate":
				w.TxRateMbps, _ = strconv.ParseFloat(strings.Fields(v)[0], 64)
			}
		}
	} else {
		return nil, errors.New("install 'iw' for Wi-Fi details")
	}
	if w.RSSI != 0 {
		// Same linear mapping Windows uses: -100 dBm = 0 %, -50 dBm = 100 %.
		w.SignalPercent = min(100, max(0, 2*(w.RSSI+100)))
	}
	w.Band = BandFromFrequency(w.FrequencyMHz)
	w.Channel = ChannelFromFrequency(w.FrequencyMHz)
	return w, nil
}
