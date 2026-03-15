package recommendation

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestCommands(t *testing.T) {
	t.Parallel()

	registry := []types.IssueClass{
		{
			Name:     "OutOfMemory",
			Commands: []string{"kubectl describe pod <pod>", "kubectl top pod <pod>"},
		},
		{
			Name:     "ConnectionRefused",
			Commands: []string{"kubectl get svc", "kubectl describe svc <service>", "shared"},
		},
		{
			Name:     "DNSFailure",
			Commands: []string{"shared", "kubectl exec -it <pod> -- nslookup <service>"},
		},
		{
			Name:     "BigClass",
			Commands: []string{"one", "two", "three", "four", "five", "six"},
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
			want:     []string{"kubectl describe pod <pod>", "kubectl top pod <pod>"},
		},
		{
			name: "two hypotheses same class deduplicates commands",
			hypotheses: []types.Hypothesis{
				{IssueClass: "OutOfMemory"},
				{IssueClass: "OutOfMemory"},
			},
			registry: registry,
			want:     []string{"kubectl describe pod <pod>", "kubectl top pod <pod>"},
		},
		{
			name: "two hypotheses different classes preserve rank order",
			hypotheses: []types.Hypothesis{
				{IssueClass: "ConnectionRefused"},
				{IssueClass: "OutOfMemory"},
			},
			registry: registry,
			want: []string{
				"kubectl get svc",
				"kubectl describe svc <service>",
				"shared",
				"kubectl describe pod <pod>",
				"kubectl top pod <pod>",
			},
		},
		{
			name: "shared command across classes is deduplicated with first occurrence",
			hypotheses: []types.Hypothesis{
				{IssueClass: "ConnectionRefused"},
				{IssueClass: "DNSFailure"},
			},
			registry: registry,
			want: []string{
				"kubectl get svc",
				"kubectl describe svc <service>",
				"shared",
				"kubectl exec -it <pod> -- nslookup <service>",
			},
		},
		{
			name: "unknown class is skipped safely",
			hypotheses: []types.Hypothesis{
				{IssueClass: "UnknownClass"},
				{IssueClass: "OutOfMemory"},
			},
			registry: registry,
			want:     []string{"kubectl describe pod <pod>", "kubectl top pod <pod>"},
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
			name: "limit enforcement returns at most five commands",
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
			want: []string{
				"kubectl get svc",
				"kubectl describe svc <service>",
				"shared",
				"kubectl exec -it <pod> -- nslookup <service>",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Commands(tt.hypotheses, tt.registry)
			if got == nil {
				t.Fatal("Commands() returned nil, want non-nil slice")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Commands() = %#v, want %#v", got, tt.want)
			}

			gotAgain := Commands(tt.hypotheses, tt.registry)
			if !reflect.DeepEqual(got, gotAgain) {
				t.Fatalf("Commands() is not deterministic: %#v != %#v", got, gotAgain)
			}
		})
	}
}
