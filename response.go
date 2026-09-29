package llm

// StopReason says why the model stopped generating.
type StopReason string

const (
	StopEndTurn   StopReason = "end_turn"   // the model finished its answer
	StopToolUse   StopReason = "tool_use"   // the model asked to run tools: see Response.ToolCalls
	StopMaxTokens StopReason = "max_tokens" // the answer was cut at Request.MaxOutputTokens
)

// Usage is the token cost of one generation.
type Usage struct {
	InputTokens  int // tokens of the request the model read
	OutputTokens int // tokens the model generated
}

// Response is the model's answer to one Request.
type Response struct {
	Text       string     // the generated text; may be empty when StopReason == StopToolUse
	StopReason StopReason
	ToolCalls  []ToolCall // set only when StopReason == StopToolUse
	Usage      Usage
}
