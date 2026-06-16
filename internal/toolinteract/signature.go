package toolinteract

import (
	"sort"
	"strings"

	"github.com/riipandi/elph/internal/runtime/toolresult"
	"github.com/riipandi/elph/pkg/core/agent"
)

// ToolCallSignature builds a stable key for tool-call deduplication.
func ToolCallSignature(call agent.ParsedToolCall) string {
	presentation := toolresult.ResolveToolRequest(call.Name, call.Parameters)
	keys := make([]string, 0, len(call.Parameters))
	for key := range call.Parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(presentation.Name)
	for _, key := range keys {
		b.WriteByte('\x00')
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(call.Parameters[key])
	}
	return b.String()
}
