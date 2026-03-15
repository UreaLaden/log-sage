package recommendation

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestNextSteps(t *testing.T) {
	t.Parallel()

	registry := []types.IssueClass{
		{
			Name:      "OutOfMemory",
			NextSteps: []string{"Inspect limits", "Check usage", "Review recent changes"},
		},
		{
			Name:      "ConnectionRefused",
			NextSteps: []string{"Verify service", "Check endpoint", "Shared step"},
		},
		{
			Name:      "DNSFailure",
			NextSteps: []string{"Shared step", "Check DNS"},
		},
		{
			Name:      "BigClass",
			NextSteps: []string{"one", "two", "three", "four", "five", "six"},
		},
	}

	tests := []struct {
		name       string
		hypotheses []types.Hypothesis
		registry   []types.IssueClass
		want       []string
	}{
		{
			name:       "nil hypotheses returns non nil empty slice",
			hypotheses: nil,
			registry:   registry,
			want:       []string{},
		},
		{
			name:       "empty hypotheses returns non nil empty slice",
			hypotheses: []types.Hypothesis{},
			registry:   registry,
			want:       []string{},
		},
		{
			name: "single hypothesis with known class",
			hypotheses: []types.Hypothesis{
				{IssueClass: "OutOfMemory"},
			},
			registry: registry,
			want:     []string{"Inspect limits", "Check usage", "Review recent changes"},
		},
		{
			name: "two hypotheses same class deduplicates steps",
			hypotheses: []types.Hypothesis{
				{IssueClass: "OutOfMemory"},
				{IssueClass: "OutOfMemory"},
			},
			registry: registry,
			want:     []string{"Inspect limits", "Check usage", "Review recent changes"},
		},
		{
			name: "two hypotheses different classes preserve rank order",
			hypotheses: []types.Hypothesis{
				{IssueClass: "ConnectionRefused"},
				{IssueClass: "OutOfMemory"},
			},
			registry: registry,
			want:     []string{"Verify service", "Check endpoint", "Shared step", "Inspect limits", "Check usage"},
		},
		{
			name: "shared step across classes is deduplicated with first occurrence",
			hypotheses: []types.Hypothesis{
				{IssueClass: "ConnectionRefused"},
				{IssueClass: "DNSFailure"},
			},
			registry: registry,
			want:     []string{"Verify service", "Check endpoint", "Shared step", "Check DNS"},
		},
		{
			name: "unknown class is skipped safely",
			hypotheses: []types.Hypothesis{
				{IssueClass: "UnknownClass"},
				{IssueClass: "OutOfMemory"},
			},
			registry: registry,
			want:     []string{"Inspect limits", "Check usage", "Review recent changes"},
		},
		{
			name: "empty registry returns non nil empty slice",
			hypotheses: []types.Hypothesis{
				{IssueClass: "OutOfMemory"},
			},
			registry: []types.IssueClass{},
			want:     []string{},
		},
		{
			name: "limit enforcement returns at most five steps",
			hypotheses: []types.Hypothesis{
				{IssueClass: "BigClass"},
			},
			registry: registry,
			want:     []string{"one", "two", "three", "four", "five"},
		},
		{
			name: "determinism",
			hypotheses: []types.Hypothesis{
				{IssueClass: "ConnectionRefused"},
				{IssueClass: "DNSFailure"},
			},
			registry: registry,
			want:     []string{"Verify service", "Check endpoint", "Shared step", "Check DNS"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NextSteps(tt.hypotheses, tt.registry)
			if got == nil {
				t.Fatal("NextSteps() returned nil, want non-nil slice")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("NextSteps() = %#v, want %#v", got, tt.want)
			}

			gotAgain := NextSteps(tt.hypotheses, tt.registry)
			if !reflect.DeepEqual(got, gotAgain) {
				t.Fatalf("NextSteps() is not deterministic: %#v != %#v", got, gotAgain)
			}
		})
	}
}
