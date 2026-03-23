package tcx

import (
	"context"
	"fmt"
	"sync"

	domain "fitgo/internal/domain/tcx"
)

// MemoryRepository stores TCX summaries in memory.
type MemoryRepository struct {
	mu        sync.RWMutex
	summaries map[string]*domain.Summary
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		summaries: make(map[string]*domain.Summary),
	}
}

func (r *MemoryRepository) Save(_ context.Context, summary *domain.Summary) error {
	if summary == nil {
		return fmt.Errorf("summary is nil")
	}
	if summary.ID == "" {
		return fmt.Errorf("summary id is empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.summaries[summary.ID] = summary
	return nil
}

func (r *MemoryRepository) GetByID(_ context.Context, id string) (*domain.Summary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summary, ok := r.summaries[id]
	if !ok {
		return nil, fmt.Errorf("summary not found")
	}
	return summary, nil
}

func (r *MemoryRepository) List(_ context.Context) ([]*domain.Summary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Summary, 0, len(r.summaries))
	for _, summary := range r.summaries {
		result = append(result, summary)
	}
	return result, nil
}
