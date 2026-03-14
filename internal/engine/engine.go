package engine

import (
	"context"

	"github.com/Urealaden/log-sage-temp/internal/normalize"
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

	ranked, err := e.scoreAndRank(ctx, hypotheses)
	if err != nil {
		return nil, err
	}

	return e.generateResult(ctx, ranked)
}

type normalizedInput []normalize.Line

type extractedSignals struct{}

func (e *engine) normalize(ctx context.Context, input types.DiagnosticInput) (normalizedInput, error) {
	lines, err := normalize.Normalize(ctx, input.Reader)
	if err != nil {
		return nil, err
	}

	return normalizedInput(lines), nil
}

func (e *engine) extractSignals(ctx context.Context, input normalizedInput) (extractedSignals, error) {
	select {
	case <-ctx.Done():
		return extractedSignals{}, ctx.Err()
	default:
	}

	_ = input

	return extractedSignals{}, nil
}

func (e *engine) detectHypotheses(ctx context.Context, signals extractedSignals) ([]types.Hypothesis, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	_ = signals

	return nil, nil
}

func (e *engine) scoreAndRank(ctx context.Context, hypotheses []types.Hypothesis) ([]types.Hypothesis, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return hypotheses, nil
}

func (e *engine) generateResult(ctx context.Context, hypotheses []types.Hypothesis) (*types.AnalysisResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return &types.AnalysisResult{
		TopCauses: hypotheses,
	}, nil
}
