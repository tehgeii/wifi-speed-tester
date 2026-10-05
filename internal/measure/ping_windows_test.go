//go:build windows && (amd64 || arm64)

package measure

import (
	"testing"
	"unsafe"
)

func TestIcmpEchoReplySize(t *testing.T) {
	if s := unsafe.Sizeof(icmpEchoReply{}); s != 40 {
		t.Errorf("ICMP_ECHO_REPLY = %d, want 40", s)
	}
}
