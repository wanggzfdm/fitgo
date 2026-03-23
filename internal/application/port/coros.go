package port

import (
	"context"

	domain "fitgo/internal/domain/coros"
)

// CorosGateway abstracts access to Coros external APIs.
type CorosGateway interface {
	Login(ctx context.Context) (string, error)
	ActivityList(ctx context.Context, size, pageNumber, modeList int) (map[string]interface{}, error)
	SportsSummary(ctx context.Context, labelId, sportType string) (*domain.SportsSummary, error)
}
