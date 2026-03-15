package scoring

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestApplyRelationships(t *testing.T) {
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
			name:  "empty slice returns empty slice",
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
			name: "candidate with no relationship remains false",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "OutOfMemory"}},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 1 {
					t.Fatalf("len(got) = %d, want 1", len(got))
				}
				if got[0].IsSymptom {
					t.Fatal("IsSymptom = true, want false")
				}
			},
		},
		{
			name: "symptom without root cause remains false",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "DNSFailure"}},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if got[0].IsSymptom {
					t.Fatal("IsSymptom = true, want false")
				}
			},
		},
		{
			name: "symptom with root cause present becomes true",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "ConnectionRefused"}},
				{Class: types.IssueClass{Name: "DNSFailure"}},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if got[0].IsSymptom {
					t.Fatal("root cause candidate incorrectly marked as symptom")
				}
				if !got[1].IsSymptom {
					t.Fatal("symptom candidate not marked as symptom")
				}
			},
		},
		{
			name: "root cause candidate is never itself marked symptom",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "DiskFull"}},
				{Class: types.IssueClass{Name: "PermissionDenied"}},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if got[0].IsSymptom {
					t.Fatal("root cause candidate incorrectly marked as symptom")
				}
			},
		},
		{
			name: "multiple symptoms can be marked simultaneously",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "ConnectionRefused"}},
				{Class: types.IssueClass{Name: "DNSFailure"}},
				{Class: types.IssueClass{Name: "TLSFailure"}},
				{Class: types.IssueClass{Name: "DependencyTimeout"}},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if got[0].IsSymptom {
					t.Fatal("ConnectionRefused should remain root cause")
				}
				if !got[1].IsSymptom || !got[2].IsSymptom || !got[3].IsSymptom {
					t.Fatalf("expected DNSFailure, TLSFailure, and DependencyTimeout to be symptoms; got %#v", got)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := cloneCandidatesForRelationships(tt.input)
			got := ApplyRelationships(tt.input)
			tt.validate(t, got)

			if !reflect.DeepEqual(tt.input, original) {
				t.Fatalf("input mutated: got %#v, want %#v", tt.input, original)
			}

			gotAgain := ApplyRelationships(tt.input)
			if !reflect.DeepEqual(got, gotAgain) {
				t.Fatalf("ApplyRelationships() is not deterministic: %#v != %#v", got, gotAgain)
			}
		})
	}
}

func cloneCandidatesForRelationships(candidates []CandidateHypothesis) []CandidateHypothesis {
	if candidates == nil {
		return nil
	}

	cloned := make([]CandidateHypothesis, len(candidates))
	copy(cloned, candidates)
	return cloned
}
