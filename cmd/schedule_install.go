package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"git.pepabo.com/windyakin/gh-auto-done/internal/scheduler"
	"github.com/spf13/cobra"
)

func newScheduleInstallCmd() *cobra.Command {
	var hostname string
	var interval int

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a scheduled job to run gh auto-done periodically",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScheduleInstall(hostname, interval)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "GitHub hostname (default: github.com)")
	cmd.Flags().IntVar(&interval, "interval", 300, "Interval in seconds between runs")

	return cmd
}

func runScheduleInstall(hostname string, interval int) error {
	if hostname == "" {
		hostname = "github.com"
	}

	ghPath, err := resolveGhPath()
	if err != nil {
		return fmt.Errorf("failed to find gh: %w", err)
	}

	config := scheduler.Config{
		GhPath:   ghPath,
		Hostname: hostname,
		Interval: interval,
		Path:     os.Getenv("PATH"),
	}

	result, err := scheduler.Install(config)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Installed %s:\n", result.ServiceType)
	for _, f := range result.ConfigFiles {
		fmt.Fprintf(os.Stderr, "  Config: %s\n", f)
	}
	fmt.Fprintf(os.Stderr, "  Log:    %s\n", result.LogInfo)
	fmt.Fprintf(os.Stderr, "  Interval: %ds\n", interval)
	fmt.Fprintf(os.Stderr, "  Host: %s\n", hostname)
	fmt.Fprintf(os.Stderr, "\nTo uninstall: gh auto-done schedule uninstall --hostname %s\n", hostname)

	return nil
}

func resolveGhPath() (string, error) {
	if p := os.Getenv("GH_PATH"); p != "" {
		return p, nil
	}

	path, err := exec.LookPath("gh")
	if err != nil {
		return "", fmt.Errorf("gh not found in PATH: %w", err)
	}

	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return path, nil
		}
		return abs, nil
	}

	return path, nil
}
