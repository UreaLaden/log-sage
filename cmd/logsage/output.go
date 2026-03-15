package main

import (
	"fmt"
	"io"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

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
