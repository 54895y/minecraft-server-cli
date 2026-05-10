package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/54895y/minecraft-server-cli/internal/config"
	"github.com/spf13/cobra"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

type appContext struct {
	build  BuildInfo
	config *config.Manager
}

func NewRootCommand(build BuildInfo) *cobra.Command {
	cfg := config.NewManager("mcserver")
	app := &appContext{build: build, config: cfg}

	cmd := &cobra.Command{
		Use:           "mcserver",
		Short:         "minecraft server cli downloads server cores, Modrinth projects, and arbitrary URLs",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cfg.Load()
		},
	}

	cmd.PersistentFlags().String("config", "", "config file path")
	_ = cfg.BindPFlag("config.file", cmd.PersistentFlags().Lookup("config"))

	cmd.AddCommand(
		app.versionCommand(),
		app.configCommand(),
		app.getCommand(),
		app.coreCommand(),
		app.modrinthCommand(),
	)

	cmd.SetErr(os.Stderr)
	cmd.SetOut(os.Stdout)
	return cmd
}

func commandContext(cmd *cobra.Command) context.Context {
	if cmd.Context() != nil {
		return cmd.Context()
	}
	return context.Background()
}

func normalizeChoice(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func printErr(cmd *cobra.Command, format string, args ...any) {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), format+"\n", args...)
}
