package scoring

// symptomOf maps an issue class name to the set of root-cause class names
// that make it a symptom. A candidate is marked IsSymptom only when at least
// one of its root-cause classes is also present in the candidate slice.
var symptomOf = map[string][]string{
	"CrashLoopBackOff":  {"OutOfMemory", "MissingEnvVar", "PermissionDenied", "DiskFull", "Panic"},
	"DNSFailure":        {"ConnectionRefused"},
	"TLSFailure":        {"ConnectionRefused", "DNSFailure"},
	"DependencyTimeout": {"ConnectionRefused", "DNSFailure"},
	"PermissionDenied":  {"DiskFull"},
}

// ApplyRelationships marks candidates as symptoms when their root-cause
// class is also present in the candidate slice.
// It returns a new slice; the input is not mutated.
func ApplyRelationships(candidates []CandidateHypothesis) []CandidateHypothesis {
	present := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		present[candidate.Class.Name] = struct{}{}
	}

	out := make([]CandidateHypothesis, 0, len(candidates))
	for _, candidate := range candidates {
		updated := candidate
		roots, ok := symptomOf[candidate.Class.Name]
		if ok {
			for _, root := range roots {
				if _, exists := present[root]; exists {
					updated.IsSymptom = true
					break
				}
			}
		}
		out = append(out, updated)
	}

	return out
}
