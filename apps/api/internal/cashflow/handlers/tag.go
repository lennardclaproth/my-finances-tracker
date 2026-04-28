package handlers

type TagByFilterHandler struct {

}

// TagByFilter applies or schedules tagging based on total matched rows and async policy.
func (s *TagByFilterHandler) TagByFilter(ctx context.Context, req BulkTagRequest) (BulkTagResult, error) {
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
