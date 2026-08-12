package cmd

import (
	"fmt"
	"os"

	"git.pepabo.com/windyakin/gh-auto-done/internal/scheduler"
	"github.com/spf13/cobra"
)

func newScheduleUninstallCmd() *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the scheduled job",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScheduleUninstall(hostname)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "GitHub hostname (default: github.com)")

	return cmd
}

func runScheduleUninstall(hostname string) error {
	if hostname == "" {
		hostname = "github.com"
	}

	if err := scheduler.Uninstall(hostname); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Uninstalled scheduled job for %s\n", hostname)

	return nil
}
