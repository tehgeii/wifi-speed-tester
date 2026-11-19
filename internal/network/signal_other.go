//go:build !windows

package network

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

var fakeSignalN atomic.Int64

// wifiSignal on Linux reads the Wi-Fi link with iw. For UI development
// without Wi-Fi, WST_FAKE_WIFI_SIGNAL may name a JSON file holding a
// []model.WiFiInfo, returned one entry per call in a loop.
func wifiSignal(ctx context.Context) (*model.WiFiInfo, error) {
	if p := os.Getenv("WST_FAKE_WIFI_SIGNAL"); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var list []model.WiFiInfo
		if err := json.Unmarshal(b, &list); err != nil || len(list) == 0 {
			return nil, errors.New("WST_FAKE_WIFI_SIGNAL: need a JSON list of Wi-Fi readings")
		}
		w := list[int(fakeSignalN.Add(1)-1)%len(list)]
		return &w, nil
	}
	info, err := detectAdapter()
	if err != nil {
		return nil, err
	}
	if info.Connection != model.ConnWiFi {
		return nil, errors.New("not connected to Wi-Fi")
	}
	return wifiInfo(info)
}
