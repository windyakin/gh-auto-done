package cmd

import "github.com/spf13/cobra"

func newScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage scheduled execution via launchd",
	}

	cmd.AddCommand(newScheduleInstallCmd())
	cmd.AddCommand(newScheduleUninstallCmd())

	return cmd
}
