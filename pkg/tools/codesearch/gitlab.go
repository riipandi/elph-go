package codesearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"resty.dev/v3"
)

func searchGitLab(ctx context.Context, client *resty.Client, query, token, apiBase string) ([]Result, error) {
	return searchGitLabAt(ctx, client, apiBase+"/search", query, token)
}

func searchGitLabAt(ctx context.Context, client *resty.Client, endpoint, query, token string) ([]Result, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if u.RawQuery == "" {
		q := u.Query()
		q.Set("scope", "blobs")
		q.Set("search", query)
		q.Set("per_page", "20")
		u.RawQuery = q.Encode()
	}

	resp, err := client.R().SetContext(ctx).
		SetHeader("PRIVATE-TOKEN", token).
		SetResponseBodyLimit(4 << 20).
		Get(u.String())
	if err != nil {
		return nil, err
	}
	data := resp.Bytes()
	if !resp.IsStatusSuccess() {
		return nil, fmt.Errorf("status %s: %s", resp.Status(), trimAPIError(data))
	}

	var items []struct {
		Basename  string `json:"basename"`
		Path      string `json:"path"`
		Ref       string `json:"ref"`
		ProjectID int    `json:"project_id"`
		Data      string `json:"data"`
		WebURL    string `json:"web_url"`
	}
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	out := make([]Result, 0, len(items))
	for _, item := range items {
		path := item.Path
		if path == "" {
			path = item.Basename
		}
		repo := fmt.Sprintf("project:%d", item.ProjectID)
		if item.Ref != "" {
			repo += "@" + item.Ref
		}
		out = append(out, Result{
			Repository: repo,
			Path:       path,
			URL:        item.WebURL,
			Snippet:    strings.TrimSpace(item.Data),
		})
	}
	return out, nil
}

func trimAPIError(data []byte) string {
	s := strings.TrimSpace(string(data))
	if len(s) > 240 {
		return s[:240] + "..."
	}
	return s
}
