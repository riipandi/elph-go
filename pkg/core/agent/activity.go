package agent

import "github.com/riipandi/elph/pkg/tools"

// Activity describes what the agent is doing during a turn.
type Activity string

const (
	ActivityIdle       Activity = ""
	ActivityConnecting Activity = "Connecting"
	ActivityLoading    Activity = "Loading"
	ActivityThinking   Activity = "Thinking"
	ActivitySearching  Activity = "Searching"
	ActivityReading    Activity = "Reading"
	ActivityWriting    Activity = "Writing"
	ActivityRunning    Activity = "Running"
	ActivityFetching   Activity = "Fetching"
	ActivityStreaming  Activity = "Streaming"
	ActivityPlanning   Activity = "Planning"
	ActivityWaiting    Activity = "Waiting"
	ActivityWorking    Activity = "Working"
)

// TurnPhases is the default ordered progression shown while a turn runs.
var TurnPhases = []Activity{
	ActivityConnecting,
	ActivityLoading,
	ActivityThinking,
	ActivitySearching,
	ActivityReading,
	ActivityWriting,
	ActivityRunning,
	ActivityStreaming,
}

var toolActivity = map[string]Activity{
	tools.Read:          ActivityReading,
	tools.ReadMediaFile: ActivityReading,
	tools.Write:         ActivityWriting,
	tools.Edit:          ActivityWriting,
	tools.Grep:          ActivitySearching,
	tools.Glob:          ActivitySearching,
	tools.CodeSearch:    ActivitySearching,
	tools.WebSearch:     ActivitySearching,
	tools.Bash:          ActivityRunning,
	tools.FetchURL:      ActivityFetching,
	tools.EnterPlanMode: ActivityPlanning,
	tools.ExitPlanMode:  ActivityPlanning,
	tools.AskUser:       ActivityWaiting,
	tools.Skill:         ActivityLoading,
	tools.TodoList:      ActivityWorking,
}

// ActivityForTool returns the indicator label for a tool call.
// Unknown tools fall back to ActivityWorking.
func ActivityForTool(tool string) Activity {
	if activity, ok := toolActivity[tool]; ok {
		return activity
	}
	return ActivityWorking
}
