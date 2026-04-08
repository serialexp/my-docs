// ABOUTME: Auto-downloads and caches the ripgrep (rg) binary for local search.
// ABOUTME: Detects OS/arch, downloads the matching release tarball, extracts rg to ~/.cache/my-docs/bin/.

package ripgrep

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const version = "15.1.0"
const downloadURL = "https://github.com/BurntSushi/ripgrep/releases/download"

// platformTargets maps GOOS/GOARCH to ripgrep release target triples.
var platformTargets = map[string]string{
	"linux/amd64":  "x86_64-unknown-linux-musl",
	"linux/arm64":  "aarch64-unknown-linux-gnu",
	"darwin/amd64": "x86_64-apple-darwin",
	"darwin/arm64": "aarch64-apple-darwin",
}

// AssetName returns the release tarball filename for the given OS/arch.
func AssetName(goos, goarch string) (string, error) {
	key := goos + "/" + goarch
	target, ok := platformTargets[key]
	if !ok {
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}
	return fmt.Sprintf("ripgrep-%s-%s.tar.gz", version, target), nil
}

func binDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "my-docs", "bin"), nil
}

// BinPath returns the path to the rg binary, downloading it if necessary.
func BinPath() (string, error) {
	dir, err := binDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, "rg")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	if err := download(path); err != nil {
		return "", err
	}
	return path, nil
}

func download(destPath string) error {
	asset, err := AssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/%s", downloadURL, version, asset)
	fmt.Fprintf(os.Stderr, "[my-docs] downloading ripgrep for local search...\n")

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("downloading ripgrep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading ripgrep: HTTP %d", resp.StatusCode)
	}

	return extractRg(resp.Body, destPath)
}

func extractRg(r io.Reader, destPath string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("decompressing ripgrep tarball: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading ripgrep tarball: %w", err)
		}

		// The binary is at <dir>/rg inside the tarball
		if hdr.Typeflag == tar.TypeReg && strings.HasSuffix(hdr.Name, "/rg") {
			if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
				return err
			}

			f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, tr); err != nil {
				os.Remove(destPath)
				return fmt.Errorf("extracting rg binary: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("rg binary not found in tarball")
}
