package core

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/54895y/minecraft-server-cli/internal/httpx"
)

const purpurAPI = "https://api.purpurmc.org/v2"

type PurpurProvider struct {
	client *httpx.Client
}

func NewPurpurProvider(client *httpx.Client) *PurpurProvider {
	return &PurpurProvider{client: client}
}

func (p *PurpurProvider) ListVersions(ctx context.Context, coreType string) ([]string, error) {
	var res purpurProjectResponse
	if err := p.client.JSON(ctx, "GET", purpurAPI+"/purpur", &res); err != nil {
		return nil, err
	}
	versions := append([]string(nil), res.Versions...)
	sort.Sort(sort.Reverse(versionStrings(versions)))
	return versions, nil
}

func (p *PurpurProvider) ListBuilds(ctx context.Context, coreType, mcVersion string) ([]Build, error) {
	var res purpurBuildsResponse
	if err := p.client.JSON(ctx, "GET", fmt.Sprintf("%s/purpur/%s", purpurAPI, mcVersion), &res); err != nil {
		return nil, err
	}
	builds := make([]Build, 0, len(res.Builds.All))
	for _, id := range res.Builds.All {
		builds = append(builds, Build{ID: id, Channel: "STABLE"})
	}
	sort.SliceStable(builds, func(i, j int) bool {
		return atoi(builds[i].ID) > atoi(builds[j].ID)
	})
	return builds, nil
}

func (p *PurpurProvider) ResolveDownload(ctx context.Context, req Request) (DownloadCandidate, error) {
	if strings.TrimSpace(req.MCVersion) == "" || req.MCVersion == "latest" {
		versions, err := p.ListVersions(ctx, "purpur")
		if err != nil {
			return DownloadCandidate{}, err
		}
		if len(versions) == 0 {
			return DownloadCandidate{}, fmt.Errorf("no Purpur versions available")
		}
		req.MCVersion = versions[0]
	}
	build := strings.TrimSpace(req.Build)
	if build == "" || build == "latest" {
		build = "latest"
	}
	url := fmt.Sprintf("%s/purpur/%s/%s/download", purpurAPI, req.MCVersion, build)
	if build == "latest" {
		url = fmt.Sprintf("%s/purpur/%s/latest/download", purpurAPI, req.MCVersion)
	}
	return DownloadCandidate{
		URL:       url,
		FileName:  fmt.Sprintf("purpur-%s-%s.jar", req.MCVersion, build),
		Build:     Build{ID: build, Channel: "STABLE"},
		MCVersion: req.MCVersion,
	}, nil
}

type purpurProjectResponse struct {
	Versions []string `json:"versions"`
}

type purpurBuildsResponse struct {
	Builds struct {
		All []string `json:"all"`
	} `json:"builds"`
}
