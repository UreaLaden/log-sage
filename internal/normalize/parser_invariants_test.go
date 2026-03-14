package normalize

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParserInvariants(t *testing.T) {
	t.Parallel()

	t.Run("plaintext", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			file string
		}{
			{name: "basic", file: "plaintext-basic.log"},
			{name: "level variants", file: "plaintext-level-variants.log"},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				lines := readNonEmptyLines(t, filepath.Join("..", "..", "testdata", "logs", tt.file))
				for i, line := range lines {
					entry := ParsePlaintext(line)
					assertParserInvariants(t, "plaintext", tt.file, i+1, line, entry)
				}
			})
		}
	})

	t.Run("json", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			file string
		}{
			{name: "basic", file: "json-basic.log"},
			{name: "alternate fields", file: "json-alt-fields.log"},
			{name: "mixed keys", file: "json-mixed-keys.log"},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				lines := readNonEmptyLines(t, filepath.Join("..", "..", "testdata", "logs", tt.file))
				for i, line := range lines {
					entry, err := ParseJSON(line)
					if err != nil {
						t.Fatalf("parser=json file=%s line=%d raw=%q: ParseJSON() error = %v", tt.file, i+1, line, err)
					}

					assertParserInvariants(t, "json", tt.file, i+1, line, entry)
				}
			})
		}
	})

	t.Run("logfmt", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			file string
		}{
			{name: "simple", file: "simple.logfmt.log"},
			{name: "timestamped", file: "timestamped.logfmt.log"},
			{name: "mixed levels", file: "mixed-levels.logfmt.log"},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				path := filepath.Join("..", "..", "testdata", "logs", "logfmt", tt.file)
				lines := readNonEmptyLines(t, path)

				file, err := os.Open(path)
				if err != nil {
					t.Fatalf("parser=logfmt file=%s: open error = %v", tt.file, err)
				}
				defer file.Close()

				entries, err := ParseLogfmt(context.Background(), file)
				if err != nil {
					t.Fatalf("parser=logfmt file=%s: ParseLogfmt() error = %v", tt.file, err)
				}

				if len(entries) != len(lines) {
					t.Fatalf("parser=logfmt file=%s: entry count = %d, want %d", tt.file, len(entries), len(lines))
				}

				for i, entry := range entries {
					assertParserInvariants(t, "logfmt", tt.file, i+1, lines[i], entry)
				}
			})
		}
	})
}

func readNonEmptyLines(t *testing.T, path string) []string {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lines := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}

	return lines
}

func assertParserInvariants(t *testing.T, parser, file string, index int, raw string, entry LogEntry) {
	t.Helper()

	if entry.Raw != raw {
		t.Fatalf("parser=%s file=%s entry=%d raw=%q: Raw = %q, want %q", parser, file, index, raw, entry.Raw, raw)
	}

	if entry.Level != "" && !isNormalizedLevel(entry.Level) {
		t.Fatalf("parser=%s file=%s entry=%d raw=%q: Level = %q violates normalization contract", parser, file, index, raw, entry.Level)
	}

	if entry.Timestamp != nil && entry.Timestamp.IsZero() {
		t.Fatalf("parser=%s file=%s entry=%d raw=%q: Timestamp must not be zero", parser, file, index, raw)
	}
}

func isNormalizedLevel(level string) bool {
	switch level {
	case "error", "warn", "info", "debug", "fatal", "panic":
		return true
	default:
		return false
	}
}
