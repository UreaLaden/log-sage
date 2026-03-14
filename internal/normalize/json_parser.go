package normalize

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var errJSONNotObject = errors.New("normalize: line is not a JSON object")

func ParseJSON(line string) (LogEntry, error) {
	entry := LogEntry{
		Raw: line,
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return entry, errJSONNotObject
	}

	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.UseNumber()

	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return entry, err
	}

	object, ok := decoded.(map[string]any)
	if !ok {
		return entry, errJSONNotObject
	}

	entry.Fields = stringifyFields(object)
	entry.Timestamp = extractJSONTimestamp(object)
	entry.Level = extractJSONLevel(object)
	entry.Message = extractJSONMessage(object)
	if entry.Message == "" {
		entry.Message = compactJSON(trimmed)
	}

	return entry, nil
}

func stringifyFields(object map[string]any) map[string]string {
	if len(object) == 0 {
		return map[string]string{}
	}

	fields := make(map[string]string, len(object))
	for key, value := range object {
		fields[key] = stringifyJSONValue(value)
	}

	return fields
}

func stringifyJSONValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case string:
		return typed
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case json.Number:
		return typed.String()
	case float64:
		return fmt.Sprintf("%v", typed)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return string(encoded)
	}
}

func extractJSONTimestamp(object map[string]any) *time.Time {
	for _, key := range []string{"timestamp", "time", "ts", "@timestamp"} {
		raw, ok := object[key]
		if !ok {
			continue
		}

		value, ok := raw.(string)
		if !ok {
			continue
		}

		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			parsed, err := time.Parse(layout, value)
			if err == nil {
				return &parsed
			}
		}
	}

	return nil
}

func extractJSONLevel(object map[string]any) string {
	for _, key := range []string{"level", "severity", "log.level"} {
		raw, ok := object[key]
		if !ok {
			continue
		}

		value, ok := raw.(string)
		if !ok {
			continue
		}

		if level := normalizeLevel(value); level != "" {
			return level
		}
	}

	return ""
}

func extractJSONMessage(object map[string]any) string {
	for _, key := range []string{"message", "msg", "log", "event"} {
		raw, ok := object[key]
		if !ok {
			continue
		}

		value, ok := raw.(string)
		if !ok {
			continue
		}

		return strings.TrimSpace(value)
	}

	return ""
}

func compactJSON(line string) string {
	var compacted bytes.Buffer
	if err := json.Compact(&compacted, []byte(line)); err != nil {
		return line
	}

	return compacted.String()
}
