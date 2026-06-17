package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"resty.dev/v3"
)

const (
	defaultHTTPTimeout       = 120 * time.Second
	streamResponseHeaderWait = 60 * time.Second
)

// NewHTTPClient returns a resty client with the default upstream timeout.
func NewHTTPClient() *resty.Client {
	return resty.New().SetTimeout(defaultHTTPTimeout)
}

// NewStreamingHTTPClient returns a resty client tuned for SSE: bounded wait for
// response headers and no overall request timeout (stall watch handles hangs).
func NewStreamingHTTPClient() *resty.Client {
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		// http.DefaultTransport is always *http.Transport in practice.
		transport := &http.Transport{
			ResponseHeaderTimeout: streamResponseHeaderWait,
		}
		return resty.New().SetTransport(transport)
	}
	transport := baseTransport.Clone()
	transport.ResponseHeaderTimeout = streamResponseHeaderWait
	return resty.New().SetTransport(transport)
}

// PostJSON sends a JSON request and decodes a JSON response.
func PostJSON(ctx context.Context, client *resty.Client, url string, headers map[string]string, body any, out any) error {
	if client == nil {
		client = NewHTTPClient()
	}

	r := client.R().SetContext(ctx)
	if body != nil {
		r.SetBody(body)
	}
	for k, v := range headers {
		r.SetHeader(k, v)
	}

	resp, err := r.Post(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return fmt.Errorf("upstream %s: %s", resp.Status(), trimBody(resp.Bytes()))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(resp.Bytes(), out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// GetJSON sends a GET request and decodes a JSON response.
func GetJSON(ctx context.Context, client *resty.Client, url string, out any) error {
	return GetJSONWithHeaders(ctx, client, url, nil, out)
}

// GetJSONWithHeaders sends a GET request with optional headers and decodes JSON.
func GetJSONWithHeaders(ctx context.Context, client *resty.Client, url string, headers map[string]string, out any) error {
	if client == nil {
		client = NewHTTPClient()
	}

	r := client.R().SetContext(ctx)
	for k, v := range headers {
		r.SetHeader(k, v)
	}

	resp, err := r.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return fmt.Errorf("upstream %s: %s", resp.Status(), trimBody(resp.Bytes()))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(resp.Bytes(), out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func trimBody(raw []byte) string {
	text := strings.TrimSpace(string(raw))
	if len(text) > 240 {
		return text[:240] + "..."
	}
	return text
}
