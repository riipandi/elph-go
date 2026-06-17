package websearch

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"resty.dev/v3"
)

func urlQueryEscape(s string) string {
	return url.QueryEscape(s)
}

func doJSON(ctx context.Context, client *resty.Client, method, rawURL string, headers map[string]string, body any, out any) error {
	r := client.R().SetContext(ctx).SetResponseBodyLimit(4 << 20)
	if body != nil {
		r.SetBody(body)
	}
	if out != nil {
		r.SetResult(out)
	}
	for k, v := range headers {
		r.SetHeader(k, v)
	}

	var resp *resty.Response
	var err error
	switch method {
	case "GET":
		resp, err = r.Get(rawURL)
	default:
		resp, err = r.Post(rawURL)
	}
	if err != nil {
		return err
	}
	if !resp.IsStatusSuccess() {
		return fmt.Errorf("status %s: %s", resp.Status(), trimHTTPErrorBody(resp.Bytes()))
	}
	return nil
}

func trimHTTPErrorBody(data []byte) string {
	s := strings.TrimSpace(string(data))
	if len(s) > 240 {
		return s[:240] + "..."
	}
	return s
}
