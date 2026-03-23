package tcx

import (
	"context"
	"fmt"

	domain "fitgo/internal/domain/tcx"
)

// QueryService handles read-side TCX use cases.
type QueryService struct {
	repo domain.SummaryRepository
}

func NewQueryService(repo domain.SummaryRepository) *QueryService {
	return &QueryService{repo: repo}
}

func (s *QueryService) GetTCXSummary(ctx context.Context, id string) (*domain.Summary, error) {
	if id == "" {
		return nil, fmt.Errorf("missing id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *QueryService) ListTCXSummaries(ctx context.Context) ([]*domain.Summary, error) {
	return s.repo.List(ctx)
}
