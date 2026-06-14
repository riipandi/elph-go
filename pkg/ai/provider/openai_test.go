package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAICompatibleComplete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		require.Equal(t, "proxy", r.Header.Get("X-Proxy"))

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, 0.7, body["temperature"])
		require.Equal(t, 0.95, body["top_p"])

		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{"content": "hello from gpt"},
			}},
		})
	}))
	defer srv.Close()

	p := NewOpenAICompatible(OpenAIOptions{
		ID:           "openai",
		APIKey:       "test-key",
		BaseURL:      srv.URL,
		DefaultModel: "gpt-test",
		Headers:      map[string]string{"X-Proxy": "proxy"},
		AuthHeader:   true,
		Temperature:  0.7,
		TopP:         0.95,
	})

	got, err := p.Complete(context.Background(), TurnRequest{
		SystemPrompt: "sys",
		UserPrompt:   "hi",
		Model:        "gpt-test",
	})
	require.NoError(t, err)
	require.Equal(t, "hello from gpt", got.Content)
}

func TestOpenAICompatibleCompleteReasoningEffort(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "medium", body["reasoning_effort"])

		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{"content": "done"},
			}},
		})
	}))
	defer srv.Close()

	p := NewOpenAICompatible(OpenAIOptions{
		APIKey:     "test-key",
		BaseURL:    srv.URL,
		AuthHeader: true,
	})

	got, err := p.Complete(context.Background(), TurnRequest{
		UserPrompt: "hi",
		Thinking: ThinkingConfig{
			Enabled:         true,
			ReasoningEffort: "medium",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "done", got.Content)
}

func TestOpenAICompatibleCompleteOpenRouterReasoning(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		reasoning, ok := body["reasoning"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "high", reasoning["effort"])

		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{"content": "done"},
			}},
		})
	}))
	defer srv.Close()

	p := NewOpenAICompatible(OpenAIOptions{
		APIKey:     "test-key",
		BaseURL:    srv.URL,
		AuthHeader: true,
	})

	got, err := p.Complete(context.Background(), TurnRequest{
		UserPrompt: "hi",
		Thinking: ThinkingConfig{
			Enabled:         true,
			ReasoningEffort: "high",
			ThinkingFormat:  ThinkingFormatOpenRouter,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "done", got.Content)
}

func TestOpenAICompatibleCompleteReasoning(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{
					"reasoning_content": "thinking step",
					"content":           "hello from reasoner",
				},
			}},
		})
	}))
	defer srv.Close()

	p := NewOpenAICompatible(OpenAIOptions{
		APIKey:     "test-key",
		BaseURL:    srv.URL,
		AuthHeader: true,
	})

	got, err := p.Complete(context.Background(), TurnRequest{UserPrompt: "hi"})
	require.NoError(t, err)
	require.Equal(t, "thinking step", got.Thinking)
	require.Equal(t, "hello from reasoner", got.Content)
}

func TestOpenAICompatibleStreamThinking(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, true, body["stream"])

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{"reasoning_content":"think "}}]}`)
		_, _ = fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{"content":"answer"}}]}`)
		_, _ = fmt.Fprintf(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	p := NewOpenAICompatible(OpenAIOptions{
		APIKey:     "test-key",
		BaseURL:    srv.URL,
		AuthHeader: true,
	})

	var thinking, content strings.Builder
	got, err := p.Complete(context.Background(), TurnRequest{
		UserPrompt: "hi",
		Stream: &TurnStream{
			OnThinking: func(chunk string) { thinking.WriteString(chunk) },
			OnContent:  func(chunk string) { content.WriteString(chunk) },
		},
	})
	require.NoError(t, err)
	require.Equal(t, "think ", thinking.String())
	require.Equal(t, "answer", content.String())
	require.Equal(t, "think", got.Thinking)
	require.Equal(t, "answer", got.Content)
}
