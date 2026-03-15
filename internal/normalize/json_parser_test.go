package normalize

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseJSON(t *testing.T) {
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
		name          string
		input         string
		wantTimestamp *time.Time
		wantLevel     string
		wantMessage   string
		wantFields    map[string]string
	}{
		{
			name:          "valid json with timestamp level and message",
			input:         `{"timestamp":"2026-03-14T12:01:02Z","level":"ERROR","message":"redis connection refused"}`,
			wantTimestamp: mustTime("2026-03-14T12:01:02Z"),
			wantLevel:     "error",
			wantMessage:   "redis connection refused",
			wantFields: map[string]string{
				"timestamp": "2026-03-14T12:01:02Z",
				"level":     "ERROR",
				"message":   "redis connection refused",
			},
		},
		{
			name:        "valid json with message only",
			input:       `{"message":"worker ready"}`,
			wantMessage: "worker ready",
			wantFields: map[string]string{
				"message": "worker ready",
			},
		},
		{
			name:          "alternate keys",
			input:         `{"time":"2026-03-14T12:01:02Z","severity":"warning","msg":"disk almost full"}`,
			wantTimestamp: mustTime("2026-03-14T12:01:02Z"),
			wantLevel:     "warn",
			wantMessage:   "disk almost full",
			wantFields: map[string]string{
				"time":     "2026-03-14T12:01:02Z",
				"severity": "warning",
				"msg":      "disk almost full",
			},
		},
		{
			name:        "invalid timestamp value",
			input:       `{"timestamp":"not-a-timestamp","level":"info","message":"still parses"}`,
			wantLevel:   "info",
			wantMessage: "still parses",
			wantFields: map[string]string{
				"timestamp": "not-a-timestamp",
				"level":     "info",
				"message":   "still parses",
			},
		},
		{
			name:        "extra fields preserved",
			input:       `{"message":"retrying","attempt":3,"service":"redis","nested":{"host":"db"}}`,
			wantMessage: "retrying",
			wantFields: map[string]string{
				"message": "retrying",
				"attempt": "3",
				"service": "redis",
				"nested":  `{"host":"db"}`,
			},
		},
		{
			name:        "large integer preserved exactly",
			input:       `{"message":"retrying","request_id":9007199254740993}`,
			wantMessage: "retrying",
			wantFields: map[string]string{
				"message":    "retrying",
				"request_id": "9007199254740993",
			},
		},
		{
			name:        "fallback message uses compact json",
			input:       ` { "service":"api", "ok":true } `,
			wantMessage: `{"service":"api","ok":true}`,
			wantFields: map[string]string{
				"service": "api",
				"ok":      "true",
			},
		},
		{
			name:        "whitespace handling around json",
			input:       "\n\t{\"log.level\":\"DEBUG\",\"event\":\" cache warmed \"}\n",
			wantLevel:   "debug",
			wantMessage: "cache warmed",
			wantFields: map[string]string{
				"log.level": "DEBUG",
				"event":     " cache warmed ",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseJSON(tt.input)
			if err != nil {
				t.Fatalf("ParseJSON() error = %v", err)
			}

			if got.Raw != tt.input {
				t.Fatalf("Raw = %q, want %q", got.Raw, tt.input)
			}

			if tt.wantTimestamp == nil {
				if got.Timestamp != nil {
					t.Fatalf("Timestamp = %v, want nil", got.Timestamp)
				}
			} else {
				if got.Timestamp == nil {
					t.Fatalf("Timestamp = nil, want %v", tt.wantTimestamp)
				}
				if !got.Timestamp.Equal(*tt.wantTimestamp) {
					t.Fatalf("Timestamp = %v, want %v", got.Timestamp, tt.wantTimestamp)
				}
			}

			if got.Level != tt.wantLevel {
				t.Fatalf("Level = %q, want %q", got.Level, tt.wantLevel)
			}

			if got.Message != tt.wantMessage {
				t.Fatalf("Message = %q, want %q", got.Message, tt.wantMessage)
			}

			for key, want := range tt.wantFields {
				if got.Fields[key] != want {
					t.Fatalf("Fields[%q] = %q, want %q", key, got.Fields[key], want)
				}
			}
		})
	}
}

func TestParseJSONRejectsNonJSONObject(t *testing.T) {
	t.Parallel()

	line := "ERROR redis connection refused"

	_, err := ParseJSON(line)
	if err == nil {
		t.Fatal("ParseJSON() error = nil, want error")
	}

	if errors.Is(err, errJSONNotObject) {
		t.Fatal("ParseJSON() returned object error for invalid JSON syntax, want decoder error")
	}

	plaintext := ParsePlaintext(line)
	if plaintext.Level != "error" || plaintext.Message != "redis connection refused" {
		t.Fatalf("ParsePlaintext() = %#v, want level=%q message=%q", plaintext, "error", "redis connection refused")
	}
}

func TestParseJSONRejectsNonObjectJSON(t *testing.T) {
	t.Parallel()

	_, err := ParseJSON(`["not","an","object"]`)
	if !errors.Is(err, errJSONNotObject) {
		t.Fatalf("ParseJSON() error = %v, want %v", err, errJSONNotObject)
	}
}

func TestParseJSONDatasets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		file              string
		wantLevels        map[string]bool
		wantMessage       string
		requiredFieldKeys []string
	}{
		{
			name:        "basic json logs",
			file:        "json-basic.log",
			wantLevels:  map[string]bool{"error": true, "info": true},
			wantMessage: "redis connection refused",
			requiredFieldKeys: []string{
				"service",
			},
		},
		{
			name:        "alternate json fields",
			file:        "json-alt-fields.log",
			wantLevels:  map[string]bool{"info": true, "warn": true},
			wantMessage: "disk almost full",
			requiredFieldKeys: []string{
				"service",
				"host",
			},
		},
		{
			name:        "mixed json field aliases",
			file:        "json-mixed-keys.log",
			wantLevels:  map[string]bool{"error": true, "warn": true, "info": true},
			wantMessage: "retrying TLS handshake",
			requiredFieldKeys: []string{
				"service",
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
			seenLevel := make(map[string]bool, len(tt.wantLevels))
			seenMessage := false

			for i := 1; scanner.Scan(); i++ {
				line := scanner.Text()
				if strings.TrimSpace(line) == "" {
					continue
				}

				entry, err := ParseJSON(line)
				if err != nil {
					t.Fatalf("file=%s line=%d parse error: %v raw=%q", tt.file, i, err, line)
				}

				if entry.Timestamp == nil {
					t.Fatalf("file=%s line=%d missing timestamp raw=%q", tt.file, i, line)
				}

				if entry.Level == "" {
					t.Fatalf("file=%s line=%d missing level raw=%q", tt.file, i, line)
				}
				if tt.wantLevels[entry.Level] {
					seenLevel[entry.Level] = true
				}

				if strings.TrimSpace(entry.Message) == "" {
					t.Fatalf("file=%s line=%d missing message raw=%q", tt.file, i, line)
				}
				if entry.Message == tt.wantMessage {
					seenMessage = true
				}

				for _, key := range tt.requiredFieldKeys {
					if entry.Fields[key] == "" {
						t.Fatalf("file=%s line=%d missing field=%s raw=%q", tt.file, i, key, line)
					}
				}
			}

			if err := scanner.Err(); err != nil {
				t.Fatalf("file=%s scan error: %v", tt.file, err)
			}

			for level := range tt.wantLevels {
				if !seenLevel[level] {
					t.Fatalf("file=%s expected to observe normalized level %q", tt.file, level)
				}
			}

			if !seenMessage {
				t.Fatalf("file=%s expected to observe message %q", tt.file, tt.wantMessage)
			}
		})
	}
}
