package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/54895y/minecraft-server-cli/internal/download"
	"github.com/54895y/minecraft-server-cli/internal/modrinth"
	"github.com/spf13/cobra"
)

func (app *appContext) modrinthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modrinth",
		Short: "Search and download Modrinth projects",
	}
	cmd.AddCommand(app.modrinthSearchCommand(), app.modrinthDownloadCommand())
	return cmd
}

func (app *appContext) modrinthSearchCommand() *cobra.Command {
	var loader string
	var mcVersion string
	var projectType string
	var limit int
	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Search Modrinth projects",
		Example: "mcserver modrinth search geyser --loader paper --mc 1.21.10 --type mod",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := modrinth.New(app.httpClient())
			results, err := client.Search(commandContext(cmd), modrinth.SearchOptions{
				Query:       args[0],
				Loader:      normalizeChoice(loader),
				GameVersion: normalizeChoice(mcVersion),
				ProjectType: normalizeProjectType(projectType),
				Limit:       limit,
			})
			if err != nil {
				return err
			}
			for _, result := range results {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%d\n", result.Slug, result.ProjectType, result.Title, result.Downloads)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&loader, "loader", "paper", "loader/category filter, or any")
	cmd.Flags().StringVar(&mcVersion, "mc", "", "Minecraft version filter")
	cmd.Flags().StringVar(&projectType, "type", "mod", "project type: mod, modpack, resourcepack, shader, datapack, or any")
	cmd.Flags().IntVar(&limit, "limit", 10, "maximum results")
	return cmd
}

func (app *appContext) modrinthDownloadCommand() *cobra.Command {
	var loader string
	var mcVersion string
	var projectType string
	var output string
	var threads int
	var resolveSearch bool
	cmd := &cobra.Command{
		Use:     "download <slug-or-id-or-query>",
		Short:   "Download a matching Modrinth project file",
		Example: "mcserver modrinth download geyser --loader paper --mc 1.21.10 -o plugins",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if threads <= 0 {
				threads = app.config.GetInt("download.threads")
			}
			client := modrinth.New(app.httpClient())
			project := args[0]
			if resolveSearch {
				resolved, err := modrinth.ResolveProject(commandContext(cmd), client, project, normalizeChoice(loader), normalizeChoice(mcVersion), normalizeProjectType(projectType))
				if err != nil {
					return err
				}
				project = resolved
			}
			versions, err := client.Versions(commandContext(cmd), modrinth.VersionsOptions{
				Project:      project,
				Loaders:      optionalList(normalizeChoice(loader)),
				GameVersions: optionalList(normalizeChoice(mcVersion)),
			})
			if err != nil {
				return err
			}
			version, file, err := modrinth.SelectFile(versions)
			if err != nil {
				return err
			}
			if output == "" {
				output = outputPath(app.config.GetString("paths.output_dir"), file.Filename)
			}
			if filepath.Ext(output) == "" {
				output = filepath.Join(output, file.Filename)
			}
			checksum := download.Checksum{}
			if file.Hashes["sha512"] != "" {
				checksum = download.Checksum{Algorithm: "sha512", Value: file.Hashes["sha512"]}
			} else if file.Hashes["sha1"] != "" {
				checksum = download.Checksum{Algorithm: "sha1", Value: file.Hashes["sha1"]}
			}
			result, err := app.downloader().Download(commandContext(cmd), download.Options{
				URL:      file.URL,
				Output:   output,
				Threads:  threads,
				Checksum: checksum,
				Progress: cmd.ErrOrStderr(),
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "saved %s (%s %s, %d bytes)\n", result.Output, project, version.VersionNumber, result.Bytes)
			return nil
		},
	}
	cmd.Flags().StringVar(&loader, "loader", "paper", "loader filter, or any")
	cmd.Flags().StringVar(&mcVersion, "mc", "", "Minecraft version filter")
	cmd.Flags().StringVar(&projectType, "type", "mod", "project type used when --search is enabled")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file or directory")
	cmd.Flags().IntVar(&threads, "threads", 0, "download threads; defaults to config download.threads")
	cmd.Flags().BoolVar(&resolveSearch, "search", false, "treat argument as a search query and auto-select a project")
	return cmd
}

func normalizeProjectType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "any":
		return "any"
	case "plugin":
		return "mod"
	default:
		return value
	}
}

func optionalList(value string) []string {
	if value == "" || value == "any" {
		return nil
	}
	return []string{value}
}
