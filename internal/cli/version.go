package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (app *appContext) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print build version information",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "mcserver %s\ncommit: %s\nbuilt: %s\n", app.build.Version, app.build.Commit, app.build.Date)
		},
	}
}
