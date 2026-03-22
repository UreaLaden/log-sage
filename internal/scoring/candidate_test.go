package scoring

import (
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestCandidateHypothesisZeroValue(t *testing.T) {
	t.Parallel()

	var candidate CandidateHypothesis

	if candidate.Class.Name != "" {
		t.Fatalf("Class.Name = %q, want empty", candidate.Class.Name)
	}
}

func TestCandidateHypothesisRetainsClassName(t *testing.T) {
	t.Parallel()

	candidate := CandidateHypothesis{
		Class: types.IssueClass{Name: "ConnectionRefused"},
	}

	if candidate.Class.Name != "DOGFOOD_FAILURE" {
		t.Fatalf("Class.Name = %q, want %q", candidate.Class.Name, "DOGFOOD_FAILURE")
	}
}
