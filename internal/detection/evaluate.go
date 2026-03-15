package detection

import "github.com/Urealaden/log-sage-temp/pkg/types"

// EvaluateRegistry runs the registered issue classes against the provided
// SignalSet and returns collected hypotheses in registry order.
func EvaluateRegistry(signals types.SignalSet) []types.Hypothesis {
	results := make([]types.Hypothesis, 0)

	for _, class := range IssueRegistry {
		hypotheses := (DefaultDetector{Class: class}).Detect(signals)
		results = append(results, hypotheses...)
	}

	return results
}
