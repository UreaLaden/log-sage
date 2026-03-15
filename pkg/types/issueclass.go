package types

// IssueClass describes a single detectable failure category. It holds the
// signal patterns used to detect the issue and the content used for
// explanation and recommendation output.
type IssueClass struct {
	// Name is the canonical issue class identifier.
	Name string

	// PrimarySignals are patterns whose matches contribute to the base
	// detection score.
	PrimarySignals []SignalPattern

	// CorroboratingSignals are patterns whose matches increase confidence
	// in an already-triggered hypothesis.
	CorroboratingSignals []SignalPattern

	// ExplanationTemplate is the human-readable explanation emitted when
	// this issue class is detected.
	ExplanationTemplate string

	// NextSteps is the ordered list of recommended debugging actions.
	NextSteps []string

	// Commands is the ordered list of suggested shell commands to aid
	// investigation.
	Commands []string
}
