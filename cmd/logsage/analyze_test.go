package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

type failAfterWriter struct {
	remaining int
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	if w.remaining == 0 {
		return 0, errors.New("write failed")
	}
	w.remaining--
	return len(p), nil
}

func TestAnalyzeCmd(t *testing.T) {
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
			name:        "empty log prints no issues",
			filename:    "empty.log",
			log:         "",
			wantContain: []string{"No issues detected."},
		},
		{
			name:     "oom log prints cause next steps and commands",
			filename: "oom.log",
			log:      "OOMKilled\n",
			wantContain: []string{
				"Top Likely Causes",
				"OutOfMemory (high confidence)",
				"Evidence:",
				"Next Steps",
				"Recommended Commands",
			},
		},
		{
			name:        "missing file returns error",
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

			cmd := newAnalyzeCmd()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
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

func TestAnalyzeCmdJSON(t *testing.T) {
	t.Parallel()

	t.Run("json with oom log is valid and contains top causes", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "oom.log")
		if err := os.WriteFile(path, []byte("OOMKilled\n"), 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		cmd := newAnalyzeCmd()
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
		if !strings.Contains(stdout.String(), "\"TopCauses\"") {
			t.Fatalf("output = %q, want TopCauses key", stdout.String())
		}
	})

	t.Run("json with empty log returns valid json with empty top causes", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "empty.log")
		if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		cmd := newAnalyzeCmd()
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
		if result.TopCauses == nil {
			t.Fatalf("TopCauses = nil, want empty slice")
		}
		if len(result.TopCauses) != 0 {
			t.Fatalf("len(TopCauses) = %d, want 0", len(result.TopCauses))
		}
	})
}

func TestAnalyzeCmdErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("stdin with valid log content succeeds", func(t *testing.T) {
		t.Parallel()

		cmd := newAnalyzeCmd()
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetIn(strings.NewReader("OOMKilled\n"))
		cmd.SetArgs([]string{"--stdin"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		output := stdout.String()
		for _, want := range []string{
			"Top Likely Causes",
			"OutOfMemory (high confidence)",
			"Next Steps",
			"Recommended Commands",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("output = %q, want substring %q", output, want)
			}
		}
	})

	t.Run("stdin with empty content prints no issues", func(t *testing.T) {
		t.Parallel()

		cmd := newAnalyzeCmd()
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--stdin"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
		if got := stdout.String(); got != "No issues detected.\n" {
			t.Fatalf("output = %q, want %q", got, "No issues detected.\n")
		}
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "oom.log")
		if err := os.WriteFile(path, []byte("OOMKilled\n"), 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		cmd := newAnalyzeCmd()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		cmd.SetContext(ctx)
		cmd.SetArgs([]string{path})

		if err := cmd.Execute(); err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
	})

	t.Run("print result write error is returned", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "oom.log")
		if err := os.WriteFile(path, []byte("OOMKilled\n"), 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		cmd := newAnalyzeCmd()
		cmd.SetOut(failingWriter{})
		cmd.SetArgs([]string{path})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "write failed") {
			t.Fatalf("error = %q, want substring %q", err.Error(), "write failed")
		}
	})

	t.Run("missing arg returns cobra arg error", func(t *testing.T) {
		t.Parallel()

		cmd := newAnalyzeCmd()
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
			t.Fatalf("error = %q, want arg validation message", err.Error())
		}
	})

	t.Run("stdin and positional arg are mutually exclusive", func(t *testing.T) {
		t.Parallel()

		cmd := newAnalyzeCmd()
		cmd.SetIn(strings.NewReader("OOMKilled\n"))
		cmd.SetArgs([]string{"--stdin", "oom.log"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "mutually exclusive") {
			t.Fatalf("error = %q, want mutual exclusivity message", err.Error())
		}
	})
}

func TestPrintResult(t *testing.T) {
	t.Parallel()

	t.Run("prints no issues detected for empty result", func(t *testing.T) {
		t.Parallel()

		var out bytes.Buffer
		if err := printResult(&out, &types.AnalysisResult{}); err != nil {
			t.Fatalf("printResult() error = %v", err)
		}
		if got := out.String(); got != "No issues detected.\n" {
			t.Fatalf("output = %q, want %q", got, "No issues detected.\n")
		}
	})

	t.Run("prints full human readable sections", func(t *testing.T) {
		t.Parallel()

		result := &types.AnalysisResult{
			TopCauses: []types.Hypothesis{
				{
					IssueClass:  "OutOfMemory",
					Confidence:  types.ConfidenceHigh,
					Explanation: "The container exceeded its memory limit.",
					Evidence: []types.Evidence{
						{Signal: "oom-killed", Occurrences: 2, Examples: []string{"OOMKilled"}},
					},
				},
			},
			RecommendedNextSteps: []string{"Inspect pod restart history."},
			RecommendedCommands:  []string{"kubectl describe pod <pod>"},
		}

		var out bytes.Buffer
		if err := printResult(&out, result); err != nil {
			t.Fatalf("printResult() error = %v", err)
		}

		output := out.String()
		for _, want := range []string{
			"Top Likely Causes\n\n",
			"1. OutOfMemory (high confidence)",
			"   The container exceeded its memory limit.",
			"\n   Evidence:\n",
			"- oom-killed: 2 occurrence(s)",
			"     > OOMKilled",
			"Next Steps\n\n",
			"Recommended Commands\n\n",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("output = %q, want substring %q", output, want)
			}
		}
	})

	t.Run("omits empty recommendation sections", func(t *testing.T) {
		t.Parallel()

		result := &types.AnalysisResult{
			TopCauses: []types.Hypothesis{
				{IssueClass: "ConnectionRefused", Confidence: types.ConfidenceMedium},
			},
		}

		var out bytes.Buffer
		if err := printResult(&out, result); err != nil {
			t.Fatalf("printResult() error = %v", err)
		}

		output := out.String()
		if strings.Contains(output, "Next Steps\n\n") {
			t.Fatalf("output = %q, did not expect Next Steps section", output)
		}
		if strings.Contains(output, "Recommended Commands\n\n") {
			t.Fatalf("output = %q, did not expect Recommended Commands section", output)
		}
		if strings.Contains(output, "Next Steps") {
			t.Fatalf("output = %q, did not expect Next Steps section", output)
		}
		if strings.Contains(output, "Recommended Commands") {
			t.Fatalf("output = %q, did not expect Recommended Commands section", output)
		}
	})

	t.Run("prints cause without explanation or evidence", func(t *testing.T) {
		t.Parallel()

		result := &types.AnalysisResult{
			TopCauses: []types.Hypothesis{
				{IssueClass: "DNSFailure", Confidence: types.ConfidenceLow},
			},
		}

		var out bytes.Buffer
		if err := printResult(&out, result); err != nil {
			t.Fatalf("printResult() error = %v", err)
		}

		output := out.String()
		if strings.Contains(output, "Evidence:") {
			t.Fatalf("output = %q, did not expect Evidence section", output)
		}
	})

	t.Run("returns writer error", func(t *testing.T) {
		t.Parallel()

		result := &types.AnalysisResult{
			TopCauses: []types.Hypothesis{
				{IssueClass: "OutOfMemory", Confidence: types.ConfidenceHigh},
			},
		}

		err := printResult(failingWriter{}, result)
		if err == nil {
			t.Fatal("printResult() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "write failed") {
			t.Fatalf("error = %q, want substring %q", err.Error(), "write failed")
		}
	})

	t.Run("returns writer errors across later sections", func(t *testing.T) {
		t.Parallel()

		result := &types.AnalysisResult{
			TopCauses: []types.Hypothesis{
				{
					IssueClass:  "OutOfMemory",
					Confidence:  types.ConfidenceHigh,
					Explanation: "The container exceeded its memory limit.",
					Evidence: []types.Evidence{
						{Signal: "oom-killed", Occurrences: 2, Examples: []string{"OOMKilled"}},
					},
				},
			},
			RecommendedNextSteps: []string{"Inspect pod restart history."},
			RecommendedCommands:  []string{"kubectl describe pod <pod>"},
		}

		for i := 0; i < 12; i++ {
			writer := &failAfterWriter{remaining: i}
			if err := printResult(writer, result); err == nil {
				t.Fatalf("printResult() error = nil for failAfter=%d, want non-nil", i)
			}
		}
	})
}

func TestRootCmdIncludesAnalyze(t *testing.T) {
	t.Parallel()

	root := newRootCmd()
	if root.Commands()[0].Use == "" {
		t.Fatal("root command has no subcommands")
	}

	found := false
	for _, cmd := range root.Commands() {
		if cmd.Use == "analyze [<file>]" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("analyze command not registered on root")
	}
}

func TestRootCmdHelpAndVersion(t *testing.T) {
	t.Parallel()

	t.Run("root help", func(t *testing.T) {
		t.Parallel()

		root := newRootCmd()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetArgs([]string{})

		if err := root.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(out.String(), "LogSage analyzes logs") {
			t.Fatalf("output = %q, want help text", out.String())
		}
	})

	t.Run("root version flag", func(t *testing.T) {
		t.Parallel()

		root := newRootCmd()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetArgs([]string{"--version"})

		if err := root.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if got := out.String(); got != "logsage version "+version+"\n" {
			t.Fatalf("output = %q, want %q", got, "logsage version "+version+"\n")
		}
	})

	t.Run("version subcommand", func(t *testing.T) {
		t.Parallel()

		root := newRootCmd()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetArgs([]string{"version"})

		if err := root.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if got := out.String(); got != version+"\n" {
			t.Fatalf("output = %q, want %q", got, version+"\n")
		}
	})

	t.Run("root version flag write error", func(t *testing.T) {
		t.Parallel()

		root := newRootCmd()
		root.SetOut(failingWriter{})
		root.SetArgs([]string{"--version"})

		err := root.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "write failed") {
			t.Fatalf("error = %q, want substring %q", err.Error(), "write failed")
		}
	})

	t.Run("version subcommand write error", func(t *testing.T) {
		t.Parallel()

		root := newRootCmd()
		root.SetOut(failingWriter{})
		root.SetArgs([]string{"version"})

		err := root.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "write failed") {
			t.Fatalf("error = %q, want substring %q", err.Error(), "write failed")
		}
	})
}

func TestRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		oldArgs := os.Args
		oldStdout := os.Stdout
		defer func() {
			os.Args = oldArgs
			os.Stdout = oldStdout
		}()

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe() error = %v", err)
		}

		os.Args = []string{"logsage", "version"}
		os.Stdout = w

		code := run()

		if err := w.Close(); err != nil {
			t.Fatalf("w.Close() error = %v", err)
		}
		if code != 0 {
			t.Fatalf("run() = %d, want 0", code)
		}
		var out bytes.Buffer
		if _, err := io.Copy(&out, r); err != nil {
			t.Fatalf("io.Copy() error = %v", err)
		}
		if got := out.String(); got != version+"\n" {
			t.Fatalf("output = %q, want %q", got, version+"\n")
		}
	})

	t.Run("help", func(t *testing.T) {
		oldArgs := os.Args
		oldStdout := os.Stdout
		defer func() {
			os.Args = oldArgs
			os.Stdout = oldStdout
		}()

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe() error = %v", err)
		}

		os.Args = []string{"logsage"}
		os.Stdout = w

		code := run()

		if err := w.Close(); err != nil {
			t.Fatalf("w.Close() error = %v", err)
		}
		if code != 0 {
			t.Fatalf("run() = %d, want 0", code)
		}
		var out bytes.Buffer
		if _, err := io.Copy(&out, r); err != nil {
			t.Fatalf("io.Copy() error = %v", err)
		}
		if !strings.Contains(out.String(), "LogSage analyzes logs") {
			t.Fatalf("output = %q, want help text", out.String())
		}
	})

	t.Run("error", func(t *testing.T) {
		oldArgs := os.Args
		oldStderr := os.Stderr
		defer func() {
			os.Args = oldArgs
			os.Stderr = oldStderr
		}()

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe() error = %v", err)
		}

		os.Args = []string{"logsage", "analyze", "/definitely/missing.log"}
		os.Stderr = w

		code := run()

		if err := w.Close(); err != nil {
			t.Fatalf("w.Close() error = %v", err)
		}
		if code != 1 {
			t.Fatalf("run() = %d, want 1", code)
		}

		var out bytes.Buffer
		if _, err := out.ReadFrom(r); err != nil {
			t.Fatalf("ReadFrom() error = %v", err)
		}
		if !strings.Contains(out.String(), "open") {
			t.Fatalf("stderr = %q, want open error", out.String())
		}
	})
}

func TestMainUsesRunExitCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		oldArgs := os.Args
		oldStdout := os.Stdout
		oldExit := exit
		defer func() {
			os.Args = oldArgs
			os.Stdout = oldStdout
			exit = oldExit
		}()

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe() error = %v", err)
		}

		os.Args = []string{"logsage", "version"}
		os.Stdout = w

		gotCode := -1
		exit = func(code int) {
			gotCode = code
		}

		main()

		if err := w.Close(); err != nil {
			t.Fatalf("w.Close() error = %v", err)
		}
		if gotCode != 0 {
			t.Fatalf("exit code = %d, want 0", gotCode)
		}
		var out bytes.Buffer
		if _, err := io.Copy(&out, r); err != nil {
			t.Fatalf("io.Copy() error = %v", err)
		}
		if got := out.String(); got != version+"\n" {
			t.Fatalf("output = %q, want %q", got, version+"\n")
		}
	})

	t.Run("error", func(t *testing.T) {
		oldArgs := os.Args
		oldStderr := os.Stderr
		oldExit := exit
		defer func() {
			os.Args = oldArgs
			os.Stderr = oldStderr
			exit = oldExit
		}()

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("os.Pipe() error = %v", err)
		}

		os.Args = []string{"logsage", "analyze", "/definitely/missing.log"}
		os.Stderr = w

		gotCode := -1
		exit = func(code int) {
			gotCode = code
		}

		main()

		if err := w.Close(); err != nil {
			t.Fatalf("w.Close() error = %v", err)
		}
		if gotCode != 1 {
			t.Fatalf("exit code = %d, want 1", gotCode)
		}
		var out bytes.Buffer
		if _, err := io.Copy(&out, r); err != nil {
			t.Fatalf("io.Copy() error = %v", err)
		}
		if !strings.Contains(out.String(), "open") {
			t.Fatalf("stderr = %q, want open error", out.String())
		}
	})
}
