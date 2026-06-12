package constants

type AgentMode string

const (
	ModeBuild AgentMode = "build" // Default mode ask for permissions
	ModePlan  AgentMode = "plan"  // Read only with some tools allowed
	ModeAsk   AgentMode = "ask"   // Read only without tools allowed (chat)
	ModeBrave AgentMode = "brave" // Skipped approval for tool calls (YOLO)
)
