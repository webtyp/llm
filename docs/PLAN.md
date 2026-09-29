---
PLAN: "feat: llm contract — Client, Streamer, TokenCounter and the message/request/response types"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 7069310062132861641
PR: https://github.com/webtyp/llm/pull/1
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Phase 1 (gate)** of
> [`AGENT_ECOSYSTEM_MASTER_PLAN.md`](https://github.com/webtyp/agent/blob/main/docs/AGENT_ECOSYSTEM_MASTER_PLAN.md).
> `webtyp/agentcontext` (phase 2) and `webtyp/agent` (phase 3) wait for this tag.

# Plan — `webtyp.com/llm`: the contract between the agent and a language model

## 0. Context

`webtyp/agent` is an orchestrator: it loops "ask the model → run the tools the model asked
for → ask again". Today the types that cross the agent ↔ model boundary (`LLMClient`,
`LLMRequest`, `LLMResponse`, `Message`, `ToolDef`, `ToolCall`) are declared **inside**
`agent/types.go` and `agent/interfaces.go`. That has two consequences:

1. Anything that implements a model (the future in-browser Go runtime) or builds a request
   (the context compiler, `webtyp/agentcontext`) must import the whole orchestrator.
2. `agentcontext` cannot exist as its own library: it would import `agent`, and `agent`
   imports it — a cycle.

This repository extracts that boundary into a **contract-only** library: interfaces and value
types, no implementation, no dependency heavier than `webtyp.com/context`. Implementations
(a Go-native model runtime running in a Web Worker) live in their own repositories.

This plan also fixes two defects found while extracting the types:

- `LLMRequest.MaxTokens` was filled with the **context window size** (8192) by
  `agent/context_window.go`, while every provider reads that field as the **output** limit.
  The field is renamed `MaxOutputTokens` so the ambiguity that caused the bug cannot be written.
- `LLMResponse.TokensUsed` (prompt + completion) was stored as the token count of a
  summary, which only depends on the completion. `Usage` now reports the two separately.

## Development rules (inline — do not assume you will read anything else)

- **This is a contract library.** It declares interfaces and value types only. No model
  runtime, no HTTP client, no mock in non-test files. A second concern is a new repository.
- **Every file compiles for the browser** under `GOOS=js GOARCH=wasm` and TinyGo.
- **Never import:**

| Never | Use instead | Why |
|---|---|---|
| `context` (stdlib) | `webtyp.com/context` | every webtyp API takes `*context.Context` |
| `fmt`, `errors`, `strings`, `strconv` | `webtyp.com/fmt` | isomorphic, small under TinyGo |
| `encoding/json` | nothing (this repo does no JSON) | reflection JSON costs ~1 MB of wasm |
| `map[K]V` | a slice | TinyGo's map runtime is a size tax |
| `os`, `log`, `net/http` | nothing | a contract never touches the environment |

- **No hardcoded strings in logic** — the role and stop-reason values are typed constants.
- Flat layout: Go files in the repository root, max 500 lines per file.
- Tests use the standard library only (`testing`). No `testify`.
- Do **not** run `gopush` or `codejob`.

## Design gate (api-design — five answers)

1. **Prior art.**
   - **OpenAI Chat Completions**: a list of role-tagged messages; the assistant message carries
     `tool_calls`, and each tool result is a `role:"tool"` message with `tool_call_id`.
   - **Anthropic Messages**: the system prompt is a separate field, not a message;
     `stop_reason` is `end_turn` / `tool_use` / `max_tokens`; `usage` reports
     `input_tokens` and `output_tokens` separately; a dedicated token-counting endpoint.
   - **Google Gemini / Genkit**: `generateContent` with `maxOutputTokens`, and a
     `countTokens` call separate from generation.
   - **langchaingo** (`llms.Model.GenerateContent`): one Go interface per model, one method.

   This contract takes Anthropic's separate `System` field and stop reasons, OpenAI's tool
   message shape, and Gemini's `MaxOutputTokens` name. Streaming follows the Gemini Go SDK
   (`GenerateContentStream`), where generating as it goes is a separate call, not a flag on
   the request. Here it is a separate interface (`Streamer`), so a runtime that cannot stream
   is not forced to fake it. Mapping to any of them, or to a local
   chat template, is then a straight field copy. It differs from all four in one way: token
   counting is its own interface (`TokenCounter`), because the in-browser runtime counts
   tokens locally with its own tokenizer, with no network round trip.
2. **Novice-name test.** `llm.Client.Generate(ctx, llm.Request)` reads as "ask the language
   model to generate an answer to this request". `llm.TokenCounter.CountTokens(text)` reads
   as itself. `llm.Streamer.GenerateStream(ctx, req, onText)` means "generate, and hand me
   the text as it comes". `MaxOutputTokens` says what it limits; the old `MaxTokens` did not, and that
   ambiguity caused a real bug (§0). Roles and stop reasons are typed constants
   (`llm.RoleTool`, `llm.StopToolUse`), not strings the caller must spell right.
3. **Complexity ledger.**
   ```
   Concepts the developer must learn   +2 (TokenCounter, Streamer) / −0   (the rest moves, not added)
   Files they must touch to do X       +0 / −0
   Lines at the call site              +0 / −0   (llm.Request instead of agent.LLMRequest)
   Ways to do the same thing           +0 / −1   after phase 3 deletes the copies in agent
   ```
4. **Where it belongs.** The boundary between "something that decides what to ask" (agent,
   agentcontext) and "something that answers" (a model runtime) is a contract owned by
   neither side — so it is its own repository, the same pattern as `webtyp/storage` between
   `orm` and the database backends. It holds no implementation, so no consumer pulls model
   code into its binary by importing it.
5. **What it deletes.** Nothing in this repository (it is new). Phase 3 deletes from
   `webtyp/agent`: `LLMClient`, `LLMRequest`, `LLMResponse`, `ToolDef`, `ToolCall` and the
   LLM-facing fields of `agent.Message`.

## Stage 1 — the types

Create exactly these files. Copy the doc comments; they are the documentation.

**`message.go`**

```go
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
```

**`request.go`**

```go
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
```

**`response.go`**

```go
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
```

**`llm.go`**

```go
// Package llm is the contract between code that asks a language model something and the
// runtime that answers. It declares interfaces and value types only; implementations live
// in their own repositories.
package llm

import "webtyp.com/context"

// Client generates one answer for one request.
type Client interface {
	Generate(ctx *context.Context, req Request) (Response, error)
}

// Streamer generates like Client, and also hands over the answer's text while the model
// produces it. onText receives each new piece in order; the pieces concatenated equal the
// returned Response.Text. The returned Response is the same one Generate would return
// (tool calls and usage arrive only there, at the end). No consumer uses it in version 1: it
// exists so the version-2 voice loop can start speaking before the answer is complete.
type Streamer interface {
	GenerateStream(ctx *context.Context, req Request, onText func(text string)) (Response, error)
}

// TokenCounter counts tokens the way one specific model does. Budgets are only right when
// they are counted with the tokenizer of the model that will read the request, so the
// runtime that implements Client for a model also implements TokenCounter for it.
type TokenCounter interface {
	CountTokens(text string) int
}
```

`go.mod` requires only `webtyp.com/context` (use `go get webtyp.com/context@latest`).

**Acceptance:**
- `grep -rn '"context"\|"fmt"\|"errors"\|"strings"\|encoding/json\|map\[' --include=*.go .` → empty.
- `go vet ./...`, `GOOS=js GOARCH=wasm go build ./...`, `gotest`, `gotest -tinygo` pass.

## Stage 2 — consumer-shaped test

**File:** `example_test.go`, package `llm_test`.

It proves the types carry one complete tool round trip the way the orchestrator uses them:

1. Declare, in the test file, `type scriptedClient struct { replies []llm.Response; seen []llm.Request }`
   whose `Generate` appends the request to `seen` and returns the next reply.
2. Script two replies: first `{StopReason: llm.StopToolUse, ToolCalls: []llm.ToolCall{{ID: "c1", Name: "clinic_hours", Input: "{}"}}}`,
   then `{StopReason: llm.StopEndTurn, Text: "Abrimos de 8 a 20"}`.
3. Write `func Example_toolRoundTrip()` that runs the loop: generate → on `StopToolUse`
   append the assistant message with its `ToolCalls` plus one `RoleTool` message with
   `ToolCallID: "c1"`, `ToolName: "clinic_hours"`, `Content: "8-20"` → generate again →
   print the final text. `// Output: Abrimos de 8 a 20`.
4. `TestRoundTrip_SecondRequestCarriesToolResult`: after the loop, `seen[1].Messages` has
   length 3 (user, assistant, tool) and the last one has `Role == llm.RoleTool` and
   `ToolCallID == "c1"`.
5. `TestStreamer_PiecesConcatenateToText`: in the test file, a `chunkedStreamer` whose
   `GenerateStream` calls `onText` with `"Abri"`, `"mos a "`, `"las 8"` and returns
   `Response{Text: "Abrimos a las 8", StopReason: llm.StopEndTurn}`. Assert that the
   concatenated pieces equal `Response.Text`, and that `var _ llm.Streamer = chunkedStreamer{}` compiles.

## Stages

| Stage | Files | Acceptance |
|---|---|---|
| 1 | `message.go`, `request.go`, `response.go`, `llm.go`, `go.mod` | builds for host, wasm and TinyGo; forbidden-import grep empty |
| 2 | `example_test.go` | example output matches; round-trip and streamer tests pass |
