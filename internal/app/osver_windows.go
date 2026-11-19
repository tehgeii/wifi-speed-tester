package app

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// verNTWorkstation is the ProductType of desktop Windows; servers report
// another value but share build numbers with desktop releases.
const verNTWorkstation = 1

// osVersion names the Windows release; Windows 11 still reports 10.0, so
// the build number tells them apart.
func osVersion() string {
	v := windows.RtlGetVersion()
	name := "Windows"
	switch {
	case v.ProductType != verNTWorkstation:
		name = "Windows Server"
	case v.MajorVersion == 10 && v.BuildNumber >= 22000:
		name = "Windows 11"
	case v.MajorVersion == 10:
		name = "Windows 10"
	}
	return fmt.Sprintf("%s (%d.%d.%d)", name, v.MajorVersion, v.MinorVersion, v.BuildNumber)
}
