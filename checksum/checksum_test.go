package checksum

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getsavvyinc/upgrade-cli/release"
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

func setupTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	srv := httptest.NewServer(handler)
	defer t.Cleanup(srv.Close)
	return srv
}

func checkSumDataHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(200)
		if r.URL.Path == "/checksums.txt" {
			io.WriteString(w, checksumData)
			return
		}
		if r.URL.Path == "/empty_checksums.txt" {
			io.WriteString(w, "")
			return
		}
		if r.URL.Path == "/malformed_checksums.txt" {
			io.WriteString(w, malformedChecksumData)
			return
		}
		// DownloadCheckSum should only be called with routes that end with checksums.txt
		t.Errorf("unexpected URL: %s", r.URL.Path)
	})
}

func TestDownloadCheckSum(t *testing.T) {
	srv := setupTestServer(t, checkSumDataHandler(t))
	ctx := context.Background()
	testSuffix := "checksums.txt"
	t.Run("ValidCheckSumFile", func(t *testing.T) {
		checksumURL := srv.URL + "/checksums.txt"
		downloader := NewCheckSumDownloader(WithAssetSuffix(testSuffix))
		checksums, err := downloader.Download(ctx, []release.Asset{
			{BrowserDownloadURL: checksumURL},
			{BrowserDownloadURL: srv.URL + "/malformed_path.txt"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, checksums)
		assert.NotEmpty(t, checksums.Checksums)
		for k, v := range checksums.Checksums {
			assert.NotEmpty(t, k)
			assert.NotEmpty(t, v)
			assert.Equal(t, strings.Join([]string{"checksum", k}, "_"), v)
		}
	})
	t.Run("InvalidCheckSumFile", func(t *testing.T) {
		testCases := []struct {
			name string
			url  string
		}{
			{
				name: "MalformedChecksumFile",
				url:  srv.URL + "/malformed_checksums.txt",
			},
			{
				name: "EmptyCheckSumFile",
				url:  srv.URL + "/empty_checksums.txt",
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				downloader := NewCheckSumDownloader(WithAssetSuffix(testSuffix))
				checksums, err := downloader.Download(ctx, []release.Asset{
					{BrowserDownloadURL: tc.url},
				})
				assert.Error(t, err)
				assert.Nil(t, checksums)
				assert.ErrorIs(t, err, ErrInvalidChecksumFile)
			})
		}
	})
	t.Run("NoCheckSumAsset", func(t *testing.T) {
		downloader := NewCheckSumDownloader(WithAssetSuffix(testSuffix))
		checksums, err := downloader.Download(ctx, []release.Asset{
			{BrowserDownloadURL: srv.URL + "/savvy_darwin_arm64"},
		})
		assert.Error(t, err)
		assert.Nil(t, checksums)
		assert.ErrorIs(t, err, ErrNoCheckSumAsset)
	})
}

func TestCheckSumValidator(t *testing.T) {
	binary := "savvy"
	const checksum = "checksum"
	checksumInfo := &Info{
		Checksums: map[string]string{
			binary + "_darwin_x86_64": checksum,
			binary + "_linux_x86_64":  checksum,
		},
	}

	testCases := []struct {
		name               string
		downloadedChecksum string
		isValid            bool
		os                 string
		arch               string
		binary             string
	}{
		{
			name:               "ValidChecksums",
			downloadedChecksum: checksum,
			os:                 "linux",
			arch:               "x86_64",
			isValid:            true,
			binary:             binary,
		},
		{
			name:               "InvalidChecksums",
			downloadedChecksum: "invalid_checksum",
			os:                 "darwin",
			arch:               "x86_64",
			isValid:            false,
			binary:             binary,
		},
		{
			name:               "InvalidOS",
			downloadedChecksum: checksum,
			os:                 "windows",
			arch:               "x86_64",
			isValid:            false,
			binary:             binary,
		},
		{
			name:               "InvalidArch",
			downloadedChecksum: checksum,
			os:                 "linux",
			arch:               "not_suppported",
			isValid:            false,
			binary:             binary,
		},
		{
			name:               "InvalidBinary",
			downloadedChecksum: checksum,
			os:                 "linux",
			arch:               "x86_64",
			isValid:            false,
			binary:             "invalid_binary",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			csv := NewCheckSumValidator(WithArch(tc.arch), WithOS(tc.os))
			isValid := csv.IsCheckSumValid(context.Background(), tc.binary, checksumInfo, tc.downloadedChecksum)
			assert.Equal(t, tc.isValid, isValid)
		})
	}
}
