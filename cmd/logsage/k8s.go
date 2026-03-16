package main

import (
	"github.com/Urealaden/log-sage-temp/internal/adapters"
	"github.com/Urealaden/log-sage-temp/internal/engine"
	"github.com/Urealaden/log-sage-temp/pkg/types"
	"github.com/spf13/cobra"
)

func newK8sCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "Analyze Kubernetes-related logs and resources",
	}

	cmd.AddCommand(newK8sPodCmd())

	return cmd
}

func newK8sPodCmd() *cobra.Command {
	return newK8sPodCmdWith(adapters.New())
}

func newK8sPodCmdWith(adapter *adapters.KubectlAdapter) *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "pod <pod-name>",
		Short: "Analyze logs for a Kubernetes pod",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reader, err := adapter.Fetch(args[0], namespace)
			if err != nil {
				return err
			}

			result, err := engine.New().Analyze(cmd.Context(), types.DiagnosticInput{
				Reader:     reader,
				SourceType: "k8s",
				Metadata: map[string]string{
					"pod":       args[0],
					"namespace": namespace,
				},
			})
			if err != nil {
				return err
			}

			return printResult(cmd.OutOrStdout(), result)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")

	return cmd
}
