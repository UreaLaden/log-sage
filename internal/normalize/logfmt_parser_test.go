package normalize

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

				wantLevels := []string{"info", "warn", "error"}
				wantMessages := []string{"starting worker", "cache nearing memory limit", "redis connection refused"}
				if entry.Raw == "" {
					t.Fatalf("file=%s entry=%d: expected raw preservation", file, index+1)
				}
				if entry.Level != wantLevels[index] {
					t.Fatalf("file=%s entry=%d raw=%q: expected level %q, got %q", file, index+1, entry.Raw, wantLevels[index], entry.Level)
				}
				if entry.Message != wantMessages[index] {
					t.Fatalf("file=%s entry=%d raw=%q: expected message %q, got %q", file, index+1, entry.Raw, wantMessages[index], entry.Message)
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

				wantLevels := []string{"debug", "warn", "panic", "fatal", "info"}
				if entry.Level != wantLevels[index] {
					t.Fatalf("file=%s entry=%d raw=%q: expected level %q, got %q", file, index+1, entry.Raw, wantLevels[index], entry.Level)
				}
				if index == 4 {
					if entry.Message != "component=server port=8080" {
						t.Fatalf("file=%s entry=%d raw=%q: expected reconstructed message, got %q", file, index+1, entry.Raw, entry.Message)
					}
				} else if entry.Message == "" {
					t.Fatalf("file=%s entry=%d raw=%q: expected message, got empty", file, index+1, entry.Raw)
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
