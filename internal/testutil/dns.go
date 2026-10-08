package testutil

import (
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// DNSServer is a UDP DNS server that answers every A query with 192.0.2.1
// after Delay. Set Fail to answer SERVFAIL, or Silent to never answer.
type DNSServer struct {
	conn   *net.UDPConn
	Delay  time.Duration
	Fail   atomic.Bool
	Silent atomic.Bool
}

// NewDNSServer listens on 127.0.0.1 on a random port.
func NewDNSServer(delay time.Duration) *DNSServer {
	c, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		panic(err)
	}
	s := &DNSServer{conn: c, Delay: delay}
	go s.serve()
	return s
}

// Port is the UDP port the server listens on.
func (s *DNSServer) Port() int { return s.conn.LocalAddr().(*net.UDPAddr).Port }

// Close stops the server.
func (s *DNSServer) Close() { s.conn.Close() }

func (s *DNSServer) serve() {
	buf := make([]byte, 1500)
	for {
		n, from, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		var m dnsmessage.Message
		if m.Unpack(buf[:n]) != nil || len(m.Questions) == 0 || s.Silent.Load() {
			continue
		}
		q := m.Questions[0]
		resp := dnsmessage.Message{
			Header:    dnsmessage.Header{ID: m.ID, Response: true, RecursionAvailable: true},
			Questions: m.Questions,
		}
		if s.Fail.Load() {
			resp.RCode = dnsmessage.RCodeServerFailure
		} else {
			resp.Answers = []dnsmessage.Resource{{
				Header: dnsmessage.ResourceHeader{Name: q.Name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET, TTL: 60},
				Body:   &dnsmessage.AResource{A: [4]byte{192, 0, 2, 1}},
			}}
		}
		out, _ := resp.Pack()
		go func() {
			time.Sleep(s.Delay)
			_, _ = s.conn.WriteToUDP(out, from)
		}()
	}
}
