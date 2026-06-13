package constants

import "testing"

func TestActivityForTool(t *testing.T) {
	tests := []struct {
		tool string
		want AgentActivity
	}{
		{ToolRead, ActivityReading},
		{ToolReadMediaFile, ActivityReading},
		{ToolWrite, ActivityWriting},
		{ToolEdit, ActivityWriting},
		{ToolGrep, ActivitySearching},
		{ToolGlob, ActivitySearching},
		{ToolCodeSearch, ActivitySearching},
		{ToolWebSearch, ActivitySearching},
		{ToolBash, ActivityRunning},
		{ToolFetchURL, ActivityFetching},
		{ToolEnterPlanMode, ActivityPlanning},
		{ToolExitPlanMode, ActivityPlanning},
		{ToolAskUser, ActivityWaiting},
		{"UnknownTool", ActivityWorking},
		{"", ActivityWorking},
	}

	for _, tc := range tests {
		if got := ActivityForTool(tc.tool); got != tc.want {
			t.Fatalf("ActivityForTool(%q) = %q, want %q", tc.tool, got, tc.want)
		}
	}
}

func TestAgentTurnPhasesOrder(t *testing.T) {
	want := []AgentActivity{
		ActivityConnecting,
		ActivityLoading,
		ActivityThinking,
		ActivitySearching,
		ActivityReading,
		ActivityWriting,
		ActivityRunning,
		ActivityStreaming,
	}
	if len(AgentTurnPhases) != len(want) {
		t.Fatalf("got %d phases, want %d", len(AgentTurnPhases), len(want))
	}
	for i, phase := range want {
		if AgentTurnPhases[i] != phase {
			t.Fatalf("phase[%d] = %q, want %q", i, AgentTurnPhases[i], phase)
		}
	}
}