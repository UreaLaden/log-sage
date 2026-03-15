package normalize

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParsePlaintext(t *testing.T) {
	t.Parallel()

	mustTime := func(value string) *time.Time {
		t.Helper()

		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			t.Fatalf("time.Parse() error = %v", err)
		}

		return &parsed
	}

	tests := []struct {
		name    string
		input   string
		want    LogEntry
		wantNil bool
	}{
		{
			name:  "timestamp and message",
			input: "2026-03-14T12:01:02Z redis connection refused",
			want: LogEntry{
				Raw:       "2026-03-14T12:01:02Z redis connection refused",
				Timestamp: mustTime("2026-03-14T12:01:02Z"),
				Message:   "redis connection refused",
			},
		},
		{
			name:  "timestamp level and message",
			input: "2026-03-14T12:01:02Z ERROR redis connection refused",
			want: LogEntry{
				Raw:       "2026-03-14T12:01:02Z ERROR redis connection refused",
				Timestamp: mustTime("2026-03-14T12:01:02Z"),
				Level:     "error",
				Message:   "redis connection refused",
			},
		},
		{
			name:  "level without timestamp",
			input: "[WARNING] disk almost full",
			want: LogEntry{
				Raw:     "[WARNING] disk almost full",
				Level:   "warn",
				Message: "disk almost full",
			},
		},
		{
			name:  "message only",
			input: "service started successfully",
			want: LogEntry{
				Raw:     "service started successfully",
				Message: "service started successfully",
			},
		},
		{
			name:  "non parseable timestamp token",
			input: "2026-03-14T12:01:02 not-a-real-rfc3339 ERROR raw failure",
			want: LogEntry{
				Raw:     "2026-03-14T12:01:02 not-a-real-rfc3339 ERROR raw failure",
				Message: "2026-03-14T12:01:02 not-a-real-rfc3339 ERROR raw failure",
			},
		},
		{
			name:  "whitespace trimming with level syntax",
			input: "  2026-03-14T12:01:02Z    level=INFO    boot complete   ",
			want: LogEntry{
				Raw:       "  2026-03-14T12:01:02Z    level=INFO    boot complete   ",
				Timestamp: mustTime("2026-03-14T12:01:02Z"),
				Level:     "info",
				Message:   "boot complete",
			},
		},
		{
			name:  "fields remain nil",
			input: "INFO hello world",
			want: LogEntry{
				Raw:     "INFO hello world",
				Level:   "info",
				Message: "hello world",
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ParsePlaintext(tt.input)

			if got.Raw != tt.want.Raw {
				t.Fatalf("Raw = %q, want %q", got.Raw, tt.want.Raw)
			}

			if tt.want.Timestamp == nil {
				if got.Timestamp != nil {
					t.Fatalf("Timestamp = %v, want nil", got.Timestamp)
				}
			} else {
				if got.Timestamp == nil {
					t.Fatalf("Timestamp = nil, want %v", tt.want.Timestamp)
				}
				if !got.Timestamp.Equal(*tt.want.Timestamp) {
					t.Fatalf("Timestamp = %v, want %v", got.Timestamp, tt.want.Timestamp)
				}
			}

			if got.Level != tt.want.Level {
				t.Fatalf("Level = %q, want %q", got.Level, tt.want.Level)
			}

			if got.Message != tt.want.Message {
				t.Fatalf("Message = %q, want %q", got.Message, tt.want.Message)
			}

			if tt.wantNil && got.Fields != nil {
				t.Fatalf("Fields = %#v, want nil", got.Fields)
			}
		})
	}
}

func TestParsePlaintextDatasets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		file     string
		validate func(t *testing.T, file string, lineNo int, entryIndex int, raw string, entry LogEntry)
	}{
		{
			name: "plaintext basic logs",
			file: "plaintext-basic.log",
			validate: func(t *testing.T, file string, lineNo int, entryIndex int, raw string, entry LogEntry) {
				t.Helper()
				_ = entryIndex

				if entry.Raw != raw {
					t.Fatalf("file=%s line=%d raw=%q: expected raw preservation, got %q", file, lineNo, raw, entry.Raw)
				}
				if entry.Timestamp == nil {
					t.Fatalf("file=%s line=%d raw=%q: expected timestamp, got nil", file, lineNo, raw)
				}
				if strings.TrimSpace(entry.Message) == "" {
					t.Fatalf("file=%s line=%d raw=%q: expected non-empty message", file, lineNo, raw)
				}
				if entry.Message != strings.TrimSpace(entry.Message) {
					t.Fatalf("file=%s line=%d raw=%q: expected trimmed message, got %q", file, lineNo, raw, entry.Message)
				}
				if entry.Level != "" && entry.Level != "error" && entry.Level != "info" {
					t.Fatalf("file=%s line=%d raw=%q: expected level error/info/empty, got %q", file, lineNo, raw, entry.Level)
				}
				if strings.Contains(entry.Message, "ERROR ") || strings.Contains(entry.Message, "INFO ") {
					t.Fatalf("file=%s line=%d raw=%q: expected message without level token, got %q", file, lineNo, raw, entry.Message)
				}
			},
		},
		{
			name: "plaintext level variants",
			file: "plaintext-level-variants.log",
			validate: func(t *testing.T, file string, lineNo int, entryIndex int, raw string, entry LogEntry) {
				t.Helper()

				expected := []struct {
					level   string
					message string
				}{
					{level: "error", message: "redis connection refused"},
					{level: "warn", message: "cache nearing memory limit"},
					{level: "debug", message: "starting worker"},
					{level: "panic", message: "runtime error: index out of range"},
					{level: "fatal", message: "unable to bind port 8080"},
				}

				if entryIndex < 1 || entryIndex > len(expected) {
					t.Fatalf("file=%s line=%d raw=%q: no expected entry for logical index %d", file, lineNo, raw, entryIndex)
				}
				want := expected[entryIndex-1]

				if entry.Raw != raw {
					t.Fatalf("file=%s line=%d raw=%q: expected raw preservation, got %q", file, lineNo, raw, entry.Raw)
				}
				if entry.Timestamp == nil {
					t.Fatalf("file=%s line=%d raw=%q: expected timestamp, got nil", file, lineNo, raw)
				}
				if entry.Level != want.level {
					t.Fatalf("file=%s line=%d raw=%q: expected level %q, got %q", file, lineNo, raw, want.level, entry.Level)
				}
				if entry.Message != want.message {
					t.Fatalf("file=%s line=%d raw=%q: expected message %q, got %q", file, lineNo, raw, want.message, entry.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join("..", "..", "testdata", "logs", tt.file)
			file, err := os.Open(path)
			if err != nil {
				t.Fatalf("file=%s open error: %v", tt.file, err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					t.Errorf("close file: %v", err)
				}
			}()

			scanner := bufio.NewScanner(file)
			entryIndex := 0
			for i := 1; scanner.Scan(); i++ {
				raw := scanner.Text()
				if strings.TrimSpace(raw) == "" {
					continue
				}

				entryIndex++
				entry := ParsePlaintext(raw)
				tt.validate(t, tt.file, i, entryIndex, raw, entry)
			}

			if err := scanner.Err(); err != nil {
				t.Fatalf("file=%s scan error: %v", tt.file, err)
			}
		})
	}
}
