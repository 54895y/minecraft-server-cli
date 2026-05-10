package core

import (
	"fmt"
	"strings"

	"github.com/54895y/minecraft-server-cli/internal/httpx"
)

func NewProvider(source string, client *httpx.Client) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "", "official":
		return NewCompositeProvider(client), nil
	case "fastmirror", "msl":
		return NewMirrorProvider(source, client), nil
	default:
		return nil, fmt.Errorf("unknown core source %q; use official, fastmirror, or msl", source)
	}
}
