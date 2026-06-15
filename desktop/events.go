package desktop

// Wails event name constants — agent callbacks and chat lifecycle
const (
	// Agent streaming events
	EventDelta      = "agent:delta"
	EventThinking   = "agent:thinking"
	EventToolCall   = "agent:toolcall"
	EventToolResult = "agent:toolresult"
	EventError      = "agent:error"
	EventUsage      = "agent:usage"
	EventCompressing = "agent:compressing"
	EventCompressDone = "agent:compress_done"

	// Agent planning events
	EventPlanning       = "agent:planning"
	EventPlanGenerated  = "agent:plan:generated"
	EventPlanStepStart  = "agent:plan:stepstart"
	EventPlanStepDone   = "agent:plan:stepdone"

	// Chat lifecycle events
	EventChatStart     = "chat:start"
	EventChatDone      = "chat:done"
	EventChatError     = "chat:error"
	EventChatCancelled = "chat:cancelled"

	// Safety confirmation event
	EventSafetyConfirm = "safety:confirm"

	// Actor streaming events
	EventActorDelta = "actor:delta"
)

