package scoring

import (
	"reflect"
	"testing"

	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestInferPhase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		candidate CandidateHypothesis
		want      types.FailurePhase
	}{
		{
			name: "image pull signals infer image_pull",
			candidate: CandidateHypothesis{
				Class: types.IssueClass{Name: "ImagePullBackOff"},
				Signals: types.SignalSet{Matches: []types.PatternMatch{
					{Text: "ERROR ErrImagePull: failed to pull image"},
				}},
			},
			want: types.FailurePhaseImagePull,
		},
		{
			name: "startup signals infer startup",
			candidate: CandidateHypothesis{
				Signals: types.SignalSet{Matches: []types.PatternMatch{
					{Text: "ERROR missing environment variable: DATABASE_URL"},
				}},
			},
			want: types.FailurePhaseStartup,
		},
		{
			name: "initialization signals infer initialization",
			candidate: CandidateHypothesis{
				Signals: types.SignalSet{Matches: []types.PatternMatch{
					{Text: "ERROR db migration failed: relation users does not exist"},
				}},
			},
			want: types.FailurePhaseInitialization,
		},
		{
			name: "runtime signals infer runtime",
			candidate: CandidateHypothesis{
				Class: types.IssueClass{Name: "ConnectionRefused"},
				Signals: types.SignalSet{Matches: []types.PatternMatch{
					{Text: "ERROR connect: connection refused"},
				}},
			},
			want: types.FailurePhaseRuntime,
		},
		{
			name: "shutdown signals infer shutdown",
			candidate: CandidateHypothesis{
				Signals: types.SignalSet{Matches: []types.PatternMatch{
					{Text: "INFO SIGTERM received, graceful shutdown in progress"},
				}},
			},
			want: types.FailurePhaseShutdown,
		},
		{
			name: "mixed signals choose earliest phase by precedence",
			candidate: CandidateHypothesis{
				Class: types.IssueClass{Name: "DependencyTimeout"},
				Signals: types.SignalSet{Matches: []types.PatternMatch{
					{Text: "ERROR context deadline exceeded while connecting"},
					{Text: "ERROR startup error: configuration incomplete"},
				}},
			},
			want: types.FailurePhaseStartup,
		},
		{
			name:      "empty input returns zero value",
			candidate: CandidateHypothesis{},
			want:      "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			first := InferPhase(tt.candidate)
			if first != tt.want {
				t.Fatalf("InferPhase() = %q, want %q", first, tt.want)
			}

			second := InferPhase(tt.candidate)
			if !reflect.DeepEqual(first, second) {
				t.Fatalf("InferPhase() is not deterministic: %q != %q", first, second)
			}
		})
	}
}
