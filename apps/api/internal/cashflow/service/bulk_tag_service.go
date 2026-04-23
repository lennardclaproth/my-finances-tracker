package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

// DefaultBulkTagAsyncCutoff is the threshold for switching filtered tag updates to async processing.
const DefaultBulkTagAsyncCutoff = 1000

var (
	// ErrBulkTagFiltersInvalid indicates malformed filter input.
	ErrBulkTagFiltersInvalid = errors.New("bulk tag filters invalid")
)

// TagByFilterMode describes whether a filtered tag request was handled synchronously or asynchronously.
type TagByFilterMode string

const (
	// TagByFilterModeSync indicates tags were updated within the request lifecycle.
	TagByFilterModeSync TagByFilterMode = "sync"
	// TagByFilterModeAsync indicates a background job was enqueued.
	TagByFilterModeAsync TagByFilterMode = "async"
)

// BulkTagFilters captures user-provided filter criteria.
type BulkTagFilters struct {
	Q           string
	Description string
	Note        string
	Source      string
	Direction   string
	Tags        string
	Untagged    *bool
	HideIgnored *bool
	From        string
	To          string
}

// BulkTagRequest is the service-level input for filtered tagging.
type BulkTagRequest struct {
	Tag       string
	AccountID *uuid.UUID
	Filters   BulkTagFilters
}

// BulkTagResult is the service-level output for filtered tagging.
type BulkTagResult struct {
	Mode         TagByFilterMode
	UpdatedCount int
	TotalMatched int
}

type bulkTagStore interface {
	CountByQuery(ctx context.Context, query storage.CashflowTransactionQuery) (int, error)
	UpdateTagByQuery(ctx context.Context, query storage.CashflowTransactionQuery, tag string) (int, error)
}

// BulkTagEnqueuer schedules large filtered updates for background execution.
type BulkTagEnqueuer interface {
	EnqueueFilter(ctx context.Context, accountID uuid.UUID, query storage.CashflowTransactionQuery, tag string) error
}

// BulkTagService encapsulates filtered-tagging orchestration and async cutoff policy.
type BulkTagService struct {
	store       bulkTagStore
	enqueuer    BulkTagEnqueuer
	asyncCutoff int
}

// NewBulkTagService constructs a bulk tag service.
func NewBulkTagService(store bulkTagStore, enqueuer BulkTagEnqueuer, asyncCutoff int) *BulkTagService {
	if asyncCutoff <= 0 {
		asyncCutoff = DefaultBulkTagAsyncCutoff
	}
	return &BulkTagService{
		store:       store,
		enqueuer:    enqueuer,
		asyncCutoff: asyncCutoff,
	}
}

// TagByFilter applies or schedules tagging based on total matched rows and async policy.
func (s *BulkTagService) TagByFilter(ctx context.Context, req BulkTagRequest) (BulkTagResult, error) {
	query, err := BuildBulkTagQuery(req.Filters)
	if err != nil {
		return BulkTagResult{}, err
	}

	total, err := s.store.CountByQuery(ctx, query)
	if err != nil {
		return BulkTagResult{}, err
	}

	if total > s.asyncCutoff && s.enqueuer != nil && req.AccountID != nil && *req.AccountID != uuid.Nil {
		if err := s.enqueuer.EnqueueFilter(ctx, *req.AccountID, query, req.Tag); err != nil {
			return BulkTagResult{}, err
		}
		return BulkTagResult{
			Mode:         TagByFilterModeAsync,
			UpdatedCount: 0,
			TotalMatched: total,
		}, nil
	}

	updated, err := s.store.UpdateTagByQuery(ctx, query, req.Tag)
	if err != nil {
		return BulkTagResult{}, err
	}
	return BulkTagResult{
		Mode:         TagByFilterModeSync,
		UpdatedCount: updated,
		TotalMatched: total,
	}, nil
}

// BuildBulkTagQuery converts API-level filters into a persistence query with semantic validation.
func BuildBulkTagQuery(filters BulkTagFilters) (storage.CashflowTransactionQuery, error) {
	var query storage.CashflowTransactionQuery
	query.Q = filters.Q
	query.Description = filters.Description
	query.Note = filters.Note
	query.Source = filters.Source

	direction, err := normalizeBulkTagDirection(filters.Direction)
	if err != nil {
		return storage.CashflowTransactionQuery{}, err
	}
	query.Direction = direction
	query.Tags = splitBulkTagValues(filters.Tags)
	if filters.Untagged != nil {
		query.Untagged = *filters.Untagged
	}
	if filters.HideIgnored != nil {
		query.HideIgnored = *filters.HideIgnored
	}
	if filters.From != "" {
		from, err := time.Parse("2006-01-02", filters.From)
		if err != nil {
			return storage.CashflowTransactionQuery{}, fmt.Errorf("%w: from must be in YYYY-MM-DD format", ErrBulkTagFiltersInvalid)
		}
		query.From = &from
	}
	if filters.To != "" {
		to, err := time.Parse("2006-01-02", filters.To)
		if err != nil {
			return storage.CashflowTransactionQuery{}, fmt.Errorf("%w: to must be in YYYY-MM-DD format", ErrBulkTagFiltersInvalid)
		}
		query.To = &to
	}
	if query.From != nil && query.To != nil && query.From.After(*query.To) {
		return storage.CashflowTransactionQuery{}, fmt.Errorf("%w: from must be before or equal to to", ErrBulkTagFiltersInvalid)
	}
	return query, nil
}

func normalizeBulkTagDirection(raw string) (string, error) {
	direction := strings.ToLower(strings.TrimSpace(raw))
	if direction == "" {
		return "", nil
	}
	if direction != "in" && direction != "out" {
		return "", fmt.Errorf("%w: direction must be either in or out", ErrBulkTagFiltersInvalid)
	}
	return direction, nil
}

func splitBulkTagValues(tags string) []string {
	if strings.TrimSpace(tags) == "" {
		return nil
	}
	raw := strings.Split(tags, ",")
	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, entry := range raw {
		tag := strings.TrimSpace(entry)
		if tag == "" {
			continue
		}
		key := strings.ToLower(tag)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, tag)
	}
	return out
}
