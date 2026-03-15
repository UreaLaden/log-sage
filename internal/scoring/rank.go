package scoring

import "sort"

// Rank returns a new slice of candidates sorted into deterministic output order.
// Non-symptom candidates appear before symptoms. Within each group candidates
// are sorted by BaseScore descending, then by Class.Name ascending as a
// tiebreaker. The input slice is not mutated.
func Rank(candidates []CandidateHypothesis) []CandidateHypothesis {
	ranked := make([]CandidateHypothesis, len(candidates))
	copy(ranked, candidates)

	sort.SliceStable(ranked, func(i, j int) bool {
		ci, cj := ranked[i], ranked[j]

		if ci.IsSymptom != cj.IsSymptom {
			return !ci.IsSymptom
		}

		if ci.BaseScore != cj.BaseScore {
			return ci.BaseScore > cj.BaseScore
		}

		return ci.Class.Name < cj.Class.Name
	})

	return ranked
}
