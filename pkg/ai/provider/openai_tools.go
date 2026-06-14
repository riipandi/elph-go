package provider

import (
	"encoding/json"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/respjson"
	"github.com/openai/openai-go/v3/shared"
)

func openAIChatTools(tools []ToolDefinition) []openai.ChatCompletionToolUnionParam {
	if len(tools) == 0 {
		return nil
	}
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		params := shared.FunctionParameters{}
		for key, value := range tool.Parameters {
			params[key] = value
		}
		out = append(out, openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        tool.Name,
			Description: openai.String(tool.Description),
			Parameters:  params,
		}))
	}
	return out
}

func openAIChatMessages(systemPrompt string, messages []ChatMessage, thinking ThinkingConfig, compat Compat) []openai.ChatCompletionMessageParamUnion {
	capacity := len(messages)
	if strings.TrimSpace(systemPrompt) != "" {
		capacity++
	}
	out := make([]openai.ChatCompletionMessageParamUnion, 0, capacity)
	if strings.TrimSpace(systemPrompt) != "" {
		if thinking.Enabled && compat.supportsDeveloperRole() {
			out = append(out, openai.DeveloperMessage(systemPrompt))
		} else {
			out = append(out, openai.SystemMessage(systemPrompt))
		}
	}
	for _, msg := range messages {
		switch msg.Role {
		case "assistant":
			asst := openai.ChatCompletionAssistantMessageParam{}
			if strings.TrimSpace(msg.Content) != "" {
				asst.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.String(msg.Content),
				}
			}
			for _, call := range msg.ToolCalls {
				args := string(call.Arguments)
				if args == "" {
					args = "{}"
				}
				asst.ToolCalls = append(asst.ToolCalls, openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: call.ID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      call.Name,
							Arguments: args,
						},
					},
				})
			}
			out = append(out, openai.ChatCompletionMessageParamUnion{OfAssistant: &asst})
		case "tool":
			out = append(out, openai.ToolMessage(msg.Content, msg.ToolCallID))
		default:
			out = append(out, openai.UserMessage(msg.Content))
		}
	}
	return out
}

func turnResultFromChatChoice(choice openai.ChatCompletionChoice) TurnResult {
	message := choice.Message
	result := TurnResult{
		Thinking: strings.TrimSpace(openAIReasoningText(message.JSON.ExtraFields, message.RawJSON())),
		Content:  strings.TrimSpace(message.Content),
	}
	for _, call := range message.ToolCalls {
		fn := call.AsFunction()
		if strings.TrimSpace(fn.ID) == "" || strings.TrimSpace(fn.Function.Name) == "" {
			continue
		}
		args := json.RawMessage(fn.Function.Arguments)
		if len(args) == 0 {
			args = json.RawMessage("{}")
		}
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:        fn.ID,
			Name:      fn.Function.Name,
			Arguments: args,
		})
	}
	switch choice.FinishReason {
	case "tool_calls":
		result.StopReason = StopReasonToolUse
	default:
		if len(result.ToolCalls) > 0 {
			result.StopReason = StopReasonToolUse
		} else {
			result.StopReason = StopReasonEndTurn
		}
	}
	return result
}

func openAIResultValid(result TurnResult) bool {
	return result.Thinking != "" || result.Content != "" || len(result.ToolCalls) > 0
}

func extraFieldString(fields map[string]respjson.Field, key string) string {
	if len(fields) == 0 {
		return ""
	}
	field, ok := fields[key]
	if !ok || !field.Valid() {
		return ""
	}
	return decodeJSONString(field.Raw())
}

func openAIReasoningText(extra map[string]respjson.Field, rawJSON string) string {
	if text := extraFieldString(extra, "reasoning_content"); text != "" {
		return text
	}
	if text := extraFieldString(extra, "reasoning"); text != "" {
		return text
	}
	var vendor struct {
		ReasoningContent string `json:"reasoning_content"`
		Reasoning        string `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &vendor); err != nil {
		return ""
	}
	if vendor.ReasoningContent != "" {
		return vendor.ReasoningContent
	}
	return vendor.Reasoning
}

func openAIStreamReasoningText(extra map[string]respjson.Field, rawJSON string) string {
	return openAIReasoningText(extra, rawJSON)
}

func decodeJSONString(raw string) string {
	if raw == "" || raw == "null" {
		return ""
	}
	var value string
	if err := json.Unmarshal([]byte(raw), &value); err == nil {
		return value
	}
	return raw
}

type openAIStreamToolAccumulator struct {
	calls map[int]*ToolCall
}

func newOpenAIStreamToolAccumulator() *openAIStreamToolAccumulator {
	return &openAIStreamToolAccumulator{calls: make(map[int]*ToolCall)}
}

func (a *openAIStreamToolAccumulator) absorbSDK(delta []openai.ChatCompletionChunkChoiceDeltaToolCall) {
	for _, item := range delta {
		idx := int(item.Index)
		call := a.calls[idx]
		if call == nil {
			call = &ToolCall{Arguments: json.RawMessage("{}")}
			a.calls[idx] = call
		}
		if item.ID != "" {
			call.ID = item.ID
		}
		if item.Function.Name != "" {
			call.Name = item.Function.Name
		}
		if item.Function.Arguments != "" {
			call.Arguments = appendJSONFragment(call.Arguments, item.Function.Arguments)
		}
	}
}

func (a *openAIStreamToolAccumulator) result() []ToolCall {
	if len(a.calls) == 0 {
		return nil
	}
	max := -1
	for idx := range a.calls {
		if idx > max {
			max = idx
		}
	}
	out := make([]ToolCall, 0, len(a.calls))
	for i := 0; i <= max; i++ {
		if call := a.calls[i]; call != nil && call.Name != "" {
			if len(call.Arguments) == 0 {
				call.Arguments = json.RawMessage("{}")
			}
			out = append(out, *call)
		}
	}
	return out
}

func appendJSONFragment(existing json.RawMessage, fragment string) json.RawMessage {
	if len(existing) == 0 || string(existing) == "{}" {
		return json.RawMessage(fragment)
	}
	return json.RawMessage(string(existing) + fragment)
}