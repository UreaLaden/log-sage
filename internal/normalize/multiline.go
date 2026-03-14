package normalize

// GroupMultiline merges continuation lines into their parent entry.
// A continuation line is any line whose Raw field begins with a space
// or horizontal tab. Continuation lines are appended to the previous
// Line.Raw separated by "\n". If the first line of the input is a
// continuation line it is treated as a standalone entry.
// The input slice is not modified.
func GroupMultiline(lines []Line) []Line {
	if len(lines) == 0 {
		return make([]Line, 0)
	}

	grouped := make([]Line, 0, len(lines))
	for _, line := range lines {
		if isContinuationLine(line.Raw) && len(grouped) > 0 {
			grouped[len(grouped)-1].Raw += "\n" + line.Raw
			continue
		}

		grouped = append(grouped, line)
	}

	return grouped
}

func isContinuationLine(raw string) bool {
	if raw == "" {
		return false
	}

	return raw[0] == ' ' || raw[0] == '\t'
}
