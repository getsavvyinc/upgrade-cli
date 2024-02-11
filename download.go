package upgrade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/getsavvyinc/upgrade-cli/release"
)

type cleanupFn func() error

type AssetDownloader interface {
	DownloadAsset(ctx context.Context, ReleaseAssets []release.Asset) (*DownloadInfo, cleanupFn, error)
}

type DownloadInfo struct {
	Checksum                 string
	DownloadedBinaryFilePath string
}

type downloader struct {
	os             string
	arch           string
	executablePath string
}

var _ AssetDownloader = (*downloader)(nil)

type AssetDownloadOpt func(*downloader)

func WithOS(os string) AssetDownloadOpt {
	return func(d *downloader) {
		d.os = os
	}
}

func WithArch(arch string) AssetDownloadOpt {
	return func(d *downloader) {
		d.arch = arch
	}
}

func NewAssetDownloader(executablePath string, opts ...AssetDownloadOpt) AssetDownloader {
	d := &downloader{
		os:             runtime.GOOS,
		arch:           runtime.GOARCH,
		executablePath: executablePath,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

var ErrNoAsset = errors.New("no asset found")

func (d *downloader) DownloadAsset(ctx context.Context, assets []release.Asset) (*DownloadInfo, cleanupFn, error) {
	// iterate through the assets and find the one that matches the os and arch
	suffix := d.os + "_" + d.arch
	for _, asset := range assets {
		if strings.HasSuffix(asset.BrowserDownloadURL, suffix) {
			return d.downloadAsset(ctx, asset.BrowserDownloadURL)
		}
	}
	return nil, nil, fmt.Errorf("%w: os:%s arch:%s", ErrNoAsset, d.os, d.arch)
}

func (d *downloader) downloadAsset(ctx context.Context, url string) (*DownloadInfo, cleanupFn, error) {
	executable := filepath.Base(d.executablePath)

	// Download the file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", executable)
	if err != nil {
		return nil, nil, err
	}
	defer tmpFile.Close()

	cleanupFn := func() error {
		return os.Remove(tmpFile.Name())
	}

	// sha256 checksum
	hasher := sha256.New()

	// Write the response body to the temporary file and hasher
	rd := io.TeeReader(resp.Body, hasher)
	_, err = io.Copy(tmpFile, rd)
	if err != nil {
		cleanupFn()
		return nil, nil, err
	}

	// Ensure the downloaded file has executable permissions
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		cleanupFn()
		return nil, nil, err
	}

	return &DownloadInfo{
		Checksum:                 hex.EncodeToString(hasher.Sum(nil)),
		DownloadedBinaryFilePath: tmpFile.Name(),
	}, cleanupFn, nil
}
