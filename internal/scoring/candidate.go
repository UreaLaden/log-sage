package scoring

import "github.com/Urealaden/log-sage-temp/pkg/types"

// CandidateHypothesis represents a detected issue class enriched with
// scoring metadata before conversion to the public types.Hypothesis output.
type CandidateHypothesis struct {
	// Class is the matched issue class from the detector registry.
	Class types.IssueClass

	// BaseScore is the raw score produced during detection.
	BaseScore float64

	// Evidence contains the signal evidence supporting this hypothesis.
	Evidence []types.Evidence

	// Signals contains the full evaluated signal set.
	// This is retained as scoring input context.
	Signals types.SignalSet

	// Phase is the inferred failure phase for this hypothesis.
	Phase types.FailurePhase

	// IsSymptom marks whether this hypothesis represents a secondary symptom
	// rather than the primary root cause.
	IsSymptom bool

	// Confidence is the mapped confidence level for this hypothesis,
	// populated by MapConfidence. Zero value means not yet mapped.
	Confidence types.ConfidenceLevel
}
