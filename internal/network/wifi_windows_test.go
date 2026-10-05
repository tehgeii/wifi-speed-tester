//go:build windows && (amd64 || arm64)

package network

import (
	"testing"
	"unsafe"
)

// The Go mirrors of the WLAN API structs must match the 64-bit Windows SDK.
func TestWlanStructSizes(t *testing.T) {
	if s := unsafe.Sizeof(wlanConnectionAttributes{}); s != 604 {
		t.Errorf("WLAN_CONNECTION_ATTRIBUTES = %d, want 604", s)
	}
	if s := unsafe.Sizeof(wlanBssEntry{}); s != 360 {
		t.Errorf("WLAN_BSS_ENTRY = %d, want 360", s)
	}
	if s := unsafe.Sizeof(wlanInterfaceInfo{}); s != 532 {
		t.Errorf("WLAN_INTERFACE_INFO = %d, want 532", s)
	}
}
