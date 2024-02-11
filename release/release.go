package release

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ReleaseInfo holds information about a release.
type ReleaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

type ReleaseGetter interface {
	GetLatestRelease(ctx context.Context) (*ReleaseInfo, error)
}

type githubReleaseGetter struct {
	repo, owner string
}

var _ ReleaseGetter = (*githubReleaseGetter)(nil)

func NewReleaseGetter(repo, owner string) *githubReleaseGetter {
	return &githubReleaseGetter{
		repo:  repo,
		owner: owner,
	}
}

func (g *githubReleaseGetter) GetLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", g.repo, g.owner)
	return getLatestRelease(ctx, url)
}

// getLatestRelease fetches the latest release from GitHub.
func getLatestRelease(ctx context.Context, url string) (*ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}
