package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Urealaden/log-sage-temp/internal/engine"
	"github.com/Urealaden/log-sage-temp/pkg/types"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var fromStdin bool
	var asJSON bool
	var quiet bool
	var topN int

	cmd := &cobra.Command{
		Use:   "analyze [<file>]",
		Short: "Analyze a log file and report likely root causes",
		Args: func(cmd *cobra.Command, args []string) error {
			if fromStdin {
				if len(args) > 0 {
					return fmt.Errorf("--stdin and a file argument are mutually exclusive")
				}
				return nil
			}
			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if asJSON && quiet {
				return fmt.Errorf("--json and --quiet are mutually exclusive")
			}

			var reader io.Reader
			if fromStdin {
				reader = cmd.InOrStdin()
			} else {
				file, err := os.Open(args[0])
				if err != nil {
					return fmt.Errorf("open %s: %w", args[0], err)
				}
				defer func() {
					closeErr := file.Close()
					if err == nil && closeErr != nil {
						err = closeErr
					}
				}()
				reader = file
			}

			result, err := engine.New().Analyze(cmd.Context(), types.DiagnosticInput{
				Reader: reader,
			})
			if err != nil {
				return err
			}
			if topN < 0 {
				return fmt.Errorf("--top must be a positive integer")
			}
			if topN > 0 && topN < len(result.TopCauses) {
				truncated := append([]types.Hypothesis(nil), result.TopCauses[:topN]...)
				result = &types.AnalysisResult{
					Summary:              result.Summary,
					TopCauses:            truncated,
					Unknowns:             result.Unknowns,
					RecommendedCommands:  result.RecommendedCommands,
					RecommendedNextSteps: result.RecommendedNextSteps,
				}
			}

			switch {
			case asJSON:
				return printJSON(cmd.OutOrStdout(), result)
			case quiet:
				return printQuiet(cmd.OutOrStdout(), result)
			default:
				return printResult(cmd.OutOrStdout(), result)
			}
		},
	}

	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "Read log input from stdin")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output results as JSON")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Output a compact one-line summary per issue")
	cmd.Flags().IntVar(&topN, "top", 0, "Limit output to top N results (0 = all)")

	return cmd
}
