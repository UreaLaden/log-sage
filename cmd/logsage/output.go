package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

var classLabels = map[string]string{
	"OutOfMemory":               "Out of memory",
	"CrashLoopBackOff":          "Container crash loop",
	"ImagePullBackOff":          "Image pull failure",
	"ConnectionRefused":         "Connection refused",
	"DNSFailure":                "DNS resolution failure",
	"TLSFailure":                "TLS/certificate failure",
	"MissingEnvVar":             "Missing environment variable",
	"PermissionDenied":          "Permission denied",
	"DiskFull":                  "Disk full",
	"DependencyTimeout":         "Dependency timeout",
	"Panic":                     "Application panic",
	"PortBindingFailure":        "Port binding failure",
	"CIPermissionDenied":        "CI permission denied",
	"MissingSecretOrAuthFailure": "Missing secret or auth failure",
	"TestFailure":               "Test failure",
}

func humanLabel(issueClass string) string {
	if label, ok := classLabels[issueClass]; ok {
		return label
	}
	return issueClass
}

// extractTestName parses "--- FAIL: TestFoo (0.00s)" and returns "TestFoo".
// Returns empty string if the pattern is not found.
func extractTestName(examples []string) string {
	for _, ex := range examples {
		if strings.HasPrefix(ex, "--- FAIL:") {
			parts := strings.Fields(ex)
			if len(parts) >= 3 {
				return parts[2]
			}
		}
	}
	return ""
}

func printResult(w io.Writer, result *types.AnalysisResult) error {
	if len(result.TopCauses) == 0 {
		_, err := fmt.Fprintf(w, "No issues detected.\n")
		return err
	}

	if _, err := fmt.Fprintf(w, "Top Likely Causes\n\n"); err != nil {
		return err
	}

	for i, cause := range result.TopCauses {
		if _, err := fmt.Fprintf(w, "%d. %s (%s confidence)\n", i+1, cause.IssueClass, cause.Confidence); err != nil {
			return err
		}
		if cause.Explanation != "" {
			if _, err := fmt.Fprintf(w, "   %s\n", cause.Explanation); err != nil {
				return err
			}
		}
		if len(cause.Evidence) > 0 {
			if _, err := fmt.Fprintf(w, "\n   Evidence:\n"); err != nil {
				return err
			}
			for _, evidence := range cause.Evidence {
				if _, err := fmt.Fprintf(w, "   - %s: %d occurrence(s)\n", evidence.Signal, evidence.Occurrences); err != nil {
					return err
				}
				if len(evidence.Examples) > 0 {
					if _, err := fmt.Fprintf(w, "     > %s\n", evidence.Examples[0]); err != nil {
						return err
					}
				}
			}
		}
		if i < len(result.TopCauses)-1 {
			if _, err := fmt.Fprintf(w, "\n"); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "\n"); err != nil {
			return err
		}
	}

	if len(result.RecommendedNextSteps) > 0 {
		if _, err := fmt.Fprintf(w, "Next Steps\n\n"); err != nil {
			return err
		}
		for _, step := range result.RecommendedNextSteps {
			if _, err := fmt.Fprintf(w, "- %s\n", step); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "\n"); err != nil {
			return err
		}
	}

	if len(result.RecommendedCommands) > 0 {
		if _, err := fmt.Fprintf(w, "Recommended Commands\n\n"); err != nil {
			return err
		}
		for _, command := range result.RecommendedCommands {
			if _, err := fmt.Fprintf(w, "- %s\n", command); err != nil {
				return err
			}
		}
	}

	return nil
}

func printJSON(w io.Writer, result *types.AnalysisResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func printQuiet(w io.Writer, result *types.AnalysisResult) error {
	if len(result.TopCauses) == 0 {
		_, err := fmt.Fprintf(w, "No issues detected.\n")
		return err
	}

	for _, cause := range result.TopCauses {
		if _, err := fmt.Fprintf(w, "%s: %s confidence\n", cause.IssueClass, cause.Confidence); err != nil {
			return err
		}
	}

	return nil
}

func printCISummary(w io.Writer, result *types.AnalysisResult) error {
	if len(result.TopCauses) == 0 {
		_, err := fmt.Fprintf(w, "No issues detected.\n")
		return err
	}

	topCause := result.TopCauses[0]
	label := humanLabel(topCause.IssueClass)

	// Build headline — extract entity when available
	headline := label
	if topCause.IssueClass == "TestFailure" {
		var examples []string
		for _, ev := range topCause.Evidence {
			examples = append(examples, ev.Examples...)
		}
		if name := extractTestName(examples); name != "" {
			headline = label + " — `" + name + "`"
		}
	}

	// Line 1: headline (parseable by head -n 1)
	if _, err := fmt.Fprintf(w, "%s\n", headline); err != nil {
		return err
	}

	// Evidence lines
	evidenceLines := ciSummaryEvidence(topCause)
	if len(evidenceLines) > 0 {
		if _, err := fmt.Fprintln(w, "Evidence:"); err != nil {
			return err
		}
		for _, line := range evidenceLines {
			if _, err := fmt.Fprintf(w, "- %s\n", line); err != nil {
				return err
			}
		}
	}

	// Recommended action (first step only)
	if len(result.RecommendedNextSteps) > 0 {
		if _, err := fmt.Fprintln(w, "Recommended Action:"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "- %s\n", result.RecommendedNextSteps[0]); err != nil {
			return err
		}
	}

	return nil
}

func ciSummaryEvidence(cause types.Hypothesis) []string {
	lines := make([]string, 0, 2)
	for _, evidence := range cause.Evidence {
		for _, example := range evidence.Examples {
			lines = append(lines, example)
			if len(lines) == 2 {
				return lines
			}
		}
		if len(evidence.Examples) == 0 && evidence.Signal != "" {
			lines = append(lines, evidence.Signal)
			if len(lines) == 2 {
				return lines
			}
		}
	}

	return lines
}
