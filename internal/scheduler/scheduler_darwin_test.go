//go:build darwin

package scheduler

import (
	"strings"
	"testing"
)

func TestLabel(t *testing.T) {
	got := label("git.pepabo.com")
	want := "com.github.cli.gh-auto-done-git-pepabo-com"
	if got != want {
		t.Errorf("label = %q, want %q", got, want)
	}
}

func TestGeneratePlist(t *testing.T) {
	t.Run("with hostname", func(t *testing.T) {
		data := plistData{
			Label:    "com.github.cli.gh-auto-done-git-pepabo-com",
			GhPath:   "/opt/homebrew/bin/gh",
			Hostname: "git.pepabo.com",
			Interval: 300,
			LogPath:  "/Users/test/Library/Logs/gh-auto-done/git-pepabo-com.log",
			Path:     "/opt/homebrew/bin:/usr/bin:/bin",
		}

		content, err := generatePlist(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		s := string(content)
		checks := []string{
			"<string>com.github.cli.gh-auto-done-git-pepabo-com</string>",
			"<string>/opt/homebrew/bin/gh</string>",
			"<string>--hostname</string>",
			"<string>git.pepabo.com</string>",
			"<integer>300</integer>",
		}
		for _, c := range checks {
			if !strings.Contains(s, c) {
				t.Errorf("plist missing %q", c)
			}
		}
	})

	t.Run("without hostname", func(t *testing.T) {
		data := plistData{
			Label:    "com.github.cli.gh-auto-done-github-com",
			GhPath:   "/opt/homebrew/bin/gh",
			Hostname: "",
			Interval: 600,
			LogPath:  "/Users/test/Library/Logs/gh-auto-done/github-com.log",
			Path:     "/usr/bin:/bin",
		}

		content, err := generatePlist(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		s := string(content)
		if strings.Contains(s, "--hostname") {
			t.Error("plist should not contain --hostname when hostname is empty")
		}
		if !strings.Contains(s, "<integer>600</integer>") {
			t.Error("plist missing interval")
		}
	})
}
