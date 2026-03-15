package scoring

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestMapConfidence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []CandidateHypothesis
		validate func(t *testing.T, got []CandidateHypothesis)
	}{
		{
			name:  "empty input returns non-nil empty slice",
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
			name: "base score at or above high threshold maps to high",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "OutOfMemory"}, BaseScore: 0.80},
				{Class: types.IssueClass{Name: "Panic"}, BaseScore: 1.20},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 2 {
					t.Fatalf("len(got) = %d, want 2", len(got))
				}
				for i, candidate := range got {
					if candidate.Confidence != types.ConfidenceHigh {
						t.Fatalf("candidate %d Confidence = %q, want %q", i, candidate.Confidence, types.ConfidenceHigh)
					}
				}
			},
		},
		{
			name: "base score at or above medium threshold maps to medium",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "DNSFailure"}, BaseScore: 0.50},
				{Class: types.IssueClass{Name: "TLSFailure"}, BaseScore: 0.79},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 2 {
					t.Fatalf("len(got) = %d, want 2", len(got))
				}
				for i, candidate := range got {
					if candidate.Confidence != types.ConfidenceMedium {
						t.Fatalf("candidate %d Confidence = %q, want %q", i, candidate.Confidence, types.ConfidenceMedium)
					}
				}
			},
		},
		{
			name: "base score at or above low threshold maps to low",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "MissingEnvVar"}, BaseScore: 0.25},
				{Class: types.IssueClass{Name: "PermissionDenied"}, BaseScore: 0.49},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 2 {
					t.Fatalf("len(got) = %d, want 2", len(got))
				}
				for i, candidate := range got {
					if candidate.Confidence != types.ConfidenceLow {
						t.Fatalf("candidate %d Confidence = %q, want %q", i, candidate.Confidence, types.ConfidenceLow)
					}
				}
			},
		},
		{
			name: "below threshold candidate is discarded",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "CrashLoopBackOff"}, BaseScore: 0.2499},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				if len(got) != 0 {
					t.Fatalf("len(got) = %d, want 0", len(got))
				}
			},
		},
		{
			name: "mixed input preserves order and discards only below threshold",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "OutOfMemory"}, BaseScore: 0.90},
				{Class: types.IssueClass{Name: "CrashLoopBackOff"}, BaseScore: 0.2499},
				{Class: types.IssueClass{Name: "DNSFailure"}, BaseScore: 0.50},
				{Class: types.IssueClass{Name: "ConnectionRefused"}, BaseScore: 0.25},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				wantNames := []string{"OutOfMemory", "DNSFailure", "ConnectionRefused"}
				wantConfidence := []types.ConfidenceLevel{
					types.ConfidenceHigh,
					types.ConfidenceMedium,
					types.ConfidenceLow,
				}

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
			name: "mapping does not change base score",
			input: []CandidateHypothesis{
				{Class: types.IssueClass{Name: "DependencyTimeout"}, BaseScore: 0.80},
				{Class: types.IssueClass{Name: "Panic"}, BaseScore: 0.50},
				{Class: types.IssueClass{Name: "DiskFull"}, BaseScore: 0.25},
			},
			validate: func(t *testing.T, got []CandidateHypothesis) {
				t.Helper()
				wantScores := []float64{0.80, 0.50, 0.25}
				if len(got) != len(wantScores) {
					t.Fatalf("len(got) = %d, want %d", len(got), len(wantScores))
				}
				for i, candidate := range got {
					if candidate.BaseScore != wantScores[i] {
						t.Fatalf("candidate %d BaseScore = %v, want %v", i, candidate.BaseScore, wantScores[i])
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
			got := MapConfidence(tt.input)
			tt.validate(t, got)

			if !reflect.DeepEqual(tt.input, original) {
				t.Fatalf("input mutated: got %#v, want %#v", tt.input, original)
			}

			gotAgain := MapConfidence(tt.input)
			if !reflect.DeepEqual(got, gotAgain) {
				t.Fatalf("MapConfidence() is not deterministic: %#v != %#v", got, gotAgain)
			}
		})
	}
}
