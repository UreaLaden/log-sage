package detection

import "testing"

func TestIssueRegistryIntegrity(t *testing.T) {
	t.Parallel()

	if len(IssueRegistry) != 15 {
		t.Fatalf("registry length = %d, want 15", len(IssueRegistry))
	}

	seenNames := make(map[string]struct{}, len(IssueRegistry))

	for _, issueClass := range IssueRegistry {
		if issueClass.Name == "" {
			t.Fatal("issue class name must not be empty")
		}

		if _, exists := seenNames[issueClass.Name]; exists {
			t.Fatalf("duplicate issue class name %q", issueClass.Name)
		}
		seenNames[issueClass.Name] = struct{}{}

		if len(issueClass.PrimarySignals) == 0 {
			t.Fatalf("issue class %q has no primary signals", issueClass.Name)
		}

		if issueClass.ExplanationTemplate == "" {
			t.Fatalf("issue class %q has empty explanation template", issueClass.Name)
		}

		if len(issueClass.NextSteps) == 0 {
			t.Fatalf("issue class %q has no next steps", issueClass.Name)
		}

		if len(issueClass.Commands) == 0 {
			t.Fatalf("issue class %q has no commands", issueClass.Name)
		}

		for _, signal := range issueClass.PrimarySignals {
			if signal.Name == "" {
				t.Fatalf("issue class %q has primary signal with empty name", issueClass.Name)
			}

			if signal.SignalType == "" {
				t.Fatalf("issue class %q primary signal %q has empty signal type", issueClass.Name, signal.Name)
			}

			if signal.MatchExpression == "" {
				t.Fatalf("issue class %q primary signal %q has empty match expression", issueClass.Name, signal.Name)
			}

			if signal.Weight <= 0 {
				t.Fatalf("issue class %q primary signal %q has non-positive weight %f", issueClass.Name, signal.Name, signal.Weight)
			}
		}
	}
}
