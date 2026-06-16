package inputui

import (
	"github.com/riipandi/elph/internal/command"
	"github.com/riipandi/elph/internal/mention"
)

// SuggestState tracks slash-command and @-mention palettes.
type SuggestState struct {
	CmdSuggestions      []command.SlashCommand
	CmdSuggestIndex     int
	ArgSuggestions      []command.ArgChoice
	ArgSuggestIndex     int
	MentionSuggestions  []mention.Entry
	MentionSuggestIndex int
	MentionIndex        []mention.Entry
	MentionIndexDir     string
	MentionIndexLoading bool
	MentionActiveQuery  string
	MentionFilterQuery  string
	MentionUserSelected bool
}
