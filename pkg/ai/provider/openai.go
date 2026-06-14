package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
	"github.com/riipandi/elph/pkg/ai/utils"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

// OpenAIOptions configures an OpenAI-compatible chat completions provider.
type OpenAIOptions struct {
	ID           string
	APIKey       string
	BaseURL      string
	DefaultModel string
	Headers      map[string]string
	AuthHeader   bool
	MaxTokens    int
	Temperature  float64
	TopP         float64
}

// OpenAICompatible calls an OpenAI-style chat completions endpoint via openai-go.
type OpenAICompatible struct {
	IDName       string
	APIKey       string
	BaseURL      string
	DefaultModel string
	Headers      map[string]string
	AuthHeader   bool
	MaxTokens    int
	Temperature  float64
	TopP         float64
	client       openai.Client
}

// NewOpenAICompatible builds a provider for a compatible HTTP API.
func NewOpenAICompatible(opts OpenAIOptions) *OpenAICompatible {
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	maxTokens := opts.MaxTokens
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}

	clientOpts := []option.RequestOption{
		option.WithBaseURL(baseURL),
		option.WithHTTPClient(utils.NewHTTPClient()),
	}
	if opts.APIKey != "" {
		clientOpts = append(clientOpts, option.WithAPIKey(opts.APIKey))
	}
	for key, value := range opts.Headers {
		clientOpts = append(clientOpts, option.WithHeader(key, value))
	}
	if opts.AuthHeader && opts.APIKey == "" && !hasHeader(opts.Headers, "Authorization") {
		clientOpts = append(clientOpts, option.WithHeader("Authorization", "Bearer "))
	}

	return &OpenAICompatible{
		IDName:       opts.ID,
		APIKey:       opts.APIKey,
		BaseURL:      baseURL,
		DefaultModel: opts.DefaultModel,
		Headers:      opts.Headers,
		AuthHeader:   opts.AuthHeader,
		MaxTokens:    maxTokens,
		Temperature:  opts.Temperature,
		TopP:         opts.TopP,
		client:       openai.NewClient(clientOpts...),
	}
}

func hasHeader(headers map[string]string, key string) bool {
	for k := range headers {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}

func (p *OpenAICompatible) ID() string {
	if p.IDName == "" {
		return "openai"
	}
	return p.IDName
}

func (p *OpenAICompatible) Complete(ctx context.Context, req TurnRequest) (TurnResult, error) {
	if p.APIKey == "" && !p.AuthHeader {
		return TurnResult{}, ErrMissingAPIKey
	}
	if req.Stream != nil {
		return p.completeStream(ctx, req)
	}
	return p.completeOnce(ctx, req)
}

func (p *OpenAICompatible) completeOnce(ctx context.Context, req TurnRequest) (TurnResult, error) {
	params := p.buildParams(req, false)
	resp, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return TurnResult{}, err
	}
	if len(resp.Choices) == 0 {
		return TurnResult{}, fmt.Errorf("%s: empty response", p.ID())
	}

	result := turnResultFromChatChoice(resp.Choices[0])
	result.Usage = turnUsageFromCompletion(resp.Usage)
	if !openAIResultValid(result) {
		return TurnResult{}, fmt.Errorf("%s: empty response", p.ID())
	}
	return result, nil
}

func (p *OpenAICompatible) completeStream(ctx context.Context, req TurnRequest) (TurnResult, error) {
	params := p.buildParams(req, true)
	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	var thinking, content strings.Builder
	var usage TurnUsage
	toolAcc := newOpenAIStreamToolAccumulator()
	var finishReason string

	for stream.Next() {
		chunk := stream.Current()
		if chunk.Usage.JSON.TotalTokens.Valid() {
			usage.InputTokens = int(chunk.Usage.PromptTokens)
			usage.OutputTokens = int(chunk.Usage.CompletionTokens)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		choice := chunk.Choices[0]
		if choice.FinishReason != "" {
			finishReason = choice.FinishReason
		}
		delta := choice.Delta
		if rc := openAIStreamReasoningText(delta.JSON.ExtraFields, delta.RawJSON()); rc != "" {
			thinking.WriteString(rc)
			req.Stream.emitThinking(rc)
		}
		if delta.Content != "" {
			content.WriteString(delta.Content)
			req.Stream.emitContent(delta.Content)
		}
		if len(delta.ToolCalls) > 0 {
			toolAcc.absorbSDK(delta.ToolCalls)
		}
	}
	if err := stream.Err(); err != nil {
		return TurnResult{}, err
	}

	result := TurnResult{
		Thinking:  strings.TrimSpace(thinking.String()),
		Content:   strings.TrimSpace(content.String()),
		Usage:     usage,
		ToolCalls: toolAcc.result(),
	}
	if finishReason == "tool_calls" || len(result.ToolCalls) > 0 {
		result.StopReason = StopReasonToolUse
	} else {
		result.StopReason = StopReasonEndTurn
	}
	if !openAIResultValid(result) {
		return TurnResult{}, fmt.Errorf("%s: empty response", p.ID())
	}
	return result, nil
}

func (p *OpenAICompatible) buildParams(req TurnRequest, stream bool) openai.ChatCompletionNewParams {
	model := req.Model
	if model == "" {
		model = p.DefaultModel
	}

	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(model),
		Messages: openAIChatMessages(req.SystemPrompt, BuildMessages(req), req.Thinking, req.Compat),
	}
	if p.Temperature != 0 {
		params.Temperature = openai.Float(p.Temperature)
	}
	if p.TopP != 0 {
		params.TopP = openai.Float(p.TopP)
	}

	maxField := "max_tokens"
	if req.Compat.MaxTokensField != "" {
		maxField = req.Compat.MaxTokensField
	}
	if p.MaxTokens > 0 {
		switch maxField {
		case "max_completion_tokens":
			params.MaxCompletionTokens = openai.Int(int64(p.MaxTokens))
		default:
			params.MaxTokens = openai.Int(int64(p.MaxTokens))
		}
	}

	applyOpenAIThinkingParams(&params, req.Thinking, req.Compat)

	if tools := openAIChatTools(req.Tools); len(tools) > 0 {
		params.Tools = tools
		params.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: openai.String(string(openai.ChatCompletionToolChoiceOptionAutoAuto)),
		}
	}
	if stream && req.Compat.supportsUsageInStreaming() {
		params.StreamOptions = openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		}
	}
	return params
}

func applyOpenAIThinkingParams(params *openai.ChatCompletionNewParams, thinking ThinkingConfig, compat Compat) {
	if !thinking.Enabled {
		return
	}
	switch thinking.ThinkingFormat {
	case ThinkingFormatOpenRouter:
		if thinking.ReasoningEffort != "" {
			params.SetExtraFields(map[string]any{
				"reasoning": map[string]any{"effort": thinking.ReasoningEffort},
			})
		}
	case ThinkingFormatQwen:
		params.SetExtraFields(map[string]any{"enable_thinking": thinking.EnableThinking})
	default:
		if compat.supportsReasoningEffort() && thinking.ReasoningEffort != "" {
			params.ReasoningEffort = shared.ReasoningEffort(thinking.ReasoningEffort)
		}
	}
}

func turnUsageFromCompletion(usage openai.CompletionUsage) TurnUsage {
	return TurnUsage{
		InputTokens:  int(usage.PromptTokens),
		OutputTokens: int(usage.CompletionTokens),
	}
}