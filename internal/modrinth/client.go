package modrinth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/54895y/minecraft-server-cli/internal/httpx"
)

const APIBase = "https://api.modrinth.com/v2"

type Client struct {
	http *httpx.Client
}

func New(httpClient *httpx.Client) *Client {
	return &Client{http: httpClient}
}

type SearchOptions struct {
	Query       string
	Loader      string
	GameVersion string
	ProjectType string
	Limit       int
}

func (c *Client) Search(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = 10
	}
	u, _ := url.Parse(APIBase + "/search")
	q := u.Query()
	q.Set("query", opts.Query)
	q.Set("limit", fmt.Sprintf("%d", opts.Limit))
	q.Set("index", "relevance")
	facets := buildFacets(opts.ProjectType, opts.Loader, opts.GameVersion)
	if len(facets) > 0 {
		b, _ := json.Marshal(facets)
		q.Set("facets", string(b))
	}
	u.RawQuery = q.Encode()

	var res struct {
		Hits []SearchResult `json:"hits"`
	}
	if err := c.http.JSON(ctx, "GET", u.String(), &res); err != nil {
		return nil, err
	}
	return res.Hits, nil
}

type VersionsOptions struct {
	Project      string
	Loaders      []string
	GameVersions []string
}

func (c *Client) Versions(ctx context.Context, opts VersionsOptions) ([]Version, error) {
	if strings.TrimSpace(opts.Project) == "" {
		return nil, fmt.Errorf("project slug or id is required")
	}
	u, _ := url.Parse(fmt.Sprintf("%s/project/%s/version", APIBase, url.PathEscape(opts.Project)))
	q := u.Query()
	if len(opts.Loaders) > 0 {
		b, _ := json.Marshal(opts.Loaders)
		q.Set("loaders", string(b))
	}
	if len(opts.GameVersions) > 0 {
		b, _ := json.Marshal(opts.GameVersions)
		q.Set("game_versions", string(b))
	}
	u.RawQuery = q.Encode()

	var versions []Version
	if err := c.http.JSON(ctx, "GET", u.String(), &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

func SelectFile(versions []Version) (Version, File, error) {
	for _, version := range versions {
		if len(version.Files) == 0 {
			continue
		}
		for _, file := range version.Files {
			if file.Primary && strings.HasSuffix(strings.ToLower(file.Filename), ".jar") {
				return version, file, nil
			}
		}
		for _, file := range version.Files {
			if strings.HasSuffix(strings.ToLower(file.Filename), ".jar") {
				return version, file, nil
			}
		}
		for _, file := range version.Files {
			if file.Primary {
				return version, file, nil
			}
		}
		return version, version.Files[0], nil
	}
	return Version{}, File{}, fmt.Errorf("no downloadable file found")
}

func ResolveProject(ctx context.Context, c *Client, query, loader, gameVersion, projectType string) (string, error) {
	results, err := c.Search(ctx, SearchOptions{
		Query:       query,
		Loader:      loader,
		GameVersion: gameVersion,
		ProjectType: projectType,
		Limit:       10,
	})
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no Modrinth project found for %q", query)
	}
	q := strings.ToLower(query)
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Downloads > results[j].Downloads
	})
	for _, result := range results {
		if strings.ToLower(result.Slug) == q || strings.ToLower(result.Title) == q {
			return result.Slug, nil
		}
	}
	return results[0].Slug, nil
}

func buildFacets(projectType, loader, gameVersion string) [][]string {
	var facets [][]string
	if projectType != "" && projectType != "any" {
		facets = append(facets, []string{fmt.Sprintf("project_type:%s", projectType)})
	}
	if loader != "" && loader != "any" {
		facets = append(facets, []string{fmt.Sprintf("categories:%s", loader)})
	}
	if gameVersion != "" && gameVersion != "any" {
		facets = append(facets, []string{fmt.Sprintf("versions:%s", gameVersion)})
	}
	return facets
}
