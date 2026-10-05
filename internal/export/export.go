// Package export renders a test result as TXT, JSON or CSV. By default
// identifying details (SSID, BSSID, IP addresses) are masked so a report can
// be shared safely.
package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Options controls rendering.
type Options struct {
	Unit             string // "Mbps" or "MB/s"
	IncludeSensitive bool   // keep SSID/BSSID/IPs unmasked
}

// MaskIP keeps the network part of an address: 192.168.1.23 -> 192.168.1.x,
// 2001:db8::1 -> 2001:db8:x:x:x:x:x:x.
func MaskIP(s string) string {
	ip := net.ParseIP(s)
	if ip == nil {
		return "hidden"
	}
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.%d.x", v4[0], v4[1], v4[2])
	}
	parts := strings.Split(ip.To16().String(), ":")
	if len(parts) > 2 {
		return parts[0] + ":" + parts[1] + ":x:x:x:x:x:x"
	}
	return "hidden"
}

// MaskText keeps the first two characters: Home_5G -> Ho*****.
func MaskText(s string) string {
	r := []rune(s)
	if len(r) <= 2 {
		return strings.Repeat("*", len(r))
	}
	return string(r[:2]) + strings.Repeat("*", len(r)-2)
}

// Sanitize returns a deep copy of r with identifying fields masked.
func Sanitize(r *model.TestResult) *model.TestResult {
	b, _ := json.Marshal(r)
	var c model.TestResult
	_ = json.Unmarshal(b, &c)
	if n := c.Network; n != nil {
		for i := range n.IPv4 {
			n.IPv4[i] = MaskIP(n.IPv4[i])
		}
		for i := range n.IPv6 {
			n.IPv6[i] = MaskIP(n.IPv6[i])
		}
		if n.Gateway != "" {
			n.Gateway = MaskIP(n.Gateway)
		}
		for i := range n.DNS {
			n.DNS[i] = MaskIP(n.DNS[i])
		}
		if w := n.WiFi; w != nil {
			w.SSID = MaskText(w.SSID)
			if w.BSSID != "" {
				w.BSSID = "hidden"
			}
		}
	}
	for i := range c.Pings {
		if c.Pings[i].IsGateway {
			c.Pings[i].Target = MaskIP(c.Pings[i].Target)
		}
	}
	for i := range c.Checks {
		if c.Checks[i].Name == "Gateway" && c.Network != nil {
			c.Checks[i].Detail = strings.ReplaceAll(c.Checks[i].Detail, r.Network.Gateway, c.Network.Gateway)
		}
	}
	return &c
}

func prepare(r *model.TestResult, o Options) *model.TestResult {
	if o.IncludeSensitive {
		return r
	}
	return Sanitize(r)
}

// JSON renders the full result.
func JSON(r *model.TestResult, o Options) ([]byte, error) {
	return json.MarshalIndent(prepare(r, o), "", "  ")
}

func speedText(s *model.SpeedResult, unit string) string {
	switch {
	case s == nil:
		return "not measured"
	case s.Error != "" && s.Mbps <= 0:
		return "failed (" + s.Error + ")"
	case s.Error != "":
		return analysis.FormatSpeed(s.Mbps, unit) + " (incomplete: " + s.Error + ")"
	}
	return analysis.FormatSpeed(s.Mbps, unit)
}

// TXT renders a human-readable report.
func TXT(r *model.TestResult, o Options) []byte {
	r = prepare(r, o)
	var b bytes.Buffer
	line := func(k, v string) { fmt.Fprintf(&b, "%-22s %s\n", k+":", v) }

	b.WriteString("WIFI SPEED + PING TEST\n")
	b.WriteString(strings.Repeat("=", 46) + "\n\n")
	line("Date", r.StartedAt.Format("02 January 2006 15:04:05"))
	mode := "General"
	if r.Mode == "gaming" {
		mode = "Gaming"
	}
	line("Mode", mode)
	if n := r.Network; n != nil {
		line("Connection", string(n.Connection))
		line("Adapter", strings.TrimSpace(n.AdapterName+" ("+n.Description+")"))
		if len(n.IPv4) > 0 {
			line("IPv4", strings.Join(n.IPv4, ", "))
		}
		if len(n.IPv6) > 0 {
			line("IPv6", strings.Join(n.IPv6, ", "))
		}
		if n.Gateway != "" {
			line("Gateway", n.Gateway)
		}
		if len(n.DNS) > 0 {
			line("DNS", strings.Join(n.DNS, ", "))
		}
		if w := n.WiFi; w != nil {
			if w.SSID != "" {
				line("SSID", w.SSID)
			}
			if w.SignalPercent > 0 {
				line("Signal", strconv.Itoa(w.SignalPercent)+"%")
			}
			if w.Band != "" {
				line("Band", w.Band)
			}
			if w.Channel > 0 {
				line("Channel", strconv.Itoa(w.Channel))
			}
			if w.RxRateMbps > 0 {
				line("Wi-Fi Link Speed", fmt.Sprintf("%.0f Mbps (radio rate, not internet speed)", w.RxRateMbps))
			}
		}
	}
	b.WriteString("\n")
	if r.Cancelled {
		b.WriteString("Test was cancelled; results are incomplete.\n\n")
	}
	if f := r.Failure; f != nil {
		fmt.Fprintf(&b, "%s\nReason: %s\nSuggestion: %s\n\n", f.Title, f.Reason, f.Suggestion)
	}
	line("Download", speedText(r.Download, o.Unit))
	line("Upload", speedText(r.Upload, o.Unit))
	if r.PingMs > 0 {
		line("Ping", analysis.FormatMs(r.PingMs))
		line("Jitter", analysis.FormatMs(r.JitterMs))
		line("Packet Loss", analysis.FormatPct(r.PacketLoss))
	} else {
		line("Ping", "not measured")
	}
	for _, s := range []*model.SpeedResult{r.Download, r.Upload} {
		if s != nil && s.LoadedSamples > 0 {
			label := "Ping During Download"
			if s == r.Upload {
				label = "Ping During Upload"
			}
			line(label, analysis.FormatMs(s.LoadedPingMs))
		}
	}
	if q := r.Quality; q != nil {
		line("Quality", string(q.Level))
	}

	if len(r.Pings) > 0 {
		b.WriteString("\nPING DETAILS\n")
		for _, p := range r.Pings {
			if p.Sent == 0 {
				fmt.Fprintf(&b, "  %s (%s): %s\n", p.Label, p.Target, p.Error)
				continue
			}
			fmt.Fprintf(&b, "  %s (%s): %d sent, %d received, loss %s", p.Label, p.Target, p.Sent, p.Received, analysis.FormatPct(p.PacketLossPct))
			if p.Received == 0 {
				b.WriteString(" (no replies; the target may block ping)")
			} else {
				fmt.Fprintf(&b, ", min %s / avg %s / max %s, jitter %s", analysis.FormatMs(p.MinMs), analysis.FormatMs(p.AvgMs), analysis.FormatMs(p.MaxMs), analysis.FormatMs(p.JitterMs))
			}
			b.WriteString("\n")
		}
	}
	if q := r.Quality; q != nil && len(q.Notes) > 0 {
		b.WriteString("\nNOTES\n")
		for _, n := range q.Notes {
			b.WriteString("  - " + n + "\n")
		}
	}
	if s := r.Download; s != nil && s.Server != "" {
		fmt.Fprintf(&b, "\nTest server: %s\n", s.Server)
	}
	b.WriteString("\nMeasured values reflect this moment, this device and this test server.\nThey are not a guarantee of your ISP plan speed.\n")
	return b.Bytes()
}

// CSVHeader is the column list shared by single-result and history exports.
var CSVHeader = []string{"date", "mode", "connection", "ssid", "download_mbps", "upload_mbps", "ping_ms", "jitter_ms", "packet_loss_pct", "download_loaded_ping_ms", "upload_loaded_ping_ms", "quality", "server"}

func csvRow(r *model.TestResult) []string {
	f := func(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }
	sp := func(s *model.SpeedResult) (string, string) {
		if s == nil || s.Mbps <= 0 {
			return "", ""
		}
		l := ""
		if s.LoadedSamples > 0 {
			l = f(s.LoadedPingMs)
		}
		return f(s.Mbps), l
	}
	d, dl := sp(r.Download)
	u, ul := sp(r.Upload)
	conn, ssid, q, server := "", "", "", ""
	if r.Network != nil {
		conn = string(r.Network.Connection)
		if r.Network.WiFi != nil {
			ssid = r.Network.WiFi.SSID
		}
	}
	if r.Quality != nil {
		q = string(r.Quality.Level)
	}
	if r.Download != nil {
		server = r.Download.Server
	}
	ping, jit, loss := "", "", ""
	if r.PingMs > 0 {
		ping, jit, loss = f(r.PingMs), f(r.JitterMs), f(r.PacketLoss)
	}
	return []string{r.StartedAt.Format(time.RFC3339), r.Mode, conn, ssid, d, u, ping, jit, loss, dl, ul, q, server}
}

// CSV renders one or more results as rows. Speeds are always in Mbps so the
// column meaning never changes.
func CSV(results []*model.TestResult, o Options) []byte {
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	_ = w.Write(CSVHeader)
	for _, r := range results {
		_ = w.Write(csvRow(prepare(r, o)))
	}
	w.Flush()
	return b.Bytes()
}
