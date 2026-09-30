package llm_test

import (
	"webtyp.com/context"
	"webtyp.com/fmt"
	"webtyp.com/llm"
)

// fixedDecider stands in for a model runtime: it always prefers the last option.
type fixedDecider struct{}

func (fixedDecider) Decide(ctx *context.Context, q llm.Question) (llm.Decision, error) {
	probs := make([]float64, len(q.Options))
	probs[len(probs)-1] = 1
	return llm.Decision{Choice: len(q.Options) - 1, Confidence: 1, Probs: probs}, nil
}

// A critic asks a closed question and acts on the chosen option, without parsing prose.
func Example_decider() {
	var critic llm.Decider = fixedDecider{}
	d, err := critic.Decide(context.Background(), llm.Question{
		Context: "User asked: ¿Hasta qué hora atendemos hoy?\nTool list_business_hours returned: 08:00-18:00\nAssistant answered: Hasta las 18:00.",
		Text:    "Does the answer use only what the tools returned?",
		Options: []string{"no", "yes"},
	})
	if err != nil {
		return
	}
	fmt.Println([]string{"no", "yes"}[d.Choice], d.Confidence)
	// Output: yes 1
}
