package port

import "context"

// AnalysisAnalyzer provides analysis output for a given activity.
type AnalysisAnalyzer interface {
	Analyze(ctx context.Context, labelId, sportType, format string) (string, error)
}
