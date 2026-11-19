package export

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/analysis"
	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// ISPReportPDF is the ISP report as a PDF: the same figures as ISPReport,
// with charts of speed and ping over time and by time of day. Like the TXT
// version it holds no network identifiers.
func ISPReportPDF(all []*model.TestResult, monitors []*model.MonitorSummary, plan *model.Plan, f ReportFilter, o Options, now time.Time) []byte {
	T := func(key string, args ...any) string { return i18n.T(o.Lang, key, args...) }
	d := buildISP(all, monitors, plan, f)
	doc := newPDF(T("isp.title"))
	spd := func(v float64) string { return analysis.FormatSpeed(v, o.Unit) }
	// Charts use the display unit; 8 bits per byte.
	conv := func(v float64) float64 {
		if o.Unit == "MB/s" {
			return v / 8
		}
		return v
	}
	unit := "Mbps"
	if o.Unit == "MB/s" {
		unit = "MB/s"
	}
	axis := func(u string) func(float64) string {
		return func(v float64) string { return axisNum(v) + " " + u }
	}

	doc.text(pdfMarginX, doc.y+18, 18, true, pdfInk, T("isp.title"))
	doc.y += 28
	doc.paragraph(T("isp.intro"), 10, false, pdfMuted)
	doc.paragraph(T("pdf.created", dateText(now, o.Lang)), 8.5, false, pdfMuted)
	doc.gap(6)

	writeMonitors := func() {
		if len(d.monitors) == 0 {
			return
		}
		doc.heading(T("rep.monTitle"), 30)
		for _, m := range d.monitors {
			doc.paragraph(m.StartedAt.Format("2006-01-02 15:04")+"  ·  "+analysis.FormatDuration(m.DurationSec, o.Lang)+":  "+m.Verdict, 9, false, pdfInk)
		}
	}
	footer := func(page, total int) (string, string) {
		return "WiFi Speed + Ping Tester · " + T("isp.title"), T("pdf.page", page, total)
	}
	end := func() []byte {
		doc.gap(14)
		doc.paragraph(T("rep.footer"), 8.5, false, pdfMuted)
		return doc.bytes(footer)
	}
	if len(d.tests) == 0 {
		doc.gap(8)
		doc.paragraph(T("isp.none"), 11, true, pdfInk)
		writeMonitors()
		return end()
	}

	planText := T("isp.planNone")
	if plan != nil && (plan.DownMbps > 0 || plan.UpMbps > 0) {
		planText = T("isp.planVal", speedOrDash(plan.DownMbps, o.Unit), speedOrDash(plan.UpMbps, o.Unit))
	}
	doc.keyValue([][2]string{
		{T("isp.period"), dateText(d.tests[0].StartedAt, o.Lang) + " – " + dateText(d.tests[len(d.tests)-1].StartedAt, o.Lang)},
		{T("isp.tests"), fmt.Sprint(len(d.tests))},
		{T("isp.connection"), d.conn},
		{T("isp.plan"), planText},
	})

	// Summary.
	doc.heading(T("isp.summary"), 80)
	var rows [][]string
	stat := func(name string, xs []float64, f func(float64) string) {
		if len(xs) > 0 {
			rows = append(rows, []string{name, f(avg(xs)), f(minOf(xs)), f(maxOf(xs))})
		}
	}
	stat("Download", d.down, spd)
	stat("Upload", d.up, spd)
	stat("Ping", d.ping, analysis.FormatMs)
	stat("Jitter", d.jit, analysis.FormatMs)
	stat("Packet Loss", d.loss, pctText)
	doc.table([]pdfCol{{T("pdf.metric"), 0.34, false}, {T("pdf.avg"), 0.22, true}, {T("pdf.lowest"), 0.22, true}, {T("pdf.highest"), 0.22, true}}, rows)
	if plan != nil && plan.DownMbps > 0 && len(d.down) > 0 {
		doc.gap(8)
		doc.keyValue([][2]string{
			{T("isp.ofPlan"), T("isp.ofPlanVal", pctText(model.PlanPercent(avg(d.down), plan.DownMbps)), d.upOfPlan())},
			{T("isp.belowPlan", d.lowPct), T("isp.belowVal", d.below, len(d.down))},
		})
	}

	// Over time: x follows the test time, so a gap of days looks like one.
	first, last := d.tests[0].StartedAt, d.tests[len(d.tests)-1].StartedAt
	span := last.Sub(first).Seconds()
	xOf := func(i int, r *model.TestResult) float64 {
		if span <= 0 {
			if len(d.tests) == 1 {
				return 0.5
			}
			return float64(i) / float64(len(d.tests)-1)
		}
		return r.StartedAt.Sub(first).Seconds() / span
	}
	var dl, ul, pg pdfSeries
	dl = pdfSeries{name: "Download", c: pdfAccent}
	ul = pdfSeries{name: "Upload", c: pdfSeries2}
	pg = pdfSeries{name: "Ping", c: pdfSeries3}
	for i, r := range d.tests {
		if v := mbps(r.Download); v > 0 {
			dl.x, dl.y = append(dl.x, xOf(i, r)), append(dl.y, conv(v))
		}
		if v := mbps(r.Upload); v > 0 {
			ul.x, ul.y = append(ul.x, xOf(i, r)), append(ul.y, conv(v))
		}
		if r.PingMs > 0 {
			pg.x, pg.y = append(pg.x, xOf(i, r)), append(pg.y, r.PingMs)
		}
	}
	from, to := first.Format("2006-01-02 15:04"), last.Format("2006-01-02 15:04")
	if len(dl.x)+len(ul.x) > 0 {
		doc.heading(T("pdf.speedChart"), 160)
		ref, refLabel := 0.0, ""
		if plan != nil && plan.DownMbps > 0 {
			ref, refLabel = conv(plan.DownMbps), T("pdf.planLine", spd(plan.DownMbps))
		}
		doc.lineChart([]pdfSeries{dl, ul}, 160, axis(unit), from, to, ref, refLabel)
	}
	if len(pg.x) > 0 {
		doc.heading(T("pdf.pingChart"), 120)
		doc.lineChart([]pdfSeries{pg}, 120, axis("ms"), from, to, 0, "")
	}

	// By time of day.
	doc.heading(T("isp.byTime"), 150)
	labels := make([]string, 4)
	values := make([]float64, 4)
	var slotRows [][]string
	for i, s := range d.slots {
		labels[i] = T(fmt.Sprintf("isp.slot%d", i))
		values[i] = -1
		if len(s.down) > 0 {
			values[i] = conv(avg(s.down))
		}
		n := max(len(s.down), len(s.ping))
		if n == 0 {
			continue
		}
		dlText, p := "—", "—"
		if len(s.down) > 0 {
			dlText = spd(avg(s.down))
		}
		if len(s.ping) > 0 {
			p = analysis.FormatMs(avg(s.ping))
		}
		slotRows = append(slotRows, []string{labels[i], fmt.Sprint(n), dlText, p})
	}
	doc.paragraph(T("pdf.timeChart"), 8.5, false, pdfMuted)
	doc.gap(4)
	doc.barChart(labels, values, pdfAccent, 130, axis(unit))
	doc.gap(8)
	doc.table([]pdfCol{{T("pdf.time"), 0.31, false}, {T("isp.tests"), 0.17, true}, {T("pdf.downAvg"), 0.28, true}, {T("pdf.pingAvg"), 0.24, true}}, slotRows)
	if d.congested() {
		doc.gap(6)
		doc.paragraph(d.peakText(T), 9.5, true, pdfWarn)
	}

	testCols := []pdfCol{{T("rep.date"), 0.22, false}, {T("isp.connection"), 0.2, false}, {"Download", 0.16, true}, {"Upload", 0.16, true}, {"Ping", 0.13, true}, {T("pdf.loss"), 0.13, true}}
	testRow := func(r *model.TestResult) []string {
		c, p, l := "", "—", "—"
		if r.Network != nil {
			c = string(r.Network.Connection)
		}
		if r.PingMs > 0 {
			p, l = analysis.FormatMs(r.PingMs), pctText(r.PacketLoss)
		}
		return []string{r.StartedAt.Format("2006-01-02 15:04"), c, speedOrDash(mbps(r.Download), o.Unit), speedOrDash(mbps(r.Upload), o.Unit), p, l}
	}
	if len(d.worst) > 0 {
		doc.heading(T("isp.worst"), 60)
		rows = rows[:0]
		for _, r := range d.worst {
			rows = append(rows, testRow(r))
		}
		doc.table(testCols, rows)
	}

	writeMonitors()

	doc.heading(T("isp.all"), 60)
	rows = rows[:0]
	for _, r := range d.tests {
		rows = append(rows, testRow(r))
	}
	doc.table(testCols, rows)
	if d.wifiTests > 0 {
		doc.gap(8)
		doc.paragraph(T("isp.wifiNote", d.wifiTests), 9, false, pdfInk)
	}
	return end()
}

// axisNum writes an axis value with up to three significant digits.
func axisNum(v float64) string {
	if v == 0 {
		return "0"
	}
	p := math.Pow(10, 2-math.Floor(math.Log10(math.Abs(v))))
	return strconv.FormatFloat(math.Round(v*p)/p, 'f', -1, 64)
}
