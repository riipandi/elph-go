package toolinteract

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const (
	DefaultAskUserQuestion   = "The agent has a question for you."
	AskUserCustomPlaceholder = "Or type your own…"
)

// AskUserFields holds parsed ask-user dialog parameters.
type AskUserFields struct {
	Question    string
	Options     []string
	AllowCustom bool
}

var askUserQuotedStrings = regexp.MustCompile(`"([^"]+)"`)

// ParseAskUserArgs normalizes ask-user tool arguments into dialog fields.
func ParseAskUserArgs(args map[string]any) AskUserFields {
	var out AskUserFields
	if args == nil {
		out.Question = DefaultAskUserQuestion
		return out
	}

	out.Question = askUserQuestionText(args)
	out.Options = askUserOptions(args)
	out.Question, out.Options = reconcileSwappedAskUserFields(out.Question, out.Options, args)
	out.AllowCustom = askUserAllowCustom(args, len(out.Options) > 0)

	if out.Question != "" && len(out.Options) == 0 {
		if opts, ok := parseJSONStringArray(out.Question); ok {
			out.Options = opts
			out.Question = ""
		} else if opts := salvageQuotedStrings(out.Question); len(opts) > 0 && strings.HasPrefix(strings.TrimSpace(out.Question), "[") {
			out.Options = opts
			out.Question = ""
		}
	}

	if len(out.Options) == 0 {
		if raw, ok := args["question"]; ok {
			out.Options = askUserOptions(map[string]any{"options": raw})
			if len(out.Options) > 0 {
				out.Question = ""
			}
		}
	}

	if strings.TrimSpace(out.Question) == "" {
		if len(out.Options) > 0 {
			out.Question = "Choose an option:"
		} else {
			out.Question = DefaultAskUserQuestion
		}
	}
	return out
}

func askUserQuestionText(args map[string]any) string {
	raw, ok := args["question"]
	if !ok || raw == nil {
		if r, ok := StringArg(args, "reason"); ok {
			return r
		}
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case []any, []string:
		return ""
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		if strings.HasPrefix(s, "[") {
			return s
		}
		return s
	}
}

func reconcileSwappedAskUserFields(question string, options []string, args map[string]any) (string, []string) {
	if len(options) > 0 || strings.TrimSpace(question) == "" {
		return question, options
	}
	raw, ok := args["options"]
	if !ok || raw == nil {
		return question, options
	}
	optText, ok := raw.(string)
	if !ok {
		return question, options
	}
	optText = strings.TrimSpace(optText)
	if optText == "" || strings.HasPrefix(optText, "[") {
		return question, options
	}
	if !strings.HasPrefix(strings.TrimSpace(question), "[") {
		return question, options
	}
	opts := options
	if len(opts) == 0 {
		if parsed, ok := parseJSONStringArray(question); ok {
			opts = parsed
		} else {
			opts = salvageQuotedStrings(question)
		}
	}
	if len(opts) == 0 {
		return question, options
	}
	return optText, opts
}

func parseJSONStringArray(s string) ([]string, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") {
		return nil, false
	}
	var opts []string
	if err := json.Unmarshal([]byte(s), &opts); err != nil {
		return nil, false
	}
	opts = trimStrings(opts)
	if len(opts) == 0 {
		return nil, false
	}
	return opts, true
}

func salvageQuotedStrings(s string) []string {
	matches := askUserQuotedStrings.FindAllStringSubmatch(s, -1)
	out := make([]string, 0, len(matches)+1)
	for _, m := range matches {
		if len(m) > 1 {
			out = append(out, strings.TrimSpace(m[1]))
		}
	}
	if tail := salvageTrailingArrayToken(s); tail != "" {
		out = append(out, tail)
	}
	return trimStrings(out)
}

func salvageTrailingArrayToken(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") {
		return ""
	}
	i := strings.LastIndex(s, ",")
	if i < 0 {
		return ""
	}
	tail := strings.TrimSpace(s[i+1:])
	if strings.HasSuffix(tail, `"`) {
		return ""
	}
	tail = strings.TrimPrefix(tail, `"`)
	tail = strings.TrimSuffix(tail, "]")
	return strings.TrimSpace(tail)
}

func askUserAllowCustom(args map[string]any, hasOptions bool) bool {
	if !hasOptions {
		return false
	}
	raw, ok := args["allowCustom"]
	if !ok || raw == nil {
		return true
	}
	switch v := raw.(type) {
	case bool:
		return v
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		return s != "false" && s != "0" && s != "no"
	default:
		return true
	}
}

func askUserOptions(args map[string]any) []string {
	raw, ok := args["options"]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return trimStrings(v)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return trimStrings(out)
	case string:
		if opts, ok := parseJSONStringArray(v); ok {
			return opts
		}
		return nil
	default:
		return nil
	}
}

func trimStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ResolveAskUserAnswer picks the submitted answer from form field values.
func ResolveAskUserAnswer(custom, choice, fallback string) string {
	if strings.TrimSpace(custom) != "" {
		return strings.TrimSpace(custom)
	}
	if strings.TrimSpace(choice) != "" {
		return strings.TrimSpace(choice)
	}
	return strings.TrimSpace(fallback)
}

// WrapAskUserQuestion word-wraps a question for the dialog width.
func WrapAskUserQuestion(question string, width int) string {
	question = strings.TrimSpace(question)
	if question == "" || width <= 0 {
		return question
	}
	return ansi.Wordwrap(question, width, "")
}