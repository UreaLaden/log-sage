package detection

import "github.com/Urealaden/log-sage-temp/pkg/types"

// Detector is implemented by each issue class that requires custom detection
// logic beyond the default pattern-matching wrapper.
type Detector interface {
	// Detect evaluates the provided SignalSet and returns zero or more
	// candidate hypotheses. Detect must not panic; an empty slice is
	// returned when no evidence supports this issue class.
	Detect(signals types.SignalSet) []types.Hypothesis
}
