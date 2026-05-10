package core

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/54895y/minecraft-server-cli/internal/httpx"
)

type MirrorProvider struct {
	source   string
	official *CompositeProvider
}

func NewMirrorProvider(source string, client *httpx.Client) *MirrorProvider {
	return &MirrorProvider{source: strings.ToLower(source), official: NewCompositeProvider(client)}
}

func (m *MirrorProvider) ListVersions(ctx context.Context, coreType string) ([]string, error) {
	return m.official.ListVersions(ctx, coreType)
}

func (m *MirrorProvider) ListBuilds(ctx context.Context, coreType, mcVersion string) ([]Build, error) {
	return m.official.ListBuilds(ctx, coreType, mcVersion)
}

func (m *MirrorProvider) ResolveDownload(ctx context.Context, req Request) (DownloadCandidate, error) {
	candidate, err := m.official.ResolveDownload(ctx, req)
	if err != nil {
		return DownloadCandidate{}, err
	}
	switch m.source {
	case "fastmirror":
		candidate.URL = fastMirrorURL(req.CoreType, candidate)
	case "msl":
		candidate.URL = mslMirrorURL(req.CoreType, candidate)
	default:
		return DownloadCandidate{}, fmt.Errorf("unknown mirror source %q", m.source)
	}
	return candidate, nil
}

func fastMirrorURL(coreType string, candidate DownloadCandidate) string {
	// FastMirror exposes a browser-oriented download service and can mirror direct URLs.
	// Keeping the original URL as a query avoids relying on page HTML selectors.
	v := url.Values{}
	v.Set("url", candidate.URL)
	v.Set("core", normalizeCore(coreType))
	v.Set("version", candidate.MCVersion)
	v.Set("build", candidate.Build.ID)
	return "https://www.fastmirror.net/api/download?" + v.Encode()
}

func mslMirrorURL(coreType string, candidate DownloadCandidate) string {
	v := url.Values{}
	v.Set("url", candidate.URL)
	v.Set("core", normalizeCore(coreType))
	v.Set("version", candidate.MCVersion)
	v.Set("build", candidate.Build.ID)
	return "https://dl.mslmc.cn/api/download?" + v.Encode()
}
