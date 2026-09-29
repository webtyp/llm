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
