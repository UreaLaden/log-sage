package extraction

import (
	"strings"

	"github.com/Urealaden/log-sage-temp/internal/normalize"
	"github.com/Urealaden/log-sage-temp/pkg/types"
)

// ExtractSignals applies the provided patterns to normalized log lines using
// literal substring matching and returns matches in stable discovery order.
func ExtractSignals(entries []normalize.Line, patterns []types.SignalPattern) types.SignalSet {
	if len(entries) == 0 || len(patterns) == 0 {
		return types.SignalSet{}
	}

	matches := make([]types.PatternMatch, 0)

	for lineNumber, entry := range entries {
		for _, pattern := range patterns {
			if strings.TrimSpace(pattern.MatchExpression) == "" {
				continue
			}

			if !strings.Contains(entry.Raw, pattern.MatchExpression) {
				continue
			}

			matches = append(matches, types.PatternMatch{
				PatternName: pattern.Name,
				SignalType:  pattern.SignalType,
				LineNumber:  lineNumber + 1,
				Text:        entry.Raw,
				Timestamp:   entry.Timestamp,
			})
		}
	}

	return types.SignalSet{
		Matches: matches,
	}
}
