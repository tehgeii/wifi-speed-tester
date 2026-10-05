//go:build windows

package measure

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The Windows ICMP helper API (iphlpapi.dll) sends echo requests without
// administrator rights, unlike raw sockets.
var (
	modIphlpapi         = windows.NewLazySystemDLL("iphlpapi.dll")
	procIcmpCreateFile  = modIphlpapi.NewProc("IcmpCreateFile")
	procIcmpCloseHandle = modIphlpapi.NewProc("IcmpCloseHandle")
	procIcmpSendEcho    = modIphlpapi.NewProc("IcmpSendEcho")
)

const (
	ipSuccess         = 0
	ipReqTimedOut     = 11010
	ipDestHostUnreach = 11003
	ipDestNetUnreach  = 11002
	ipTTLExpired      = 11013
)

// icmpEchoReply mirrors ICMP_ECHO_REPLY on 64-bit Windows.
type icmpEchoReply struct {
	Address       uint32
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       struct {
		Ttl         uint8
		Tos         uint8
		Flags       uint8
		OptionsSize uint8
		OptionsData uintptr
	}
}

type winPinger struct{}

func newPlatformPinger() (Pinger, error) {
	if err := procIcmpSendEcho.Find(); err != nil {
		return nil, fmt.Errorf("ICMP API unavailable: %w", err)
	}
	return winPinger{}, nil
}

func (winPinger) Close() error { return nil }

func (winPinger) Ping(ctx context.Context, ip net.IP, timeout time.Duration) (time.Duration, error) {
	v4 := ip.To4()
	if v4 == nil {
		return 0, errors.New("only IPv4 ping targets are supported")
	}
	type res struct {
		rtt time.Duration
		err error
	}
	ch := make(chan res, 1)
	// IcmpSendEcho blocks for up to timeout; run it off the caller's goroutine
	// so cancellation returns immediately.
	go func() {
		rtt, err := sendEcho(v4, timeout)
		ch <- res{rtt, err}
	}()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case r := <-ch:
		return r.rtt, r.err
	}
}

func sendEcho(v4 net.IP, timeout time.Duration) (time.Duration, error) {
	h, _, err := procIcmpCreateFile.Call()
	if windows.Handle(h) == windows.InvalidHandle {
		return 0, fmt.Errorf("IcmpCreateFile: %w", err)
	}
	defer procIcmpCloseHandle.Call(h)

	payload := []byte("WiFiSpeedTester-ping-payload-32b")
	reply := make([]byte, int(unsafe.Sizeof(icmpEchoReply{}))+len(payload)+64)
	dst := binary.LittleEndian.Uint32(v4) // IPAddr is stored in network order

	start := time.Now()
	n, _, callErr := procIcmpSendEcho.Call(
		h,
		uintptr(dst),
		uintptr(unsafe.Pointer(&payload[0])),
		uintptr(len(payload)),
		0,
		uintptr(unsafe.Pointer(&reply[0])),
		uintptr(len(reply)),
		uintptr(timeout.Milliseconds()),
	)
	elapsed := time.Since(start)
	if n == 0 {
		if errno, ok := callErr.(windows.Errno); ok {
			return 0, icmpStatusError(uint32(errno))
		}
		return 0, ErrTimeout
	}
	r := (*icmpEchoReply)(unsafe.Pointer(&reply[0]))
	if r.Status != ipSuccess {
		return 0, icmpStatusError(r.Status)
	}
	// RoundTripTime only has millisecond resolution (0 on fast LANs); the
	// wall-clock time around the call is sub-millisecond and never lower.
	rtt := time.Duration(r.RoundTripTime) * time.Millisecond
	if elapsed > rtt && elapsed < rtt+time.Millisecond {
		rtt = elapsed
	} else if rtt == 0 {
		rtt = elapsed
	}
	return rtt, nil
}

func icmpStatusError(code uint32) error {
	switch code {
	case ipReqTimedOut:
		return ErrTimeout
	case ipDestHostUnreach:
		return errors.New("destination host unreachable")
	case ipDestNetUnreach:
		return errors.New("destination network unreachable")
	case ipTTLExpired:
		return errors.New("TTL expired in transit")
	}
	return fmt.Errorf("ICMP error %d", code)
}
