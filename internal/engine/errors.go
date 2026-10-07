package engine

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"strings"
	"syscall"

	"github.com/tehgeii/wifi-speed-tester/internal/i18n"
	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Explain turns a low-level error into a reason and a suggestion a
// non-technical user can act on. The raw error is kept as Technical.
func Explain(lang i18n.Lang, title string, err error) *model.Failure {
	f := &model.Failure{Title: title, Technical: err.Error()}
	set := func(key string, args ...any) {
		f.Reason = i18n.T(lang, "err."+key+".r", args...)
		f.Suggestion = i18n.T(lang, "err."+key+".s")
	}
	var statusErr *measure.HTTPStatusError
	var dnsErr *net.DNSError
	var netErr net.Error
	var certErr *x509.UnknownAuthorityError
	var hostErr x509.HostnameError
	var invalidCert x509.CertificateInvalidError
	msg := strings.ToLower(err.Error())

	switch {
	case errors.Is(err, context.Canceled):
		set("cancelled")
	case errors.As(err, &statusErr) && statusErr.StatusCode == 429:
		set("ratelimit")
	case errors.As(err, &statusErr):
		set("http", statusErr.Status)
	case strings.Contains(msg, "proxyconnect") || strings.HasSuffix(msg, ": forbidden"):
		set("proxy")
	case errors.As(err, &dnsErr):
		set("dns")
	case errors.As(err, &certErr) || errors.As(err, &hostErr) || errors.As(err, &invalidCert) || strings.Contains(msg, "certificate") || strings.Contains(msg, "tls"):
		set("tls")
	case errors.Is(err, syscall.ECONNRESET) || strings.Contains(msg, "connection reset") || strings.Contains(msg, "forcibly closed"):
		set("reset")
	case errors.Is(err, syscall.ECONNREFUSED) || strings.Contains(msg, "refused"):
		set("refused")
	case errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()):
		set("timeout")
	case errors.Is(err, measure.ErrNoData):
		set("nodata")
	case strings.Contains(msg, "network is unreachable") || strings.Contains(msg, "no route"):
		set("unreachable")
	default:
		set("other")
	}
	return f
}
