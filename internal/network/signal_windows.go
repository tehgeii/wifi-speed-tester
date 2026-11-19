//go:build windows

package network

import (
	"context"

	"github.com/tehgeii/wifi-speed-tester/internal/model"
)

func wifiSignal(ctx context.Context) (*model.WiFiInfo, error) {
	// No adapter name: the first connected Wi-Fi interface is used.
	return wifiInfo(&model.NetworkInfo{})
}
