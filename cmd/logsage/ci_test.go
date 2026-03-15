package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestCICmd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		log         string
		filename    string
		missingFile bool
		wantErr     bool
		wantContain []string
	}{
		{
			name:     "ci log file with oom content succeeds",
			filename: "ci.log",
			log:      "OOMKilled\n",
			wantContain: []string{
				"Top Likely Causes",
				"OutOfMemory",
			},
		},
		{
			name:        "missing file returns open error",
			filename:    "missing.log",
			missingFile: true,
			wantErr:     true,
			wantContain: []string{"open"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, tt.filename)
			if !tt.missingFile {
				if err := os.WriteFile(path, []byte(tt.log), 0o644); err != nil {
					t.Fatalf("os.WriteFile() error = %v", err)
				}
			}

			cmd := newCICmd()
			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetArgs([]string{path})

			err := cmd.Execute()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Execute() error = nil, want non-nil")
				}
				for _, want := range tt.wantContain {
					if !strings.Contains(err.Error(), want) {
						t.Fatalf("error = %q, want substring %q", err.Error(), want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v, want nil", err)
			}

			output := stdout.String()
			for _, want := range tt.wantContain {
				if !strings.Contains(output, want) {
					t.Fatalf("output = %q, want substring %q", output, want)
				}
			}
		})
	}
}

func TestCICmdJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "ci.log")
	if err := os.WriteFile(path, []byte("OOMKilled\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cmd := newCICmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--json", path})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	var result types.AnalysisResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil; output=%q", err, stdout.String())
	}
	if len(result.TopCauses) == 0 {
		t.Fatalf("TopCauses = %v, want non-empty", result.TopCauses)
	}
	if result.TopCauses[0].IssueClass != "OutOfMemory" {
		t.Fatalf("TopCauses[0].IssueClass = %q, want %q", result.TopCauses[0].IssueClass, "OutOfMemory")
	}
}

func TestCICmdQuiet(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "ci.log")
	if err := os.WriteFile(path, []byte("OOMKilled\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cmd := newCICmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--quiet", path})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	if got := stdout.String(); got != "OutOfMemory: high confidence\n" {
		t.Fatalf("output = %q, want %q", got, "OutOfMemory: high confidence\n")
	}
}

func TestCICmdTop(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "ci.log")
	log := "CrashLoopBackOff\nOOMKilled\n"
	if err := os.WriteFile(path, []byte(log), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cmd := newCICmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--top", "1", path})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "OutOfMemory (high confidence)") {
		t.Fatalf("output = %q, want top-ranked cause", output)
	}
	if strings.Contains(output, "CrashLoopBackOff") {
		t.Fatalf("output = %q, did not expect lower-ranked cause", output)
	}
}

func TestCICmdRootRegistration(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	sub, _, err := cmd.Find([]string{"ci"})
	if err != nil {
		t.Fatalf("Find() error = %v, want nil", err)
	}
	if sub == nil {
		t.Fatal("Find() returned nil command, want ci subcommand")
	}
	if sub.Name() != "ci" {
		t.Fatalf("subcommand name = %q, want %q", sub.Name(), "ci")
	}
}

func TestCICmdNoArg(t *testing.T) {
	t.Parallel()

	cmd := newCICmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Fatalf("error = %q, want arg validation message", err.Error())
	}
}
