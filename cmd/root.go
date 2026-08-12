package cmd

import (
	"context"
	"fmt"
	"os"

	"git.pepabo.com/windyakin/gh-auto-done/internal/github"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var hostname string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "gh-auto-done",
		Short: "Automatically mark notifications as done for closed/merged PRs and Issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), hostname, dryRun)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "GitHub hostname (e.g. git.pepabo.com)")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview what would be done without making changes")

	cmd.AddCommand(newScheduleCmd())

	return cmd
}

func Execute() error {
	return NewRootCmd().Execute()
}

func run(ctx context.Context, hostname string, dryRun bool) error {
	client, err := github.NewClient(hostname)
	if err != nil {
		return err
	}

	notifications, err := client.ListNotifications(ctx)
	if err != nil {
		return err
	}

	if len(notifications) == 0 {
		fmt.Fprintln(os.Stderr, "No notifications found.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Checking %d notifications...\n", len(notifications))

	var doneCount, skipCount, errorCount int

	for _, n := range notifications {
		if n.Subject.Type != "PullRequest" && n.Subject.Type != "Issue" {
			skipCount++
			continue
		}

		if n.Subject.URL == "" {
			skipCount++
			continue
		}

		state, err := client.GetSubjectState(ctx, n.Subject.URL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [WARN] %s %s %q (failed to get state: %v)\n",
				n.Repository.FullName, n.Subject.Type, n.Subject.Title, err)
			errorCount++
			continue
		}

		if state.State != "closed" {
			skipCount++
			continue
		}

		if dryRun {
			fmt.Fprintf(os.Stderr, "  [DRY RUN] %s %s %q\n",
				n.Repository.FullName, n.Subject.Type, n.Subject.Title)
			doneCount++
			continue
		}

		if err := client.MarkThreadAsDone(ctx, n.ID); err != nil {
			fmt.Fprintf(os.Stderr, "  [WARN] %s %s %q (failed to mark as done: %v)\n",
				n.Repository.FullName, n.Subject.Type, n.Subject.Title, err)
			errorCount++
			continue
		}

		fmt.Fprintf(os.Stderr, "  [DONE] %s %s %q\n",
			n.Repository.FullName, n.Subject.Type, n.Subject.Title)
		doneCount++
	}

	fmt.Fprintf(os.Stderr, "Summary: %d marked as done, %d skipped, %d errors\n",
		doneCount, skipCount, errorCount)

	return nil
}
