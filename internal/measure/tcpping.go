package measure

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"
)

// TCPPinger measures the time to open a TCP connection (SYN to SYN-ACK,
// about one round trip). It is the fallback when a network blocks ICMP; the
// values can read slightly higher than ICMP ping.
type TCPPinger struct{ Port int }

func (t TCPPinger) Ping(ctx context.Context, ip net.IP, timeout time.Duration) (time.Duration, error) {
	d := net.Dialer{Timeout: timeout}
	start := time.Now()
	c, err := d.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), strconv.Itoa(t.Port)))
	if err != nil {
		var ne net.Error
		if ctx.Err() == nil && errors.As(err, &ne) && ne.Timeout() {
			return 0, ErrTimeout
		}
		return 0, err
	}
	rtt := time.Since(start)
	c.Close()
	return rtt, nil
}

func (TCPPinger) Close() error { return nil }
