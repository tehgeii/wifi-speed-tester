package app

import (
	"context"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/update"
)

// IssuesURL is where problem reports go. It is inside this project's
// GitHub pages, so openURL allows it.
var IssuesURL = "https://github.com/" + update.Repo + "/issues/new"

var (
	ipv4RE = regexp.MustCompile(`\b(\d{1,3})\.(\d{1,3})\.\d{1,3}\.\d{1,3}\b`)
	ipv6RE = regexp.MustCompile(`\b[0-9a-fA-F]{1,4}(?::[0-9a-fA-F]{0,4}){2,7}\b`)
)

// scrubAddresses hides IP addresses in error texts before they are shared
// in a public issue: IPv4 keeps its first two numbers, IPv6 none.
func scrubAddresses(s string) string {
	s = ipv4RE.ReplaceAllString(s, "$1.$2.x.x")
	return ipv6RE.ReplaceAllString(s, "x:x:x")
}

// problemReport describes this setup for a bug report: app and Windows
// versions, connection type, language and the last test's outcome. It holds
// no IP address, Wi-Fi name, BSSID or adapter name.
func (a *App) problemReport() map[string]any {
	lang := a.lang()
	T := func(key string, args ...any) string { return i18n.T(lang, key, args...) }
	var b strings.Builder
	line := func(k, v string) { b.WriteString(k + ": " + v + "\n") }
	line(T("prob.app"), "WiFi Speed + Ping Tester "+Version)
	line(T("prob.os"), osVersion()+" ("+runtime.GOARCH+")")
	conn := T("prob.unknown")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if n, err := a.eng.Detect(ctx); err == nil && n != nil && n.Connection != "" {
		conn = string(n.Connection)
	}
	line(T("prob.conn"), conn)
	line(T("prob.lang"), string(lang))

	a.mu.Lock()
	last := a.last
	a.mu.Unlock()
	title := T("prob.title")
	switch {
	case last == nil:
		line(T("prob.lastTest"), T("prob.noTest"))
	case last.Failure != nil:
		line(T("prob.lastTest"), last.Mode+", "+last.StartedAt.Format("2006-01-02 15:04"))
		line(T("prob.error"), scrubAddresses(last.Failure.Title))
		if last.Failure.Technical != "" {
			line(T("prob.details"), scrubAddresses(last.Failure.Technical))
		}
		title += ": " + scrubAddresses(last.Failure.Title)
	default:
		level := "—"
		if last.Quality != nil {
			level = string(last.Quality.Level)
		}
		line(T("prob.lastTest"), last.Mode+", "+last.StartedAt.Format("2006-01-02 15:04")+", "+level)
	}
	info := strings.TrimRight(b.String(), "\n")
	body := T("prob.body") + "\n\n\n\n" + T("prob.infoHead") + "\n```\n" + info + "\n```\n"
	q := url.Values{"title": {title}, "body": {body}}
	return map[string]any{"text": info, "url": IssuesURL + "?" + q.Encode()}
}

// copyProblem puts the report on the clipboard when the host can; the
// page copies it otherwise.
func (a *App) copyProblem(text string) (bool, error) {
	if a.host.CopyText == nil {
		return false, nil
	}
	return true, a.host.CopyText(text)
}
