package upgrade

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// checksumData is a sample checksum file for testing
// It contains the checksums, one or more spaces and binary_os_arch pairs
// NOTE: we intentionally have an extra space at the beg of each line
const checksumData = ` checksum_savvy_darwin_arm64  savvy_darwin_arm64
 checksum_savvy_darwin_x86_64 savvy_darwin_x86_64
 checksum_savvy_linux_arm64 savvy_linux_arm64
 checksum_savvy_linux_i386 savvy_linux_i386
 checksum_savvy_linux_x86_64  savvy_linux_x86_64
`

const malformedChecksumData = `6796a0fb64d0c78b2de5410a94749a 3bfb77291747c1835fbd427e8bf00f6af3  savvy_darwin_arm64
`

func setupTestServer(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(200)
		if r.URL.Path == "/checksums.txt" {
			io.WriteString(w, checksumData)
			return
		}
		if r.URL.Path == "/malformed_checksums.txt" {
			t.Log("sending malformed checksums")
			io.WriteString(w, malformedChecksumData)
			return
		}
		// DownloadCheckSum should only be called with routes that end with checksums.txt
		t.Errorf("unexpected URL: %s", r.URL.Path)
	}))
	defer t.Cleanup(srv.Close)
	return srv
}

func TestDownloadCheckSum(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()
	testSuffix := "checksums.txt"
	t.Run("TestDownloadCheckSum_ValidCheckSumFile", func(t *testing.T) {
		checksumURL := srv.URL + "/checksums.txt"
		downloader := NewCheckSumDownloader(WithAssetSuffix(testSuffix))
		checksums, err := downloader.DownloadCheckSum(ctx, []ReleaseAsset{
			{BrowserDownloadURL: checksumURL},
			{BrowserDownloadURL: srv.URL + "/malformed_path.txt"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, checksums)
		assert.NotEmpty(t, checksums.checksums)
		for k, v := range checksums.checksums {
			assert.NotEmpty(t, k)
			assert.NotEmpty(t, v)
			assert.Equal(t, strings.Join([]string{"checksum", k}, "_"), v)
		}
	})
	t.Run("TestDownloadCheckSum_InvalidCheckSumFile", func(t *testing.T) {
		malformedChecksumURL := srv.URL + "/malformed_checksums.txt"
		downloader := NewCheckSumDownloader(WithAssetSuffix(testSuffix))
		checksums, err := downloader.DownloadCheckSum(ctx, []ReleaseAsset{
			{BrowserDownloadURL: malformedChecksumURL},
		})
		assert.Error(t, err)
		assert.Nil(t, checksums)
		assert.ErrorIs(t, err, ErrInvalidChecksumFile)
	})
	t.Run("TestDownloadCheckSum_NoCheckSumAsset", func(t *testing.T) {
		downloader := NewCheckSumDownloader(WithAssetSuffix(testSuffix))
		checksums, err := downloader.DownloadCheckSum(ctx, []ReleaseAsset{
			{BrowserDownloadURL: srv.URL + "/savvy_darwin_arm64"},
		})
		assert.Error(t, err)
		assert.Nil(t, checksums)
		assert.ErrorIs(t, err, ErrNoCheckSumAsset)
	})
}
