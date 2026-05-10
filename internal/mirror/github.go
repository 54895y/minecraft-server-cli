package mirror

import (
	"net/url"
	"strings"
)

var GitHubProxyBases = map[string]string{
	"none":     "",
	"gh-proxy": "https://gh-proxy.com/",
	"akams":    "https://github.akams.cn/",
	"gitproxy": "https://gitproxy.dev/",
	"gitwarp":  "https://www.gitwarp.com/",
}

func RewriteGitHubURL(rawURL, proxyName, customBase string) string {
	proxyName = strings.ToLower(strings.TrimSpace(proxyName))
	if proxyName == "" || proxyName == "none" {
		return rawURL
	}
	if !isGitHubURL(rawURL) {
		return rawURL
	}
	base := customBase
	if base == "" {
		base = GitHubProxyBases[proxyName]
	}
	if base == "" {
		return rawURL
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return base + rawURL
}

func isGitHubURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "github.com" || host == "raw.githubusercontent.com" || host == "objects.githubusercontent.com" || strings.HasSuffix(host, ".githubusercontent.com")
}
