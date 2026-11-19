package export

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf16"
)

// A4 portrait, in points (1/72 inch).
const (
	pdfW, pdfH  = 595.28, 841.89
	pdfMarginX  = 48.0
	pdfTop      = 54.0
	pdfBottom   = 64.0
	pdfContentW = pdfW - 2*pdfMarginX
)

type rgb [3]float64

var (
	pdfInk     = rgb{0.12, 0.14, 0.17}
	pdfMuted   = rgb{0.40, 0.45, 0.51}
	pdfRule    = rgb{0.85, 0.87, 0.90}
	pdfZebra   = rgb{0.955, 0.965, 0.98}
	pdfAccent  = rgb{0.12, 0.44, 0.92} // download
	pdfSeries2 = rgb{0.51, 0.31, 0.85} // upload
	pdfSeries3 = rgb{0.85, 0.42, 0.05} // ping
	pdfWarn    = rgb{0.80, 0.16, 0.20}
)

// pdfDoc writes a small PDF 1.4 document with the standard Helvetica fonts
// (not embedded, WinAnsi encoding): text, lines, rectangles, tables and
// simple charts. That is all the reports need, so the exe needs no PDF
// library. Positions are points from the top-left corner of the page; y is
// the top of the next free line.
type pdfDoc struct {
	title     string
	pages     []*bytes.Buffer
	cur       *bytes.Buffer
	y         float64
	onNewPage func() // repeats a table header after a page break
}

func newPDF(title string) *pdfDoc {
	d := &pdfDoc{title: title}
	d.newPage()
	return d
}

func (d *pdfDoc) newPage() {
	d.cur = &bytes.Buffer{}
	d.pages = append(d.pages, d.cur)
	d.y = pdfTop
}

// need starts a new page unless h more points fit on this one.
func (d *pdfDoc) need(h float64) {
	if d.y+h <= pdfH-pdfBottom {
		return
	}
	d.newPage()
	if d.onNewPage != nil {
		d.onNewPage()
	}
}

func (d *pdfDoc) op(format string, args ...any) { fmt.Fprintf(d.cur, format+"\n", args...) }

// num writes a coordinate with at most two decimals.
func num(v float64) string { return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64) }

func col(c rgb) string { return num(c[0]) + " " + num(c[1]) + " " + num(c[2]) }

// text draws s with its baseline at y.
func (d *pdfDoc) text(x, y, size float64, bold bool, c rgb, s string) {
	font := "F1"
	if bold {
		font = "F2"
	}
	d.op("BT /%s %s Tf %s rg %s %s Td (%s) Tj ET", font, num(size), col(c), num(x), num(pdfH-y), pdfString(s))
}

func (d *pdfDoc) textRight(right, y, size float64, bold bool, c rgb, s string) {
	d.text(right-textWidth(s, size, bold), y, size, bold, c, s)
}

func (d *pdfDoc) rect(x, y, w, h float64, c rgb) {
	d.op("%s rg %s %s %s %s re f", col(c), num(x), num(pdfH-y-h), num(w), num(h))
}

func (d *pdfDoc) line(x1, y1, x2, y2, width float64, c rgb, dashed bool) {
	dash := "[] 0 d"
	if dashed {
		dash = "[3 3] 0 d"
	}
	d.op("%s %s RG %s w %s %s m %s %s l S", dash, col(c), num(width), num(x1), num(pdfH-y1), num(x2), num(pdfH-y2))
}

// dot draws a filled circle (four Bézier quarters).
func (d *pdfDoc) dot(x, y, r float64, c rgb) {
	k, Y := 0.5523*r, pdfH-y
	d.op("%s rg %s %s m", col(c), num(x+r), num(Y))
	d.op("%s %s %s %s %s %s c", num(x+r), num(Y+k), num(x+k), num(Y+r), num(x), num(Y+r))
	d.op("%s %s %s %s %s %s c", num(x-k), num(Y+r), num(x-r), num(Y+k), num(x-r), num(Y))
	d.op("%s %s %s %s %s %s c", num(x-r), num(Y-k), num(x-k), num(Y-r), num(x), num(Y-r))
	d.op("%s %s %s %s %s %s c f", num(x+k), num(Y-r), num(x+r), num(Y-k), num(x+r), num(Y))
}

// ---- flowing content -----------------------------------------------------------

func (d *pdfDoc) gap(h float64) { d.y += h }

// paragraph writes wrapped text across the content width.
func (d *pdfDoc) paragraph(s string, size float64, bold bool, c rgb) {
	d.paragraphAt(pdfMarginX, pdfContentW, s, size, bold, c)
}

func (d *pdfDoc) paragraphAt(x, w float64, s string, size float64, bold bool, c rgb) {
	lead := size * 1.4
	for _, l := range wrapText(s, size, bold, w) {
		d.need(lead)
		d.text(x, d.y+size, size, bold, c, l)
		d.y += lead
	}
}

// heading starts a section, keeping at least keep points of it together.
func (d *pdfDoc) heading(s string, keep float64) {
	d.gap(14)
	d.need(26 + keep)
	d.text(pdfMarginX, d.y+11, 10.5, true, pdfAccent, s)
	d.y += 17
	d.line(pdfMarginX, d.y, pdfMarginX+pdfContentW, d.y, 0.6, pdfRule, false)
	d.y += 8
}

// keyValue writes "key   value" rows, wrapping long values.
func (d *pdfDoc) keyValue(rows [][2]string) {
	const size, keyW = 9.5, 150.0
	for _, r := range rows {
		lines := wrapText(r[1], size, false, pdfContentW-keyW)
		d.need(size * 1.5 * float64(len(lines)))
		d.text(pdfMarginX, d.y+size, size, true, pdfMuted, r[0])
		for _, l := range lines {
			d.text(pdfMarginX+keyW, d.y+size, size, false, pdfInk, l)
			d.y += size * 1.5
		}
	}
}

type pdfCol struct {
	title string
	frac  float64 // share of the content width
	right bool
}

// table draws a striped table; after a page break the header repeats.
func (d *pdfDoc) table(cols []pdfCol, rows [][]string) {
	const size, rowH, pad = 8.5, 15.0, 5.0
	cell := func(i int, s string, bold bool, c rgb) {
		x := pdfMarginX
		for _, cc := range cols[:i] {
			x += cc.frac * pdfContentW
		}
		w := cols[i].frac*pdfContentW - 2*pad
		s = fitText(s, size, bold, w)
		if cols[i].right {
			d.textRight(x+cols[i].frac*pdfContentW-pad, d.y+10.5, size, bold, c, s)
		} else {
			d.text(x+pad, d.y+10.5, size, bold, c, s)
		}
	}
	header := func() {
		for i, c := range cols {
			cell(i, c.title, true, pdfMuted)
		}
		d.y += rowH
		d.line(pdfMarginX, d.y, pdfMarginX+pdfContentW, d.y, 0.6, pdfRule, false)
	}
	d.need(rowH * 2)
	header()
	d.onNewPage = header
	for i, r := range rows {
		d.need(rowH)
		if i%2 == 1 {
			d.rect(pdfMarginX, d.y, pdfContentW, rowH, pdfZebra)
		}
		for j, s := range r {
			cell(j, s, false, pdfInk)
		}
		d.y += rowH
	}
	d.onNewPage = nil
}

// ---- charts --------------------------------------------------------------------

type pdfSeries struct {
	name string
	c    rgb
	x, y []float64 // x in 0..1
}

// lineChart plots series over a shared 0..1 x axis. ref, when > 0, draws a
// dashed reference line (the plan) labelled refLabel.
func (d *pdfDoc) lineChart(series []pdfSeries, h float64, yFmt func(float64) string, xFrom, xTo string, ref float64, refLabel string) {
	const left, bottom, legendH = 44.0, 16.0, 16.0
	d.need(h + legendH)
	// Legend.
	x := pdfMarginX + left
	for _, s := range series {
		d.rect(x, d.y+3, 9, 9, s.c)
		d.text(x+13, d.y+11, 8.5, false, pdfInk, s.name)
		x += 13 + textWidth(s.name, 8.5, false) + 16
	}
	d.y += legendH
	top, plotH := d.y, h-bottom
	plotX, plotW := pdfMarginX+left, pdfContentW-left
	maxV := ref
	for _, s := range series {
		for _, v := range s.y {
			maxV = math.Max(maxV, v)
		}
	}
	maxV = niceCeil(maxV)
	py := func(v float64) float64 { return top + plotH - v/maxV*plotH }
	for _, f := range []float64{0, 0.5, 1} {
		yy := py(maxV * f)
		d.line(plotX, yy, plotX+plotW, yy, 0.5, pdfRule, false)
		d.textRight(plotX-6, yy+3, 7.5, false, pdfMuted, yFmt(maxV*f))
	}
	d.text(plotX, top+h-3, 7.5, false, pdfMuted, xFrom)
	d.textRight(plotX+plotW, top+h-3, 7.5, false, pdfMuted, xTo)
	if ref > 0 {
		d.line(plotX, py(ref), plotX+plotW, py(ref), 0.8, pdfWarn, true)
		d.textRight(plotX+plotW, py(ref)-3, 7.5, false, pdfWarn, refLabel)
	}
	for _, s := range series {
		if len(s.x) > 1 {
			d.op("[] 0 d %s RG 1.4 w 1 j", col(s.c))
			for i := range s.x {
				verb := "l"
				if i == 0 {
					verb = "m"
				}
				d.op("%s %s %s", num(plotX+s.x[i]*plotW), num(pdfH-py(s.y[i])), verb)
			}
			d.op("S")
		}
		if len(s.x) <= 40 {
			for i := range s.x {
				d.dot(plotX+s.x[i]*plotW, py(s.y[i]), 1.8, s.c)
			}
		}
	}
	d.y = top + h
}

// barChart draws one bar per label; a negative value means no data.
func (d *pdfDoc) barChart(labels []string, values []float64, c rgb, h float64, vFmt func(float64) string) {
	const bottom = 16.0
	d.need(h)
	top, plotH := d.y, h-bottom-12
	maxV := 0.0
	for _, v := range values {
		maxV = math.Max(maxV, v)
	}
	maxV = niceCeil(maxV)
	slot := pdfContentW / float64(len(labels))
	bw := math.Min(70, slot*0.55)
	base := top + 12 + plotH
	d.line(pdfMarginX, base, pdfMarginX+pdfContentW, base, 0.6, pdfRule, false)
	for i, l := range labels {
		cx := pdfMarginX + slot*(float64(i)+0.5)
		lw := textWidth(l, 8, false)
		d.text(cx-lw/2, base+12, 8, false, pdfMuted, l)
		v := values[i]
		label := "—"
		bh := 0.0
		if v >= 0 {
			bh = v / maxV * plotH
			d.rect(cx-bw/2, base-bh, bw, bh, c)
			label = vFmt(v)
		}
		w := textWidth(label, 8, true)
		d.text(cx-w/2, base-bh-4, 8, true, pdfInk, label)
	}
	d.y = top + h
}

// niceCeil rounds v up to 1, 2, 2.5 or 5 times a power of ten.
func niceCeil(v float64) float64 {
	if v <= 0 {
		return 1
	}
	p := math.Pow(10, math.Floor(math.Log10(v)))
	for _, m := range []float64{1, 2, 2.5, 5, 10} {
		if m*p >= v-1e-9 {
			return m * p
		}
	}
	return 10 * p
}

// ---- output --------------------------------------------------------------------

// bytes finishes the document; footer gives the left and right footer
// text of each page.
func (d *pdfDoc) bytes(footer func(page, total int) (string, string)) []byte {
	for i, p := range d.pages {
		d.cur = p
		l, r := footer(i+1, len(d.pages))
		d.line(pdfMarginX, pdfH-46, pdfMarginX+pdfContentW, pdfH-46, 0.5, pdfRule, false)
		d.text(pdfMarginX, pdfH-34, 8, false, pdfMuted, l)
		d.textRight(pdfMarginX+pdfContentW, pdfH-34, 8, false, pdfMuted, r)
	}
	var out bytes.Buffer
	var offsets []int
	obj := func(body string) {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", len(offsets), body)
	}
	out.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	// 1 catalog, 2 page tree, 3-4 fonts, 5 info, then a page and its
	// content stream for each page.
	kids := make([]string, len(d.pages))
	for i := range d.pages {
		kids[i] = fmt.Sprintf("%d 0 R", 6+2*i)
	}
	obj("<< /Type /Catalog /Pages 2 0 R >>")
	obj(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(d.pages)))
	obj("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")
	obj("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>")
	obj(fmt.Sprintf("<< /Title %s /Producer %s /Creator %s >>", pdfTextString(d.title), pdfTextString("WiFi Speed + Ping Tester"), pdfTextString("WiFi Speed + Ping Tester")))
	for i, p := range d.pages {
		obj(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %s %s] /Resources << /Font << /F1 3 0 R /F2 4 0 R >> >> /Contents %d 0 R >>", num(pdfW), num(pdfH), 7+2*i))
		var z bytes.Buffer
		w := zlib.NewWriter(&z)
		_, _ = w.Write(p.Bytes())
		_ = w.Close()
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n<< /Length %d /Filter /FlateDecode >>\nstream\n", len(offsets), z.Len())
		out.Write(z.Bytes())
		out.WriteString("\nendstream\nendobj\n")
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets)+1)
	for _, o := range offsets {
		fmt.Fprintf(&out, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R /Info 5 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets)+1, xref)
	return out.Bytes()
}

// ---- text encoding and metrics ---------------------------------------------------

// winAnsi maps the Unicode characters of WinAnsiEncoding (Windows-1252)
// outside Latin-1.
var winAnsi = map[rune]byte{
	'€': 0x80, '‚': 0x82, 'ƒ': 0x83, '„': 0x84, '…': 0x85, '†': 0x86, '‡': 0x87, 'ˆ': 0x88, '‰': 0x89,
	'Š': 0x8A, '‹': 0x8B, 'Œ': 0x8C, 'Ž': 0x8E, '‘': 0x91, '’': 0x92, '“': 0x93, '”': 0x94, '•': 0x95,
	'–': 0x96, '—': 0x97, '˜': 0x98, '™': 0x99, 'š': 0x9A, '›': 0x9B, 'œ': 0x9C, 'ž': 0x9E, 'Ÿ': 0x9F,
}

// pdfReplace spells the symbols the app uses that the standard fonts lack.
var pdfReplace = map[rune]string{
	'≥': ">=", '≤': "<=", '≈': "~", '↓': "v", '↑': "^", '→': "->", '★': "*", '✓': "OK", '✗': "x", '−': "-",
	' ': " ", ' ': " ", '\t': " ",
}

// encodeWinAnsi converts s to the bytes the standard fonts draw.
func encodeWinAnsi(s string) []byte {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 0x20 && r < 0x7f, r >= 0xa0 && r <= 0xff:
			out = append(out, byte(r))
		case winAnsi[r] != 0:
			out = append(out, winAnsi[r])
		case pdfReplace[r] != "":
			out = append(out, pdfReplace[r]...)
		default:
			out = append(out, '?')
		}
	}
	return out
}

// pdfString escapes s for a literal string in a content stream.
func pdfString(s string) string {
	var b strings.Builder
	for _, c := range encodeWinAnsi(s) {
		if c == '(' || c == ')' || c == '\\' {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	return b.String()
}

// pdfTextString encodes s as UTF-16 for the document information.
func pdfTextString(s string) string {
	var b strings.Builder
	b.WriteString("<FEFF")
	for _, u := range utf16.Encode([]rune(s)) {
		fmt.Fprintf(&b, "%04X", u)
	}
	b.WriteString(">")
	return b.String()
}

// Glyph widths of Helvetica and Helvetica-Bold for codes 32-126, in 1/1000
// of the font size (from the Adobe font metrics).
var helvW = [95]int{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278,
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556,
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778,
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556,
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556,
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584,
}

var helvBoldW = [95]int{
	278, 333, 474, 556, 556, 889, 722, 238, 333, 333, 389, 584, 278, 333, 278, 278,
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 333, 333, 584, 584, 584, 611,
	975, 722, 722, 722, 722, 667, 611, 778, 722, 278, 556, 722, 611, 833, 722, 778,
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 333, 278, 333, 584, 556,
	333, 556, 611, 556, 611, 556, 333, 611, 611, 278, 278, 556, 278, 889, 611, 611,
	611, 611, 389, 556, 333, 611, 556, 778, 556, 556, 500, 389, 280, 389, 584,
}

// highW holds the widths used above 126 (the same in both weights, close
// enough for layout); other codes count as 556.
var highW = map[byte]int{0x85: 1000, 0x91: 278, 0x92: 278, 0x93: 500, 0x94: 500, 0x95: 350, 0x96: 556, 0x97: 1000, 0xa0: 278, 0xb0: 400, 0xb7: 278, 0xd7: 584}

func textWidth(s string, size float64, bold bool) float64 {
	tbl := &helvW
	if bold {
		tbl = &helvBoldW
	}
	w := 0
	for _, c := range encodeWinAnsi(s) {
		switch {
		case c >= 32 && c <= 126:
			w += tbl[c-32]
		case highW[c] != 0:
			w += highW[c]
		default:
			w += 556
		}
	}
	return float64(w) * size / 1000
}

// wrapText breaks s into lines no wider than w; "\n" forces a break.
func wrapText(s string, size float64, bold bool, w float64) []string {
	var out []string
	for _, para := range strings.Split(s, "\n") {
		line := ""
		for _, word := range strings.Fields(para) {
			try := word
			if line != "" {
				try = line + " " + word
			}
			if line != "" && textWidth(try, size, bold) > w {
				out = append(out, line)
				line = word
			} else {
				line = try
			}
		}
		out = append(out, line)
	}
	return out
}

// fitText shortens s with an ellipsis so it fits in w.
func fitText(s string, size float64, bold bool, w float64) string {
	if textWidth(s, size, bold) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && textWidth(string(r)+"…", size, bold) > w {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}
