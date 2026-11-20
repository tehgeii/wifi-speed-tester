// Package update checks GitHub Releases for a newer version. It only runs
// when the user has the check enabled or presses "Check for updates", and it
// sends nothing but a plain request for the latest release. Install, run
// only when the user presses "Update now", downloads the portable zip for
// this build and replaces the executable (see install.go).
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Repo is the GitHub repository releases are published to.
const Repo = "tehgeii/wifi-speed-tester"

// ReleasesPage is where users download new versions.
const ReleasesPage = "https://github.com/" + Repo + "/releases/latest"

// downloadPrefix is where release files may come from; anything else in
// the API answer is ignored. A variable so tests can use a local server.
var downloadPrefix = "https://github.com/" + Repo + "/releases/download/"

// Info describes the latest published release.
type Info struct {
	Latest  string `json:"latest"`  // e.g. "1.1.0"
	Current string `json:"current"` // this build
	Newer   bool   `json:"newer"`   // Latest is newer than Current
	URL     string `json:"url"`     // release page
	// CanInstall is set by the app when "Update now" can work: the release
	// has a zip and checksums for this build and the window can restart.
	CanInstall bool `json:"canInstall"`

	ZipName string `json:"-"` // e.g. WiFiSpeedTester-Portable-2.1.0-win-x64.zip
	ZipURL  string `json:"-"`
	SumsURL string `json:"-"` // SHA256SUMS.txt
}

// Installable reports whether the release has the files Install needs.
func (i Info) Installable() bool { return i.Newer && i.ZipURL != "" && i.SumsURL != "" }

// assetSuffix names the zip for an architecture, as scripts/build.sh does.
func assetSuffix(goarch string) string {
	switch goarch {
	case "amd64":
		return "-win-x64.zip"
	case "arm64":
		return "-win-arm64.zip"
	}
	return ""
}

// Check asks GitHub for the latest release of Repo. goarch picks the zip
// for this build (runtime.GOARCH).
func Check(ctx context.Context, client *http.Client, apiBase, current, goarch string) (Info, error) {
	info := Info{Current: current, URL: ReleasesPage}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/repos/"+Repo+"/releases/latest", nil)
	if err != nil {
		return info, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "WiFiSpeedTester/"+current)
	resp, err := client.Do(req)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return info, fmt.Errorf("GitHub: %s", resp.Status)
	}
	var body struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return info, err
	}
	info.Latest = strings.TrimPrefix(body.TagName, "v")
	if strings.HasPrefix(body.HTMLURL, "https://github.com/"+Repo+"/") {
		info.URL = body.HTMLURL
	}
	info.Newer = Newer(info.Latest, current)
	suffix := assetSuffix(goarch)
	for _, a := range body.Assets {
		if !strings.HasPrefix(a.URL, downloadPrefix) || strings.ContainsAny(a.Name, `/\`) {
			continue
		}
		switch {
		case a.Name == "SHA256SUMS.txt":
			info.SumsURL = a.URL
		case suffix != "" && strings.HasPrefix(a.Name, "WiFiSpeedTester-Portable-") && strings.HasSuffix(a.Name, suffix):
			info.ZipName, info.ZipURL = a.Name, a.URL
		}
	}
	return info, nil
}

// Newer reports whether version a is newer than b. Development builds
// ("dev", "0.0.0-<sha>") are never considered up to date with a release.
func Newer(a, b string) bool {
	pa, okA := parse(a)
	pb, okB := parse(b)
	if !okA {
		return false
	}
	if !okB {
		return true
	}
	for i := 0; i < 3; i++ {
		if pa.n[i] != pb.n[i] {
			return pa.n[i] > pb.n[i]
		}
	}
	// 1.2.0 is newer than 1.2.0-beta.1
	return pa.pre == "" && pb.pre != ""
}

type version struct {
	n   [3]int
	pre string
}

func parse(s string) (version, bool) {
	var v version
	s = strings.TrimPrefix(s, "v")
	core, pre, _ := strings.Cut(s, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return v, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return v, false
		}
		v.n[i] = n
	}
	v.pre = pre
	return v, v.n != [3]int{} // 0.0.0-<sha> is a CI build, not a release
}
