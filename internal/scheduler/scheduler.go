package scheduler

import "strings"

type Config struct {
	GhPath   string
	Hostname string
	Interval int
	Path     string
}

type InstallResult struct {
	ServiceType string
	ConfigFiles []string
	LogInfo     string
}

func SanitizeHostname(hostname string) string {
	return strings.ReplaceAll(hostname, ".", "-")
}

func ServiceName(hostname string) string {
	return "gh-auto-done-" + SanitizeHostname(hostname)
}
