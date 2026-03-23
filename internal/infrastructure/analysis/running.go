package analysis

import (
	"context"

	"fitgo/internal/infrastructure/analysis/running"
)

// RunningAnalyzer adapts the running analyzer to the application interface.
type RunningAnalyzer struct{}

func NewRunningAnalyzer() *RunningAnalyzer {
	return &RunningAnalyzer{}
}

func (a *RunningAnalyzer) Analyze(_ context.Context, labelId, sportType, format string) (string, error) {
	return running.RunAnalyzerWithFormat(labelId, sportType, format)
}
