package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var exit = os.Exit

func main() {
	exit(run())
}

func run() int {
	if err := newRootCmd().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func newRootCmd() *cobra.Command {
	var showVersion bool

	rootCmd := &cobra.Command{
		Use:           "logsage",
		Short:         "LogSage analyzes logs to identify likely root causes.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
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
	rootCmd.AddCommand(newAnalyzeCmd())
	rootCmd.AddCommand(newCICmd())
	rootCmd.AddCommand(newK8sCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the logsage version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), Version); err != nil {
				return err
			}
			return nil
		},
	}
}
