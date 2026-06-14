package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/riipandi/elph/pkg/ai/utils"
)

const defaultAnthropicBaseURL = "https://api.anthropic.com"

// AnthropicOptions configures an Anthropic Messages API provider.
type AnthropicOptions struct {
	ID          string
	APIKey      string
	Model       string
	BaseURL     string
	Headers     map[string]string
	MaxTokens   int
	Temperature float64
	TopP        float64
}

// Anthropic calls the Anthropic Messages API via anthropic-sdk-go.
type Anthropic struct {
	IDName      string
	APIKey      string
	Model       string
	BaseURL     string
	Headers     map[string]string
	MaxTokens   int
	Temperature float64
	TopP        float64
	client      anthropic.Client
}

// NewAnthropic builds an Anthropic provider from explicit settings.
func NewAnthropic(opts AnthropicOptions) *Anthropic {
	maxTokens := opts.MaxTokens
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}

	clientOpts := []option.RequestOption{
		option.WithBaseURL(normalizeAnthropicBaseURL(opts.BaseURL)),
		option.WithHTTPClient(utils.NewHTTPClient()),
	}
	if opts.APIKey != "" {
		clientOpts = append(clientOpts, option.WithAPIKey(opts.APIKey))
	}
	for key, value := range opts.Headers {
		clientOpts = append(clientOpts, option.WithHeader(key, value))
	}

	return &Anthropic{
		IDName:      opts.ID,
		APIKey:      opts.APIKey,
		Model:       opts.Model,
		BaseURL:     strings.TrimRight(opts.BaseURL, "/"),
		Headers:     opts.Headers,
		MaxTokens:   maxTokens,
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		client:      anthropic.NewClient(clientOpts...),
	}
}

func normalizeAnthropicBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return defaultAnthropicBaseURL
	}
	if strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}
	return baseURL
}

func (p *Anthropic) ID() string {
	if p.IDName == "" {
		return "anthropic"
	}
	return p.IDName
}

func (p *Anthropic) Complete(ctx context.Context, req TurnRequest) (TurnResult, error) {
	if p.APIKey == "" {
		return TurnResult{}, ErrMissingAPIKey
	}
	if req.Stream != nil {
		return p.completeStream(ctx, req)
	}
	return p.completeOnce(ctx, req)
}

func (p *Anthropic) completeOnce(ctx context.Context, req TurnRequest) (TurnResult, error) {
	params := p.buildParams(req)
	resp, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return TurnResult{}, err
	}

	result := turnResultFromAnthropicMessage(resp)
	if !anthropicResultValid(result) {
		return TurnResult{}, fmt.Errorf("%s: empty response", p.ID())
	}
	return result, nil
}

func (p *Anthropic) completeStream(ctx context.Context, req TurnRequest) (TurnResult, error) {
	params := p.buildParams(req)
	stream := p.client.Messages.NewStreaming(ctx, params)

	var thinking, content strings.Builder
	var usage TurnUsage
	var toolCalls []ToolCall
	var currentTool *ToolCall
	var toolInput strings.Builder

	for stream.Next() {
		event := stream.Current()
		switch variant := event.AsAny().(type) {
		case anthropic.MessageStartEvent:
			usage.InputTokens = int(variant.Message.Usage.InputTokens)
		case anthropic.MessageDeltaEvent:
			if variant.Usage.OutputTokens > 0 {
				usage.OutputTokens = int(variant.Usage.OutputTokens)
			}
		case anthropic.ContentBlockStartEvent:
			if variant.ContentBlock.Type == "tool_use" {
				currentTool = &ToolCall{
					ID:   variant.ContentBlock.ID,
					Name: variant.ContentBlock.Name,
				}
				toolInput.Reset()
				if variant.ContentBlock.Input != nil {
					raw, _ := json.Marshal(variant.ContentBlock.Input)
					toolInput.Write(raw)
				}
			}
		case anthropic.ContentBlockDeltaEvent:
			delta := variant.Delta
			switch delta.Type {
			case "thinking_delta":
				if delta.Thinking != "" {
					thinking.WriteString(delta.Thinking)
					req.Stream.emitThinking(delta.Thinking)
				}
			case "text_delta":
				if delta.Text != "" {
					content.WriteString(delta.Text)
					req.Stream.emitContent(delta.Text)
				}
			case "input_json_delta":
				if delta.PartialJSON != "" {
					toolInput.WriteString(delta.PartialJSON)
				}
			}
		case anthropic.ContentBlockStopEvent:
			if currentTool != nil {
				args := json.RawMessage(toolInput.String())
				if len(args) == 0 {
					args = json.RawMessage("{}")
				}
				currentTool.Arguments = args
				toolCalls = append(toolCalls, *currentTool)
				currentTool = nil
			}
		}
	}
	if err := stream.Err(); err != nil {
		return TurnResult{}, err
	}

	result := TurnResult{
		Thinking:  strings.TrimSpace(thinking.String()),
		Content:   strings.TrimSpace(content.String()),
		Usage:     usage,
		ToolCalls: toolCalls,
	}
	if len(result.ToolCalls) > 0 {
		result.StopReason = StopReasonToolUse
	} else {
		result.StopReason = StopReasonEndTurn
	}
	if !anthropicResultValid(result) {
		return TurnResult{}, fmt.Errorf("%s: empty response", p.ID())
	}
	return result, nil
}

func (p *Anthropic) buildParams(req TurnRequest) anthropic.MessageNewParams {
	model := req.Model
	if model == "" {
		model = p.Model
	}

	params := anthropic.MessageNewParams{
		Model:      anthropic.Model(model),
		MaxTokens:  int64(p.MaxTokens),
		Messages:   anthropicMessages(BuildMessages(req)),
		Temperature: anthropic.Float(p.Temperature),
		TopP:        anthropic.Float(p.TopP),
	}
	if strings.TrimSpace(req.SystemPrompt) != "" {
		params.System = []anthropic.TextBlockParam{{Text: req.SystemPrompt}}
	}
	if tools := anthropicTools(req.Tools); len(tools) > 0 {
		params.Tools = tools
	}
	applyAnthropicThinkingParams(&params, req.Thinking)
	return params
}

func applyAnthropicThinkingParams(params *anthropic.MessageNewParams, thinking ThinkingConfig) {
	if !thinking.Enabled {
		return
	}
	if thinking.Adaptive {
		params.Thinking = anthropic.ThinkingConfigParamUnion{
			OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{},
		}
		if thinking.AdaptiveEffort != "" {
			params.OutputConfig = anthropic.OutputConfigParam{
				Effort: anthropic.OutputConfigEffort(thinking.AdaptiveEffort),
			}
		}
		return
	}
	if thinking.BudgetTokens > 0 {
		params.Thinking = anthropic.ThinkingConfigParamOfEnabled(int64(thinking.BudgetTokens))
	}
}