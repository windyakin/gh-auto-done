package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/cli/go-gh/v2/pkg/api"
)

var linkNextPattern = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

type Client struct {
	rest *api.RESTClient
}

func NewClient(hostname string) (*Client, error) {
	opts := api.ClientOptions{}
	if hostname != "" {
		opts.Host = hostname
	}
	rest, err := api.NewRESTClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %w", err)
	}
	return &Client{rest: rest}, nil
}

func (c *Client) ListNotifications(ctx context.Context) ([]Notification, error) {
	var all []Notification
	path := "notifications?per_page=100"

	for path != "" {
		resp, err := c.rest.RequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list notifications: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var page []Notification
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("failed to parse notifications: %w", err)
		}
		all = append(all, page...)

		path = nextPageURL(resp.Header.Get("Link"))
	}

	return all, nil
}

func nextPageURL(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	m := linkNextPattern.FindStringSubmatch(linkHeader)
	if m == nil {
		return ""
	}
	return m[1]
}

func (c *Client) GetSubjectState(ctx context.Context, subjectURL string) (*SubjectState, error) {
	var state SubjectState
	err := c.rest.DoWithContext(ctx, http.MethodGet, subjectURL, nil, &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (c *Client) MarkThreadAsDone(ctx context.Context, threadID string) error {
	path := fmt.Sprintf("notifications/threads/%s", threadID)
	resp, err := c.rest.RequestWithContext(ctx, http.MethodPatch, path, nil)
	if err != nil {
		return fmt.Errorf("failed to mark thread %s as done: %w", threadID, err)
	}
	resp.Body.Close()
	return nil
}
