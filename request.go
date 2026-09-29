package llm

// ToolDef describes a tool the model may ask to run.
type ToolDef struct {
	Name        string // unique identifier
	Description string // what the tool does — this is what the model reads to decide
	InputSchema string // JSON Schema of the arguments
}

// Request is everything the model sees for one generation.
type Request struct {
	System          string    // instructions that stay the same across the conversation
	Messages        []Message // the conversation, oldest first
	Tools           []ToolDef // the tools the model may ask to run in this generation
	MaxOutputTokens int       // the most tokens the model may generate for this answer
}
