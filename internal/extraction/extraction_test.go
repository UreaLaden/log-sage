package extraction

import (
	"testing"

	"github.com/Urealaden/log-sage-temp/internal/normalize"
	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestExtractSignals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entries  []normalize.Line
		patterns []types.SignalPattern
		expected []types.PatternMatch
	}{
		{
			name: "no patterns returns empty signal set",
			entries: []normalize.Line{
				{Raw: "connection refused"},
			},
			expected: nil,
		},
		{
			name: "no entries returns empty signal set",
			patterns: []types.SignalPattern{
				{Name: "conn-refused", SignalType: "ConnectionRefused", MatchExpression: "connection refused"},
			},
			expected: nil,
		},
		{
			name: "single pattern matches one entry",
			entries: []normalize.Line{
				{Raw: "dial tcp: connection refused"},
			},
			patterns: []types.SignalPattern{
				{Name: "conn-refused", SignalType: "ConnectionRefused", MatchExpression: "connection refused"},
			},
			expected: []types.PatternMatch{
				{
					PatternName: "conn-refused",
					SignalType:  "ConnectionRefused",
					LineNumber:  1,
					Text:        "dial tcp: connection refused",
				},
			},
		},
		{
			name: "single pattern matches multiple entries",
			entries: []normalize.Line{
				{Raw: "redis: connection refused"},
				{Raw: "postgres: connection refused"},
			},
			patterns: []types.SignalPattern{
				{Name: "conn-refused", SignalType: "ConnectionRefused", MatchExpression: "connection refused"},
			},
			expected: []types.PatternMatch{
				{
					PatternName: "conn-refused",
					SignalType:  "ConnectionRefused",
					LineNumber:  1,
					Text:        "redis: connection refused",
				},
				{
					PatternName: "conn-refused",
					SignalType:  "ConnectionRefused",
					LineNumber:  2,
					Text:        "postgres: connection refused",
				},
			},
		},
		{
			name: "multiple patterns match one entry in pattern order",
			entries: []normalize.Line{
				{Raw: "panic: dial tcp: connection refused"},
			},
			patterns: []types.SignalPattern{
				{Name: "panic", SignalType: "Panic", MatchExpression: "panic:"},
				{Name: "conn-refused", SignalType: "ConnectionRefused", MatchExpression: "connection refused"},
			},
			expected: []types.PatternMatch{
				{
					PatternName: "panic",
					SignalType:  "Panic",
					LineNumber:  1,
					Text:        "panic: dial tcp: connection refused",
				},
				{
					PatternName: "conn-refused",
					SignalType:  "ConnectionRefused",
					LineNumber:  1,
					Text:        "panic: dial tcp: connection refused",
				},
			},
		},
		{
			name: "non matching patterns produce no matches",
			entries: []normalize.Line{
				{Raw: "service started successfully"},
			},
			patterns: []types.SignalPattern{
				{Name: "panic", SignalType: "Panic", MatchExpression: "panic:"},
			},
			expected: nil,
		},
		{
			name: "output order follows entry order then pattern order",
			entries: []normalize.Line{
				{Raw: "panic: crash"},
				{Raw: "connection refused and panic:"},
			},
			patterns: []types.SignalPattern{
				{Name: "conn-refused", SignalType: "ConnectionRefused", MatchExpression: "connection refused"},
				{Name: "panic", SignalType: "Panic", MatchExpression: "panic:"},
			},
			expected: []types.PatternMatch{
				{
					PatternName: "panic",
					SignalType:  "Panic",
					LineNumber:  1,
					Text:        "panic: crash",
				},
				{
					PatternName: "conn-refused",
					SignalType:  "ConnectionRefused",
					LineNumber:  2,
					Text:        "connection refused and panic:",
				},
				{
					PatternName: "panic",
					SignalType:  "Panic",
					LineNumber:  2,
					Text:        "connection refused and panic:",
				},
			},
		},
		{
			name: "source context is preserved in pattern match",
			entries: []normalize.Line{
				{Raw: "2026-03-14T19:00:00Z connection refused", Timestamp: "2026-03-14T19:00:00Z"},
			},
			patterns: []types.SignalPattern{
				{Name: "conn-refused", SignalType: "ConnectionRefused", MatchExpression: "connection refused"},
			},
			expected: []types.PatternMatch{
				{
					PatternName: "conn-refused",
					SignalType:  "ConnectionRefused",
					LineNumber:  1,
					Text:        "2026-03-14T19:00:00Z connection refused",
					Timestamp:   "2026-03-14T19:00:00Z",
				},
			},
		},
		{
			name: "blank match expressions do not create false positives",
			entries: []normalize.Line{
				{Raw: "panic: crash"},
			},
			patterns: []types.SignalPattern{
				{Name: "blank", SignalType: "Noise", MatchExpression: ""},
				{Name: "spaces", SignalType: "Noise", MatchExpression: "   "},
			},
			expected: nil,
		},
		{
			name: "empty entry content does not panic",
			entries: []normalize.Line{
				{Raw: ""},
			},
			patterns: []types.SignalPattern{
				{Name: "panic", SignalType: "Panic", MatchExpression: "panic:"},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ExtractSignals(tt.entries, tt.patterns)
			if len(got.Matches) != len(tt.expected) {
				t.Fatalf("match count = %d, want %d", len(got.Matches), len(tt.expected))
			}

			for i := range tt.expected {
				if got.Matches[i] != tt.expected[i] {
					t.Fatalf("match[%d] = %#v, want %#v", i, got.Matches[i], tt.expected[i])
				}
			}
		})
	}
}
