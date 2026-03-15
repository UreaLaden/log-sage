package main

import (
	"strings"
	"testing"
)

func TestK8sPodCmd(t *testing.T) {
	t.Parallel()

	t.Run("requires a pod name", func(t *testing.T) {
		t.Parallel()

		cmd := newK8sPodCmd()
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
			t.Fatalf("error = %q, want arg validation message", err.Error())
		}
	})

	t.Run("returns scaffold error with pod name", func(t *testing.T) {
		t.Parallel()

		cmd := newK8sPodCmd()
		cmd.SetArgs([]string{"demo-pod"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
		if err.Error() != "k8s pod analysis is not implemented yet" {
			t.Fatalf("error = %q, want %q", err.Error(), "k8s pod analysis is not implemented yet")
		}
	})
}

func TestK8sCmd(t *testing.T) {
	t.Parallel()

	t.Run("registers pod subcommand", func(t *testing.T) {
		t.Parallel()

		cmd := newK8sCmd()
		sub, _, err := cmd.Find([]string{"pod"})
		if err != nil {
			t.Fatalf("Find() error = %v, want nil", err)
		}
		if sub == nil {
			t.Fatal("Find() returned nil command, want pod subcommand")
		}
		if sub.Name() != "pod" {
			t.Fatalf("subcommand name = %q, want %q", sub.Name(), "pod")
		}
	})

	t.Run("is registered on root", func(t *testing.T) {
		t.Parallel()

		cmd := newRootCmd()
		sub, _, err := cmd.Find([]string{"k8s"})
		if err != nil {
			t.Fatalf("Find() error = %v, want nil", err)
		}
		if sub == nil {
			t.Fatal("Find() returned nil command, want k8s subcommand")
		}
		if sub.Name() != "k8s" {
			t.Fatalf("subcommand name = %q, want %q", sub.Name(), "k8s")
		}
	})
}
