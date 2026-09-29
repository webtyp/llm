# Agent Guide — `webtyp/llm`

Constraints for agents working on this library. **Read this before any change.**
The current work order, when one exists, is [docs/PLAN.md](docs/PLAN.md). The ecosystem plan
is [`agent/docs/AGENT_ECOSYSTEM_MASTER_PLAN.md`](https://github.com/webtyp/agent/blob/main/docs/AGENT_ECOSYSTEM_MASTER_PLAN.md).

---

## What this library is

The contract between the orchestrator and a language model. It contains **interfaces and value
types only**. It must never contain:

- a model runtime, a tokenizer, weights, or an HTTP client (each belongs in its own repository),
- a mock or fake in a non-test file,
- any dependency heavier than `webtyp.com/context`.

If a change needs any of those, it belongs in another repository.

---

## The builds that define "done"

```bash
go vet ./...
gotest
gotest -tinygo
GOOS=js GOARCH=wasm go build ./...
```

---

## Never import these

| Never | Use instead | Why |
|---|---|---|
| `context` (stdlib) | `webtyp.com/context` | every webtyp API takes `*context.Context` |
| `fmt`, `errors`, `strings`, `strconv` | `webtyp.com/fmt` | isomorphic, small under TinyGo |
| `encoding/json` | nothing, since this repo does no JSON | reflection-based JSON costs ~1 MB of wasm |
| `time` | `webtyp.com/time` | |
| `map[K]V` | a slice | TinyGo's map runtime is a size tax on every binary |
| `os`, `log`, `net/http` | nothing | a contract never touches the environment |

---

## Common mistakes to avoid

- Adding a field "for one provider". A field only one runtime understands is a leak. Map it
  inside that runtime instead.
- Using a string where a typed constant exists (`"tool"` instead of `llm.RoleTool`).
- Reusing `MaxOutputTokens` as the context size. They are different limits. The context
  budget belongs to `webtyp/agentcontext`.
- Publishing with anything other than `gopush`.
