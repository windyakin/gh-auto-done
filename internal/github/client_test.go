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

func TestListNotificationsPagination(t *testing.T) {
	page1 := `[{"id":"1","subject":{"title":"PR 1","url":"https://api.github.com/repos/owner/repo/pulls/1","type":"PullRequest"},"repository":{"full_name":"owner/repo"}}]`
	page2 := `[{"id":"2","subject":{"title":"PR 2","url":"https://api.github.com/repos/owner/repo/pulls/2","type":"PullRequest"},"repository":{"full_name":"owner/repo"}}]`

	var requestCount int
	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		requestCount++
		switch {
		case r.URL.Path == "/notifications" && r.URL.Query().Get("page") == "":
			resp := jsonResponseWithBody(200, page1)
			resp.Header.Set("Link", `<https://api.github.com/notifications?per_page=100&page=2>; rel="next"`)
			return resp, nil
		case r.URL.Path == "/notifications" && r.URL.Query().Get("page") == "2":
			return jsonResponseWithBody(200, page2), nil
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			return jsonResponseWithBody(200, "[]"), nil
		}
	})

	notifications, err := client.ListNotifications(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifications) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(notifications))
	}
	if requestCount != 2 {
		t.Errorf("expected 2 requests, got %d", requestCount)
	}
	if notifications[0].ID != "1" {
		t.Errorf("expected first notification ID 1, got %s", notifications[0].ID)
	}
	if notifications[1].ID != "2" {
		t.Errorf("expected second notification ID 2, got %s", notifications[1].ID)
	}
}

func TestNextPageURL(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "with next link",
			header: `<https://api.github.com/notifications?page=2>; rel="next", <https://api.github.com/notifications?page=5>; rel="last"`,
			want:   "https://api.github.com/notifications?page=2",
		},
		{
			name:   "no next link",
			header: `<https://api.github.com/notifications?page=5>; rel="last"`,
			want:   "",
		},
		{
			name:   "empty header",
			header: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextPageURL(tt.header)
			if got != tt.want {
				t.Errorf("nextPageURL(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
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
	var patchCalled bool

	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/notifications/threads/123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		patchCalled = true
		return &http.Response{
			StatusCode: 205,
			Header:     http.Header{},
			Body:       http.NoBody,
		}, nil
	})

	err := client.MarkThreadAsDone(context.Background(), "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !patchCalled {
		t.Error("expected PATCH to be called")
	}
}
