package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/Urealaden/log-sage-temp/internal/normalize"
	"github.com/Urealaden/log-sage-temp/pkg/types"
)

func TestEngineAnalyze(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		log              string
		cancelBeforeCall bool
		wantClasses      []string
		wantConfidence   types.ConfidenceLevel
		wantNextSteps    bool
		wantErr          bool
	}{
		{
			name:        "empty input",
			log:         "",
			wantClasses: nil,
		},
		{
			name:           "oomkilled",
			log:            "OOMKilled\n",
			wantClasses:    []string{"OutOfMemory"},
			wantConfidence: types.ConfidenceHigh,
			wantNextSteps:  true,
		},
		{
			name:        "connection refused",
			log:         "dial tcp 10.0.0.8:5432: connect: connection refused\n",
			wantClasses: []string{"ConnectionRefused"},
		},
		{
			name:        "crashloopbackoff and oom ranks root cause first",
			log:         "CrashLoopBackOff\nOOMKilled\n",
			wantClasses: []string{"OutOfMemory", "CrashLoopBackOff"},
		},
		{
			name:             "context already cancelled",
			log:              "OOMKilled\n",
			cancelBeforeCall: true,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tt.cancelBeforeCall {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			result, err := New().Analyze(ctx, types.DiagnosticInput{
				Reader: strings.NewReader(tt.log),
			})
			if tt.wantErr {
				if err == nil {
					t.Fatal("Analyze() error = nil, want non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Analyze() error = %v, want nil", err)
			}
			if result == nil {
				t.Fatal("Analyze() result = nil, want non-nil")
			}

			if len(tt.wantClasses) == 0 {
				if len(result.TopCauses) != 0 {
					t.Fatalf("len(result.TopCauses) = %d, want 0", len(result.TopCauses))
				}
				return
			}

			if len(result.TopCauses) < len(tt.wantClasses) {
				t.Fatalf("len(result.TopCauses) = %d, want at least %d", len(result.TopCauses), len(tt.wantClasses))
			}

			for i, wantClass := range tt.wantClasses {
				if result.TopCauses[i].IssueClass != wantClass {
					t.Fatalf("TopCauses[%d].IssueClass = %q, want %q", i, result.TopCauses[i].IssueClass, wantClass)
				}
			}

			if tt.wantConfidence != "" && result.TopCauses[0].Confidence != tt.wantConfidence {
				t.Fatalf("TopCauses[0].Confidence = %q, want %q", result.TopCauses[0].Confidence, tt.wantConfidence)
			}
			if tt.wantNextSteps && len(result.RecommendedNextSteps) == 0 {
				t.Fatal("RecommendedNextSteps is empty, want non-empty")
			}
		})
	}
}

func TestEngineStagesRespectCancellation(t *testing.T) {
	t.Parallel()

	e := &engine{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := e.extractSignals(ctx, normalizedInput{{Raw: "OOMKilled"}}); err == nil {
		t.Fatal("extractSignals() error = nil, want context cancellation error")
	}

	if _, err := e.detectHypotheses(ctx, types.SignalSet{}); err == nil {
		t.Fatal("detectHypotheses() error = nil, want context cancellation error")
	}

	if _, err := e.scoreAndRank(ctx, nil, types.SignalSet{}); err == nil {
		t.Fatal("scoreAndRank() error = nil, want context cancellation error")
	}

	if _, err := e.generateResult(ctx, nil); err == nil {
		t.Fatal("generateResult() error = nil, want context cancellation error")
	}
}

func TestEngineExtractSignalsUsesRegistryPatterns(t *testing.T) {
	t.Parallel()

	e := &engine{}
	signals, err := e.extractSignals(context.Background(), normalizedInput{
		normalize.Line{Raw: "OOMKilled"},
		normalize.Line{Raw: "CrashLoopBackOff"},
	})
	if err != nil {
		t.Fatalf("extractSignals() error = %v, want nil", err)
	}
	if len(signals.Matches) < 2 {
		t.Fatalf("len(signals.Matches) = %d, want at least 2", len(signals.Matches))
	}
}
