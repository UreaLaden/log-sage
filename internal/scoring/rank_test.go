package scoring

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestRank(t *testing.T) {
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
			name:  "empty input returns non-nil empty slice",
			input: []CandidateHypothesis{},
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
			name: "single candidate is returned unchanged",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "OutOfMemory"}, BaseScore: 1.0},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				want := []CandidateHypothesis{
					{Class: types.IssueClass{Name: "OutOfMemory"}, BaseScore: 1.0},
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("got %#v, want %#v", got, want)
				}
			},
		},
		{
			name: "non symptoms sort by score descending",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "DNSFailure"}, BaseScore: 0.5},
				{Class: types.IssueClass{Name: "OutOfMemory"}, BaseScore: 0.9},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				want := []string{"OutOfMemory", "DNSFailure"}
				for i, name := range want {
					if got[i].Class.Name != name {
						t.Fatalf("candidate %d Class.Name = %q, want %q", i, got[i].Class.Name, name)
					}
				}
			},
		},
		{
			name: "non symptom ranks above symptom regardless of score",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "CrashLoopBackOff"}, BaseScore: 1.5, IsSymptom: true},
				{Class: types.IssueClass{Name: "OutOfMemory"}, BaseScore: 0.5, IsSymptom: false},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if got[0].Class.Name != "OutOfMemory" || got[0].IsSymptom {
					t.Fatalf("first candidate = %#v, want non-symptom OutOfMemory", got[0])
				}
				if got[1].Class.Name != "CrashLoopBackOff" || !got[1].IsSymptom {
					t.Fatalf("second candidate = %#v, want symptom CrashLoopBackOff", got[1])
				}
			},
		},
		{
			name: "symptoms sort among themselves by score descending",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "DependencyTimeout"}, BaseScore: 0.6, IsSymptom: true},
				{Class: types.IssueClass{Name: "DNSFailure"}, BaseScore: 0.8, IsSymptom: true},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				want := []string{"DNSFailure", "DependencyTimeout"}
				for i, name := range want {
					if got[i].Class.Name != name {
						t.Fatalf("candidate %d Class.Name = %q, want %q", i, got[i].Class.Name, name)
					}
				}
			},
		},
		{
			name: "equal scores use class name tiebreaker",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "TLSFailure"}, BaseScore: 0.8},
				{Class: types.IssueClass{Name: "DNSFailure"}, BaseScore: 0.8},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				want := []string{"DNSFailure", "TLSFailure"}
				for i, name := range want {
					if got[i].Class.Name != name {
						t.Fatalf("candidate %d Class.Name = %q, want %q", i, got[i].Class.Name, name)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := cloneCandidatesForRelationships(tt.input)
			got := Rank(tt.input)
			tt.validate(t, got)

			if !reflect.DeepEqual(tt.input, original) {
				t.Fatalf("input mutated: got %#v, want %#v", tt.input, original)
			}

			gotAgain := Rank(tt.input)
			if !reflect.DeepEqual(got, gotAgain) {
				t.Fatalf("Rank() is not deterministic: %#v != %#v", got, gotAgain)
			}
		})
	}
}
