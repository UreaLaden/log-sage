package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.0.0-dev"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var showVersion bool

	rootCmd := &cobra.Command{
		Use:           "logsage",
		Short:         "LogSage analyzes logs to identify likely root causes.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), cmd.Version); err != nil {
					return err
				}
				return nil
			}

			return cmd.Help()
		},
	}

	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print the logsage version")
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the logsage version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), version); err != nil {
				return err
			}
			return nil
		},
	}
}
