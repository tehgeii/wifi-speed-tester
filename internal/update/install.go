package update

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Limits on what is downloaded and unpacked; a release zip is about 3.5 MB.
const (
	maxSums = 64 << 10
	maxZip  = 64 << 20
	maxExe  = 128 << 20
)

// ExeName is the program inside the release zip.
const ExeName = "WiFiSpeedTester.exe"

// Files from the zip that are refreshed next to the program when present
// there, so the notes and licences match the new version. The user's own
// configuration file is never touched (the zip only has the example).
var docFiles = []string{"README.txt", "LICENSE.txt", "THIRD_PARTY_NOTICES.txt", "WiFiSpeedTester.config.example.json"}

// Progress reports a stage ("download", "verify", "install") and, while
// downloading, the bytes received so far and the total (0 when unknown).
type Progress func(stage string, done, total int64)

// Install downloads the release zip, checks it against SHA256SUMS.txt from
// the same release, and swaps the executable at exePath: the running one is
// renamed to <exe>.old (Windows allows renaming a running program, not
// deleting it) and the new one takes its name. The new version starts on the
// next launch; Cleanup removes the .old file then. On any error the original
// executable stays in place and temporary files are removed.
//
// The checksum protects against a broken or incomplete download. Both files
// come from the same GitHub release, so it is not a signature.
func Install(ctx context.Context, client *http.Client, info Info, exePath string, progress Progress) error {
	if progress == nil {
		progress = func(string, int64, int64) {}
	}
	if !info.Installable() {
		return errors.New("this release has no download for this version of Windows")
	}
	if !strings.HasPrefix(info.ZipURL, downloadPrefix) || !strings.HasPrefix(info.SumsURL, downloadPrefix) {
		return errors.New("refusing to download from outside the project's releases")
	}

	progress("download", 0, 0)
	sums, err := fetch(ctx, client, info.SumsURL, maxSums, nil)
	if err != nil {
		return fmt.Errorf("SHA256SUMS.txt: %w", err)
	}
	want, ok := findSum(sums, info.ZipName)
	if !ok {
		return fmt.Errorf("SHA256SUMS.txt has no checksum for %s", info.ZipName)
	}
	data, err := fetch(ctx, client, info.ZipURL, maxZip, func(done, total int64) { progress("download", done, total) })
	if err != nil {
		return fmt.Errorf("%s: %w", info.ZipName, err)
	}

	progress("verify", 0, 0)
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != want {
		return errors.New("the download is damaged (checksum mismatch); nothing was changed")
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("the download is not a valid zip: %w", err)
	}
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[strings.ToLower(path.Base(f.Name))] = f
	}
	exeFile := files[strings.ToLower(ExeName)]
	if exeFile == nil {
		return errors.New("the download has no " + ExeName)
	}

	progress("install", 0, 0)
	newPath := exePath + ".new"
	if err := unpack(exeFile, newPath); err != nil {
		os.Remove(newPath)
		return fmt.Errorf("could not write next to the program (%w); download the new version manually", err)
	}
	oldPath := exePath + ".old"
	os.Remove(oldPath) // left over from an earlier update
	if err := os.Rename(exePath, oldPath); err != nil {
		os.Remove(newPath)
		return fmt.Errorf("could not replace the program: %w", err)
	}
	if err := os.Rename(newPath, exePath); err != nil {
		os.Rename(oldPath, exePath)
		os.Remove(newPath)
		return fmt.Errorf("could not replace the program: %w", err)
	}

	dir := filepath.Dir(exePath)
	for _, name := range docFiles {
		f := files[strings.ToLower(name)]
		dst := filepath.Join(dir, name)
		if f == nil {
			continue
		}
		if _, err := os.Stat(dst); err != nil {
			continue // only refresh files the user kept
		}
		if unpack(f, dst+".new") == nil {
			if os.Rename(dst+".new", dst) != nil {
				os.Remove(dst + ".new")
			}
		} else {
			os.Remove(dst + ".new")
		}
	}
	return nil
}

// Cleanup removes what an update left behind: the previous program, once it
// has exited, and a half-written new one.
func Cleanup(exePath string) error {
	os.Remove(exePath + ".new")
	err := os.Remove(exePath + ".old")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func fetch(ctx context.Context, client *http.Client, url string, limit int64, progress func(done, total int64)) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WiFiSpeedTester-update")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}
	if resp.ContentLength > limit {
		return nil, errors.New("the file is unexpectedly large")
	}
	var r io.Reader = io.LimitReader(resp.Body, limit+1)
	if progress != nil {
		r = &countingReader{r: r, total: max(resp.ContentLength, 0), report: progress}
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("the file is unexpectedly large")
	}
	return data, nil
}

// countingReader reports progress about every 256 KB.
type countingReader struct {
	r           io.Reader
	done, total int64
	last        int64
	report      func(done, total int64)
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.done += int64(n)
	if c.done-c.last >= 256<<10 || (err == io.EOF && c.done != c.last) {
		c.last = c.done
		c.report(c.done, c.total)
	}
	return n, err
}

// findSum looks name up in sha256sum output ("<hex>  <name>" or "<hex> *<name>").
func findSum(sums []byte, name string) (string, bool) {
	sc := bufio.NewScanner(bytes.NewReader(sums))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name && len(fields[0]) == 64 {
			return strings.ToLower(fields[0]), true
		}
	}
	return "", false
}

func unpack(f *zip.File, dst string) error {
	if f.UncompressedSize64 > maxExe {
		return errors.New("file in the zip is too large")
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	n, err := io.Copy(out, io.LimitReader(rc, maxExe+1))
	if err == nil && n > maxExe {
		err = errors.New("file in the zip is too large")
	}
	if err == nil {
		err = out.Sync()
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}
