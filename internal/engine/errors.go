package engine

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"strings"
	"syscall"

	"github.com/tehgeii/wifi-speed-tester/internal/measure"
	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Explain turns a low-level error into a reason and a suggestion a
// non-technical user can act on. The raw error is kept as Technical.
func Explain(title string, err error) *model.Failure {
	f := &model.Failure{Title: title, Technical: err.Error()}
	var statusErr *measure.HTTPStatusError
	var dnsErr *net.DNSError
	var netErr net.Error
	var certErr *x509.UnknownAuthorityError
	var hostErr x509.HostnameError
	var invalidCert x509.CertificateInvalidError
	msg := strings.ToLower(err.Error())

	switch {
	case errors.Is(err, context.Canceled):
		f.Reason = "The test was cancelled."
		f.Suggestion = "Press Start Test to run it again."
	case errors.As(err, &statusErr) && statusErr.StatusCode == 429:
		f.Reason = "The test server is limiting requests because too many tests were run recently."
		f.Suggestion = "Wait a minute and try again, or add another test server in the config file."
	case errors.As(err, &statusErr):
		f.Reason = "The test server returned an error (" + statusErr.Status + ")."
		f.Suggestion = "Try again later or select another test server."
	case strings.Contains(msg, "proxyconnect") || strings.HasSuffix(msg, ": forbidden"):
		f.Reason = "A proxy or firewall blocked the connection to the test server."
		f.Suggestion = "Try another network, or ask your network administrator to allow the test server."
	case errors.As(err, &dnsErr):
		f.Reason = "The test server's name could not be looked up (DNS)."
		f.Suggestion = "Check that you are connected to the internet and that your DNS server works. Try again in a moment."
	case errors.As(err, &certErr) || errors.As(err, &hostErr) || errors.As(err, &invalidCert) || strings.Contains(msg, "certificate") || strings.Contains(msg, "tls"):
		f.Reason = "A secure connection to the test server could not be established."
		f.Suggestion = "Check that the system date and time are correct. A proxy, captive portal or antivirus that inspects HTTPS can also cause this."
	case errors.Is(err, syscall.ECONNRESET) || strings.Contains(msg, "connection reset") || strings.Contains(msg, "forcibly closed"):
		f.Reason = "The connection was reset during the test."
		f.Suggestion = "Your connection may be unstable, or a firewall/VPN interrupted it. Try again."
	case errors.Is(err, syscall.ECONNREFUSED) || strings.Contains(msg, "refused"):
		f.Reason = "The test server refused the connection."
		f.Suggestion = "Try again or select another test server."
	case errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()):
		f.Reason = "The test server did not respond in time."
		f.Suggestion = "Try again or select another test server. A firewall, VPN or a very slow connection can also cause this."
	case errors.Is(err, measure.ErrNoData):
		f.Reason = "No data could be transferred to or from the test server."
		f.Suggestion = "Try again. If it keeps failing, a firewall or proxy may be blocking speed tests."
	case strings.Contains(msg, "network is unreachable") || strings.Contains(msg, "no route"):
		f.Reason = "The network is unreachable."
		f.Suggestion = "Check that Wi-Fi or the cable is connected, then try again."
	default:
		f.Reason = "The test could not be completed."
		f.Suggestion = "Try again. If the problem continues, try another network or test server."
	}
	return f
}
