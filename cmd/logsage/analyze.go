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

			if asJSON {
				return printJSON(cmd.OutOrStdout(), result)
			}

			return printResult(cmd.OutOrStdout(), result)
		},
	}

	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "Read log input from stdin")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output results as JSON")

	return cmd
}
