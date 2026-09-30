package llm

import "webtyp.com/context"

// Decider answers a closed question with a probability for every option, instead of free
// text. It is how a program asks a model for a judgment it can act on without parsing prose:
// a critic's "does this answer use only what the tools returned?", a router's "which tool fits?".
// A runtime implements it by reading the model's probabilities for the option letters in one
// forward pass, the technique documented in webtyp/agenteval docs/JUDGE.md.
type Decider interface {
	Decide(ctx *context.Context, q Question) (Decision, error)
}

// Question is a closed question about a text.
type Question struct {
	Context string   // what the model reads before the question
	Text    string   // the question itself
	Options []string // 2 to 10 possible answers; yes/no questions use exactly {"no", "yes"}
}

// Decision is the answer to a Question.
type Decision struct {
	Choice     int       // index into Question.Options of the most probable option
	Confidence float64   // the probability of Choice
	Probs      []float64 // one per option, in the order given; they sum to 1
}
