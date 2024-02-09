package upgrade

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
)

type Upgrader interface {
	Upgrade(ctx context.Context, currentVersion string) error
}

type upgrader struct {
	executablePath     string
	repo               string
	owner              string
	downloader         AssetDownloader
	checksumDownloader CheckSumDownloader
	checksumValidator  CheckSumValidator
}

var _ Upgrader = (*upgrader)(nil)

type Opt func(*upgrader)

func WithAssetDownloader(d AssetDownloader) Opt {
	return func(u *upgrader) {
		u.downloader = d
	}
}

func WithCheckSumDownloader(c CheckSumDownloader) Opt {
	return func(u *upgrader) {
		u.checksumDownloader = c
	}
}

func WithCheckSumValidator(c CheckSumValidator) Opt {
	return func(u *upgrader) {
		u.checksumValidator = c
	}
}

func NewUpgrader(owner string, repo string, executablePath string, opts ...Opt) Upgrader {
	u := &upgrader{
		repo:               repo,
		owner:              owner,
		executablePath:     executablePath,
		downloader:         NewAssetDownloader(executablePath),
		checksumDownloader: NewCheckSumDownloader(),
		checksumValidator:  NewCheckSumValidator(),
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

var ErrInvalidCheckSum = errors.New("invalid checksum")

func (u *upgrader) Upgrade(ctx context.Context, currentVersion string) error {
	curr, err := version.NewVersion(currentVersion)
	if err != nil {
		return err
	}

	releaseInfo, err := getLatestRelease(ctx, u.owner, u.repo)
	if err != nil {
		return err
	}

	latest, err := version.NewVersion(releaseInfo.TagName)
	if err != nil {
		return err
	}

	if latest.LessThanOrEqual(curr) {
		return nil
	}

	// from the releaseInfo, download the binary for the architecture

	downloadInfo, cleanup, err := u.downloader.DownloadAsset(ctx, releaseInfo.Assets)
	if err != nil {
		return err
	}

	if cleanup != nil {
		defer cleanup()
	}

	// download the checksum file
	checksumInfo, err := u.checksumDownloader.DownloadCheckSum(ctx, releaseInfo.Assets)
	if err != nil {
		return err
	}

	executableName := filepath.Base(u.executablePath)
	// verify the checksum
	if !u.checksumValidator.IsCheckSumValid(ctx, executableName, checksumInfo, downloadInfo.Checksum) {
		return ErrInvalidCheckSum
	}

	if err := replaceBinary(downloadInfo.DownloadedBinaryFilePath, u.executablePath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// replaceBinary replaces the current executable with the downloaded update.
func replaceBinary(tmpFilePath, currentBinaryPath string) error {
	// Replace the current binary with the new binary
	if err := os.Rename(tmpFilePath, currentBinaryPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}