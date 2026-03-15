package engine

import (
	"context"

	"github.com/Urealaden/log-sage-temp/internal/detection"
	"github.com/Urealaden/log-sage-temp/internal/extraction"
	"github.com/Urealaden/log-sage-temp/internal/normalize"
	"github.com/Urealaden/log-sage-temp/internal/recommendation"
	"github.com/Urealaden/log-sage-temp/internal/scoring"
	"github.com/Urealaden/log-sage-temp/pkg/types"
)

type Engine interface {
	Analyze(ctx context.Context, input types.DiagnosticInput) (*types.AnalysisResult, error)
}

type engine struct{}

func New() Engine {
	return &engine{}
}

func (e *engine) Analyze(ctx context.Context, input types.DiagnosticInput) (*types.AnalysisResult, error) {
	normalized, err := e.normalize(ctx, input)
	if err != nil {
		return nil, err
	}

	signals, err := e.extractSignals(ctx, normalized)
	if err != nil {
		return nil, err
	}

	hypotheses, err := e.detectHypotheses(ctx, signals)
	if err != nil {
		return nil, err
	}

	ranked, err := e.scoreAndRank(ctx, hypotheses, signals)
	if err != nil {
		return nil, err
	}

	return e.generateResult(ctx, ranked)
}

type normalizedInput []normalize.Line

func (e *engine) normalize(ctx context.Context, input types.DiagnosticInput) (normalizedInput, error) {
	lines, err := normalize.Normalize(ctx, input.Reader)
	if err != nil {
		return nil, err
	}

	return normalizedInput(lines), nil
}

func (e *engine) extractSignals(ctx context.Context, input normalizedInput) (types.SignalSet, error) {
	select {
	case <-ctx.Done():
		return types.SignalSet{}, ctx.Err()
	default:
	}

	allPatterns := make([]types.SignalPattern, 0)
	for _, class := range detection.IssueRegistry {
		allPatterns = append(allPatterns, class.PrimarySignals...)
		allPatterns = append(allPatterns, class.CorroboratingSignals...)
	}

	return extraction.ExtractSignals([]normalize.Line(input), allPatterns), nil
}

func (e *engine) detectHypotheses(ctx context.Context, signals types.SignalSet) ([]types.Hypothesis, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return detection.EvaluateRegistry(signals), nil
}

func (e *engine) scoreAndRank(ctx context.Context, hypotheses []types.Hypothesis, signals types.SignalSet) ([]types.Hypothesis, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	candidates := scoring.BuildCandidates(hypotheses, signals)
	candidates = scoring.ApplyCorroboration(candidates)
	candidates = scoring.ApplyRelationships(candidates)
	candidates = scoring.MapConfidence(candidates)
	candidates = scoring.Rank(candidates)

	result := make([]types.Hypothesis, len(candidates))
	for i, candidate := range candidates {
		result[i] = types.Hypothesis{
			IssueClass:  candidate.Class.Name,
			Confidence:  candidate.Confidence,
			Phase:       candidate.Phase,
			Score:       candidate.BaseScore,
			Evidence:    candidate.Evidence,
			Explanation: candidate.Class.ExplanationTemplate,
		}
	}

	return result, nil
}

func (e *engine) generateResult(ctx context.Context, hypotheses []types.Hypothesis) (*types.AnalysisResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	nextSteps := recommendation.NextSteps(hypotheses, detection.IssueRegistry)

	return &types.AnalysisResult{
		TopCauses:            hypotheses,
		RecommendedNextSteps: nextSteps,
	}, nil
}
