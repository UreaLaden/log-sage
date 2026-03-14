package normalize

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type errReader struct {
	used bool
}

func (e *errReader) Read(p []byte) (int, error) {
	if e.used {
		return 0, errors.New("simulated read error")
	}

	e.used = true
	n := copy(p, "level=info msg=ok\n")
	return n, nil
}

func TestParseLogfmtDatasets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		file       string
		wantCount  int
		validateAt func(t *testing.T, file string, index int, entry LogEntry)
	}{
		{
			name:      "simple logfmt",
			file:      "simple.logfmt.log",
			wantCount: 3,
			validateAt: func(t *testing.T, file string, index int, entry LogEntry) {
				t.Helper()

				type want struct {
					level   string
					message string
				}

				wantEntries := []want{
					{level: "info", message: "starting worker"},
					{level: "warn", message: "cache nearing memory limit"},
					{level: "error", message: "redis connection refused"},
				}
				if index >= len(wantEntries) {
					t.Fatalf("file=%s entry=%d raw=%q: unexpected extra parsed entry", file, index+1, entry.Raw)
				}
				if entry.Raw == "" {
					t.Fatalf("file=%s entry=%d: expected raw preservation", file, index+1)
				}
				if entry.Level != wantEntries[index].level {
					t.Fatalf("file=%s entry=%d raw=%q: expected level %q, got %q", file, index+1, entry.Raw, wantEntries[index].level, entry.Level)
				}
				if entry.Message != wantEntries[index].message {
					t.Fatalf("file=%s entry=%d raw=%q: expected message %q, got %q", file, index+1, entry.Raw, wantEntries[index].message, entry.Message)
				}
				if entry.Timestamp != nil {
					t.Fatalf("file=%s entry=%d raw=%q: expected nil timestamp, got %v", file, index+1, entry.Raw, entry.Timestamp)
				}
			},
		},
		{
			name:      "timestamped logfmt",
			file:      "timestamped.logfmt.log",
			wantCount: 3,
			validateAt: func(t *testing.T, file string, index int, entry LogEntry) {
				t.Helper()
				if entry.Timestamp == nil {
					t.Fatalf("file=%s entry=%d raw=%q: expected timestamp, got nil", file, index+1, entry.Raw)
				}
				if entry.Message == "" {
					t.Fatalf("file=%s entry=%d raw=%q: expected message, got empty", file, index+1, entry.Raw)
				}
			},
		},
		{
			name:      "mixed levels logfmt",
			file:      "mixed-levels.logfmt.log",
			wantCount: 5,
			validateAt: func(t *testing.T, file string, index int, entry LogEntry) {
				t.Helper()

				type want struct {
					level   string
					message string
				}

				wantEntries := []want{
					{level: "debug", message: "starting worker"},
					{level: "warn", message: "cache nearing memory limit"},
					{level: "panic", message: "runtime error: index out of range"},
					{level: "fatal", message: "unable to bind port 8080"},
					{level: "info", message: "component=server port=8080"},
				}
				if index >= len(wantEntries) {
					t.Fatalf("file=%s entry=%d raw=%q: unexpected extra parsed entry", file, index+1, entry.Raw)
				}
				if entry.Level != wantEntries[index].level {
					t.Fatalf("file=%s entry=%d raw=%q: expected level %q, got %q", file, index+1, entry.Raw, wantEntries[index].level, entry.Level)
				}
				if entry.Message != wantEntries[index].message {
					t.Fatalf("file=%s entry=%d raw=%q: expected message %q, got %q", file, index+1, entry.Raw, wantEntries[index].message, entry.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join("..", "..", "testdata", "logs", "logfmt", tt.file)
			file, err := os.Open(path)
			if err != nil {
				t.Fatalf("file=%s open error: %v", tt.file, err)
			}
			defer file.Close()

			entries, err := ParseLogfmt(context.Background(), file)
			if err != nil {
				t.Fatalf("file=%s parse error: %v", tt.file, err)
			}

			if len(entries) != tt.wantCount {
				t.Fatalf("file=%s: expected %d entries, got %d", tt.file, tt.wantCount, len(entries))
			}

			for i, entry := range entries {
				tt.validateAt(t, tt.file, i, entry)
			}
		})
	}
}

func TestParseLogfmtEdgeCases(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(strings.Join([]string{
		`level=info msg="starting worker" component=server`,
		`level=warn msg=cache_nearing_memory_limit`,
		`level=error brokenpair msg="redis connection refused"`,
		`level=info component=server port=8080`,
		``,
	}, "\n"))

	entries, err := ParseLogfmt(context.Background(), input)
	if err != nil {
		t.Fatalf("ParseLogfmt() error = %v", err)
	}

	if len(entries) != 4 {
		t.Fatalf("ParseLogfmt() entry count = %d, want 4", len(entries))
	}

	if entries[0].Fields["component"] != "server" {
		t.Fatalf("entries[0].Fields[component] = %q, want %q", entries[0].Fields["component"], "server")
	}

	if entries[1].Message != "cache_nearing_memory_limit" {
		t.Fatalf("entries[1].Message = %q, want %q", entries[1].Message, "cache_nearing_memory_limit")
	}

	if entries[2].Message != "redis connection refused" {
		t.Fatalf("entries[2].Message = %q, want %q", entries[2].Message, "redis connection refused")
	}

	if entries[3].Message != "component=server port=8080" {
		t.Fatalf("entries[3].Message = %q, want %q", entries[3].Message, "component=server port=8080")
	}
}

func TestParseLogfmtCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ParseLogfmt(ctx, strings.NewReader(`level=info msg="starting worker"`))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ParseLogfmt() error = %v, want %v", err, context.Canceled)
	}
}

func TestParseLogfmtCanceledContextEmptyInput(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ParseLogfmt(ctx, strings.NewReader(""))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ParseLogfmt() error = %v, want %v", err, context.Canceled)
	}
}

func TestParseLogfmtTrailingBackslash(t *testing.T) {
	t.Parallel()

	entries, err := ParseLogfmt(context.Background(), strings.NewReader(`level=info msg="trailing\`))
	if err != nil {
		t.Fatalf("ParseLogfmt() unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("ParseLogfmt() entry count = %d, want 1", len(entries))
	}

	if entries[0].Raw == "" {
		t.Fatalf("ParseLogfmt() entry.Raw must not be empty")
	}
}

func TestParseLogfmtScannerError(t *testing.T) {
	t.Parallel()

	_, err := ParseLogfmt(context.Background(), &errReader{})
	if err == nil {
		t.Fatal("ParseLogfmt() error = nil, want non-nil")
	}

	if errors.Is(err, io.EOF) {
		t.Fatalf("ParseLogfmt() error = %v, want non-EOF error", err)
	}
}

func TestParseLogfmtMalformedLineStillReturnsEntry(t *testing.T) {
	t.Parallel()

	entries, err := ParseLogfmt(context.Background(), strings.NewReader("brokenpair\n"))
	if err != nil {
		t.Fatalf("ParseLogfmt() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("ParseLogfmt() entry count = %d, want 1", len(entries))
	}

	if entries[0].Raw != "brokenpair" {
		t.Fatalf("entries[0].Raw = %q, want %q", entries[0].Raw, "brokenpair")
	}

	if entries[0].Fields != nil {
		t.Fatalf("entries[0].Fields = %#v, want nil", entries[0].Fields)
	}

	if entries[0].Message != "" {
		t.Fatalf("entries[0].Message = %q, want empty", entries[0].Message)
	}
}

func TestParseLogfmtIgnoresEmptyKeyAndUnknownLevel(t *testing.T) {
	t.Parallel()

	entries, err := ParseLogfmt(context.Background(), strings.NewReader(`=orphan level=trace msg="still running"`))
	if err != nil {
		t.Fatalf("ParseLogfmt() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("ParseLogfmt() entry count = %d, want 1", len(entries))
	}

	if entries[0].Level != "" {
		t.Fatalf("entries[0].Level = %q, want empty", entries[0].Level)
	}

	if entries[0].Message != "still running" {
		t.Fatalf("entries[0].Message = %q, want %q", entries[0].Message, "still running")
	}

	if _, ok := entries[0].Fields[""]; ok {
		t.Fatalf("entries[0].Fields contained empty key: %#v", entries[0].Fields)
	}
}

func TestDetectTimestampEdgeCases(t *testing.T) {
	t.Parallel()

	if got := detectTimestamp(""); got != "" {
		t.Fatalf("detectTimestamp(\"\") = %q, want empty", got)
	}

	if got := detectTimestamp("   "); got != "" {
		t.Fatalf("detectTimestamp(whitespace) = %q, want empty", got)
	}

	want := "2026-03-14T12:01:02.123456789Z"
	if got := detectTimestamp(want + " worker started"); got != want {
		t.Fatalf("detectTimestamp(RFC3339Nano line) = %q, want %q", got, want)
	}
}
