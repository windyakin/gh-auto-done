package scheduler

import "testing"

func TestSanitizeHostname(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"github.com", "github-com"},
		{"git.pepabo.com", "git-pepabo-com"},
		{"ghe.example.co.jp", "ghe-example-co-jp"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeHostname(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeHostname(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestServiceName(t *testing.T) {
	tests := []struct {
		hostname string
		want     string
	}{
		{"github.com", "gh-auto-done-github-com"},
		{"git.pepabo.com", "gh-auto-done-git-pepabo-com"},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			got := ServiceName(tt.hostname)
			if got != tt.want {
				t.Errorf("ServiceName(%q) = %q, want %q", tt.hostname, got, tt.want)
			}
		})
	}
}
