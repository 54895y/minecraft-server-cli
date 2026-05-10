package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func (app *appContext) configCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and write mcserver configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the active config path",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), app.config.Path())
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Print all active settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := json.MarshalIndent(app.config.AllSettings(), "", "  ")
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <key>",
		Short: "Print one setting",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), app.config.GetString(args[0]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "set <key> <value>",
		Short:   "Persist one setting",
		Example: "mcserver config set download.threads 16",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			app.config.Set(args[0], parseConfigValue(args[1]))
			if err := app.config.Write(); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "saved %s\n", app.config.Path())
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Remove the persisted config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.config.Reset(); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", app.config.Path())
			return nil
		},
	})

	return cmd
}

func parseConfigValue(value string) any {
	trimmed := strings.TrimSpace(value)
	if i, err := strconv.Atoi(trimmed); err == nil {
		return i
	}
	if b, err := strconv.ParseBool(trimmed); err == nil {
		return b
	}
	if strings.Contains(trimmed, ",") {
		parts := strings.Split(trimmed, ",")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			if p := strings.TrimSpace(part); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	return value
}
