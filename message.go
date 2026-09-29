package llm

// Role says who wrote a Message.
type Role string

const (
	RoleSystem    Role = "system"    // instructions or context injected by the application
	RoleUser      Role = "user"      // the person talking to the model
	RoleAssistant Role = "assistant" // the model's own previous output
	RoleTool      Role = "tool"      // the result of running a tool the model asked for
)

// Message is one entry of the conversation sent to the model.
type Message struct {
	Role    Role
	Content string

	// ToolCalls is set only when Role == RoleAssistant and the model asked to run tools.
	ToolCalls []ToolCall

	// ToolCallID and ToolName are set only when Role == RoleTool: the ToolCall.ID this
	// result answers, and the tool that produced Content.
	ToolCallID string
	ToolName   string
}

// ToolCall is one tool invocation the model asked for.
type ToolCall struct {
	ID    string // correlates the call with the RoleTool message that answers it
	Name  string // matches ToolDef.Name
	Input string // raw JSON arguments, validated against ToolDef.InputSchema before running
}
