package coros

import (
	"context"

	"fitgo/internal/application/port"
	domain "fitgo/internal/domain/coros"
)

// QueryService handles read-side Coros use cases.
type QueryService struct {
	gateway port.CorosGateway
}

func NewQueryService(gateway port.CorosGateway) *QueryService {
	return &QueryService{gateway: gateway}
}

func (s *QueryService) Login(ctx context.Context) (string, error) {
	return s.gateway.Login(ctx)
}

func (s *QueryService) ActivityList(ctx context.Context, size, pageNumber, modeList int) (map[string]interface{}, error) {
	return s.gateway.ActivityList(ctx, size, pageNumber, modeList)
}

func (s *QueryService) SportsSummary(ctx context.Context, labelId, sportType string) (*domain.SportsSummary, error) {
	return s.gateway.SportsSummary(ctx, labelId, sportType)
}
