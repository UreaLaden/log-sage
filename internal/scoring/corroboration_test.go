package scoring

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestApplyCorroboration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []CandidateHypothesis
		validate func(t *testing.T, got []CandidateHypothesis)
	}{
		{
			name:  "nil input returns non-nil empty slice",
			input: nil,
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
			name: "single evidence item leaves score unchanged",
			input: []CandidateHypothesis{
				{
					Class:     types.IssueClass{Name: "ConnectionRefused"},
					BaseScore: 1.0,
					Evidence:  []types.Evidence{{Signal: "connection-refused", Occurrences: 1}},
				},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}
				if got[0].BaseScore != 1.0 {
					t.Fatalf("BaseScore = %v, want 1.0", got[0].BaseScore)
				}
			},
		},
		{
			name: "multiple evidence items increase score correctly",
			input: []CandidateHypothesis{
				{
					Class:     types.IssueClass{Name: "ImagePullBackOff"},
					BaseScore: 1.0,
					Evidence: []types.Evidence{
						{Signal: "imagepullbackoff", Occurrences: 1},
						{Signal: "errimagepull", Occurrences: 1},
						{Signal: "pull-access-denied", Occurrences: 1},
					},
				},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}
				if got[0].BaseScore != 1.5 {
					t.Fatalf("BaseScore = %v, want 1.5", got[0].BaseScore)
				}
			},
		},
		{
			name: "candidate order is preserved",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "MissingEnvVar"}, BaseScore: 0.8, Evidence: []types.Evidence{{Signal: "required-env", Occurrences: 1}, {Signal: "startup-error", Occurrences: 1}}},
				{Class: types.IssueClass{Name: "DiskFull"}, BaseScore: 0.9, Evidence: []types.Evidence{{Signal: "no-space-left", Occurrences: 1}}},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 2 {
					t.Fatalf("len(got) = %d, want 2", len(got))
				}
				if got[0].Class.Name != "MissingEnvVar" || got[1].Class.Name != "DiskFull" {
					t.Fatalf("order = [%q, %q], want [MissingEnvVar, DiskFull]", got[0].Class.Name, got[1].Class.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := cloneCandidates(tt.input)
			got := ApplyCorroboration(tt.input)
			tt.validate(t, got)

			if !reflect.DeepEqual(tt.input, original) {
				t.Fatalf("input mutated: got %#v, want %#v", tt.input, original)
			}

			gotAgain := ApplyCorroboration(tt.input)
			if !reflect.DeepEqual(got, gotAgain) {
				t.Fatalf("ApplyCorroboration() is not deterministic: %#v != %#v", got, gotAgain)
			}
		})
	}
}

func cloneCandidates(candidates []CandidateHypothesis) []CandidateHypothesis {
	if candidates == nil {
		return nil
	}

	cloned := make([]CandidateHypothesis, len(candidates))
	for i, candidate := range candidates {
		cloned[i] = candidate
		if candidate.Evidence != nil {
			evidence := make([]types.Evidence, len(candidate.Evidence))
			for j, item := range candidate.Evidence {
				evidence[j] = item
				if item.Examples != nil {
					examples := make([]string, len(item.Examples))
					copy(examples, item.Examples)
					evidence[j].Examples = examples
				}
			}
			cloned[i].Evidence = evidence
		}
		if candidate.Signals.Matches != nil {
			matches := make([]types.PatternMatch, len(candidate.Signals.Matches))
			copy(matches, candidate.Signals.Matches)
			cloned[i].Signals = types.SignalSet{Matches: matches}
		}
	}

	return cloned
}
