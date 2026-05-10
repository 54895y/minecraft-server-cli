package core

import "context"

type Provider interface {
	ListVersions(ctx context.Context, coreType string) ([]string, error)
	ListBuilds(ctx context.Context, coreType, mcVersion string) ([]Build, error)
	ResolveDownload(ctx context.Context, req Request) (DownloadCandidate, error)
}

type Request struct {
	CoreType  string
	MCVersion string
	Build     string
	Channel   string
	Source    string
}

type Build struct {
	ID      string
	Channel string
}

type DownloadCandidate struct {
	URL       string
	FileName  string
	Size      int64
	Checksums map[string]string
	Build     Build
	MCVersion string
}
