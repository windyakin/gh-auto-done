package github

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newTestClient(t *testing.T, handler roundTripFunc) *Client {
	t.Helper()
	rest, err := api.NewRESTClient(api.ClientOptions{
		Transport:        handler,
		AuthToken:        "test-token",
		UnixDomainSocket: "",
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return &Client{rest: rest}
}

func jsonResponseWithBody(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestListNotifications(t *testing.T) {
	body := `[
		{"id":"1","subject":{"title":"Fix bug","url":"https://api.github.com/repos/owner/repo/pulls/1","type":"PullRequest"},"repository":{"full_name":"owner/repo"}},
		{"id":"2","subject":{"title":"Add feature","url":"https://api.github.com/repos/owner/repo/issues/2","type":"Issue"},"repository":{"full_name":"owner/repo"}},
		{"id":"3","subject":{"title":"v1.0","url":"https://api.github.com/repos/owner/repo/releases/3","type":"Release"},"repository":{"full_name":"owner/repo"}}
	]`

	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/notifications" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		return jsonResponseWithBody(200, body), nil
	})

	notifications, err := client.ListNotifications(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifications) != 3 {
		t.Fatalf("expected 3 notifications, got %d", len(notifications))
	}
	if notifications[0].Subject.Type != "PullRequest" {
		t.Errorf("expected PullRequest, got %s", notifications[0].Subject.Type)
	}
	if notifications[1].Subject.Type != "Issue" {
		t.Errorf("expected Issue, got %s", notifications[1].Subject.Type)
	}
	if notifications[2].Subject.Type != "Release" {
		t.Errorf("expected Release, got %s", notifications[2].Subject.Type)
	}
}

func TestGetSubjectState(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantState string
	}{
		{
			name:      "closed PR",
			body:      `{"state":"closed","merged":true}`,
			wantState: "closed",
		},
		{
			name:      "open PR",
			body:      `{"state":"open"}`,
			wantState: "open",
		},
		{
			name:      "closed issue",
			body:      `{"state":"closed","state_reason":"completed"}`,
			wantState: "closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
				return jsonResponseWithBody(200, tt.body), nil
			})

			state, err := client.GetSubjectState(context.Background(), "https://api.github.com/repos/owner/repo/pulls/1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if state.State != tt.wantState {
				t.Errorf("expected state %q, got %q", tt.wantState, state.State)
			}
		})
	}
}

func TestMarkThreadAsDone(t *testing.T) {
	var called bool

	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/notifications/threads/123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		called = true
		return &http.Response{
			StatusCode: 204,
			Header:     http.Header{},
			Body:       http.NoBody,
		}, nil
	})

	err := client.MarkThreadAsDone(context.Background(), "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected DELETE to be called")
	}
}
