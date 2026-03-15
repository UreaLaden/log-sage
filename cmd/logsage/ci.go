package main

import (
	"fmt"
	"os"

	"github.com/Urealaden/log-sage-temp/internal/engine"
	"github.com/Urealaden/log-sage-temp/pkg/types"
	"github.com/spf13/cobra"
)

func newCICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ci <log-file>",
		Short: "Analyze a CI log file and report likely root causes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
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

			result, err := engine.New().Analyze(cmd.Context(), types.DiagnosticInput{
				Reader:     file,
				SourceType: "ci",
			})
			if err != nil {
				return err
			}

			return printResult(cmd.OutOrStdout(), result)
		},
	}
}
