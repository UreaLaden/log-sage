package scoring

// ApplyCorroboration boosts candidate BaseScore when multiple evidence items
// support the same hypothesis. It returns a new slice without mutating the
// input slice.
func ApplyCorroboration(candidates []CandidateHypothesis) []CandidateHypothesis {
	result := make([]CandidateHypothesis, len(candidates))
	copy(result, candidates)

	for i, candidate := range result {
		if len(candidate.Evidence) <= 1 {
			continue
		}

		candidate.BaseScore += float64(len(candidate.Evidence)-1) * 0.25
		result[i] = candidate
	}

	return result
}
