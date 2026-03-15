package recommendation

import "github.com/Urealaden/log-sage-temp/pkg/types"

// NextSteps returns a deduplicated list of recommended next steps derived from
// the provided ranked hypotheses. Steps are collected in hypothesis rank order;
// duplicates are skipped on a first-seen basis. The registry parameter is the
// source of next-step text for each issue class.
func NextSteps(hypotheses []types.Hypothesis, registry []types.IssueClass) []string {
	lookup := make(map[string]types.IssueClass, len(registry))
	for _, class := range registry {
		lookup[class.Name] = class
	}

	seen := map[string]struct{}{}
	steps := make([]string, 0, 5)

	for _, hypothesis := range hypotheses {
		class, ok := lookup[hypothesis.IssueClass]
		if !ok {
			continue
		}

		for _, step := range class.NextSteps {
			if _, exists := seen[step]; exists {
				continue
			}

			seen[step] = struct{}{}
			steps = append(steps, step)

			if len(steps) >= 5 {
				return steps
			}
		}
	}

	return steps
}
