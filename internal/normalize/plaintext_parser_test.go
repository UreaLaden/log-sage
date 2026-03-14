package normalize

import (
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
