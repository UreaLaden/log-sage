package normalize

import (
	"strings"
	"time"
)

type LogEntry struct {
	Raw       string
	Timestamp *time.Time
	Level     string
	Message   string
	Fields    map[string]string
}

func ParsePlaintext(line string) LogEntry {
	entry := LogEntry{
		Raw: line,
	}

	remaining := strings.TrimSpace(line)
	if remaining == "" {
		return entry
	}

	if ts, rest, ok := parseLeadingTimestamp(remaining); ok {
		entry.Timestamp = ts
		remaining = rest
	}

	if level, rest, ok := parseLevelToken(remaining); ok {
		entry.Level = level
		remaining = rest
	}

	entry.Message = strings.TrimSpace(remaining)

	return entry
}

func parseLeadingTimestamp(line string) (*time.Time, string, bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, line, false
	}

	token := fields[0]
	parsed, err := time.Parse(time.RFC3339, token)
	if err != nil {
		return nil, line, false
	}

	rest := strings.TrimSpace(strings.TrimPrefix(line, token))

	return &parsed, rest, true
}

func parseLevelToken(line string) (string, string, bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", line, false
	}

	first := fields[0]
	switch {
	case strings.HasPrefix(strings.ToLower(first), "level="):
		level := normalizeLevel(strings.TrimPrefix(first, first[:len("level=")]))
		if level == "" {
			return "", line, false
		}

		rest := strings.TrimSpace(strings.TrimPrefix(line, first))

		return level, rest, true
	default:
		level := normalizeLevel(strings.Trim(first, "[]"))
		if level == "" {
			return "", line, false
		}

		rest := strings.TrimSpace(strings.TrimPrefix(line, first))

		return level, rest, true
	}
}

func normalizeLevel(token string) string {
	switch strings.ToLower(token) {
	case "error":
		return "error"
	case "warn", "warning":
		return "warn"
	case "info":
		return "info"
	case "debug":
		return "debug"
	case "fatal":
		return "fatal"
	case "panic":
		return "panic"
	default:
		return ""
	}
}
