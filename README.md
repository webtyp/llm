# llm
<img src="docs/img/badges.svg">

The contract between code that asks a language model something and the runtime that answers:
`Client`, `Streamer`, `TokenCounter`, and the message, request and response types that cross between them.
It is a contract only. There is no model in this repository, so importing it adds no model
code to your binary.

## Getting started

**I want to call a model.** Take an `llm.Client` from your composition root and send it an
`llm.Request`:

```go
resp, err := client.Generate(ctx, llm.Request{
	System:          "You are the receptionist of a clinic.",
	Messages:        []llm.Message{{Role: llm.RoleUser, Content: "¿A qué hora abren?"}},
	Tools:           tools,
	MaxOutputTokens: 256,
})
```

**I want to answer a tool call.** Append the assistant message with its `ToolCalls`, then
one message per call:
`llm.Message{Role: llm.RoleTool, ToolCallID: call.ID, ToolName: call.Name, Content: output}`.

**I want to budget a prompt.** Ask for an `llm.TokenCounter`, which is the tokenizer of the
model that will read the prompt.

**I want to implement a model runtime.** Implement `Client` and `TokenCounter`, plus `Streamer` if it can stream. Render
`Request` into your model's chat template inside `Generate`.

| I want to… | Use |
|---|---|
| ask a model | `Client.Generate` |
| count tokens for a budget | `TokenCounter.CountTokens` |
| receive the answer while it is generated | `Streamer.GenerateStream` (version 2 voice loop) |
| offer tools | `Request.Tools` (`ToolDef`) |
| know why the model stopped | `Response.StopReason` (`StopEndTurn`, `StopToolUse`, `StopMaxTokens`) |

## Documentation

- [Architecture](docs/ARCHITECTURE.md): what the contract is, who implements it, and where the model runs.
- [Small browser LLMs](docs/SMALL_LLM_BROWSER_MODEL.md): candidate models under 1B parameters (research).
- [Efficient SLMs](docs/EFFICIENT_SLM.md): background on small models and quantization (research).
- [Agent guide](AGENTS.md): rules for anyone changing this library.
