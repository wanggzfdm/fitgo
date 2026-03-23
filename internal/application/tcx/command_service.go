package tcx

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"time"

	domain "fitgo/internal/domain/tcx"
)

// CommandService handles write-side TCX use cases.
type CommandService struct {
	repo domain.SummaryRepository
}

func NewCommandService(repo domain.SummaryRepository) *CommandService {
	return &CommandService{repo: repo}
}

// UploadTCX validates, parses, and stores a TCX summary.
func (s *CommandService) UploadTCX(ctx context.Context, file multipart.File, filename string) (*domain.Summary, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read tcx content: %w", err)
	}

	if err := domain.Validate(content); err != nil {
		return nil, fmt.Errorf("invalid tcx content: %w", err)
	}

	summary, err := domain.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("parse tcx content: %w", err)
	}

	if summary.ID == "" {
		summary.ID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	if summary.Filename == "" {
		summary.Filename = filename
	}

	if err := s.repo.Save(ctx, summary); err != nil {
		return nil, fmt.Errorf("store tcx summary: %w", err)
	}

	return summary, nil
}
