package constants

import "testing"

func TestMessageStyleKindsDiffer(t *testing.T) {
	user := MessageStyle(MessageUser).GetForeground()
	tool := MessageStyle(MessageTool).GetForeground()
	if user == tool {
		t.Fatal("user and tool foreground colors should differ")
	}
}

func TestThinkingUsesDimText(t *testing.T) {
	if MessageStyle(MessageThinking).GetForeground() != DimText {
		t.Fatal("thinking messages should use dim gray foreground")
	}
}