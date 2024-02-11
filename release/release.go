package release

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Info holds information about a release.
type Info struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Getter interface {
	GetLatestRelease(ctx context.Context) (*Info, error)
}

type githubReleaseGetter struct {
	repo, owner string
}

var _ Getter = (*githubReleaseGetter)(nil)

func NewReleaseGetter(repo, owner string) *githubReleaseGetter {
	return &githubReleaseGetter{
		repo:  repo,
		owner: owner,
	}
}

func (g *githubReleaseGetter) GetLatestRelease(ctx context.Context) (*Info, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", g.owner, g.repo)
	return getLatestRelease(ctx, url)
}

// getLatestRelease fetches the latest release from GitHub.
func getLatestRelease(ctx context.Context, url string) (*Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release Info
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}
