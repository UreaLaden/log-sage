package recommendation

import "github.com/Urealaden/log-sage-temp/pkg/types"

// Commands returns a deduplicated list of recommended debugging commands
// derived from the provided ranked hypotheses. Commands are collected in
// hypothesis rank order; duplicates are skipped on a first-seen basis.
// At most 5 commands are returned.
func Commands(hypotheses []types.Hypothesis, registry []types.IssueClass) []string {
	lookup := make(map[string]types.IssueClass, len(registry))
	for _, class := range registry {
		lookup[class.Name] = class
	}

	seen := map[string]struct{}{}
	commands := make([]string, 0, 5)

	for _, hypothesis := range hypotheses {
		class, ok := lookup[hypothesis.IssueClass]
		if !ok {
			continue
		}

		for _, command := range class.Commands {
			if _, exists := seen[command]; exists {
				continue
			}

			seen[command] = struct{}{}
			commands = append(commands, command)

			if len(commands) >= 5 {
				return commands
			}
		}
	}

	return commands
}
