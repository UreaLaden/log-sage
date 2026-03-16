package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestPrintCISummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		result      *types.AnalysisResult
		wantExact   string
		wantContain []string
		wantAbsent  []string
	}{
		{
			name: "oom result renders top cause confidence evidence and action",
			result: &types.AnalysisResult{
				TopCauses: []types.Hypothesis{
					{
						IssueClass: "OutOfMemory",
						Confidence: types.ConfidenceHigh,
						Evidence: []types.Evidence{
							{
								Signal:      "oom",
								Occurrences: 2,
								Examples: []string{
									"container terminated with OOMKilled",
									"kernel: out of memory",
								},
							},
						},
					},
				},
				RecommendedNextSteps: []string{"Increase memory limit"},
			},
			wantContain: []string{
				"Top Cause: OutOfMemory (high confidence)",
				"Evidence:",
				"- container terminated with OOMKilled",
				"- kernel: out of memory",
				"Recommended Action:",
				"- Increase memory limit",
			},
		},
		{
			name:      "no issues",
			result:    &types.AnalysisResult{},
			wantExact: "No issues detected.\n",
		},
		{
			name: "caps evidence lines at two",
			result: &types.AnalysisResult{
				TopCauses: []types.Hypothesis{
					{
						IssueClass: "Panic",
						Confidence: types.ConfidenceMedium,
						Evidence: []types.Evidence{
							{
								Examples: []string{"first", "second", "third"},
							},
						},
					},
				},
			},
			wantContain: []string{
				"- first",
				"- second",
			},
			wantAbsent: []string{"- third"},
		},
		{
			name: "falls back to signal when examples empty",
			result: &types.AnalysisResult{
				TopCauses: []types.Hypothesis{
					{
						IssueClass: "DNSFailure",
						Confidence: types.ConfidenceMedium,
						Evidence: []types.Evidence{
							{
								Signal: "lookup api.internal: no such host",
							},
						},
					},
				},
			},
			wantContain: []string{
				"Evidence:",
				"- lookup api.internal: no such host",
			},
		},
		{
			name: "omits recommended action when no next steps",
			result: &types.AnalysisResult{
				TopCauses: []types.Hypothesis{
					{
						IssueClass: "ConnectionRefused",
						Confidence: types.ConfidenceLow,
						Evidence: []types.Evidence{
							{
								Examples: []string{"dial tcp 10.0.0.8:5432: connect: connection refused"},
							},
						},
					},
				},
			},
			wantContain: []string{
				"Top Cause: ConnectionRefused (low confidence)",
			},
			wantAbsent: []string{"Recommended Action:"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			if err := printCISummary(&buf, tt.result); err != nil {
				t.Fatalf("printCISummary() error = %v", err)
			}

			got := buf.String()
			if tt.wantExact != "" && got != tt.wantExact {
				t.Fatalf("output = %q, want %q", got, tt.wantExact)
			}
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Fatalf("output = %q, want substring %q", got, want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Fatalf("output = %q, did not expect substring %q", got, absent)
				}
			}
		})
	}
}
