package normalize

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGroupMultilineDataset(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "testdata", "logs", "stacktrace-sample.log")
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open stacktrace fixture: %v", err)
	}
	defer file.Close()

	lines, err := Normalize(context.Background(), file)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	grouped := GroupMultiline(lines)
	if len(grouped) != 1 {
		t.Fatalf("GroupMultiline() entry count = %d, want 1", len(grouped))
	}

	if grouped[0].Timestamp != "2026-03-14T12:00:01Z" {
		t.Fatalf("grouped[0].Timestamp = %q, want %q", grouped[0].Timestamp, "2026-03-14T12:00:01Z")
	}

	if !strings.Contains(grouped[0].Raw, "panic: redis unavailable") {
		t.Fatalf("grouped[0].Raw missing parent content: %q", grouped[0].Raw)
	}

	if !strings.Contains(grouped[0].Raw, "    goroutine 1 [running]:") {
		t.Fatalf("grouped[0].Raw missing continuation content: %q", grouped[0].Raw)
	}

	if !strings.Contains(grouped[0].Raw, "\n    main.main()") {
		t.Fatalf("grouped[0].Raw missing newline join semantics: %q", grouped[0].Raw)
	}
}

func TestGroupMultilineUnit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []Line
		want  []Line
	}{
		{
			name: "no continuation lines",
			input: []Line{
				{Raw: "first", Timestamp: "2026-03-14T12:00:01Z"},
				{Raw: "second", Timestamp: "2026-03-14T12:00:02Z"},
			},
			want: []Line{
				{Raw: "first", Timestamp: "2026-03-14T12:00:01Z"},
				{Raw: "second", Timestamp: "2026-03-14T12:00:02Z"},
			},
		},
		{
			name: "one parent with multiple continuations",
			input: []Line{
				{Raw: "panic: boom", Timestamp: "2026-03-14T12:00:01Z"},
				{Raw: "  frame one"},
				{Raw: "\tframe two"},
			},
			want: []Line{
				{Raw: "panic: boom\n  frame one\n\tframe two", Timestamp: "2026-03-14T12:00:01Z"},
			},
		},
		{
			name: "multiple independent groups",
			input: []Line{
				{Raw: "first", Timestamp: "2026-03-14T12:00:01Z"},
				{Raw: "  detail one"},
				{Raw: "second", Timestamp: "2026-03-14T12:00:02Z"},
				{Raw: "\tdetail two"},
			},
			want: []Line{
				{Raw: "first\n  detail one", Timestamp: "2026-03-14T12:00:01Z"},
				{Raw: "second\n\tdetail two", Timestamp: "2026-03-14T12:00:02Z"},
			},
		},
		{
			name: "first line is a continuation",
			input: []Line{
				{Raw: "  orphan detail"},
				{Raw: "root", Timestamp: "2026-03-14T12:00:02Z"},
			},
			want: []Line{
				{Raw: "  orphan detail"},
				{Raw: "root", Timestamp: "2026-03-14T12:00:02Z"},
			},
		},
		{
			name:  "empty input returns non-nil empty slice",
			input: nil,
			want:  []Line{},
		},
		{
			name: "single non-continuation line",
			input: []Line{
				{Raw: "single", Timestamp: "2026-03-14T12:00:01Z"},
			},
			want: []Line{
				{Raw: "single", Timestamp: "2026-03-14T12:00:01Z"},
			},
		},
		{
			name: "tab-prefixed continuation lines",
			input: []Line{
				{Raw: "panic: tab", Timestamp: "2026-03-14T12:00:01Z"},
				{Raw: "\tstack line"},
			},
			want: []Line{
				{Raw: "panic: tab\n\tstack line", Timestamp: "2026-03-14T12:00:01Z"},
			},
		},
		{
			name: "mixed space and tab continuation lines",
			input: []Line{
				{Raw: "panic: mixed", Timestamp: "2026-03-14T12:00:01Z"},
				{Raw: "  frame one"},
				{Raw: "\tframe two"},
				{Raw: "next", Timestamp: "2026-03-14T12:00:02Z"},
			},
			want: []Line{
				{Raw: "panic: mixed\n  frame one\n\tframe two", Timestamp: "2026-03-14T12:00:01Z"},
				{Raw: "next", Timestamp: "2026-03-14T12:00:02Z"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GroupMultiline(tt.input)
			if got == nil {
				t.Fatal("GroupMultiline() returned nil slice")
			}

			if len(got) != len(tt.want) {
				t.Fatalf("GroupMultiline() len = %d, want %d", len(got), len(tt.want))
			}

			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("GroupMultiline()[%d] = %#v, want %#v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGroupMultilineDoesNotMutate(t *testing.T) {
	t.Parallel()

	input := []Line{
		{Raw: "panic: boom", Timestamp: "2026-03-14T12:00:01Z"},
		{Raw: "  frame one"},
		{Raw: "\tframe two"},
	}
	original := append([]Line(nil), input...)

	grouped := GroupMultiline(input)
	if len(grouped) != 1 {
		t.Fatalf("GroupMultiline() len = %d, want 1", len(grouped))
	}

	for i := range original {
		if input[i] != original[i] {
			t.Fatalf("input[%d] mutated: got %#v want %#v", i, input[i], original[i])
		}
	}
}
