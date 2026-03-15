package scoring

import "github.com/Urealaden/log-sage-temp/pkg/types"

// MapConfidence maps each candidate's BaseScore to a ConfidenceLevel and
// discards candidates below the minimum threshold.
// It returns a new slice in the same order; the input is not mutated.
func MapConfidence(candidates []CandidateHypothesis) []CandidateHypothesis {
	out := make([]CandidateHypothesis, 0, len(candidates))

	for _, c := range candidates {
		candidate := c

		switch {
		case candidate.BaseScore >= 0.80:
			candidate.Confidence = types.ConfidenceHigh
		case candidate.BaseScore >= 0.50:
			candidate.Confidence = types.ConfidenceMedium
		case candidate.BaseScore >= 0.25:
			candidate.Confidence = types.ConfidenceLow
		default:
			continue
		}

		out = append(out, candidate)
	}

	return out
}
