package export

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// MonitorTXT renders a stability monitor run.
func MonitorTXT(s *model.MonitorSummary, o Options) []byte {
	T := func(key string, args ...any) string { return i18n.T(o.Lang, key, args...) }
	var b bytes.Buffer
	line := func(k, v string) { fmt.Fprintf(&b, "%-22s %s\n", k+":", v) }
	b.WriteString(T("rep.monTitle") + "\n" + strings.Repeat("=", 46) + "\n\n")
	line(T("rep.date"), dateText(s.StartedAt, o.Lang))
	line(T("rep.duration"), analysis.FormatDuration(s.DurationSec, o.Lang))
	target := s.Target
	if !o.IncludeSensitive {
		target = maskIfPrivate(target)
	}
	line(T("rep.target"), strings.TrimSpace(s.TargetLabel+" ("+target+")"))
	if n := s.Network; n != nil {
		line(T("rep.connection"), string(n.Connection))
	}
	b.WriteString("\n")
	line(T("rep.probes"), T("rep.probesVal", s.Sent, s.Received, analysis.FormatPct(math.Round(s.LossPct*10)/10)))
	if s.Received > 0 {
		line(T("rep.latency"), T("rep.latencyVal", analysis.FormatMs(s.AvgMs), analysis.FormatMs(s.MinMs), analysis.FormatMs(s.MaxMs), analysis.FormatMs(s.P95Ms), analysis.FormatMs(s.JitterMs)))
	}
	line(T("rep.spikes"), fmt.Sprint(s.Spikes))
	line(T("rep.outages"), T("rep.outagesVal", s.Outages, analysis.FormatDuration(s.LongestOutage, o.Lang)))
	if s.GatewaySent > 0 {
		line(T("rep.router"), fmt.Sprintf("%d / %d", s.GatewayLost, s.GatewaySent))
	}
	b.WriteString("\n" + T("rep.verdict") + ": " + s.Verdict + "\n")
	for _, n := range s.Notes {
		b.WriteString("  - " + n + "\n")
	}
	b.WriteString("\n" + T("rep.footer") + "\n")
	return b.Bytes()
}

// maskIfPrivate hides private (LAN) addresses; public hosts stay visible.
func maskIfPrivate(h string) string {
	if strings.Count(h, ".") == 3 && (strings.HasPrefix(h, "10.") || strings.HasPrefix(h, "192.168.") || strings.HasPrefix(h, "172.")) {
		return MaskIP(h)
	}
	return h
}

// ReportFilter selects the tests that go into an ISP report.
type ReportFilter struct {
	Since      time.Time // zero = all
	Connection string    // "" = all
}

// ispData is what the ISP report shows, worked out once for the TXT and
// the PDF versions.
type ispData struct {
	tests    []*model.TestResult // in time order
	monitors []*model.MonitorSummary
	conn     string
	plan     *model.Plan

	down, up, ping, jit, loss []float64

	lowPct     float64 // tests below lowPct % of the plan (download)…
	below      int     // …and how many
	slots      [4]ispSlot
	fast, slow int // slot indexes, -1 when unknown
	peakDrop   float64
	worst      []*model.TestResult // slowest downloads first, up to 5
	wifiTests  int
}

type ispSlot struct{ down, ping []float64 }

// buildISP selects the tests for the report and computes its figures.
// Quick Ping results are left out of the speed figures.
func buildISP(all []*model.TestResult, monitors []*model.MonitorSummary, plan *model.Plan, f ReportFilter) *ispData {
	d := &ispData{plan: plan, fast: -1, slow: -1}
	for _, r := range all {
		if r.Mode == "quick" || r.Cancelled || !r.StartedAt.After(f.Since) {
			continue
		}
		if f.Connection != "" && (r.Network == nil || string(r.Network.Connection) != f.Connection) {
			continue
		}
		if (r.Download == nil || r.Download.Mbps <= 0) && r.PingMs <= 0 {
			continue
		}
		d.tests = append(d.tests, r)
	}
	sort.Slice(d.tests, func(a, b int) bool { return d.tests[a].StartedAt.Before(d.tests[b].StartedAt) })
	// Stability monitor runs in the period are evidence too, even without
	// speed tests.
	for _, m := range monitors {
		if len(d.monitors) < 10 && m.StartedAt.After(f.Since) && (f.Connection == "" || (m.Network != nil && string(m.Network.Connection) == f.Connection)) {
			d.monitors = append(d.monitors, m)
		}
	}
	if len(d.tests) == 0 {
		return d
	}
	d.conn = f.Connection
	if d.conn == "" {
		d.conn = connectionsOf(d.tests)
	}

	speedOf := func(s *model.SpeedResult) float64 {
		if v := mbps(s); v > 0 {
			return v
		}
		return -1
	}
	ifPing := func(v func(*model.TestResult) float64) func(*model.TestResult) float64 {
		return func(r *model.TestResult) float64 {
			if r.PingMs > 0 {
				return v(r)
			}
			return -1
		}
	}
	d.down = collect(d.tests, func(r *model.TestResult) float64 { return speedOf(r.Download) })
	d.up = collect(d.tests, func(r *model.TestResult) float64 { return speedOf(r.Upload) })
	d.ping = collect(d.tests, ifPing(func(r *model.TestResult) float64 { return r.PingMs }))
	d.jit = collect(d.tests, ifPing(func(r *model.TestResult) float64 { return r.JitterMs }))
	d.loss = collect(d.tests, ifPing(func(r *model.TestResult) float64 { return r.PacketLoss }))

	if plan != nil && plan.DownMbps > 0 {
		low := plan.LowPct
		if low <= 0 {
			low = 50
		}
		d.lowPct = low
		for _, v := range d.down {
			if model.PlanPercent(v, plan.DownMbps) < low {
				d.below++
			}
		}
	}

	// Time of day: four 6-hour slots.
	for _, r := range d.tests {
		i := r.StartedAt.Hour() / 6
		if v := mbps(r.Download); v > 0 {
			d.slots[i].down = append(d.slots[i].down, v)
		}
		if r.PingMs > 0 {
			d.slots[i].ping = append(d.slots[i].ping, r.PingMs)
		}
	}
	for i, s := range d.slots {
		if len(s.down) == 0 {
			continue
		}
		if d.fast < 0 || avg(s.down) > avg(d.slots[d.fast].down) {
			d.fast = i
		}
		if d.slow < 0 || avg(s.down) < avg(d.slots[d.slow].down) {
			d.slow = i
		}
	}
	if d.fast >= 0 && d.slow >= 0 && d.fast != d.slow {
		d.peakDrop = (1 - avg(d.slots[d.slow].down)/avg(d.slots[d.fast].down)) * 100
	}

	for _, r := range d.tests {
		if mbps(r.Download) > 0 {
			d.worst = append(d.worst, r)
		}
	}
	sort.SliceStable(d.worst, func(a, c int) bool { return mbps(d.worst[a].Download) < mbps(d.worst[c].Download) })
	d.worst = d.worst[:min(5, len(d.worst))]
	d.wifiTests = countConn(d.tests, model.ConnWiFi)
	return d
}

// congested reports whether the slowest time slot is clearly slower.
func (d *ispData) congested() bool { return d.peakDrop >= 25 }

// ISPReport summarizes many speed tests into one document a user can send
// to their ISP: averages, % of plan, a time-of-day breakdown, the slowest
// tests and the latest stability monitor runs in the period.
func ISPReport(all []*model.TestResult, monitors []*model.MonitorSummary, plan *model.Plan, f ReportFilter, o Options) []byte {
	T := func(key string, args ...any) string { return i18n.T(o.Lang, key, args...) }
	d := buildISP(all, monitors, plan, f)

	var b bytes.Buffer
	line := func(k, v string) { fmt.Fprintf(&b, "%-26s %s\n", k+":", v) }
	b.WriteString(T("isp.title") + "\n" + strings.Repeat("=", 52) + "\n" + T("isp.intro") + "\n\n")
	writeMonitors := func() {
		if len(d.monitors) == 0 {
			return
		}
		b.WriteString("\n" + T("rep.monTitle") + "\n")
		for _, m := range d.monitors {
			fmt.Fprintf(&b, "  %s  %s: %s\n", m.StartedAt.Format("2006-01-02 15:04"), analysis.FormatDuration(m.DurationSec, o.Lang), m.Verdict)
		}
	}
	if len(d.tests) == 0 {
		b.WriteString(T("isp.none") + "\n")
		writeMonitors()
		b.WriteString("\n" + T("rep.footer") + "\n")
		return b.Bytes()
	}
	line(T("isp.period"), dateText(d.tests[0].StartedAt, o.Lang)+" – "+dateText(d.tests[len(d.tests)-1].StartedAt, o.Lang))
	line(T("isp.tests"), fmt.Sprint(len(d.tests)))
	line(T("isp.connection"), d.conn)
	if plan != nil && (plan.DownMbps > 0 || plan.UpMbps > 0) {
		line(T("isp.plan"), T("isp.planVal", speedOrDash(plan.DownMbps, o.Unit), speedOrDash(plan.UpMbps, o.Unit)))
	} else {
		line(T("isp.plan"), T("isp.planNone"))
	}

	b.WriteString("\n" + T("isp.summary") + "\n")
	spd := func(v float64) string { return analysis.FormatSpeed(v, o.Unit) }
	stat := func(name string, xs []float64, f func(float64) string) {
		if len(xs) > 0 {
			line(name, T("isp.avgMinMax", f(avg(xs)), f(minOf(xs)), f(maxOf(xs))))
		}
	}
	stat("Download", d.down, spd)
	stat("Upload", d.up, spd)
	stat("Ping", d.ping, analysis.FormatMs)
	stat("Jitter", d.jit, analysis.FormatMs)
	stat("Packet Loss", d.loss, pctText)
	if plan != nil && plan.DownMbps > 0 && len(d.down) > 0 {
		line(T("isp.ofPlan"), T("isp.ofPlanVal", pctText(model.PlanPercent(avg(d.down), plan.DownMbps)), d.upOfPlan()))
		line(T("isp.belowPlan", d.lowPct), T("isp.belowVal", d.below, len(d.down)))
	}

	b.WriteString("\n" + T("isp.byTime") + "\n")
	for i, s := range d.slots {
		n := max(len(s.down), len(s.ping))
		if n == 0 {
			continue
		}
		dl, p := "—", "—"
		if len(s.down) > 0 {
			dl = spd(avg(s.down))
		}
		if len(s.ping) > 0 {
			p = analysis.FormatMs(avg(s.ping))
		}
		b.WriteString("  " + T("isp.timeRow", T(fmt.Sprintf("isp.slot%d", i)), n, dl, p) + "\n")
	}
	if d.congested() {
		b.WriteString("  " + d.peakText(T) + "\n")
	}

	row := func(r *model.TestResult) string {
		c := ""
		if r.Network != nil {
			c = string(r.Network.Connection)
		}
		p, l := "—", "—"
		if r.PingMs > 0 {
			p, l = analysis.FormatMs(r.PingMs), pctText(r.PacketLoss)
		}
		return "  " + T("isp.row", r.StartedAt.Format("2006-01-02 15:04"), c, speedOrDash(mbps(r.Download), o.Unit), speedOrDash(mbps(r.Upload), o.Unit), p, l)
	}
	if len(d.worst) > 0 {
		b.WriteString("\n" + T("isp.worst") + "\n")
		for _, r := range d.worst {
			b.WriteString(row(r) + "\n")
		}
	}

	writeMonitors()

	b.WriteString("\n" + T("isp.all") + "\n")
	for _, r := range d.tests {
		b.WriteString(row(r) + "\n")
	}
	if d.wifiTests > 0 {
		b.WriteString("\n" + T("isp.wifiNote", d.wifiTests) + "\n")
	}
	b.WriteString("\n" + T("rep.footer") + "\n")
	return b.Bytes()
}

// upOfPlan is the average upload as a share of the plan, or a dash.
func (d *ispData) upOfPlan() string {
	if d.plan == nil || d.plan.UpMbps <= 0 || len(d.up) == 0 {
		return "—"
	}
	return pctText(model.PlanPercent(avg(d.up), d.plan.UpMbps))
}

func (d *ispData) peakText(T func(string, ...any) string) string {
	return T("isp.peak", T(fmt.Sprintf("isp.slot%d", d.slow)), d.peakDrop, T(fmt.Sprintf("isp.slot%d", d.fast)))
}

func mbps(s *model.SpeedResult) float64 {
	if s == nil || s.Mbps <= 0 {
		return 0
	}
	return s.Mbps
}

// collect gathers f over rs. f returns a negative value for "not
// measured"; a speed of 0 also means not measured.
func collect(rs []*model.TestResult, f func(*model.TestResult) float64) []float64 {
	var out []float64
	for _, r := range rs {
		if v := f(r); v >= 0 {
			out = append(out, v)
		}
	}
	return out
}

func avg(xs []float64) float64 {
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func minOf(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs {
		m = math.Min(m, x)
	}
	return m
}

func maxOf(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs {
		m = math.Max(m, x)
	}
	return m
}

func pctText(v float64) string { return analysis.FormatPct(math.Round(v*10) / 10) }

func speedOrDash(v float64, unit string) string {
	if v <= 0 {
		return "—"
	}
	return analysis.FormatSpeed(v, unit)
}

func connectionsOf(rs []*model.TestResult) string {
	seen := map[string]bool{}
	var out []string
	for _, r := range rs {
		if r.Network != nil && !seen[string(r.Network.Connection)] {
			seen[string(r.Network.Connection)] = true
			out = append(out, string(r.Network.Connection))
		}
	}
	return strings.Join(out, ", ")
}

func countConn(rs []*model.TestResult, c model.ConnectionType) int {
	n := 0
	for _, r := range rs {
		if r.Network != nil && r.Network.Connection == c {
			n++
		}
	}
	return n
}
