package app

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// osVersion names the Windows release; Windows 11 still reports 10.0, so
// the build number tells them apart.
func osVersion() string {
	v := windows.RtlGetVersion()
	name := "Windows 10"
	if v.MajorVersion == 10 && v.BuildNumber >= 22000 {
		name = "Windows 11"
	} else if v.MajorVersion != 10 {
		name = "Windows"
	}
	return fmt.Sprintf("%s (%d.%d.%d)", name, v.MajorVersion, v.MinorVersion, v.BuildNumber)
}
