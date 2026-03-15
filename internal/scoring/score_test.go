package scoring

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestBuildCandidates(t *testing.T) {
	t.Parallel()

	sharedSignals := types.SignalSet{
		Matches: []types.PatternMatch{
			{Text: "ERROR ErrImagePull: failed to pull image", SignalType: "ImagePullBackOff"},
		},
	}
	multipleSignals := types.SignalSet{
		Matches: []types.PatternMatch{
			{Text: "startup error: configuration incomplete"},
			{Text: "connect: connection refused"},
		},
	}

	tests := []struct {
		name       string
		hypotheses []types.Hypothesis
		signals    types.SignalSet
		validate   func(t *testing.T, got []CandidateHypothesis)
	}{
		{
			name:       "empty hypotheses returns non-nil empty slice",
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
			},
		},
		{
			name: "single hypothesis populates candidate fields",
			hypotheses: []types.Hypothesis{
				{
					IssueClass: "ImagePullBackOff",
					Score:      1.9,
					Evidence: []types.Evidence{
						{Signal: "imagepullbackoff", Occurrences: 1, Examples: []string{"ImagePullBackOff: back-off pulling image"}},
					},
				},
			},
			signals: sharedSignals,
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}
				candidate := got[0]
				if candidate.Class.Name != "ImagePullBackOff" {
					t.Fatalf("Class.Name = %q, want %q", candidate.Class.Name, "ImagePullBackOff")
				}
				if candidate.BaseScore != 1.9 {
					t.Fatalf("BaseScore = %v, want 1.9", candidate.BaseScore)
				}
				if !reflect.DeepEqual(candidate.Evidence, gotInputEvidence()) {
					t.Fatalf("Evidence = %#v, want %#v", candidate.Evidence, gotInputEvidence())
				}
				if !reflect.DeepEqual(candidate.Signals, sharedSignals) {
					t.Fatalf("Signals = %#v, want %#v", candidate.Signals, sharedSignals)
				}
				if candidate.Phase != types.FailurePhaseImagePull {
					t.Fatalf("Phase = %q, want %q", candidate.Phase, types.FailurePhaseImagePull)
				}
				if candidate.IsSymptom {
					t.Fatal("IsSymptom = true, want false")
				}
			},
		},
		{
			name: "multiple hypotheses preserve order and propagate signals",
			hypotheses: []types.Hypothesis{
				{IssueClass: "MissingEnvVar", Score: 1.0},
				{IssueClass: "ConnectionRefused", Score: 0.8},
			},
			signals: multipleSignals,
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 2 {
					t.Fatalf("len(got) = %d, want 2", len(got))
				}
				if got[0].Class.Name != "MissingEnvVar" || got[1].Class.Name != "ConnectionRefused" {
					t.Fatalf("candidate order = [%q, %q], want [MissingEnvVar, ConnectionRefused]", got[0].Class.Name, got[1].Class.Name)
				}
				for i, candidate := range got {
					if !reflect.DeepEqual(candidate.Signals, multipleSignals) {
						t.Fatalf("candidate %d Signals = %#v, want %#v", i, candidate.Signals, multipleSignals)
					}
					if candidate.IsSymptom {
						t.Fatalf("candidate %d IsSymptom = true, want false", i)
					}
				}
				if got[0].Phase != types.FailurePhaseStartup {
					t.Fatalf("first Phase = %q, want %q", got[0].Phase, types.FailurePhaseStartup)
				}
				if got[1].Phase != types.FailurePhaseStartup {
					t.Fatalf("second Phase = %q, want %q", got[1].Phase, types.FailurePhaseStartup)
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
			got := BuildCandidates(tt.hypotheses, tt.signals)
			tt.validate(t, got)

			if !reflect.DeepEqual(tt.hypotheses, originalHypotheses) {
				t.Fatalf("hypotheses mutated: got %#v, want %#v", tt.hypotheses, originalHypotheses)
			}
			if !reflect.DeepEqual(tt.signals, originalSignals) {
				t.Fatalf("signals mutated: got %#v, want %#v", tt.signals, originalSignals)
			}

			gotAgain := BuildCandidates(tt.hypotheses, tt.signals)
			if !reflect.DeepEqual(got, gotAgain) {
				t.Fatalf("BuildCandidates() is not deterministic: %#v != %#v", got, gotAgain)
			}
		})
	}
}

func TestBuildCandidatesInputIsolation(t *testing.T) {
	t.Parallel()

	hypotheses := []types.Hypothesis{
		{
			IssueClass: "OutOfMemory",
			Score:      1.5,
			Evidence: []types.Evidence{
				{Signal: "oom-killed", Occurrences: 2, Examples: []string{"OOMKilled: container exceeded memory limit"}},
			},
		},
	}
	signals := types.SignalSet{
		Matches: []types.PatternMatch{{PatternName: "oom-killed", Text: "OOMKilled"}},
	}

	got := BuildCandidates(hypotheses, signals)

	// Mutate inputs after the call.
	hypotheses[0].Evidence[0].Signal = "mutated-evidence"
	hypotheses[0].Evidence[0].Examples[0] = "mutated-example"
	signals.Matches[0].PatternName = "mutated-pattern"
	signals.Matches[0].Text = "mutated-signal"

	if got[0].Evidence[0].Signal != "oom-killed" {
		t.Fatalf("candidate Evidence was aliased to input: got %q, want %q", got[0].Evidence[0].Signal, "oom-killed")
	}
	if got[0].Evidence[0].Examples[0] != "OOMKilled: container exceeded memory limit" {
		t.Fatalf("candidate Evidence examples were aliased to input: got %q, want %q", got[0].Evidence[0].Examples[0], "OOMKilled: container exceeded memory limit")
	}
	if got[0].Signals.Matches[0].PatternName != "oom-killed" {
		t.Fatalf("candidate Signals pattern was aliased to input: got %q, want %q", got[0].Signals.Matches[0].PatternName, "oom-killed")
	}
	if got[0].Signals.Matches[0].Text != "OOMKilled" {
		t.Fatalf("candidate Signals was aliased to input: got %q, want %q", got[0].Signals.Matches[0].Text, "OOMKilled")
	}
}

func gotInputEvidence() []types.Evidence {
	return []types.Evidence{
		{Signal: "imagepullbackoff", Occurrences: 1, Examples: []string{"ImagePullBackOff: back-off pulling image"}},
	}
}

func cloneHypotheses(hypotheses []types.Hypothesis) []types.Hypothesis {
	if hypotheses == nil {
		return nil
	}

	cloned := make([]types.Hypothesis, len(hypotheses))
	for i, hypothesis := range hypotheses {
		cloned[i] = hypothesis
		if hypothesis.Evidence != nil {
			evidence := make([]types.Evidence, len(hypothesis.Evidence))
			for j, item := range hypothesis.Evidence {
				evidence[j] = item
				if item.Examples != nil {
					examples := make([]string, len(item.Examples))
					copy(examples, item.Examples)
					evidence[j].Examples = examples
				}
			}
			cloned[i].Evidence = evidence
		}
	}

	return cloned
}
