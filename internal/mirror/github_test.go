package mirror

import "testing"

func TestRewriteGitHubURL(t *testing.T) {
	got := RewriteGitHubURL("https://github.com/owner/repo/releases/download/v1/app.zip", "gh-proxy", "")
	want := "https://gh-proxy.com/https://github.com/owner/repo/releases/download/v1/app.zip"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRewriteGitHubURLSkipsNonGitHub(t *testing.T) {
	raw := "https://example.com/file.zip"
	if got := RewriteGitHubURL(raw, "gh-proxy", ""); got != raw {
		t.Fatalf("got %q want %q", got, raw)
	}
}
