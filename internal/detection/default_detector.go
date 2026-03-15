package detection

import "github.com/Urealaden/log-sage-temp/pkg/types"

// DefaultDetector wraps an IssueClass and implements Detector using the
// class's primary and corroborating signal patterns.
type DefaultDetector struct {
	Class types.IssueClass
}

// Detect evaluates the provided SignalSet and returns a single hypothesis when
// at least one primary signal for the wrapped issue class is present.
func (d DefaultDetector) Detect(signals types.SignalSet) []types.Hypothesis {
	if len(signals.Matches) == 0 {
		return nil
	}

	score := 0.0
	evidence := make([]types.Evidence, 0)
	hasPrimaryMatch := false

	for _, pattern := range d.Class.PrimarySignals {
		occurrences := 0
		examples := make([]string, 0, 3)

		for _, match := range signals.Matches {
			if match.PatternName != pattern.Name {
				continue
			}

			occurrences++
			if len(examples) < 3 {
				examples = append(examples, match.Text)
			}
		}

		if occurrences == 0 {
			continue
		}

		hasPrimaryMatch = true
		score += pattern.Weight
		evidence = append(evidence, types.Evidence{
			Signal:      pattern.Name,
			Occurrences: occurrences,
			Examples:    examples,
		})
	}

	if !hasPrimaryMatch {
		return nil
	}

	for _, pattern := range d.Class.CorroboratingSignals {
		if hasPatternMatch(signals.Matches, pattern.Name) {
			score += pattern.Weight
		}
	}

	confidence, ok := confidenceFromScore(score)
	if !ok {
		return nil
	}

	return []types.Hypothesis{
		{
			IssueClass:  d.Class.Name,
			Confidence:  confidence,
			Score:       score,
			Evidence:    evidence,
			Explanation: d.Class.ExplanationTemplate,
		},
	}
}

func hasPatternMatch(matches []types.PatternMatch, patternName string) bool {
	for _, match := range matches {
		if match.PatternName == patternName {
			return true
		}
	}

	return false
}

func confidenceFromScore(score float64) (types.ConfidenceLevel, bool) {
	switch {
	case score >= 0.80:
		return types.ConfidenceHigh, true
	case score >= 0.50:
		return types.ConfidenceMedium, true
	case score >= 0.25:
		return types.ConfidenceLow, true
	default:
		return "", false
	}
}
