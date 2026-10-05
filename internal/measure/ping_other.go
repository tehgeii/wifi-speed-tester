//go:build !windows

package measure

import (
	"context"
	"errors"
	"math/rand/v2"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// sockPinger uses an unprivileged ICMP datagram socket ("udp4"), falling back
// to a raw socket when running as root. Used for development and the CLI on
// Linux/macOS; the shipped product is the Windows build.
type sockPinger struct{}

func newPlatformPinger() (Pinger, error) { return sockPinger{}, nil }

func (sockPinger) Close() error { return nil }

func (sockPinger) Ping(ctx context.Context, ip net.IP, timeout time.Duration) (time.Duration, error) {
	v4 := ip.To4()
	if v4 == nil {
		return 0, errors.New("only IPv4 ping targets are supported")
	}
	network, addr := "udp4", net.Addr(&net.UDPAddr{IP: v4})
	conn, err := icmp.ListenPacket(network, "0.0.0.0")
	if err != nil {
		network, addr = "ip4:icmp", &net.IPAddr{IP: v4}
		conn, err = icmp.ListenPacket(network, "0.0.0.0")
		if err != nil {
			return 0, err
		}
	}
	defer conn.Close()

	seq := rand.IntN(0xffff)
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: seq, Data: []byte("WiFiSpeedTester-ping-payload-32b")},
	}
	b, err := msg.Marshal(nil)
	if err != nil {
		return 0, err
	}
	deadline := time.Now().Add(timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Now()) })
	defer stop()

	start := time.Now()
	if _, err := conn.WriteTo(b, addr); err != nil {
		return 0, err
	}
	buf := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if ctx.Err() != nil {
				return 0, ctx.Err()
			}
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				return 0, ErrTimeout
			}
			return 0, err
		}
		rm, err := icmp.ParseMessage(1, buf[:n])
		if err != nil {
			continue
		}
		if echo, ok := rm.Body.(*icmp.Echo); ok && rm.Type == ipv4.ICMPTypeEchoReply && echo.Seq == seq {
			return time.Since(start), nil
		}
	}
}
