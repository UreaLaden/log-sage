package scoring

import "github.com/Urealaden/log-sage-temp/pkg/types"

// BuildCandidates converts a flat slice of detected hypotheses into
// CandidateHypothesis values enriched with scoring context.
// The input order is preserved. signals is the SignalSet that was used
// to produce hypotheses and is stored on each candidate for use by
// later scoring passes.
func BuildCandidates(hypotheses []types.Hypothesis, signals types.SignalSet) []CandidateHypothesis {
	candidates := make([]CandidateHypothesis, 0, len(hypotheses))
	for _, hypothesis := range hypotheses {
		candidate := CandidateHypothesis{
			Class:     types.IssueClass{Name: hypothesis.IssueClass},
			BaseScore: hypothesis.Score,
			Evidence:  hypothesis.Evidence,
			Signals:   signals,
			IsSymptom: false,
		}
		candidate.Phase = InferPhase(candidate)
		candidates = append(candidates, candidate)
	}

	return candidates
}
