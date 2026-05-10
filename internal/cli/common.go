package cli

import (
	"path/filepath"
	"time"

	"github.com/54895y/minecraft-server-cli/internal/download"
	"github.com/54895y/minecraft-server-cli/internal/httpx"
)

func (app *appContext) httpClient() *httpx.Client {
	timeout, err := time.ParseDuration(app.config.GetString("download.timeout"))
	if err != nil {
		timeout = 30 * time.Second
	}
	return httpx.New(app.config.GetString("download.user_agent"), timeout, app.config.GetInt("download.retries"))
}

func (app *appContext) downloader() *download.Downloader {
	return download.New(app.httpClient())
}

func outputPath(base, fallbackName string) string {
	if base == "" {
		base = "."
	}
	if fallbackName == "" {
		fallbackName = "download.bin"
	}
	if ext := filepath.Ext(base); ext != "" {
		return base
	}
	return filepath.Join(base, fallbackName)
}
