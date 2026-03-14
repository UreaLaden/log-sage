package normalize

import (
	"bufio"
	"context"
	"io"
	"strings"
	"time"
)

type Line struct {
	Raw       string
	Timestamp string
}

func Normalize(ctx context.Context, r io.Reader) ([]Line, error) {
	scanner := bufio.NewScanner(r)
	lines := make([]Line, 0)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		raw := strings.TrimRight(scanner.Text(), "\r\n")
		if raw == "" {
			continue
		}

		lines = append(lines, Line{
			Raw:       raw,
			Timestamp: detectTimestamp(raw),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return lines, nil
}

func detectTimestamp(line string) string {
	if line == "" {
		return ""
	}

	prefix := strings.Fields(line)
	if len(prefix) == 0 {
		return ""
	}

	candidate := prefix[0]
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if _, err := time.Parse(layout, candidate); err == nil {
			return candidate
		}
	}

	return ""
}
