//go:build linux

package scheduler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const serviceTemplate = `[Unit]
Description=Mark closed/merged GitHub notifications as done ({{.Hostname}})

[Service]
Type=oneshot
ExecStart={{.GhPath}} auto-done{{if .Hostname}} --hostname {{.Hostname}}{{end}}
Environment=PATH={{.Path}}
`

const timerTemplate = `[Unit]
Description=Run gh auto-done periodically ({{.Hostname}})

[Timer]
OnBootSec=60
OnUnitActiveSec={{.Interval}}s
Persistent=false

[Install]
WantedBy=timers.target
`

type unitData struct {
	GhPath   string
	Hostname string
	Interval int
	Path     string
}

func unitDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "systemd", "user"), nil
}

func unitPaths(hostname string) (service string, timer string, err error) {
	dir, err := unitDir()
	if err != nil {
		return "", "", err
	}
	name := ServiceName(hostname)
	return filepath.Join(dir, name+".service"), filepath.Join(dir, name+".timer"), nil
}

func renderTemplate(tmplStr string, data unitData) ([]byte, error) {
	tmpl, err := template.New("unit").Parse(tmplStr)
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
	servicePath, timerPath, err := unitPaths(config.Hostname)
	if err != nil {
		return nil, err
	}

	dir, err := unitDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create systemd user directory: %w", err)
	}

	data := unitData{
		GhPath:   config.GhPath,
		Hostname: config.Hostname,
		Interval: config.Interval,
		Path:     config.Path,
	}

	serviceContent, err := renderTemplate(serviceTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate service unit: %w", err)
	}

	timerContent, err := renderTemplate(timerTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate timer unit: %w", err)
	}

	if err := os.WriteFile(servicePath, serviceContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write service unit: %w", err)
	}

	if err := os.WriteFile(timerPath, timerContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write timer unit: %w", err)
	}

	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return nil, fmt.Errorf("failed to reload systemd: %w", err)
	}

	timerUnit := ServiceName(config.Hostname) + ".timer"
	if err := exec.Command("systemctl", "--user", "enable", "--now", timerUnit).Run(); err != nil {
		return nil, fmt.Errorf("failed to enable timer: %w", err)
	}

	return &InstallResult{
		ServiceType: "systemd timer",
		ConfigFiles: []string{servicePath, timerPath},
		LogInfo:     fmt.Sprintf("journalctl --user -u %s", ServiceName(config.Hostname)),
	}, nil
}

func Uninstall(hostname string) error {
	servicePath, timerPath, err := unitPaths(hostname)
	if err != nil {
		return err
	}

	timerUnit := ServiceName(hostname) + ".timer"
	_ = exec.Command("systemctl", "--user", "disable", "--now", timerUnit).Run()

	for _, path := range []string{timerPath, servicePath} {
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
		}
	}

	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()

	return nil
}
