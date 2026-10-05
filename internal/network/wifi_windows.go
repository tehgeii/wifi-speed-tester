//go:build windows

package network

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

// Native Wifi API (wlanapi.dll). Read-only queries of the current connection.
var (
	modWlanapi                = windows.NewLazySystemDLL("wlanapi.dll")
	procWlanOpenHandle        = modWlanapi.NewProc("WlanOpenHandle")
	procWlanCloseHandle       = modWlanapi.NewProc("WlanCloseHandle")
	procWlanQueryInterface    = modWlanapi.NewProc("WlanQueryInterface")
	procWlanGetNetworkBssList = modWlanapi.NewProc("WlanGetNetworkBssList")
	procWlanFreeMemory        = modWlanapi.NewProc("WlanFreeMemory")
	procWlanEnumInterfaces    = modWlanapi.NewProc("WlanEnumInterfaces")
)

const (
	wlanOpcodeCurrentConnection = 7
	wlanOpcodeChannelNumber     = 8
	wlanStateConnected          = 1
)

type dot11SSID struct {
	Length uint32
	SSID   [32]byte
}

type wlanAssociationAttributes struct {
	SSID          dot11SSID
	BssType       uint32
	BSSID         [6]byte
	PhyType       uint32
	PhyIndex      uint32
	SignalQuality uint32
	RxRateKbps    uint32
	TxRateKbps    uint32
}

type wlanSecurityAttributes struct {
	SecurityEnabled int32
	OneXEnabled     int32
	AuthAlgorithm   uint32
	CipherAlgorithm uint32
}

type wlanConnectionAttributes struct {
	State       uint32
	Mode        uint32
	ProfileName [256]uint16
	Assoc       wlanAssociationAttributes
	Security    wlanSecurityAttributes
}

type wlanRateSet struct {
	Length uint32
	Rates  [126]uint16
}

type wlanBssEntry struct {
	SSID            dot11SSID
	PhyID           uint32
	BSSID           [6]byte
	BssType         uint32
	PhyType         uint32
	RSSI            int32
	LinkQuality     uint32
	InRegDomain     uint8
	BeaconPeriod    uint16
	Timestamp       uint64
	HostTimestamp   uint64
	Capability      uint16
	ChCenterFreqKHz uint32
	RateSet         wlanRateSet
	IEOffset        uint32
	IESize          uint32
}

type wlanInterfaceInfo struct {
	GUID        windows.GUID
	Description [256]uint16
	State       uint32
}

func wlanErr(op string, rc uintptr) error {
	if rc == 0 {
		return nil
	}
	if windows.Errno(rc) == windows.ERROR_ACCESS_DENIED {
		return errors.New("access denied (on Windows 11, allow apps to use Location to read Wi-Fi details)")
	}
	return fmt.Errorf("%s: %w", op, windows.Errno(rc))
}

func phyName(t uint32) string {
	switch t {
	case 4:
		return "802.11a"
	case 5:
		return "802.11b"
	case 6:
		return "802.11g"
	case 7:
		return "Wi-Fi 4 (802.11n)"
	case 8:
		return "Wi-Fi 5 (802.11ac)"
	case 10:
		return "Wi-Fi 6/6E (802.11ax)"
	case 11:
		return "Wi-Fi 7 (802.11be)"
	}
	return ""
}

func wifiInfo(net *model.NetworkInfo) (*model.WiFiInfo, error) {
	if err := procWlanOpenHandle.Find(); err != nil {
		return nil, errors.New("WLAN service API not available")
	}
	var negotiated uint32
	var h windows.Handle
	if rc, _, _ := procWlanOpenHandle.Call(2, 0, uintptr(unsafe.Pointer(&negotiated)), uintptr(unsafe.Pointer(&h))); rc != 0 {
		return nil, wlanErr("WlanOpenHandle", rc)
	}
	defer procWlanCloseHandle.Call(uintptr(h), 0)

	guid, ok := adapterGUID(net.AdapterName)
	if !ok {
		g, err := firstConnectedInterface(h)
		if err != nil {
			return nil, err
		}
		guid = g
	}

	var size uint32
	var data unsafe.Pointer
	rc, _, _ := procWlanQueryInterface.Call(uintptr(h), uintptr(unsafe.Pointer(&guid)), wlanOpcodeCurrentConnection, 0,
		uintptr(unsafe.Pointer(&size)), uintptr(unsafe.Pointer(&data)), 0)
	if rc != 0 {
		return nil, wlanErr("WlanQueryInterface", rc)
	}
	defer procWlanFreeMemory.Call(uintptr(data))
	if uintptr(size) < unsafe.Sizeof(wlanConnectionAttributes{}) {
		return nil, errors.New("unexpected WLAN connection data")
	}
	attr := (*wlanConnectionAttributes)(data)
	if attr.State != wlanStateConnected {
		return nil, errors.New("Wi-Fi adapter is not connected")
	}
	a := attr.Assoc
	ssidLen := min(int(a.SSID.Length), 32)
	w := &model.WiFiInfo{
		SSID:          string(a.SSID.SSID[:ssidLen]),
		BSSID:         macString(a.BSSID),
		SignalPercent: int(a.SignalQuality),
		PhyType:       phyName(a.PhyType),
		RxRateMbps:    float64(a.RxRateKbps) / 1000,
		TxRateMbps:    float64(a.TxRateKbps) / 1000,
	}
	if w.SSID == "" {
		w.Note = "SSID hidden by Windows (Location permission may be required)."
	}

	var chPtr unsafe.Pointer
	if rc, _, _ := procWlanQueryInterface.Call(uintptr(h), uintptr(unsafe.Pointer(&guid)), wlanOpcodeChannelNumber, 0,
		uintptr(unsafe.Pointer(&size)), uintptr(unsafe.Pointer(&chPtr)), 0); rc == 0 && chPtr != nil {
		w.Channel = int(*(*uint32)(chPtr))
		procWlanFreeMemory.Call(uintptr(chPtr))
	}

	// The BSS list gives the exact centre frequency and RSSI of the access
	// point we are associated with.
	ssid := a.SSID
	var list unsafe.Pointer
	if rc, _, _ := procWlanGetNetworkBssList.Call(uintptr(h), uintptr(unsafe.Pointer(&guid)), uintptr(unsafe.Pointer(&ssid)),
		uintptr(a.BssType), uintptr(attr.Security.SecurityEnabled), 0, uintptr(unsafe.Pointer(&list))); rc == 0 && list != nil {
		count := *(*uint32)(unsafe.Add(list, 4))
		entries := unsafe.Slice((*wlanBssEntry)(unsafe.Add(list, 8)), count)
		for _, e := range entries {
			if e.BSSID == a.BSSID {
				w.FrequencyMHz = int(e.ChCenterFreqKHz / 1000)
				w.RSSI = int(e.RSSI)
				break
			}
		}
		procWlanFreeMemory.Call(uintptr(list))
	}

	if w.FrequencyMHz > 0 {
		w.Band = BandFromFrequency(w.FrequencyMHz)
		if w.Channel == 0 {
			w.Channel = ChannelFromFrequency(w.FrequencyMHz)
		}
	}
	if w.Band == "" {
		w.Band = BandFromChannel(w.Channel)
	}
	return w, nil
}

func firstConnectedInterface(h windows.Handle) (windows.GUID, error) {
	var list unsafe.Pointer
	if rc, _, _ := procWlanEnumInterfaces.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&list))); rc != 0 {
		return windows.GUID{}, wlanErr("WlanEnumInterfaces", rc)
	}
	defer procWlanFreeMemory.Call(uintptr(list))
	count := *(*uint32)(list)
	items := unsafe.Slice((*wlanInterfaceInfo)(unsafe.Add(list, 8)), count)
	for _, it := range items {
		if it.State == wlanStateConnected {
			return it.GUID, nil
		}
	}
	return windows.GUID{}, errors.New("no connected Wi-Fi interface")
}

func macString(b [6]byte) string {
	parts := make([]string, 6)
	for i, x := range b {
		parts[i] = fmt.Sprintf("%02X", x)
	}
	return strings.Join(parts, ":")
}
