package scoring

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func runPipeline(hypotheses []types.Hypothesis, signals types.SignalSet) []CandidateHypothesis {
	candidates := BuildCandidates(hypotheses, signals)
	candidates = ApplyCorroboration(candidates)
	candidates = ApplyRelationships(candidates)
	candidates = MapConfidence(candidates)
	candidates = Rank(candidates)
	return candidates
}

func TestScoringPipeline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hypotheses []types.Hypothesis
		signals    types.SignalSet
		validate   func(t *testing.T, got []CandidateHypothesis)
	}{
		{
			name:       "empty input returns non nil empty slice through full pipeline",
			hypotheses: nil,
			signals:    types.SignalSet{},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if got == nil {
					t.Fatal("got nil, want non-nil empty slice")
				}
				if len(got) != 0 {
					t.Fatalf("len(got) = %d, want 0", len(got))
				}

				built := BuildCandidates(nil, types.SignalSet{})
				if built == nil {
					t.Fatal("BuildCandidates returned nil, want non-nil empty slice")
				}
				corroborated := ApplyCorroboration(built)
				if corroborated == nil {
					t.Fatal("ApplyCorroboration returned nil, want non-nil empty slice")
				}
				related := ApplyRelationships(corroborated)
				if related == nil {
					t.Fatal("ApplyRelationships returned nil, want non-nil empty slice")
				}
				mapped := MapConfidence(related)
				if mapped == nil {
					t.Fatal("MapConfidence returned nil, want non-nil empty slice")
				}
				ranked := Rank(mapped)
				if ranked == nil {
					t.Fatal("Rank returned nil, want non-nil empty slice")
				}
			},
		},
		{
			name: "single high confidence root cause",
			hypotheses: []types.Hypothesis{
				{
					IssueClass: "OutOfMemory",
					Score:      0.90,
					Evidence: []types.Evidence{
						{Signal: "oom-killed", Occurrences: 1, Examples: []string{"OOMKilled: container exceeded memory limit"}},
					},
				},
			},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "oom-killed", SignalType: "OutOfMemory", Text: "OOMKilled: container exceeded memory limit"},
				},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}
				if got[0].Class.Name != "OutOfMemory" {
					t.Fatalf("Class.Name = %q, want %q", got[0].Class.Name, "OutOfMemory")
				}
				if got[0].Confidence != types.ConfidenceHigh {
					t.Fatalf("Confidence = %q, want %q", got[0].Confidence, types.ConfidenceHigh)
				}
				if got[0].IsSymptom {
					t.Fatal("IsSymptom = true, want false")
				}
			},
		},
		{
			name: "mixed scores discard below threshold and preserve ranked survivors",
			hypotheses: []types.Hypothesis{
				{IssueClass: "ConnectionRefused", Score: 0.75},
				{IssueClass: "DNSFailure", Score: 0.10},
				{IssueClass: "TLSFailure", Score: 0.50},
			},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "connection-refused", SignalType: "ConnectionRefused", Text: "dial tcp: connection refused"},
				},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				wantNames := []string{"ConnectionRefused", "TLSFailure"}
				wantConfidence := []types.ConfidenceLevel{types.ConfidenceMedium, types.ConfidenceMedium}
				if len(got) != len(wantNames) {
					t.Fatalf("len(got) = %d, want %d", len(got), len(wantNames))
				}
				for i, candidate := range got {
					if candidate.Class.Name != wantNames[i] {
						t.Fatalf("candidate %d Class.Name = %q, want %q", i, candidate.Class.Name, wantNames[i])
					}
					if candidate.Confidence != wantConfidence[i] {
						t.Fatalf("candidate %d Confidence = %q, want %q", i, candidate.Confidence, wantConfidence[i])
					}
				}
			},
		},
		{
			name: "relationship adjustment marks symptom and affects rank",
			hypotheses: []types.Hypothesis{
				{IssueClass: "CrashLoopBackOff", Score: 0.70},
				{IssueClass: "OutOfMemory", Score: 0.60},
			},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "oom-killed", SignalType: "OutOfMemory", Text: "OOMKilled: container exceeded memory limit"},
				},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 2 {
					t.Fatalf("len(got) = %d, want 2", len(got))
				}
				if got[0].Class.Name != "OutOfMemory" || got[0].IsSymptom {
					t.Fatalf("first candidate = %#v, want non-symptom OutOfMemory", got[0])
				}
				if got[1].Class.Name != "CrashLoopBackOff" {
					t.Fatalf("second Class.Name = %q, want %q", got[1].Class.Name, "CrashLoopBackOff")
				}
				if !got[1].IsSymptom {
					t.Fatal("CrashLoopBackOff should be marked as symptom")
				}
			},
		},
		{
			name: "corroboration boost promotes confidence tier",
			hypotheses: []types.Hypothesis{
				{
					IssueClass: "DependencyTimeout",
					Score:      0.65,
					Evidence: []types.Evidence{
						{Signal: "timeout after", Occurrences: 1, Examples: []string{"timeout after 30s"}},
						{Signal: "context deadline exceeded", Occurrences: 1, Examples: []string{"context deadline exceeded"}},
					},
				},
			},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "timeout-after", SignalType: "DependencyTimeout", Text: "timeout after 30s"},
				},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}
				if got[0].Confidence != types.ConfidenceHigh {
					t.Fatalf("Confidence = %q, want %q", got[0].Confidence, types.ConfidenceHigh)
				}
				if got[0].BaseScore < 0.80 {
					t.Fatalf("BaseScore = %v, want corroborated score >= 0.80", got[0].BaseScore)
				}
			},
		},
		{
			name: "deterministic ordering for ties uses class name",
			hypotheses: []types.Hypothesis{
				{IssueClass: "TLSFailure", Score: 0.50},
				{IssueClass: "DNSFailure", Score: 0.50},
				{IssueClass: "ConnectionRefused", Score: 0.50},
			},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "connection-refused", SignalType: "ConnectionRefused", Text: "connection refused"},
				},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				wantNames := []string{"ConnectionRefused", "DNSFailure", "TLSFailure"}
				if len(got) != len(wantNames) {
					t.Fatalf("len(got) = %d, want %d", len(got), len(wantNames))
				}
				for i, candidate := range got {
					if candidate.Class.Name != wantNames[i] {
						t.Fatalf("candidate %d Class.Name = %q, want %q", i, candidate.Class.Name, wantNames[i])
					}
				}
			},
		},
		{
			name: "full realistic mixed case",
			hypotheses: []types.Hypothesis{
				{
					IssueClass: "ConnectionRefused",
					Score:      0.55,
					Evidence: []types.Evidence{
						{Signal: "connection refused", Occurrences: 2, Examples: []string{"dial tcp: connection refused"}},
						{Signal: "dial tcp", Occurrences: 1, Examples: []string{"dial tcp 10.0.0.8:5432"}},
					},
				},
				{
					IssueClass: "DNSFailure",
					Score:      0.40,
					Evidence: []types.Evidence{
						{Signal: "no such host", Occurrences: 1, Examples: []string{"lookup db: no such host"}},
					},
				},
				{
					IssueClass: "CrashLoopBackOff",
					Score:      0.30,
					Evidence: []types.Evidence{
						{Signal: "back-off restarting failed container", Occurrences: 1, Examples: []string{"Back-off restarting failed container"}},
					},
				},
				{
					IssueClass: "OutOfMemory",
					Score:      0.10,
					Evidence: []types.Evidence{
						{Signal: "oom-killed", Occurrences: 1, Examples: []string{"OOMKilled"}},
					},
				},
			},
			signals: types.SignalSet{
				Matches: []types.PatternMatch{
					{PatternName: "connection-refused", SignalType: "ConnectionRefused", Text: "dial tcp: connection refused"},
					{PatternName: "no-such-host", SignalType: "DNSFailure", Text: "lookup db: no such host"},
					{PatternName: "crashloopbackoff", SignalType: "CrashLoopBackOff", Text: "Back-off restarting failed container"},
				},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				wantNames := []string{"ConnectionRefused", "DNSFailure", "CrashLoopBackOff"}
				wantConfidence := []types.ConfidenceLevel{
					types.ConfidenceHigh,
					types.ConfidenceLow,
					types.ConfidenceLow,
				}
				wantSymptom := []bool{false, true, true}
				if len(got) != len(wantNames) {
					t.Fatalf("len(got) = %d, want %d", len(got), len(wantNames))
				}
				for i, candidate := range got {
					if candidate.Class.Name != wantNames[i] {
						t.Fatalf("candidate %d Class.Name = %q, want %q", i, candidate.Class.Name, wantNames[i])
					}
					if candidate.Confidence != wantConfidence[i] {
						t.Fatalf("candidate %d Confidence = %q, want %q", i, candidate.Confidence, wantConfidence[i])
					}
					if candidate.IsSymptom != wantSymptom[i] {
						t.Fatalf("candidate %d IsSymptom = %t, want %t", i, candidate.IsSymptom, wantSymptom[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			originalHypotheses := cloneHypotheses(tt.hypotheses)
			originalSignals := cloneSignalSet(tt.signals)

			got := runPipeline(tt.hypotheses, tt.signals)
			tt.validate(t, got)

			if !reflect.DeepEqual(tt.hypotheses, originalHypotheses) {
				t.Fatalf("hypotheses mutated: got %#v, want %#v", tt.hypotheses, originalHypotheses)
			}
			if !reflect.DeepEqual(tt.signals, originalSignals) {
				t.Fatalf("signals mutated: got %#v, want %#v", tt.signals, originalSignals)
			}

			gotAgain := runPipeline(tt.hypotheses, tt.signals)
			if !reflect.DeepEqual(got, gotAgain) {
				t.Fatalf("runPipeline() is not deterministic: %#v != %#v", got, gotAgain)
			}
		})
	}
}
