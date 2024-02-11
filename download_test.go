package upgrade

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getsavvyinc/upgrade-cli/release"
	"github.com/stretchr/testify/assert"
)

// downloadData is the content of the file that is downloaded in the tests.
// It's sha256 hash is: 88fd602a930bc7c0bb78c385f3cb70e976a0cdc3517020be32f19aae8c8eba17
// NOTE: There is no newline at the end of the file.
const downloadData = `#!/bin/sh

echo "Hello, World!"`

const downloadDataChecksum = "88fd602a930bc7c0bb78c385f3cb70e976a0cdc3517020be32f19aae8c8eba17"

func setupTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	srv := httptest.NewServer(handler)
	defer t.Cleanup(srv.Close)
	return srv
}

func downloadDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(200)
	io.WriteString(w, downloadData)
}

func shouldNeverBeCalled(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected URL: %s", r.URL.Path)
	})
}

func TestAssetDownloader(t *testing.T) {
	const executablePath = "savvy"
	t.Run("TryDownloadingMissingAsset", func(t *testing.T) {
		srv := setupTestServer(t, shouldNeverBeCalled(t))
		ctx := context.Background()
		downloader := NewAssetDownloader(executablePath)
		asset, cleanupFn, err := downloader.DownloadAsset(ctx, []release.Asset{
			{BrowserDownloadURL: srv.URL + "/nonexistent"},
		})
		assert.ErrorIs(t, err, ErrNoAsset)
		assert.Nil(t, asset)
		assert.Nil(t, cleanupFn)
	})
	t.Run("EnsureDownloadedDoesntChangeContent", func(t *testing.T) {
		srv := setupTestServer(t, http.HandlerFunc(downloadDataHandler))
		ctx := context.Background()
		downloader := NewAssetDownloader(executablePath, WithOS("os"), WithArch("arch"))
		asset, cleanupFn, err := downloader.DownloadAsset(ctx, []release.Asset{
			{BrowserDownloadURL: srv.URL + "/download_os_arch"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, asset)
		assert.NotNil(t, cleanupFn)

		assert.Equal(t, downloadDataChecksum, asset.Checksum)
		t.Run("VerifyCleanup", func(t *testing.T) {
			tmpFile := asset.DownloadedBinaryFilePath
			assert.FileExists(t, tmpFile)
			assert.NoError(t, cleanupFn())
			assert.NoFileExists(t, tmpFile)
		})
	})
}
