package export

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// checkPDF verifies the file structure: header, trailer, and that every
// xref entry points at its object.
func checkPDF(t *testing.T, b []byte) {
	t.Helper()
	if !bytes.HasPrefix(b, []byte("%PDF-1.4\n")) || !bytes.HasSuffix(b, []byte("%%EOF\n")) {
		t.Fatal("not a complete PDF")
	}
	m := regexp.MustCompile(`startxref\n(\d+)\n`).FindSubmatch(b)
	if m == nil {
		t.Fatal("no startxref")
	}
	at, _ := strconv.Atoi(string(m[1]))
	if !bytes.HasPrefix(b[at:], []byte("xref\n0 ")) {
		t.Fatalf("startxref %d does not point at the xref table", at)
	}
	lines := strings.Split(string(b[at:]), "\n")
	n, _ := strconv.Atoi(strings.Fields(lines[1])[1])
	for i := 1; i < n; i++ {
		e := lines[2+i]
		if len(e) != 19 { // 20 bytes with the newline
			t.Fatalf("xref entry %d has the wrong length: %q", i, e)
		}
		off, _ := strconv.Atoi(e[:10])
		if want := fmt.Sprintf("%d 0 obj\n", i); !bytes.HasPrefix(b[off:], []byte(want)) {
			t.Fatalf("xref entry %d points at %q", i, b[off:off+12])
		}
	}
}

var pdfSeq int

// pdfText extracts the text with pdftotext when it is installed (CI sets
// WST_NEED_PDFTOTEXT so it cannot be skipped there).
func pdfText(t *testing.T, b []byte) string {
	t.Helper()
	if _, err := exec.LookPath("pdftotext"); err != nil {
		if os.Getenv("WST_NEED_PDFTOTEXT") != "" {
			t.Fatal("pdftotext is not installed")
		}
		t.Skip("pdftotext is not installed")
	}
	f := filepath.Join(t.TempDir(), "r.pdf")
	if err := os.WriteFile(f, b, 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("pdftotext", "-layout", f, "-").CombinedOutput()
	if err != nil {
		t.Fatalf("pdftotext: %v\n%s", err, out)
	}
	if dir := os.Getenv("WST_PDF_OUT"); dir != "" { // to look at it
		pdfSeq++
		_ = os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s-%d.pdf", t.Name(), pdfSeq)), b, 0o644)
	}
	return string(out)
}

func TestISPReportPDF(t *testing.T) {
	day := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	all := []*model.TestResult{
		test(day.Add(9*time.Hour), model.ConnEthernet, 48, 10, 15),
		test(day.Add(10*time.Hour), model.ConnEthernet, 46, 10, 16),
		test(day.Add(20*time.Hour), model.ConnEthernet, 12, 9, 60),
		test(day.Add(21*time.Hour), model.ConnWiFi, 20, 8, 40),
	}
	all[0].Network.IPv4 = []string{"192.168.1.23"}
	all[0].Network.WiFi = &model.WiFiInfo{SSID: "SecretHome"}
	plan := &model.Plan{DownMbps: 50, UpMbps: 10, LowPct: 50}
	mons := []*model.MonitorSummary{{StartedAt: day.Add(20 * time.Hour), DurationSec: 600, Verdict: "dropped 3 times (router fine)"}}
	b := ISPReportPDF(all, mons, plan, ReportFilter{Since: day.Add(-time.Hour)}, Options{Unit: "Mbps", Lang: i18n.EN}, day.Add(30*time.Hour))
	checkPDF(t, b)
	txt := pdfText(t, b)
	for _, want := range []string{
		"INTERNET QUALITY REPORT", "Created 02 October 2026", "50.0 Mbps down / 10.0 Mbps up",
		"Download", "31.5 Mbps", "12.0 Mbps", "48.0 Mbps", "download 63% · upload 92.5%", "2 of 4 tests",
		"DOWNLOAD AND UPLOAD OVER TIME", "plan 50.0 Mbps", "PING OVER TIME", "BY TIME OF DAY", "18:00–24:00",
		"congestion at busy hours", "SLOWEST TESTS", "dropped 3 times (router fine)", "ALL TESTS", "2026-10-01 21:00",
		"1 of these tests were over Wi-Fi", "not a guarantee", "Page 1 of",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("PDF text is missing %q\n%s", want, txt)
		}
	}
	for _, secret := range []string{"192.168", "SecretHome"} {
		if strings.Contains(txt, secret) || bytes.Contains(b, []byte(secret)) {
			t.Errorf("PDF contains %q", secret)
		}
	}
}

func TestISPReportPDFPagesAndLanguage(t *testing.T) {
	day := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	var all []*model.TestResult
	for i := 0; i < 90; i++ {
		all = append(all, test(day.Add(time.Duration(i*3)*time.Hour), model.ConnWiFi, float64(20+i%40), 8, float64(15+i%30)))
	}
	b := ISPReportPDF(all, nil, nil, ReportFilter{}, Options{Unit: "MB/s", Lang: i18n.ID}, day)
	checkPDF(t, b)
	txt := pdfText(t, b)
	if n := strings.Count(txt, "\f"); n < 3 {
		t.Errorf("90 tests fit on %d pages, want a table across several", n)
	}
	// The table header repeats on every page it continues on.
	if strings.Count(txt, "Tanggal") < 3 {
		t.Errorf("table header not repeated:\n%s", txt)
	}
	for _, want := range []string{"LAPORAN KUALITAS INTERNET", "belum diisi", "MB/s", "Halaman 1 dari", "SEMUA TES", "Rata-rata"} {
		if !strings.Contains(txt, want) {
			t.Errorf("PDF text is missing %q", want)
		}
	}

	empty := ISPReportPDF(nil, []*model.MonitorSummary{{StartedAt: day, DurationSec: 300, Verdict: "stable"}}, nil, ReportFilter{}, Options{Lang: i18n.EN}, day)
	checkPDF(t, empty)
	if txt := pdfText(t, empty); !strings.Contains(txt, "No speed tests") || !strings.Contains(txt, "stable") {
		t.Errorf("empty report:\n%s", txt)
	}
}

func TestPDFText(t *testing.T) {
	if got := string(encodeWinAnsi("≥ 25 Mbps · é – ok ★ 你")); got != ">= 25 Mbps \xb7 \xe9 \x96 ok * ?" {
		t.Errorf("encodeWinAnsi = %q", got)
	}
	if got := pdfString(`a (b) \ c`); got != `a \(b\) \\ c` {
		t.Errorf("pdfString = %q", got)
	}
	if w := textWidth("Wi", 10, false); w != 9.44+2.22 {
		t.Errorf("width = %v", w)
	}
	lines := wrapText("one two three four five six", 10, false, 50)
	if len(lines) < 3 || strings.Join(lines, " ") != "one two three four five six" {
		t.Errorf("wrap = %q", lines)
	}
	for _, l := range lines {
		if textWidth(l, 10, false) > 50 && strings.Contains(l, " ") {
			t.Errorf("line too wide: %q", l)
		}
	}
	if s := fitText("a long connection name", 9, false, 40); !strings.HasSuffix(s, "…") || textWidth(s, 9, false) > 40 {
		t.Errorf("fit = %q", s)
	}
	if niceCeil(31) != 50 || niceCeil(0) != 1 || niceCeil(200) != 200 || niceCeil(2.1) != 2.5 {
		t.Error("niceCeil")
	}
	if axisNum(12.5) != "12.5" || axisNum(1234) != "1230" || axisNum(0.125) != "0.125" {
		t.Errorf("axisNum %s %s %s", axisNum(12.5), axisNum(1234), axisNum(0.125))
	}
}
