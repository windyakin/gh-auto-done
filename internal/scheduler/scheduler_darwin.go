//go:build darwin

package scheduler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.GhPath}}</string>
        <string>auto-done</string>
{{- if .Hostname}}
        <string>--hostname</string>
        <string>{{.Hostname}}</string>
{{- end}}
    </array>
    <key>StartInterval</key>
    <integer>{{.Interval}}</integer>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>{{.Path}}</string>
    </dict>
</dict>
</plist>
`

type plistData struct {
	Label    string
	GhPath   string
	Hostname string
	Interval int
	LogPath  string
	Path     string
}

func label(hostname string) string {
	return "com.github.cli." + ServiceName(hostname)
}

func plistPath(hostname string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", label(hostname)+".plist"), nil
}

func logDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Logs", "gh-auto-done"), nil
}

func logPath(hostname string) (string, error) {
	dir, err := logDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SanitizeHostname(hostname)+".log"), nil
}

func generatePlist(data plistData) ([]byte, error) {
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Install(config Config) (*InstallResult, error) {
	pp, err := plistPath(config.Hostname)
	if err != nil {
		return nil, err
	}

	ld, err := logDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(ld, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	lp, err := logPath(config.Hostname)
	if err != nil {
		return nil, err
	}

	data := plistData{
		Label:    label(config.Hostname),
		GhPath:   config.GhPath,
		Hostname: config.Hostname,
		Interval: config.Interval,
		LogPath:  lp,
		Path:     config.Path,
	}

	content, err := generatePlist(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plist: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(pp), 0755); err != nil {
		return nil, fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	if err := os.WriteFile(pp, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to write plist: %w", err)
	}

	bootout(pp)

	if err := bootstrap(pp); err != nil {
		return nil, fmt.Errorf("failed to bootstrap launchd agent: %w", err)
	}

	return &InstallResult{
		ServiceType: "launchd agent",
		ConfigFiles: []string{pp},
		LogInfo:     lp,
	}, nil
}

func Uninstall(hostname string) error {
	pp, err := plistPath(hostname)
	if err != nil {
		return err
	}

	if _, err := os.Stat(pp); os.IsNotExist(err) {
		return fmt.Errorf("plist not found: %s", pp)
	}

	bootout(pp)

	if err := os.Remove(pp); err != nil {
		return fmt.Errorf("failed to remove plist: %w", err)
	}

	return nil
}

func bootstrap(path string) error {
	domain := fmt.Sprintf("gui/%d", os.Getuid())
	return exec.Command("launchctl", "bootstrap", domain, path).Run()
}

func bootout(path string) {
	domain := fmt.Sprintf("gui/%d", os.Getuid())
	_ = exec.Command("launchctl", "bootout", domain, path).Run()
}
