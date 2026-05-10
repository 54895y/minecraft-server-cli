package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/54895y/minecraft-server-cli/internal/core"
	"github.com/54895y/minecraft-server-cli/internal/download"
	"github.com/spf13/cobra"
)

func (app *appContext) coreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "core",
		Short: "Download Minecraft server cores",
	}
	cmd.AddCommand(app.coreListCommand(), app.coreBuildsCommand(), app.coreDownloadCommand())
	return cmd
}

func (app *appContext) coreListCommand() *cobra.Command {
	var coreType string
	var source string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Minecraft versions for a core",
		RunE: func(cmd *cobra.Command, args []string) error {
			if source == "" {
				source = app.config.GetString("core.source")
			}
			provider, err := core.NewProvider(source, app.httpClient())
			if err != nil {
				return err
			}
			versions, err := provider.ListVersions(commandContext(cmd), coreType)
			if err != nil {
				return err
			}
			for _, version := range versions {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), version)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&coreType, "type", "paper", "core type: paper, folia, velocity, purpur")
	cmd.Flags().StringVar(&source, "source", "", "source: official, fastmirror, msl")
	return cmd
}

func (app *appContext) coreBuildsCommand() *cobra.Command {
	var coreType string
	var source string
	var mcVersion string
	cmd := &cobra.Command{
		Use:   "builds",
		Short: "List builds for a core version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mcVersion == "" {
				return fmt.Errorf("--mc is required")
			}
			if source == "" {
				source = app.config.GetString("core.source")
			}
			provider, err := core.NewProvider(source, app.httpClient())
			if err != nil {
				return err
			}
			builds, err := provider.ListBuilds(commandContext(cmd), coreType, mcVersion)
			if err != nil {
				return err
			}
			for _, build := range builds {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", build.ID, build.Channel)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&coreType, "type", "paper", "core type: paper, folia, velocity, purpur")
	cmd.Flags().StringVar(&source, "source", "", "source: official, fastmirror, msl")
	cmd.Flags().StringVar(&mcVersion, "mc", "", "Minecraft version")
	return cmd
}

func (app *appContext) coreDownloadCommand() *cobra.Command {
	var coreType string
	var source string
	var mcVersion string
	var build string
	var channel string
	var output string
	var threads int

	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Resolve and download a Minecraft server core",
		Example: "mcserver core download --type paper --mc 1.21.10 --build latest -o server.jar",
		RunE: func(cmd *cobra.Command, args []string) error {
			if source == "" {
				source = app.config.GetString("core.source")
			}
			if build == "" {
				build = "latest"
			}
			if channel == "" {
				channel = "STABLE"
			}
			if threads <= 0 {
				threads = app.config.GetInt("download.threads")
			}

			provider, err := core.NewProvider(source, app.httpClient())
			if err != nil {
				return err
			}
			candidate, err := provider.ResolveDownload(commandContext(cmd), core.Request{
				CoreType:  coreType,
				MCVersion: mcVersion,
				Build:     build,
				Channel:   strings.ToUpper(channel),
				Source:    source,
			})
			if err != nil {
				return err
			}
			if output == "" {
				output = outputPath(app.config.GetString("paths.output_dir"), candidate.FileName)
			}
			if filepath.Ext(output) == "" {
				output = filepath.Join(output, candidate.FileName)
			}
			checksum := download.Checksum{}
			if candidate.Checksums["sha256"] != "" {
				checksum = download.Checksum{Algorithm: "sha256", Value: candidate.Checksums["sha256"]}
			}
			result, err := app.downloader().Download(commandContext(cmd), download.Options{
				URL:      candidate.URL,
				Output:   output,
				Threads:  threads,
				Checksum: checksum,
				Progress: cmd.ErrOrStderr(),
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "saved %s (%s %s build %s, %d bytes)\n", result.Output, coreType, candidate.MCVersion, candidate.Build.ID, result.Bytes)
			return nil
		},
	}
	cmd.Flags().StringVar(&coreType, "type", "paper", "core type: paper, folia, velocity, purpur")
	cmd.Flags().StringVar(&source, "source", "", "source: official, fastmirror, msl")
	cmd.Flags().StringVar(&mcVersion, "mc", "latest", "Minecraft version, or latest")
	cmd.Flags().StringVar(&build, "build", "latest", "build id, or latest")
	cmd.Flags().StringVar(&channel, "channel", "STABLE", "PaperMC channel: STABLE, EXPERIMENTAL, ANY")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file or directory")
	cmd.Flags().IntVar(&threads, "threads", 0, "download threads; defaults to config download.threads")
	return cmd
}
