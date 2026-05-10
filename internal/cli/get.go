package cli

import (
	"fmt"
	"net/url"
	"path"

	"github.com/54895y/minecraft-server-cli/internal/download"
	"github.com/54895y/minecraft-server-cli/internal/mirror"
	"github.com/spf13/cobra"
)

func (app *appContext) getCommand() *cobra.Command {
	var output string
	var threads int
	var sha1sum string
	var sha256sum string
	var sha512sum string
	var githubProxy string
	var githubProxyURL string

	cmd := &cobra.Command{
		Use:     "get <url>",
		Short:   "Download any URL with the built-in downloader",
		Example: "mcserver get https://example.com/server.jar -o server.jar --threads 16",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rawURL := mirror.RewriteGitHubURL(args[0], githubProxy, githubProxyURL)
			if output == "" {
				output = outputPath(app.config.GetString("paths.output_dir"), fileNameFromURL(args[0]))
			}
			if threads <= 0 {
				threads = app.config.GetInt("download.threads")
			}
			checksum := checksumFromFlags(sha1sum, sha256sum, sha512sum)
			result, err := app.downloader().Download(commandContext(cmd), download.Options{
				URL:      rawURL,
				Output:   output,
				Threads:  threads,
				Checksum: checksum,
				Progress: cmd.ErrOrStderr(),
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "saved %s (%d bytes)\n", result.Output, result.Bytes)
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file or directory")
	cmd.Flags().IntVar(&threads, "threads", 0, "download threads; defaults to config download.threads")
	cmd.Flags().StringVar(&sha1sum, "sha1", "", "expected SHA1 checksum")
	cmd.Flags().StringVar(&sha256sum, "sha256", "", "expected SHA256 checksum")
	cmd.Flags().StringVar(&sha512sum, "sha512", "", "expected SHA512 checksum")
	cmd.Flags().StringVar(&githubProxy, "github-proxy", "", "GitHub proxy name: none, gh-proxy, akams, gitproxy, gitwarp, custom")
	cmd.Flags().StringVar(&githubProxyURL, "github-proxy-url", "", "custom GitHub proxy base URL")
	return cmd
}

func fileNameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "download.bin"
	}
	name := path.Base(u.Path)
	if name == "." || name == "/" || name == "" {
		return "download.bin"
	}
	return name
}

func checksumFromFlags(sha1sum, sha256sum, sha512sum string) download.Checksum {
	switch {
	case sha512sum != "":
		return download.Checksum{Algorithm: "sha512", Value: sha512sum}
	case sha256sum != "":
		return download.Checksum{Algorithm: "sha256", Value: sha256sum}
	case sha1sum != "":
		return download.Checksum{Algorithm: "sha1", Value: sha1sum}
	default:
		return download.Checksum{}
	}
}
