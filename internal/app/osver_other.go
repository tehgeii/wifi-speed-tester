//go:build !windows

package app

import "runtime"

func osVersion() string { return runtime.GOOS }
