//go:build windows

package scheduler

import (
	"fmt"
	"os/exec"
)

func taskName(hostname string) string {
	return `gh-auto-done\` + SanitizeHostname(hostname)
}

func Install(config Config) (*InstallResult, error) {
	intervalMinutes := config.Interval / 60
	if intervalMinutes < 1 {
		intervalMinutes = 1
	}

	tr := fmt.Sprintf(`"%s" auto-done`, config.GhPath)
	if config.Hostname != "" {
		tr += fmt.Sprintf(` --hostname %s`, config.Hostname)
	}

	name := taskName(config.Hostname)

	cmd := exec.Command("schtasks", "/create",
		"/tn", name,
		"/tr", tr,
		"/sc", "minute",
		"/mo", fmt.Sprintf("%d", intervalMinutes),
		"/f",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create scheduled task: %w\n%s", err, output)
	}

	return &InstallResult{
		ServiceType: "Windows scheduled task",
		ConfigFiles: []string{name},
		LogInfo:     "Task Scheduler (taskschd.msc)",
	}, nil
}

func Uninstall(hostname string) error {
	name := taskName(hostname)

	cmd := exec.Command("schtasks", "/delete",
		"/tn", name,
		"/f",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete scheduled task: %w\n%s", err, output)
	}

	return nil
}
