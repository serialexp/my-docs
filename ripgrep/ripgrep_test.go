// ABOUTME: Tests for ripgrep binary auto-download.
// ABOUTME: Covers OS/arch to asset name mapping and tarball extraction.

package ripgrep

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestAssetName(t *testing.T) {
	tests := []struct {
		goos, goarch string
		want         string
		wantErr      bool
	}{
		{"linux", "amd64", "ripgrep-" + version + "-x86_64-unknown-linux-musl.tar.gz", false},
		{"linux", "arm64", "ripgrep-" + version + "-aarch64-unknown-linux-gnu.tar.gz", false},
		{"darwin", "amd64", "ripgrep-" + version + "-x86_64-apple-darwin.tar.gz", false},
		{"darwin", "arm64", "ripgrep-" + version + "-aarch64-apple-darwin.tar.gz", false},
		{"windows", "amd64", "", true},
		{"freebsd", "amd64", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"/"+tt.goarch, func(t *testing.T) {
			got, err := AssetName(tt.goos, tt.goarch)
			if tt.wantErr {
				if err == nil {
					t.Errorf("AssetName(%q, %q) = %q, want error", tt.goos, tt.goarch, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("AssetName(%q, %q) error = %v", tt.goos, tt.goarch, err)
			}
			if got != tt.want {
				t.Errorf("AssetName(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
			}
		})
	}
}

func TestExtractRg(t *testing.T) {
	// Build a fake tarball with a rg binary inside
	fakeContent := []byte("#!/bin/fake rg binary")
	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	hdr := &tar.Header{
		Name: "ripgrep-15.1.0-x86_64-unknown-linux-musl/rg",
		Mode: 0o755,
		Size: int64(len(fakeContent)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(fakeContent); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()

	// Extract to temp dir
	dir := t.TempDir()
	destPath := filepath.Join(dir, "rg")

	if err := extractRg(&buf, destPath); err != nil {
		t.Fatalf("extractRg() error = %v", err)
	}

	// Verify the binary was extracted
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("reading extracted binary: %v", err)
	}
	if string(data) != string(fakeContent) {
		t.Errorf("extracted content = %q, want %q", string(data), string(fakeContent))
	}

	// Verify it's executable
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("extracted binary is not executable: mode = %v", info.Mode())
	}
}

func TestExtractRg_NoBinary(t *testing.T) {
	// Build a tarball without an rg binary
	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	hdr := &tar.Header{
		Name: "ripgrep-15.1.0/README.md",
		Mode: 0o644,
		Size: 5,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()

	dir := t.TempDir()
	destPath := filepath.Join(dir, "rg")

	err := extractRg(&buf, destPath)
	if err == nil {
		t.Error("extractRg() should return error when rg binary not found in tarball")
	}
}
