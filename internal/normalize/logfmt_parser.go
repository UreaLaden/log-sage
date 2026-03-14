package normalize

import (
	"bufio"
	"context"
	"io"
	"strings"
	"time"
)

type logfmtField struct {
	key   string
	value string
}

// ParseLogfmt reads logfmt-formatted lines from r and returns a slice of
// LogEntry values. Empty lines are skipped. ctx cancellation is respected
// between lines; if ctx is done the function returns immediately with
// ctx.Err().
func ParseLogfmt(ctx context.Context, r io.Reader) ([]LogEntry, error) {
	scanner := bufio.NewScanner(r)
	entries := make([]LogEntry, 0)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		raw := strings.TrimRight(scanner.Text(), "\r\n")
		if strings.TrimSpace(raw) == "" {
			continue
		}

		entries = append(entries, parseLogfmtLine(raw))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return entries, nil
}

func parseLogfmtLine(line string) LogEntry {
	entry := LogEntry{
		Raw: line,
	}

	fields := tokenizeLogfmt(line)
	if len(fields) == 0 {
		return entry
	}

	entry.Fields = make(map[string]string, len(fields))
	for _, field := range fields {
		entry.Fields[field.key] = field.value
	}

	entry.Timestamp = extractLogfmtTimestamp(fields)
	entry.Level = extractLogfmtLevel(fields)
	entry.Message = extractLogfmtMessage(fields)
	if entry.Message == "" {
		entry.Message = reconstructLogfmtMessage(fields)
	}

	return entry
}

func tokenizeLogfmt(line string) []logfmtField {
	fields := make([]logfmtField, 0)
	i := 0

	for i < len(line) {
		for i < len(line) && line[i] == ' ' {
			i++
		}
		if i >= len(line) {
			break
		}

		keyStart := i
		for i < len(line) && line[i] != '=' && line[i] != ' ' {
			i++
		}
		if i >= len(line) || line[i] != '=' {
			for i < len(line) && line[i] != ' ' {
				i++
			}
			continue
		}

		key := line[keyStart:i]
		i++ // skip '='
		if key == "" {
			continue
		}

		value := ""
		if i < len(line) && line[i] == '"' {
			i++
			var builder strings.Builder
			closed := false
			for i < len(line) {
				switch line[i] {
				case '\\':
					if i+1 < len(line) {
						builder.WriteByte(line[i+1])
						i += 2
						continue
					}
					i++
				case '"':
					i++
					closed = true
					break
				default:
					builder.WriteByte(line[i])
					i++
				}
				if closed {
					break
				}
			}
			value = builder.String()
		} else {
			valueStart := i
			for i < len(line) && line[i] != ' ' {
				i++
			}
			value = line[valueStart:i]
		}

		fields = append(fields, logfmtField{key: key, value: value})
	}

	return fields
}

func extractLogfmtTimestamp(fields []logfmtField) *time.Time {
	for _, key := range []string{"timestamp", "time", "ts"} {
		for _, field := range fields {
			if field.key != key {
				continue
			}

			for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
				parsed, err := time.Parse(layout, field.value)
				if err == nil {
					return &parsed
				}
			}
		}
	}

	return nil
}

func extractLogfmtLevel(fields []logfmtField) string {
	for _, field := range fields {
		if field.key != "level" {
			continue
		}
		if level := normalizeLevel(field.value); level != "" {
			return level
		}
	}

	return ""
}

func extractLogfmtMessage(fields []logfmtField) string {
	for _, key := range []string{"msg", "message"} {
		for _, field := range fields {
			if field.key == key {
				return strings.TrimSpace(field.value)
			}
		}
	}

	return ""
}

func reconstructLogfmtMessage(fields []logfmtField) string {
	parts := make([]string, 0)
	for _, field := range fields {
		switch field.key {
		case "timestamp", "time", "ts", "level", "msg", "message":
			continue
		default:
			parts = append(parts, field.key+"="+field.value)
		}
	}

	return strings.Join(parts, " ")
}
