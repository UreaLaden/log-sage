package detection

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestEvaluateRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		signals  types.SignalSet
		validate func(t *testing.T, got []types.Hypothesis)
	}{
		{
			name:    "empty signal set returns empty non nil result",
			signals: types.SignalSet{},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if got == nil {
					t.Fatal("got nil, want empty slice")
				}
				if len(got) != 0 {
					t.Fatalf("len(got) = %d, want 0", len(got))
				}
			},
		},
		{
			name: "no matching registry entries returns empty result",
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "unrelated", Text: "nothing to see"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if got == nil {
					t.Fatal("got nil, want empty slice")
				}
				if len(got) != 0 {
					t.Fatalf("len(got) = %d, want 0", len(got))
				}
			},
		},
		{
			name: "one matching issue class returns one hypothesis",
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "connection-refused", Text: "connection refused"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}
				if got[0].IssueClass != "ConnectionRefused" {
					t.Fatalf("IssueClass = %q, want %q", got[0].IssueClass, "ConnectionRefused")
				}
			},
		},
		{
			name: "multiple matching issue classes return hypotheses in registry order",
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "connection-refused", Text: "connection refused"},
					{PatternName: "panic-prefix", Text: "panic: boom"},
					{PatternName: "goroutine", Text: "goroutine 1 [running]:"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if len(got) != 2 {
					t.Fatalf("len(got) = %d, want 2", len(got))
				}
				if got[0].IssueClass != "ConnectionRefused" {
					t.Fatalf("got[0].IssueClass = %q, want %q", got[0].IssueClass, "ConnectionRefused")
				}
				if got[1].IssueClass != "Panic" {
					t.Fatalf("got[1].IssueClass = %q, want %q", got[1].IssueClass, "Panic")
				}
			},
		},
		{
			name: "non matching entries do not block later matching entries",
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "unrelated", Text: "unrelated"},
					{PatternName: "panic-prefix", Text: "panic: boom"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}
				if got[0].IssueClass != "Panic" {
					t.Fatalf("IssueClass = %q, want %q", got[0].IssueClass, "Panic")
				}
			},
		},
		{
			name: "repeated calls are deterministic and input is not mutated",
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "connection-refused", Text: "connection refused"},
					{PatternName: "panic-prefix", Text: "panic: boom"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				again := EvaluateRegistry(types.SignalSet{
					Matches: []types.PatternMatch{
						{PatternName: "connection-refused", Text: "connection refused"},
						{PatternName: "panic-prefix", Text: "panic: boom"},
					},
				})
				if !reflect.DeepEqual(got, again) {
					t.Fatalf("results differ across repeated calls: %#v != %#v", got, again)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			originalSignals := cloneSignalSet(tt.signals)
			got := EvaluateRegistry(tt.signals)
			tt.validate(t, got)

			if !reflect.DeepEqual(tt.signals, originalSignals) {
				t.Fatalf("signals mutated: got %#v, want %#v", tt.signals, originalSignals)
			}
		})
	}
}
