package update

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// release serves a fake GitHub release: the API answer, the zip and the
// checksums, from a local server standing in for github.com.
type release struct {
	srv     *httptest.Server
	zip     []byte
	sums    string
	zipName string
}

func makeZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func newRelease(t *testing.T, zipData []byte) *release {
	t.Helper()
	rel := &release{zip: zipData, zipName: "WiFiSpeedTester-Portable-2.1.0-win-x64.zip"}
	sum := sha256.Sum256(zipData)
	rel.sums = hex.EncodeToString(sum[:]) + "  " + rel.zipName + "\n" + strings.Repeat("0", 64) + "  WiFiSpeedTester-Portable-2.1.0-win-arm64.zip\n"
	rel.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dl := "/" + Repo + "/releases/download/v2.1.0/"
		switch r.URL.Path {
		case "/repos/" + Repo + "/releases/latest":
			base := rel.srv.URL + dl
			fmt.Fprintf(w, `{"tag_name":"v2.1.0","html_url":"https://github.com/%s/releases/tag/v2.1.0","assets":[
				{"name":"WiFiSpeedTester-Portable-2.1.0-win-arm64.zip","browser_download_url":"%sWiFiSpeedTester-Portable-2.1.0-win-arm64.zip"},
				{"name":"%s","browser_download_url":"%s%s"},
				{"name":"SHA256SUMS.txt","browser_download_url":"%sSHA256SUMS.txt"},
				{"name":"evil.zip","browser_download_url":"https://example.com/evil.zip"}]}`,
				Repo, base, rel.zipName, base, rel.zipName, base)
		case dl + rel.zipName:
			w.Write(rel.zip)
		case dl + "SHA256SUMS.txt":
			w.Write([]byte(rel.sums))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(rel.srv.Close)
	old := downloadPrefix
	downloadPrefix = rel.srv.URL + "/" + Repo + "/releases/download/"
	t.Cleanup(func() { downloadPrefix = old })
	return rel
}

func (rel *release) check(t *testing.T, goarch string) Info {
	t.Helper()
	info, err := Check(context.Background(), rel.srv.Client(), rel.srv.URL, "2.0.1", goarch)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

// installDir is a portable folder with the running program and its notes.
func installDir(t *testing.T) (dir, exe string) {
	dir = t.TempDir()
	exe = filepath.Join(dir, ExeName)
	os.WriteFile(exe, []byte("old program"), 0o755)
	os.WriteFile(filepath.Join(dir, "README.txt"), []byte("old readme"), 0o644)
	os.WriteFile(filepath.Join(dir, "WiFiSpeedTester.config.json"), []byte(`{"mine":true}`), 0o644)
	return dir, exe
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

func TestCheckPicksAssetsForThisBuild(t *testing.T) {
	rel := newRelease(t, []byte("zip"))
	info := rel.check(t, "amd64")
	if !info.Installable() || info.ZipName != rel.zipName || !strings.HasSuffix(info.SumsURL, "/SHA256SUMS.txt") {
		t.Fatalf("amd64: %+v", info)
	}
	if info := rel.check(t, "arm64"); !strings.HasSuffix(info.ZipURL, "-win-arm64.zip") {
		t.Fatalf("arm64: %+v", info)
	}
	if info := rel.check(t, "386"); info.Installable() {
		t.Fatalf("no zip for 386, but installable: %+v", info)
	}
}

func TestInstall(t *testing.T) {
	rel := newRelease(t, makeZip(t, map[string]string{
		ExeName: "new program", "README.txt": "new readme", "LICENSE.txt": "licence",
		"WiFiSpeedTester.config.example.json": "{}",
	}))
	info := rel.check(t, "amd64")
	dir, exe := installDir(t)
	var stages []string
	err := Install(context.Background(), rel.srv.Client(), info, exe, func(stage string, done, total int64) {
		if len(stages) == 0 || stages[len(stages)-1] != stage {
			stages = append(stages, stage)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(stages, ","); got != "download,verify,install" {
		t.Errorf("stages = %s", got)
	}
	if read(t, exe) != "new program" || read(t, exe+".old") != "old program" {
		t.Errorf("exe = %q, old = %q", read(t, exe), read(t, exe+".old"))
	}
	if read(t, filepath.Join(dir, "README.txt")) != "new readme" {
		t.Error("README.txt not refreshed")
	}
	if _, err := os.Stat(filepath.Join(dir, "LICENSE.txt")); err == nil {
		t.Error("a file the user had removed came back")
	}
	if read(t, filepath.Join(dir, "WiFiSpeedTester.config.json")) != `{"mine":true}` {
		t.Error("the user's configuration was touched")
	}
	if err := Cleanup(exe); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(exe + ".old"); err == nil {
		t.Error(".old not removed")
	}
	if err := Cleanup(exe); err != nil {
		t.Fatalf("second cleanup: %v", err)
	}
}

// Every failure leaves the original program in place and no temporary file.
func TestInstallFailuresKeepTheProgram(t *testing.T) {
	good := map[string]string{ExeName: "new program"}
	cases := []struct {
		name  string
		zip   []byte
		setup func(rel *release, info *Info)
		want  string
	}{
		{"checksum mismatch", makeZip(t, good), func(rel *release, _ *Info) {
			rel.zip = makeZip(t, map[string]string{ExeName: "tampered"})
		}, "checksum mismatch"},
		{"no checksum for the zip", makeZip(t, good), func(rel *release, _ *Info) {
			rel.sums = strings.Repeat("0", 64) + "  other.zip\n"
		}, "no checksum"},
		{"zip without the program", makeZip(t, map[string]string{"README.txt": "x"}), nil, "has no " + ExeName},
		{"not a zip", []byte("not a zip at all"), nil, "not a valid zip"},
		{"URL outside the releases", makeZip(t, good), func(_ *release, info *Info) {
			info.ZipURL = "https://example.com/" + info.ZipName
		}, "refusing"},
		{"server error", makeZip(t, good), func(_ *release, info *Info) {
			info.ZipURL += ".missing"
		}, "404"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rel := newRelease(t, c.zip)
			info := rel.check(t, "amd64")
			if c.setup != nil {
				c.setup(rel, &info)
			}
			dir, exe := installDir(t)
			err := Install(context.Background(), rel.srv.Client(), info, exe, nil)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
			if read(t, exe) != "old program" {
				t.Errorf("program changed: %q", read(t, exe))
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 3 {
				var names []string
				for _, e := range entries {
					names = append(names, e.Name())
				}
				t.Errorf("leftover files: %v", names)
			}
		})
	}
}

func TestInstallSizeLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(bytes.Repeat([]byte("x"), 2000))
	}))
	defer srv.Close()
	if _, err := fetch(context.Background(), srv.Client(), srv.URL, 1000, nil); err == nil || !strings.Contains(err.Error(), "large") {
		t.Fatalf("err = %v", err)
	}
	if b, err := fetch(context.Background(), srv.Client(), srv.URL, 2000, nil); err != nil || len(b) != 2000 {
		t.Fatalf("at the limit: %d bytes, %v", len(b), err)
	}
}

func TestInstallUnwritableFolder(t *testing.T) {
	if runtime.GOOS == "windows" || os.Getuid() == 0 {
		t.Skip("a read-only folder is not enforced for root or on Windows")
	}
	rel := newRelease(t, makeZip(t, map[string]string{ExeName: "new program"}))
	info := rel.check(t, "amd64")
	dir, exe := installDir(t)
	os.Chmod(dir, 0o555)
	defer os.Chmod(dir, 0o755)
	err := Install(context.Background(), rel.srv.Client(), info, exe, nil)
	if err == nil || !strings.Contains(err.Error(), "manually") || read(t, exe) != "old program" {
		t.Fatalf("err = %v, exe = %q", err, read(t, exe))
	}
}
