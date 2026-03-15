package detection

import (
	"math"
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestDefaultDetectorDetect(t *testing.T) {
	t.Parallel()

	baseClass := types.IssueClass{
		Name: "ConnectionRefused",
		PrimarySignals: []types.SignalPattern{
			{Name: "conn-refused", SignalType: "ConnectionRefused", MatchExpression: "connection refused", Weight: 0.6},
			{Name: "dial-tcp", SignalType: "ConnectionRefused", MatchExpression: "dial tcp", Weight: 0.3},
		},
		CorroboratingSignals: []types.SignalPattern{
			{Name: "retry-loop", SignalType: "ConnectionRefused", MatchExpression: "retrying", Weight: 0.2},
		},
		ExplanationTemplate: "The application attempted to connect to a service that refused the connection.",
	}

	tests := []struct {
		name     string
		detector DefaultDetector
		signals  types.SignalSet
		validate func(t *testing.T, got []types.Hypothesis)
	}{
		{
			name:     "empty signal set returns nil",
			detector: DefaultDetector{Class: baseClass},
			signals:  types.SignalSet{},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if got != nil {
					t.Fatalf("got %#v, want nil", got)
				}
			},
		},
		{
			name:     "no primary match returns nil",
			detector: DefaultDetector{Class: baseClass},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "unrelated", Text: "hello"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if got != nil {
					t.Fatalf("got %#v, want nil", got)
				}
			},
		},
		{
			name:     "corroborating only returns nil",
			detector: DefaultDetector{Class: baseClass},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "retry-loop", Text: "retrying connection"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if got != nil {
					t.Fatalf("got %#v, want nil", got)
				}
			},
		},
		{
			name:     "single primary match emits one hypothesis",
			detector: DefaultDetector{Class: baseClass},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "conn-refused", Text: "connection refused"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}

				hyp := got[0]
				if hyp.IssueClass != "ConnectionRefused" {
					t.Fatalf("IssueClass = %q, want %q", hyp.IssueClass, "ConnectionRefused")
				}
				if hyp.Score != 0.6 {
					t.Fatalf("Score = %v, want 0.6", hyp.Score)
				}
				if hyp.Confidence != types.ConfidenceMedium {
					t.Fatalf("Confidence = %q, want %q", hyp.Confidence, types.ConfidenceMedium)
				}
				if hyp.Explanation != baseClass.ExplanationTemplate {
					t.Fatalf("Explanation = %q, want %q", hyp.Explanation, baseClass.ExplanationTemplate)
				}
				if hyp.Phase != "" {
					t.Fatalf("Phase = %q, want zero value", hyp.Phase)
				}
			},
		},
		{
			name:     "multiple primary matches still emit one hypothesis with grouped evidence",
			detector: DefaultDetector{Class: baseClass},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "conn-refused", Text: "connection refused to redis"},
					{PatternName: "conn-refused", Text: "connection refused to postgres"},
					{PatternName: "conn-refused", Text: "connection refused to api"},
					{PatternName: "conn-refused", Text: "connection refused to cache"},
					{PatternName: "dial-tcp", Text: "dial tcp 10.0.0.1:5432"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}

				hyp := got[0]
				if math.Abs(hyp.Score-0.9) > 1e-9 {
					t.Fatalf("Score = %v, want approximately 0.9", hyp.Score)
				}
				if hyp.Confidence != types.ConfidenceHigh {
					t.Fatalf("Confidence = %q, want %q", hyp.Confidence, types.ConfidenceHigh)
				}
				if len(hyp.Evidence) != 2 {
					t.Fatalf("len(Evidence) = %d, want 2", len(hyp.Evidence))
				}

				first := hyp.Evidence[0]
				if first.Signal != "conn-refused" || first.Occurrences != 4 {
					t.Fatalf("first evidence = %#v, want signal conn-refused with 4 occurrences", first)
				}
				if len(first.Examples) != 3 {
					t.Fatalf("len(first.Examples) = %d, want 3", len(first.Examples))
				}

				second := hyp.Evidence[1]
				if second.Signal != "dial-tcp" || second.Occurrences != 1 {
					t.Fatalf("second evidence = %#v, want signal dial-tcp with 1 occurrence", second)
				}
			},
		},
		{
			name:     "corroborating signal adds to score",
			detector: DefaultDetector{Class: baseClass},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "conn-refused", Text: "connection refused"},
					{PatternName: "retry-loop", Text: "retrying connection"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}
				if got[0].Score != 0.8 {
					t.Fatalf("Score = %v, want 0.8", got[0].Score)
				}
				if got[0].Confidence != types.ConfidenceHigh {
					t.Fatalf("Confidence = %q, want %q", got[0].Confidence, types.ConfidenceHigh)
				}
			},
		},
		{
			name: "score below threshold after primary match returns nil",
			detector: DefaultDetector{Class: types.IssueClass{
				Name: "LowSignal",
				PrimarySignals: []types.SignalPattern{
					{Name: "weak", SignalType: "LowSignal", MatchExpression: "weak", Weight: 0.2},
				},
				ExplanationTemplate: "weak",
			}},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "weak", Text: "weak evidence"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				if got != nil {
					t.Fatalf("got %#v, want nil", got)
				}
			},
		},
		{
			name:     "deterministic repeated calls return identical results",
			detector: DefaultDetector{Class: baseClass},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "conn-refused", Text: "connection refused"},
					{PatternName: "dial-tcp", Text: "dial tcp"},
					{PatternName: "retry-loop", Text: "retrying"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				gotAgain := (DefaultDetector{Class: baseClass}).Detect(types.SignalSet{
					Matches: []types.PatternMatch{
						{PatternName: "conn-refused", Text: "connection refused"},
						{PatternName: "dial-tcp", Text: "dial tcp"},
						{PatternName: "retry-loop", Text: "retrying"},
					},
				})

				if !reflect.DeepEqual(got, gotAgain) {
					t.Fatalf("results differ across repeated calls: %#v != %#v", got, gotAgain)
				}
			},
		},
		{
			name:     "detector does not mutate input signal set",
			detector: DefaultDetector{Class: baseClass},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "conn-refused", Text: "connection refused"},
					{PatternName: "retry-loop", Text: "retrying"},
				},
			},
			validate: func(t *testing.T, got []types.Hypothesis) {
				t.Helper()
				_ = got
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			originalSignals := cloneSignalSet(tt.signals)
			got := tt.detector.Detect(tt.signals)
			tt.validate(t, got)

			if !reflect.DeepEqual(tt.signals, originalSignals) {
				t.Fatalf("signals mutated: got %#v, want %#v", tt.signals, originalSignals)
			}
		})
	}
}

func cloneSignalSet(signals types.SignalSet) types.SignalSet {
	if signals.Matches == nil {
		return types.SignalSet{}
	}

	clonedMatches := make([]types.PatternMatch, len(signals.Matches))
	copy(clonedMatches, signals.Matches)

	return types.SignalSet{
		Matches: clonedMatches,
	}
}
