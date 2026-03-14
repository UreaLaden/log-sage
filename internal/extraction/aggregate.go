package extraction

import "github.com/Urealaden/log-sage-temp/pkg/types"

// SignalSummary is the aggregated view of a SignalSet produced by
// AggregateSignals.
type SignalSummary struct {
	TotalMatches   int
	CountByType    map[string]int
	PatternsByType map[string][]string
}

// AggregateSignals condenses a SignalSet into a summary grouped by signal type.
func AggregateSignals(ss types.SignalSet) SignalSummary {
	summary := SignalSummary{
		TotalMatches:   len(ss.Matches),
		CountByType:    make(map[string]int),
		PatternsByType: make(map[string][]string),
	}

	seenPatternsByType := make(map[string]map[string]struct{})

	for _, match := range ss.Matches {
		summary.CountByType[match.SignalType]++

		seenPatterns, ok := seenPatternsByType[match.SignalType]
		if !ok {
			seenPatterns = make(map[string]struct{})
			seenPatternsByType[match.SignalType] = seenPatterns
		}

		if _, ok := seenPatterns[match.PatternName]; ok {
			continue
		}

		summary.PatternsByType[match.SignalType] = append(
			summary.PatternsByType[match.SignalType],
			match.PatternName,
		)
		seenPatterns[match.PatternName] = struct{}{}
	}

	return summary
}
