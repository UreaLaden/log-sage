package main

import (
	"fmt"

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
	return &cobra.Command{
		Use:   "pod <pod-name>",
		Short: "Analyze logs for a Kubernetes pod",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("k8s pod analysis is not implemented yet")
		},
	}
}
