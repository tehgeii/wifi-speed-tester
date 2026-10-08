//go:build windows

package network

import (
	"context"
	"errors"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

var procWlanScan = modWlanapi.NewProc("WlanScan")

const dot11BssTypeAny = 3

// scanWiFi lists the access points Windows can see (read-only). With rescan
// it first asks the adapter for a fresh scan, which takes a few seconds.
func scanWiFi(ctx context.Context, rescan bool) ([]model.WiFiNetwork, error) {
	if err := procWlanOpenHandle.Find(); err != nil {
		return nil, errors.New("WLAN service API not available")
	}
	var negotiated uint32
	var h windows.Handle
	if rc, _, _ := procWlanOpenHandle.Call(2, 0, uintptr(unsafe.Pointer(&negotiated)), uintptr(unsafe.Pointer(&h))); rc != 0 {
		return nil, wlanErr("WlanOpenHandle", rc)
	}
	defer procWlanCloseHandle.Call(uintptr(h), 0)

	guid, err := scanInterface(h)
	if err != nil {
		return nil, err
	}
	if rescan {
		if rc, _, _ := procWlanScan.Call(uintptr(h), uintptr(unsafe.Pointer(&guid)), 0, 0, 0); rc == 0 {
			select { // the scan completes asynchronously; results settle in ~4 s
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(4 * time.Second):
			}
		}
	}

	// The BSSID we are associated with, to mark it in the list.
	var current [6]byte
	var size uint32
	var data unsafe.Pointer
	if rc, _, _ := procWlanQueryInterface.Call(uintptr(h), uintptr(unsafe.Pointer(&guid)), wlanOpcodeCurrentConnection, 0,
		uintptr(unsafe.Pointer(&size)), uintptr(unsafe.Pointer(&data)), 0); rc == 0 && data != nil {
		if uintptr(size) >= unsafe.Sizeof(wlanConnectionAttributes{}) {
			if attr := (*wlanConnectionAttributes)(data); attr.State == wlanStateConnected {
				current = attr.Assoc.BSSID
			}
		}
		procWlanFreeMemory.Call(uintptr(data))
	}

	var list unsafe.Pointer
	if rc, _, _ := procWlanGetNetworkBssList.Call(uintptr(h), uintptr(unsafe.Pointer(&guid)), 0, dot11BssTypeAny, 0, 0,
		uintptr(unsafe.Pointer(&list))); rc != 0 {
		return nil, wlanErr("WlanGetNetworkBssList", rc)
	}
	defer procWlanFreeMemory.Call(uintptr(list))
	count := *(*uint32)(unsafe.Add(list, 4))
	entries := unsafe.Slice((*wlanBssEntry)(unsafe.Add(list, 8)), count)
	out := make([]model.WiFiNetwork, 0, len(entries))
	for _, e := range entries {
		n := min(int(e.SSID.Length), 32)
		mhz := int(e.ChCenterFreqKHz / 1000)
		out = append(out, model.WiFiNetwork{
			SSID:          strings.ToValidUTF8(string(e.SSID.SSID[:n]), "?"),
			BSSID:         macString(e.BSSID),
			RSSI:          int(e.RSSI),
			SignalPercent: int(e.LinkQuality),
			Channel:       ChannelFromFrequency(mhz),
			Band:          BandFromFrequency(mhz),
			FrequencyMHz:  mhz,
			Connected:     current != [6]byte{} && e.BSSID == current,
		})
	}
	return out, nil
}

// scanInterface picks the connected Wi-Fi adapter, or the first one.
func scanInterface(h windows.Handle) (windows.GUID, error) {
	var list unsafe.Pointer
	if rc, _, _ := procWlanEnumInterfaces.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&list))); rc != 0 {
		return windows.GUID{}, wlanErr("WlanEnumInterfaces", rc)
	}
	defer procWlanFreeMemory.Call(uintptr(list))
	count := *(*uint32)(list)
	items := unsafe.Slice((*wlanInterfaceInfo)(unsafe.Add(list, 8)), count)
	if len(items) == 0 {
		return windows.GUID{}, errors.New("no Wi-Fi adapter found")
	}
	for _, it := range items {
		if it.State == wlanStateConnected {
			return it.GUID, nil
		}
	}
	return items[0].GUID, nil
}
