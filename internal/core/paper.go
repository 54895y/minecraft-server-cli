package core

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/54895y/minecraft-server-cli/internal/httpx"
)

const paperAPI = "https://fill.papermc.io/v3"

var paperProjects = map[string]string{
	"paper":    "paper",
	"folia":    "folia",
	"velocity": "velocity",
}

type CompositeProvider struct {
	client *httpx.Client
	purpur *PurpurProvider
}

func NewCompositeProvider(client *httpx.Client) *CompositeProvider {
	return &CompositeProvider{client: client, purpur: NewPurpurProvider(client)}
}

func (p *CompositeProvider) ListVersions(ctx context.Context, coreType string) ([]string, error) {
	if normalizeCore(coreType) == "purpur" {
		return p.purpur.ListVersions(ctx, coreType)
	}
	project, err := paperProject(coreType)
	if err != nil {
		return nil, err
	}
	var res paperProjectResponse
	if err := p.client.JSON(ctx, "GET", fmt.Sprintf("%s/projects/%s", paperAPI, project), &res); err != nil {
		return nil, err
	}
	versions := flattenVersions(res.Versions)
	sort.Sort(sort.Reverse(versionStrings(versions)))
	return versions, nil
}

func (p *CompositeProvider) ListBuilds(ctx context.Context, coreType, mcVersion string) ([]Build, error) {
	if normalizeCore(coreType) == "purpur" {
		return p.purpur.ListBuilds(ctx, coreType, mcVersion)
	}
	project, err := paperProject(coreType)
	if err != nil {
		return nil, err
	}
	var res []paperBuild
	if err := p.client.JSON(ctx, "GET", fmt.Sprintf("%s/projects/%s/versions/%s/builds", paperAPI, project, mcVersion), &res); err != nil {
		return nil, err
	}
	builds := make([]Build, 0, len(res))
	for _, b := range res {
		builds = append(builds, Build{ID: strconv.Itoa(b.ID), Channel: b.Channel})
	}
	sort.SliceStable(builds, func(i, j int) bool {
		return atoi(builds[i].ID) > atoi(builds[j].ID)
	})
	return builds, nil
}

func (p *CompositeProvider) ResolveDownload(ctx context.Context, req Request) (DownloadCandidate, error) {
	if normalizeCore(req.CoreType) == "purpur" {
		return p.purpur.ResolveDownload(ctx, req)
	}
	project, err := paperProject(req.CoreType)
	if err != nil {
		return DownloadCandidate{}, err
	}
	if strings.TrimSpace(req.MCVersion) == "" || req.MCVersion == "latest" {
		versions, err := p.ListVersions(ctx, req.CoreType)
		if err != nil {
			return DownloadCandidate{}, err
		}
		if len(versions) == 0 {
			return DownloadCandidate{}, fmt.Errorf("no versions available for %s", req.CoreType)
		}
		req.MCVersion = selectLatestMinecraftVersion(versions)
	}

	var res []paperBuild
	if err := p.client.JSON(ctx, "GET", fmt.Sprintf("%s/projects/%s/versions/%s/builds", paperAPI, project, req.MCVersion), &res); err != nil {
		return DownloadCandidate{}, err
	}
	build, err := selectPaperBuild(res, req.Build, req.Channel)
	if err != nil {
		return DownloadCandidate{}, err
	}
	download, ok := build.Downloads["server:default"]
	if !ok || download.URL == "" {
		return DownloadCandidate{}, fmt.Errorf("build %d has no server:default download", build.ID)
	}
	return DownloadCandidate{
		URL:       download.URL,
		FileName:  download.Name,
		Size:      download.Size,
		Checksums: download.Checksums,
		Build:     Build{ID: strconv.Itoa(build.ID), Channel: build.Channel},
		MCVersion: req.MCVersion,
	}, nil
}

type paperProjectResponse struct {
	Versions map[string][]string `json:"versions"`
}

type paperBuild struct {
	ID        int                      `json:"id"`
	Channel   string                   `json:"channel"`
	Downloads map[string]paperDownload `json:"downloads"`
}

type paperDownload struct {
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Size      int64             `json:"size"`
	Checksums map[string]string `json:"checksums"`
}

func paperProject(coreType string) (string, error) {
	project := paperProjects[normalizeCore(coreType)]
	if project == "" {
		return "", fmt.Errorf("unsupported official core %q; supported: paper, folia, velocity, purpur", coreType)
	}
	return project, nil
}

func normalizeCore(coreType string) string {
	return strings.ToLower(strings.TrimSpace(coreType))
}

func flattenVersions(groups map[string][]string) []string {
	seen := map[string]bool{}
	var versions []string
	for _, list := range groups {
		for _, version := range list {
			if !seen[version] {
				seen[version] = true
				versions = append(versions, version)
			}
		}
	}
	return versions
}

func selectLatestMinecraftVersion(versions []string) string {
	for _, version := range versions {
		if strings.HasPrefix(version, "1.") {
			return version
		}
	}
	if len(versions) == 0 {
		return ""
	}
	return versions[0]
}

func selectPaperBuild(builds []paperBuild, buildID, channel string) (paperBuild, error) {
	buildID = strings.TrimSpace(buildID)
	channel = strings.ToUpper(strings.TrimSpace(channel))
	if channel == "" {
		channel = "STABLE"
	}
	sort.SliceStable(builds, func(i, j int) bool {
		return builds[i].ID > builds[j].ID
	})
	for _, build := range builds {
		if buildID != "" && buildID != "latest" && strconv.Itoa(build.ID) != buildID {
			continue
		}
		if channel != "ANY" && strings.ToUpper(build.Channel) != channel {
			continue
		}
		return build, nil
	}
	return paperBuild{}, fmt.Errorf("no matching build found")
}

type versionStrings []string

func (v versionStrings) Len() int      { return len(v) }
func (v versionStrings) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v versionStrings) Less(i, j int) bool {
	return compareVersion(v[i], v[j]) < 0
}

func compareVersion(a, b string) int {
	as := splitVersion(a)
	bs := splitVersion(b)
	for i := 0; i < len(as) || i < len(bs); i++ {
		av, bv := 0, 0
		if i < len(as) {
			av = as[i]
		}
		if i < len(bs) {
			bv = bs[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return strings.Compare(a, b)
}

func splitVersion(v string) []int {
	parts := strings.FieldsFunc(v, func(r rune) bool {
		return r == '.' || r == '-' || r == '_'
	})
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		n, _ := strconv.Atoi(strings.TrimLeftFunc(part, func(r rune) bool { return r < '0' || r > '9' }))
		out = append(out, n)
	}
	return out
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
