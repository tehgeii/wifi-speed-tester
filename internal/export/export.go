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
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Options controls rendering.
type Options struct {
	Unit             string    // "Mbps" or "MB/s"
	IncludeSensitive bool      // keep SSID/BSSID/IPs unmasked
	Lang             i18n.Lang // report language (TXT)
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
		if c.Checks[i].Key == "gateway" && c.Network != nil {
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

func speedText(s *model.SpeedResult, unit string, T func(string, ...any) string) string {
	switch {
	case s == nil:
		return T("rep.notMeasured")
	case s.Error != "" && s.Mbps <= 0:
		return T("rep.failed", s.Error)
	case s.Error != "":
		return T("rep.incomplete", analysis.FormatSpeed(s.Mbps, unit), s.Error)
	}
	return analysis.FormatSpeed(s.Mbps, unit)
}

func dateText(t time.Time, lang i18n.Lang) string {
	if lang == i18n.ID {
		return t.Format("02-01-2006 15:04:05")
	}
	return t.Format("02 January 2006 15:04:05")
}

// TXT renders a human-readable report in o.Lang.
func TXT(r *model.TestResult, o Options) []byte {
	r = prepare(r, o)
	T := func(key string, args ...any) string { return i18n.T(o.Lang, key, args...) }
	var b bytes.Buffer
	line := func(k, v string) { fmt.Fprintf(&b, "%-22s %s\n", k+":", v) }

	b.WriteString(T("rep.title") + "\n")
	b.WriteString(strings.Repeat("=", 46) + "\n\n")
	line(T("rep.date"), dateText(r.StartedAt, o.Lang))
	line(T("rep.mode"), analysis.ProfileLabel(r.Mode, o.Lang))
	if n := r.Network; n != nil {
		line(T("rep.connection"), string(n.Connection))
		line(T("rep.adapter"), strings.TrimSpace(n.AdapterName+" ("+n.Description+")"))
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
				line(T("rep.signal"), strconv.Itoa(w.SignalPercent)+"%")
			}
			if w.Band != "" {
				line(T("rep.band"), w.Band)
			}
			if w.Channel > 0 {
				line(T("rep.channel"), strconv.Itoa(w.Channel))
			}
			if w.RxRateMbps > 0 {
				line(T("rep.linkSpeed"), T("rep.linkSpeedVal", w.RxRateMbps))
			}
		}
	}
	b.WriteString("\n")
	if r.Cancelled {
		b.WriteString(T("rep.cancelled") + "\n\n")
	}
	if f := r.Failure; f != nil {
		fmt.Fprintf(&b, "%s\n%s: %s\n%s: %s\n\n", f.Title, T("rep.reason"), f.Reason, T("rep.suggestion"), f.Suggestion)
	}
	if r.Mode != "quick" {
		line(T("metric.download"), speedText(r.Download, o.Unit, T))
		line(T("metric.upload"), speedText(r.Upload, o.Unit, T))
	}
	if r.PingMs > 0 {
		line(T("metric.ping"), analysis.FormatMs(r.PingMs))
		line(T("metric.jitter"), analysis.FormatMs(r.JitterMs))
		line(T("metric.packetLoss"), analysis.FormatPct(r.PacketLoss))
	} else {
		line(T("metric.ping"), T("rep.notMeasured"))
	}
	for _, s := range []*model.SpeedResult{r.Download, r.Upload} {
		if s != nil && s.LoadedSamples > 0 {
			label := T("rep.pingDownload")
			if s == r.Upload {
				label = T("rep.pingUpload")
			}
			line(label, analysis.FormatMs(s.LoadedPingMs))
		}
	}
	if q := r.Quality; q != nil {
		line(T("rep.quality"), T("level."+string(q.Level)))
	}

	if len(r.Pings) > 0 {
		b.WriteString("\n" + T("rep.pingDetails") + "\n")
		for _, p := range r.Pings {
			if p.Sent == 0 {
				fmt.Fprintf(&b, "  %s (%s): %s\n", p.Label, p.Target, p.Error)
				continue
			}
			fmt.Fprintf(&b, "  %s (%s): %s", p.Label, p.Target, T("rep.pingLine", p.Sent, p.Received, analysis.FormatPct(p.PacketLossPct)))
			if p.Received == 0 {
				b.WriteString(T("rep.noReplies"))
			} else {
				b.WriteString(", " + T("rep.pingStats", analysis.FormatMs(p.MinMs), analysis.FormatMs(p.AvgMs), analysis.FormatMs(p.MaxMs), analysis.FormatMs(p.JitterMs)))
			}
			b.WriteString("\n")
		}
	}
	if q := r.Quality; q != nil {
		for _, sec := range []struct {
			title string
			items []string
		}{{T("rep.notes"), q.Notes}, {T("rep.tips"), q.Tips}} {
			if len(sec.items) == 0 {
				continue
			}
			b.WriteString("\n" + sec.title + "\n")
			for _, n := range sec.items {
				b.WriteString("  - " + n + "\n")
			}
		}
	}
	if s := r.Download; s != nil && s.Server != "" {
		b.WriteString("\n" + T("rep.server", s.Server) + "\n")
	}
	b.WriteString("\n" + T("rep.footer") + "\n")
	return b.Bytes()
}

// Share renders a short summary for chat apps. It never includes the SSID
// or addresses.
func Share(r *model.TestResult, unit string, lang i18n.Lang) string {
	T := func(key string, args ...any) string { return i18n.T(lang, key, args...) }
	sp := func(s *model.SpeedResult) string {
		if s == nil || s.Mbps <= 0 {
			return "—"
		}
		return analysis.FormatSpeed(s.Mbps, unit)
	}
	lines := []string{T("share.head") + " (" + analysis.ProfileLabel(r.Mode, lang) + ")"}
	if r.Mode != "quick" {
		lines = append(lines, T("share.speeds", sp(r.Download), sp(r.Upload)))
	}
	if r.PingMs > 0 {
		lines = append(lines, T("share.latency", analysis.FormatMs(r.PingMs), analysis.FormatMs(r.JitterMs), analysis.FormatPct(r.PacketLoss)))
	}
	q := string(model.QualityUnknown)
	if r.Quality != nil {
		q = string(r.Quality.Level)
	}
	ql := T("share.quality", T("level."+q))
	if n := r.Network; n != nil {
		ql += " · " + string(n.Connection)
		if n.WiFi != nil && n.WiFi.Band != "" {
			ql += " " + n.WiFi.Band
		}
	}
	lines = append(lines, ql, dateText(r.StartedAt, lang))
	return strings.Join(lines, "\n")
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
