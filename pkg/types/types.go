package types

import "io"

type DiagnosticInput struct {
	Reader     io.Reader
	SourceType string
	Metadata   map[string]string
}

type AnalysisResult struct {
	Summary              string
	TopCauses            []Hypothesis
	Unknowns             []string
	RecommendedCommands  []string
	RecommendedNextSteps []string
}

type Hypothesis struct {
	IssueClass  string
	Confidence  ConfidenceLevel
	Phase       FailurePhase
	Score       float64
	Evidence    []Evidence
	Explanation string
}

type Evidence struct {
	Signal      string
	Occurrences int
	Examples    []string
}

type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "high"
	ConfidenceMedium ConfidenceLevel = "medium"
	ConfidenceLow    ConfidenceLevel = "low"
)

type FailurePhase string

const (
	FailurePhaseImagePull      FailurePhase = "image_pull"
	FailurePhaseStartup        FailurePhase = "startup"
	FailurePhaseInitialization FailurePhase = "initialization"
	FailurePhaseRuntime        FailurePhase = "runtime"
	FailurePhaseShutdown       FailurePhase = "shutdown"
)
