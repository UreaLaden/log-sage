package adapters

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestKubectlAdapterFetch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		runner      Runner
		wantErr     bool
		wantContain []string
	}{
		{
			name: "concatenates logs and describe output",
			runner: func(name string, args ...string) ([]byte, error) {
				switch args[0] {
				case "logs":
					return []byte("OOMKilled\n"), nil
				case "describe":
					return []byte("Status: OOMKilled\n"), nil
				default:
					return nil, fmt.Errorf("unexpected subcommand: %s", args[0])
				}
			},
			wantContain: []string{"OOMKilled", "Status: OOMKilled"},
		},
		{
			name: "kubectl logs error propagates",
			runner: func(name string, args ...string) ([]byte, error) {
				if args[0] == "logs" {
					return nil, fmt.Errorf("pod not found")
				}
				return []byte("ok"), nil
			},
			wantErr: true,
		},
		{
			name: "kubectl describe error propagates",
			runner: func(name string, args ...string) ([]byte, error) {
				if args[0] == "describe" {
					return nil, fmt.Errorf("pod not found")
				}
				return []byte("ok"), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewWithRunner(tt.runner)
			reader, err := adapter.Fetch("demo-pod", "default")
			if tt.wantErr {
				if err == nil {
					t.Fatal("Fetch() error = nil, want non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Fetch() error = %v, want nil", err)
			}

			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("io.ReadAll() error = %v", err)
			}
			got := string(data)
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Fatalf("output = %q, want substring %q", got, want)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	adapter := New()
	if adapter == nil {
		t.Fatal("New() = nil, want non-nil")
	}
	if adapter.runner == nil {
		t.Fatal("New().runner = nil, want non-nil")
	}
}

func TestDefaultRunner(t *testing.T) {
	t.Parallel()

	out, err := defaultRunner("sh", "-c", "printf ok")
	if err != nil {
		t.Fatalf("defaultRunner() error = %v, want nil", err)
	}
	if got := string(out); got != "ok" {
		t.Fatalf("defaultRunner() output = %q, want %q", got, "ok")
	}
}
