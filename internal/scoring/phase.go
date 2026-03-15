package scoring

import (
	"strings"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

type phaseRule struct {
	phase   types.FailurePhase
	matches func(CandidateHypothesis) bool
}

var phasePrecedence = []phaseRule{
	{phase: types.FailurePhaseImagePull, matches: matchesImagePull},
	{phase: types.FailurePhaseStartup, matches: matchesStartup},
	{phase: types.FailurePhaseInitialization, matches: matchesInitialization},
	{phase: types.FailurePhaseRuntime, matches: matchesRuntime},
	{phase: types.FailurePhaseShutdown, matches: matchesShutdown},
}

// InferPhase deterministically infers the earliest applicable failure phase for
// a candidate. It returns the zero value when no phase can be inferred.
func InferPhase(candidate CandidateHypothesis) types.FailurePhase {
	for _, rule := range phasePrecedence {
		if rule.matches(candidate) {
			return rule.phase
		}
	}

	return ""
}

func matchesImagePull(candidate CandidateHypothesis) bool {
	return candidateHasAny(candidate,
		"imagepullbackoff",
		"errimagepull",
		"pull access denied",
	)
}

func matchesStartup(candidate CandidateHypothesis) bool {
	return candidateHasClass(candidate,
		"CrashLoopBackOff",
		"MissingEnvVar",
		"PortBindingFailure",
	) || candidateHasAny(candidate,
		"panic on boot",
		"missing environment variable",
		"required env",
		"config not found",
		"startup error",
		"panic: getenv",
		"crashloopbackoff",
		"back-off restarting failed container",
		"address already in use",
		"bind: ",
	)
}

func matchesInitialization(candidate CandidateHypothesis) bool {
	return candidateHasAny(candidate,
		"db migration",
		"migration failed",
		"migration error",
		"failed to migrate",
		"schema setup",
		"schema initialization",
		"schema init",
		"database schema",
	)
}

func matchesRuntime(candidate CandidateHypothesis) bool {
	return candidateHasClass(candidate,
		"OutOfMemory",
		"ConnectionRefused",
		"DNSFailure",
		"TLSFailure",
		"DependencyTimeout",
		"Panic",
		"DiskFull",
	) || candidateHasAny(candidate,
		"connection refused",
		"econnrefused",
		"dial tcp",
		"no such host",
		"nxdomain",
		"temporary failure in name resolution",
		"x509:",
		"tls:",
		"context deadline exceeded",
		"timeout after",
		"network timeout",
		"i/o timeout",
		"oomkilled",
		"out of memory",
		"cannot allocate memory",
		"exit code 137",
		"runtime error:",
		"panic:",
		"no space left on device",
		"enospc",
		"disk quota exceeded",
	)
}

func matchesShutdown(candidate CandidateHypothesis) bool {
	return candidateHasAny(candidate,
		"graceful shutdown",
		"sigterm",
		"received signal terminated",
		"received sigterm",
		"shutdown complete",
		"shutting down",
	)
}

func candidateHasClass(candidate CandidateHypothesis, classNames ...string) bool {
	for _, className := range classNames {
		if candidate.Class.Name == className {
			return true
		}
	}

	return false
}

func candidateHasAny(candidate CandidateHypothesis, phrases ...string) bool {
	haystacks := candidateHaystacks(candidate)
	for _, phrase := range phrases {
		for _, haystack := range haystacks {
			if strings.Contains(haystack, phrase) {
				return true
			}
		}
	}

	return false
}

func candidateHaystacks(candidate CandidateHypothesis) []string {
	haystacks := make([]string, 0, len(candidate.Signals.Matches)*3+1)
	if candidate.Class.Name != "" {
		haystacks = append(haystacks, strings.ToLower(candidate.Class.Name))
	}

	for _, match := range candidate.Signals.Matches {
		if match.PatternName != "" {
			haystacks = append(haystacks, strings.ToLower(match.PatternName))
		}
		if match.SignalType != "" {
			haystacks = append(haystacks, strings.ToLower(match.SignalType))
		}
		if match.Text != "" {
			haystacks = append(haystacks, strings.ToLower(match.Text))
		}
	}

	return haystacks
}
