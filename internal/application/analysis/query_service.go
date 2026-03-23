package analysis

import (
	"context"

	"fitgo/internal/application/port"
)

// QueryService handles read-side analysis use cases.
type QueryService struct {
	analyzer port.AnalysisAnalyzer
}

func NewQueryService(analyzer port.AnalysisAnalyzer) *QueryService {
	return &QueryService{analyzer: analyzer}
}

func (s *QueryService) Analyze(ctx context.Context, labelId, sportType, format string) (string, error) {
	return s.analyzer.Analyze(ctx, labelId, sportType, format)
}
