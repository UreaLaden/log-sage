package extraction

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestAggregateSignals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		set  types.SignalSet
		want SignalSummary
	}{
		{
			name: "empty signal set returns zero summary with empty maps",
			set:  types.SignalSet{},
			want: SignalSummary{
				TotalMatches:   0,
				CountByType:    map[string]int{},
				PatternsByType: map[string][]string{},
			},
		},
		{
			name: "single match adds one count and one pattern",
			set: types.SignalSet{
				Matches: []types.PatternMatch{
					{SignalType: "ConnectionRefused", PatternName: "conn-refused"},
				},
			},
			want: SignalSummary{
				TotalMatches: 1,
				CountByType: map[string]int{
					"ConnectionRefused": 1,
				},
				PatternsByType: map[string][]string{
					"ConnectionRefused": {"conn-refused"},
				},
			},
		},
		{
			name: "multiple matches same type count correctly",
			set: types.SignalSet{
				Matches: []types.PatternMatch{
					{SignalType: "ConnectionRefused", PatternName: "conn-refused"},
					{SignalType: "ConnectionRefused", PatternName: "conn-refused-alt"},
					{SignalType: "ConnectionRefused", PatternName: "conn-refused-alt"},
				},
			},
			want: SignalSummary{
				TotalMatches: 3,
				CountByType: map[string]int{
					"ConnectionRefused": 3,
				},
				PatternsByType: map[string][]string{
					"ConnectionRefused": {"conn-refused", "conn-refused-alt"},
				},
			},
		},
		{
			name: "duplicate pattern names within same type are deduplicated in first seen order",
			set: types.SignalSet{
				Matches: []types.PatternMatch{
					{SignalType: "Panic", PatternName: "panic"},
					{SignalType: "Panic", PatternName: "runtime-error"},
					{SignalType: "Panic", PatternName: "panic"},
					{SignalType: "Panic", PatternName: "runtime-error"},
				},
			},
			want: SignalSummary{
				TotalMatches: 4,
				CountByType: map[string]int{
					"Panic": 4,
				},
				PatternsByType: map[string][]string{
					"Panic": {"panic", "runtime-error"},
				},
			},
		},
		{
			name: "multiple types remain independent",
			set: types.SignalSet{
				Matches: []types.PatternMatch{
					{SignalType: "Panic", PatternName: "panic"},
					{SignalType: "ConnectionRefused", PatternName: "conn-refused"},
					{SignalType: "Panic", PatternName: "runtime-error"},
					{SignalType: "ConnectionRefused", PatternName: "dial-tcp"},
				},
			},
			want: SignalSummary{
				TotalMatches: 4,
				CountByType: map[string]int{
					"Panic":             2,
					"ConnectionRefused": 2,
				},
				PatternsByType: map[string][]string{
					"Panic":             {"panic", "runtime-error"},
					"ConnectionRefused": {"conn-refused", "dial-tcp"},
				},
			},
		},
		{
			name: "first seen ordering is preserved",
			set: types.SignalSet{
				Matches: []types.PatternMatch{
					{SignalType: "Panic", PatternName: "runtime-error"},
					{SignalType: "Panic", PatternName: "panic"},
					{SignalType: "Panic", PatternName: "runtime-error"},
					{SignalType: "Panic", PatternName: "goroutine"},
				},
			},
			want: SignalSummary{
				TotalMatches: 4,
				CountByType: map[string]int{
					"Panic": 4,
				},
				PatternsByType: map[string][]string{
					"Panic": {"runtime-error", "panic", "goroutine"},
				},
			},
		},
		{
			name: "duplicate pattern names across different types remain independent per type",
			set: types.SignalSet{
				Matches: []types.PatternMatch{
					{SignalType: "Panic", PatternName: "shared"},
					{SignalType: "ConnectionRefused", PatternName: "shared"},
					{SignalType: "Panic", PatternName: "shared"},
				},
			},
			want: SignalSummary{
				TotalMatches: 3,
				CountByType: map[string]int{
					"Panic":             2,
					"ConnectionRefused": 1,
				},
				PatternsByType: map[string][]string{
					"Panic":             {"shared"},
					"ConnectionRefused": {"shared"},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := AggregateSignals(tt.set)

			if got.TotalMatches != tt.want.TotalMatches {
				t.Fatalf("TotalMatches = %d, want %d", got.TotalMatches, tt.want.TotalMatches)
			}

			if !reflect.DeepEqual(got.CountByType, tt.want.CountByType) {
				t.Fatalf("CountByType = %#v, want %#v", got.CountByType, tt.want.CountByType)
			}

			if !reflect.DeepEqual(got.PatternsByType, tt.want.PatternsByType) {
				t.Fatalf("PatternsByType = %#v, want %#v", got.PatternsByType, tt.want.PatternsByType)
			}
		})
	}
}
