package scoring

import "github.com/Urealaden/log-sage-temp/pkg/types"

// BuildCandidates converts a flat slice of detected hypotheses into
// CandidateHypothesis values enriched with scoring context.
// The input order is preserved. signals is the SignalSet that was used
// to produce hypotheses and is stored on each candidate for use by
// later scoring passes.
//
// Each candidate owns its Evidence and Signals slices — mutations to the
// original hypotheses or signals after this call do not affect the returned
// candidates.
func BuildCandidates(hypotheses []types.Hypothesis, signals types.SignalSet) []CandidateHypothesis {
	candidates := make([]CandidateHypothesis, 0, len(hypotheses))
	for _, hypothesis := range hypotheses {
		candidate := CandidateHypothesis{
			Class:     types.IssueClass{Name: hypothesis.IssueClass},
			BaseScore: hypothesis.Score,
			Evidence:  cloneEvidence(hypothesis.Evidence),
			Signals:   cloneSignalSet(signals),
			IsSymptom: false,
		}
		candidate.Phase = InferPhase(candidate)
		candidates = append(candidates, candidate)
	}

	return candidates
}

func cloneEvidence(evidence []types.Evidence) []types.Evidence {
	if evidence == nil {
		return nil
	}

	cloned := make([]types.Evidence, len(evidence))
	for i, item := range evidence {
		cloned[i] = item
		if item.Examples != nil {
			examples := make([]string, len(item.Examples))
			copy(examples, item.Examples)
			cloned[i].Examples = examples
		}
	}

	return cloned
}

func cloneSignalSet(signals types.SignalSet) types.SignalSet {
	if signals.Matches == nil {
		return types.SignalSet{}
	}

	clonedMatches := make([]types.PatternMatch, len(signals.Matches))
	copy(clonedMatches, signals.Matches)

	return types.SignalSet{
		Matches: clonedMatches,
	}
}
