package recommendation

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/internal/detection"
	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestNextStepsAndCommandsConsistency(t *testing.T) {
	t.Parallel()

	registry := []types.IssueClass{
		{
			Name:      "Alpha",
			NextSteps: []string{"step-a1", "shared-step"},
			Commands:  []string{"cmd-a1", "shared-cmd"},
		},
		{
			Name:      "Beta",
			NextSteps: []string{"step-b1", "shared-step"},
			Commands:  []string{"cmd-b1", "shared-cmd"},
		},
		{
			Name:      "Gamma",
			NextSteps: []string{"step-g1", "step-g2"},
			Commands:  []string{"cmd-g1", "cmd-g2"},
		},
		{
			Name:      "Big",
			NextSteps: []string{"s1", "s2", "s3", "s4", "s5", "s6"},
			Commands:  []string{"c1", "c2", "c3", "c4", "c5", "c6"},
		},
	}

	tests := []struct {
		name         string
		hypotheses   []types.Hypothesis
		registry     []types.IssueClass
		wantSteps    []string
		wantCommands []string
	}{
		{
			name:         "nil hypotheses",
			hypotheses:   nil,
			registry:     registry,
			wantSteps:    []string{},
			wantCommands: []string{},
		},
		{
			name:         "empty hypotheses",
			hypotheses:   []types.Hypothesis{},
			registry:     registry,
			wantSteps:    []string{},
			wantCommands: []string{},
		},
		{
			name: "single hypothesis",
			hypotheses: []types.Hypothesis{
				{IssueClass: "Alpha"},
			},
			registry:     registry,
			wantSteps:    []string{"step-a1", "shared-step"},
			wantCommands: []string{"cmd-a1", "shared-cmd"},
		},
		{
			name: "shared content guard is independent per function",
			hypotheses: []types.Hypothesis{
				{IssueClass: "Alpha"},
			},
			registry: []types.IssueClass{
				{
					Name:      "Alpha",
					NextSteps: []string{"shared"},
					Commands:  []string{"shared"},
				},
			},
			wantSteps:    []string{"shared"},
			wantCommands: []string{"shared"},
		},
		{
			name: "multiple hypotheses follow rank order",
			hypotheses: []types.Hypothesis{
				{IssueClass: "Beta"},
				{IssueClass: "Alpha"},
			},
			registry:     registry,
			wantSteps:    []string{"step-b1", "shared-step", "step-a1"},
			wantCommands: []string{"cmd-b1", "shared-cmd", "cmd-a1"},
		},
		{
			name: "duplicate steps across classes removed preserving order",
			hypotheses: []types.Hypothesis{
				{IssueClass: "Alpha"},
				{IssueClass: "Beta"},
			},
			registry:     registry,
			wantSteps:    []string{"step-a1", "shared-step", "step-b1"},
			wantCommands: []string{"cmd-a1", "shared-cmd", "cmd-b1"},
		},
		{
			name: "mixed registry ordering does not affect hypothesis ordering",
			hypotheses: []types.Hypothesis{
				{IssueClass: "Gamma"},
				{IssueClass: "Alpha"},
			},
			registry: []types.IssueClass{
				registry[2],
				registry[0],
				registry[1],
			},
			wantSteps:    []string{"step-g1", "step-g2", "step-a1", "shared-step"},
			wantCommands: []string{"cmd-g1", "cmd-g2", "cmd-a1", "shared-cmd"},
		},
		{
			name: "limit enforcement",
			hypotheses: []types.Hypothesis{
				{IssueClass: "Big"},
			},
			registry:     registry,
			wantSteps:    []string{"s1", "s2", "s3", "s4", "s5"},
			wantCommands: []string{"c1", "c2", "c3", "c4", "c5"},
		},
		{
			name: "determinism",
			hypotheses: []types.Hypothesis{
				{IssueClass: "Alpha"},
				{IssueClass: "Gamma"},
			},
			registry:     registry,
			wantSteps:    []string{"step-a1", "shared-step", "step-g1", "step-g2"},
			wantCommands: []string{"cmd-a1", "shared-cmd", "cmd-g1", "cmd-g2"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotSteps := NextSteps(tt.hypotheses, tt.registry)
			if gotSteps == nil {
				t.Fatal("NextSteps() returned nil, want non-nil slice")
			}
			if !reflect.DeepEqual(gotSteps, tt.wantSteps) {
				t.Fatalf("NextSteps() = %#v, want %#v", gotSteps, tt.wantSteps)
			}

			gotCommands := Commands(tt.hypotheses, tt.registry)
			if gotCommands == nil {
				t.Fatal("Commands() returned nil, want non-nil slice")
			}
			if !reflect.DeepEqual(gotCommands, tt.wantCommands) {
				t.Fatalf("Commands() = %#v, want %#v", gotCommands, tt.wantCommands)
			}

			gotStepsAgain := NextSteps(tt.hypotheses, tt.registry)
			if !reflect.DeepEqual(gotSteps, gotStepsAgain) {
				t.Fatalf("NextSteps() is not deterministic: %#v != %#v", gotSteps, gotStepsAgain)
			}

			gotCommandsAgain := Commands(tt.hypotheses, tt.registry)
			if !reflect.DeepEqual(gotCommands, gotCommandsAgain) {
				t.Fatalf("Commands() is not deterministic: %#v != %#v", gotCommands, gotCommandsAgain)
			}
		})
	}
}

func TestRegistryBackedRecommendations(t *testing.T) {
	t.Parallel()

	for _, class := range detection.IssueRegistry {
		class := class
		t.Run(class.Name, func(t *testing.T) {
			t.Parallel()

			hypotheses := []types.Hypothesis{{IssueClass: class.Name}}

			steps := NextSteps(hypotheses, detection.IssueRegistry)
			if steps == nil {
				t.Fatal("NextSteps() returned nil, want non-nil slice")
			}
			if len(class.NextSteps) > 0 && len(steps) == 0 {
				t.Fatal("NextSteps() returned empty slice for class with configured steps")
			}

			commands := Commands(hypotheses, detection.IssueRegistry)
			if commands == nil {
				t.Fatal("Commands() returned nil, want non-nil slice")
			}
			if len(class.Commands) > 0 && len(commands) == 0 {
				t.Fatal("Commands() returned empty slice for class with configured commands")
			}
		})
	}
}
