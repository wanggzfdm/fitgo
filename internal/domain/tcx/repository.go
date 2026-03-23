package tcx

import "context"

// SummaryRepository stores and retrieves TCX summaries.
type SummaryRepository interface {
	Save(ctx context.Context, summary *Summary) error
	GetByID(ctx context.Context, id string) (*Summary, error)
	List(ctx context.Context) ([]*Summary, error)
}
