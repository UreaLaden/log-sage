package normalize

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(
		"2026-03-14T14:00:00Z service started\n" +
			"\n" +
			"plain log line\r\n" +
			"\r\n" +
			"2026-03-14T14:00:01-04:00 request handled",
	)

	got, err := Normalize(context.Background(), input)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	want := []Line{
		{
			Raw:       "2026-03-14T14:00:00Z service started",
			Timestamp: "2026-03-14T14:00:00Z",
		},
		{
			Raw:       "plain log line",
			Timestamp: "",
		},
		{
			Raw:       "2026-03-14T14:00:01-04:00 request handled",
			Timestamp: "2026-03-14T14:00:01-04:00",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Normalize() mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestNormalizeCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Normalize(ctx, strings.NewReader("line one\nline two\n"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Normalize() error = %v, want %v", err, context.Canceled)
	}
}
