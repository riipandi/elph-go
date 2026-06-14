package provider

// ProviderProgressPhase identifies which provider command phase is running.
type ProviderProgressPhase string

const (
	ProviderProgressConnect ProviderProgressPhase = "connect"
	ProviderProgressSync    ProviderProgressPhase = "sync"
)

// ProviderProgressAction describes what happened to one provider.
type ProviderProgressAction string

const (
	ProviderProgressWorking   ProviderProgressAction = "working"
	ProviderProgressCreated   ProviderProgressAction = "created"
	ProviderProgressBackfill  ProviderProgressAction = "backfilled"
	ProviderProgressUnchanged ProviderProgressAction = "unchanged"
	ProviderProgressSynced    ProviderProgressAction = "synced"
	ProviderProgressSkipped   ProviderProgressAction = "skipped"
	ProviderProgressFetchMeta ProviderProgressAction = "fetch_metadata"
)

// ProviderProgressEvent is one progress tick for CLI/TUI reporting.
type ProviderProgressEvent struct {
	Phase      ProviderProgressPhase
	ProviderID string
	Label      string
	Index      int
	Total      int
	Action     ProviderProgressAction
	Detail     string
}

// ProviderProgressReporter receives per-provider progress updates.
type ProviderProgressReporter func(ProviderProgressEvent)

func reportProviderProgress(reporter ProviderProgressReporter, evt ProviderProgressEvent) {
	if reporter == nil {
		return
	}
	reporter(evt)
}
