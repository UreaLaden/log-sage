package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

type corpusExpected struct {
	IssueClasses     []string `json:"issue_classes"`
	MinConfidence    string   `json:"min_confidence"`
	RequiredEvidence []string `json:"required_evidence"`
}

func TestCorpusIncidents(t *testing.T) {
	t.Parallel()

	const incidentsDir = "../../testdata/incidents"

	entries, err := os.ReadDir(incidentsDir)
	if err != nil {
		t.Fatalf("os.ReadDir(%q) error = %v", incidentsDir, err)
	}
	if len(entries) == 0 {
		t.Fatal("no incident directories found under testdata/incidents")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entry := entry
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(incidentsDir, entry.Name())

			expected := loadCorpusExpected(t, filepath.Join(dir, "expected.json"))

			logFile, err := os.Open(filepath.Join(dir, "logs.txt"))
			if err != nil {
				t.Fatalf("os.Open(logs.txt) error = %v", err)
			}
			t.Cleanup(func() {
				if closeErr := logFile.Close(); closeErr != nil {
					t.Errorf("logFile.Close() error = %v", closeErr)
				}
			})

			result, err := New().Analyze(context.Background(), types.DiagnosticInput{Reader: logFile})
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}
			if result == nil {
				t.Fatal("Analyze() result = nil, want non-nil")
			}

			if len(expected.IssueClasses) == 0 {
				if len(result.TopCauses) != 0 {
					classes := make([]string, len(result.TopCauses))
					for i, cause := range result.TopCauses {
						classes[i] = cause.IssueClass
					}
					t.Fatalf("TopCauses = %v, want empty for negative sample", classes)
				}
				return
			}

			causesByClass := make(map[string]types.Hypothesis, len(result.TopCauses))
			for _, cause := range result.TopCauses {
				if _, exists := causesByClass[cause.IssueClass]; !exists {
					causesByClass[cause.IssueClass] = cause
				}
			}

			minConfidence := types.ConfidenceLevel(expected.MinConfidence)
			for _, className := range expected.IssueClasses {
				cause, ok := causesByClass[className]
				if !ok {
					t.Errorf("issue class %q not found in TopCauses", className)
					continue
				}
				if minConfidence != "" && !corpusConfidenceMet(cause.Confidence, minConfidence) {
					t.Errorf("issue class %q confidence = %q, want at least %q", className, cause.Confidence, minConfidence)
				}
			}

			allEvidence := corpusEvidenceText(result.TopCauses)
			for _, required := range expected.RequiredEvidence {
				if !strings.Contains(allEvidence, required) {
					t.Errorf("required evidence %q not found in result", required)
				}
			}
		})
	}
}

func loadCorpusExpected(t *testing.T, path string) corpusExpected {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}

	var expected corpusExpected
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", path, err)
	}

	return expected
}

func corpusEvidenceText(hypotheses []types.Hypothesis) string {
	var allText strings.Builder

	for _, hypothesis := range hypotheses {
		for _, evidence := range hypothesis.Evidence {
			allText.WriteString(evidence.Signal)
			allText.WriteString("\n")
			for _, example := range evidence.Examples {
				allText.WriteString(example)
				allText.WriteString("\n")
			}
		}
	}

	return allText.String()
}

func corpusConfidenceMet(actual, min types.ConfidenceLevel) bool {
	order := map[types.ConfidenceLevel]int{
		types.ConfidenceLow:    1,
		types.ConfidenceMedium: 2,
		types.ConfidenceHigh:   3,
	}

	return order[actual] >= order[min]
}
