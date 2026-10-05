//go:build windows

package network

import (
	"errors"
	"net"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

const (
	ifTypeEthernet  = 6
	ifTypeIEEE80211 = 71
	ifTypeTunnel    = 131
	ifTypePropVirt  = 53
	ifTypeWWANPP    = 243
	ifTypeWWANPP2   = 244
)

func adapters() ([]*windows.IpAdapterAddresses, error) {
	flags := uint32(windows.GAA_FLAG_INCLUDE_GATEWAYS | windows.GAA_FLAG_SKIP_MULTICAST | windows.GAA_FLAG_SKIP_ANYCAST)
	size := uint32(16 << 10)
	for i := 0; i < 3; i++ {
		buf := make([]byte, size)
		first := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0]))
		err := windows.GetAdaptersAddresses(windows.AF_UNSPEC, flags, 0, first, &size)
		if err == windows.ERROR_BUFFER_OVERFLOW {
			continue
		}
		if err != nil {
			return nil, err
		}
		var out []*windows.IpAdapterAddresses
		for a := first; a != nil; a = a.Next {
			out = append(out, a)
		}
		return out, nil
	}
	return nil, errors.New("GetAdaptersAddresses: buffer too small")
}

func detectAdapter() (*model.NetworkInfo, error) {
	list, err := adapters()
	if err != nil {
		return nil, err
	}
	// Ask Windows which interface it would use to reach the internet.
	var best uint32
	_ = windows.GetBestInterfaceEx(&windows.SockaddrInet4{Addr: [4]byte{1, 1, 1, 1}}, &best)

	var chosen *windows.IpAdapterAddresses
	for _, a := range list {
		if a.OperStatus == windows.IfOperStatusUp && best != 0 && a.IfIndex == best {
			chosen = a
			break
		}
	}
	if chosen == nil {
		// Fall back to the up adapter with a gateway and the lowest metric.
		for _, a := range list {
			if a.OperStatus != windows.IfOperStatusUp || a.FirstGatewayAddress == nil {
				continue
			}
			if chosen == nil || a.Ipv4Metric < chosen.Ipv4Metric {
				chosen = a
			}
		}
	}
	if chosen == nil {
		return nil, errors.New("no active network adapter with a default gateway was found")
	}

	info := &model.NetworkInfo{
		AdapterName: windows.UTF16PtrToString(chosen.FriendlyName),
		Description: windows.UTF16PtrToString(chosen.Description),
	}
	base := model.ConnUnknown
	switch chosen.IfType {
	case ifTypeIEEE80211:
		base = model.ConnWiFi
	case ifTypeEthernet:
		base = model.ConnEthernet
	case ifTypeWWANPP, ifTypeWWANPP2:
		base = model.ConnCellular
	case ifTypeTunnel, ifTypePropVirt:
		base = model.ConnVPN
	}
	info.Connection = classify(base, info.Description)
	if chosen.IfType == ifTypeIEEE80211 {
		info.Connection = model.ConnWiFi // a Wi-Fi radio is always reported as Wi-Fi
	}

	var ips []net.IP
	for u := chosen.FirstUnicastAddress; u != nil; u = u.Next {
		ips = append(ips, u.Address.IP())
	}
	splitAddrs(ips, info)
	for g := chosen.FirstGatewayAddress; g != nil; g = g.Next {
		ip := g.Address.IP()
		if ip == nil {
			continue
		}
		if ip.To4() != nil {
			info.Gateway = ip.String()
			break
		}
		if info.Gateway == "" {
			info.Gateway = ip.String()
		}
	}
	for d := chosen.FirstDnsServerAddress; d != nil; d = d.Next {
		if ip := d.Address.IP(); ip != nil && !ip.IsLinkLocalUnicast() {
			info.DNS = append(info.DNS, ip.String())
		}
	}
	return info, nil
}

// adapterGUID returns the interface GUID used by the WLAN API for the given
// adapter friendly name.
func adapterGUID(friendlyName string) (windows.GUID, bool) {
	list, err := adapters()
	if err != nil {
		return windows.GUID{}, false
	}
	for _, a := range list {
		if windows.UTF16PtrToString(a.FriendlyName) == friendlyName {
			g, err := windows.GUIDFromString(windows.BytePtrToString(a.AdapterName))
			return g, err == nil
		}
	}
	return windows.GUID{}, false
}
