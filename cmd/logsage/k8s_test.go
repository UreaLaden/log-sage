package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Urealaden/log-sage-temp/internal/adapters"
)

func TestK8sPodCmd(t *testing.T) {
	t.Parallel()

	t.Run("requires a pod name", func(t *testing.T) {
		t.Parallel()

		cmd := newK8sPodCmdWith(adapters.NewWithRunner(func(name string, args ...string) ([]byte, error) {
			return nil, nil
		}))
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
			t.Fatalf("error = %q, want arg validation message", err.Error())
		}
	})

	t.Run("adapter error propagates", func(t *testing.T) {
		t.Parallel()

		cmd := newK8sPodCmdWith(adapters.NewWithRunner(func(name string, args ...string) ([]byte, error) {
			if args[0] == "logs" {
				return nil, fmt.Errorf("pod not found")
			}
			return []byte("ok"), nil
		}))
		cmd.SetArgs([]string{"demo-pod"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "pod not found") {
			t.Fatalf("error = %q, want propagated adapter error", err.Error())
		}
	})

	t.Run("successful analysis uses adapter output", func(t *testing.T) {
		t.Parallel()

		cmd := newK8sPodCmdWith(adapters.NewWithRunner(func(name string, args ...string) ([]byte, error) {
			switch args[0] {
			case "logs":
				return []byte("OOMKilled\n"), nil
			case "describe":
				return []byte("Status: OOMKilled\n"), nil
			default:
				return nil, fmt.Errorf("unexpected subcommand: %s", args[0])
			}
		}))
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetArgs([]string{"demo-pod"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		output := stdout.String()
		for _, want := range []string{
			"Top Likely Causes",
			"OutOfMemory (high confidence)",
			"Recommended Commands",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("output = %q, want substring %q", output, want)
			}
		}
	})

	t.Run("namespace flag defaults to default", func(t *testing.T) {
		t.Parallel()

		var seen []string
		cmd := newK8sPodCmdWith(adapters.NewWithRunner(func(name string, args ...string) ([]byte, error) {
			seen = append([]string(nil), args...)
			switch args[0] {
			case "logs":
				return []byte("OOMKilled\n"), nil
			case "describe":
				return []byte("Status: OOMKilled\n"), nil
			default:
				return nil, fmt.Errorf("unexpected subcommand: %s", args[0])
			}
		}))

		flag := cmd.Flag("namespace")
		if flag == nil {
			t.Fatal("namespace flag = nil, want non-nil")
		}
		if flag.DefValue != "default" {
			t.Fatalf("namespace default = %q, want %q", flag.DefValue, "default")
		}

		cmd.SetArgs([]string{"demo-pod"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
		if len(seen) == 0 {
			t.Fatal("runner was not invoked")
		}
		if !strings.Contains(strings.Join(seen, " "), "default") {
			t.Fatalf("runner args = %v, want default namespace", seen)
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
