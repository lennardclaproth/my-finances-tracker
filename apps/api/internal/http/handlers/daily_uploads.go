package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	marketdataParsers "github.com/lennardclaproth/my-finances-tracker/internal/marketdata/parsers"
)

type dailyUploadStore interface {
	Create(ctx context.Context, upload *marketdata.DailyUpload) error
	FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.DailyUpload, error)
}

type dailyUploadListingStore interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error)
}

type dailyUploadProviderStore interface {
	GetByName(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error)
}

type dailyUploadFileWriter interface {
	WriteCsv(r io.Reader) (string, error)
}

type dailyUploadFileRemover interface {
	Remove(path string) error
}

type dailyUploadEnqueuer interface {
	Enqueue(ctx context.Context, uploadID uuid.UUID) error
}

type uploadDailiesMultipartRequest struct {
	File      multipart.File `multipart:"file"`
	Filename  string         `multipart:"filename"`
	ListingID string         `form:"listing_id"`
}

func (r uploadDailiesMultipartRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if strings.TrimSpace(r.ListingID) == "" {
		problems["listing_id"] = "listing_id is required"
	}
	return problems
}

// UploadDailiesFile uploads a listing daily data file to be processed asynchronously.
//
// @Summary Upload listing daily data file
// @Description Upload a .csv/.txt file with listing daily values and process it asynchronously based on listing source parser.
// @Tags dailies
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Daily data file"
// @Param listing_id formData string true "Listing UUID"
// @Success 202 {object} api.DailyUploadAcceptedResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/dailies/upload [post]
func UploadDailiesFile(
	log logging.Logger,
	uploadStore dailyUploadStore,
	listingStore dailyUploadListingStore,
	providerStore dailyUploadProviderStore,
	fileWriter dailyUploadFileWriter,
	fileRemover dailyUploadFileRemover,
	enqueuer dailyUploadEnqueuer,
) http.Handler {
	policy := marketdata.NewManualUploadPolicy(
		listingStore,
		providerStore,
		func(source marketdata.Source) error {
			_, err := marketdataParsers.CreateDailyParser(source)
			return err
		},
	)
	endpoint := func(ctx context.Context, req uploadDailiesMultipartRequest) (status int, res any, err error) {
		defer func() {
			if closeErr := req.File.Close(); closeErr != nil {
				log.Warn(ctx, "failed closing daily upload file", "error", closeErr.Error())
			}
		}()

		if !isAllowedDailyUploadFilename(req.Filename) {
			return http.StatusBadRequest, map[string]string{"file": "file must have .csv or .txt extension"}, nil
		}

		listingID, err := uuid.Parse(strings.TrimSpace(req.ListingID))
		if err != nil {
			return http.StatusBadRequest, map[string]string{"listing_id": "listing_id must be a valid UUID"}, nil
		}

		listing, err := policy.ValidateListing(ctx, listingID)
		if err != nil {
			switch {
			case errors.Is(err, marketdata.ErrManualUploadListingNotFound):
				return http.StatusNotFound, map[string]string{"listing_id": "listing not found"}, nil
			case errors.Is(err, marketdata.ErrManualUploadProviderUnavailable):
				return http.StatusUnprocessableEntity, map[string]string{"listing_id": "provider unavailable for listing source"}, nil
			case errors.Is(err, marketdata.ErrManualUploadProviderNotManual):
				return http.StatusUnprocessableEntity, map[string]string{"listing_id": "provider does not support manual daily uploads"}, nil
			case errors.Is(err, marketdata.ErrManualUploadParserUnavailable):
				return http.StatusUnprocessableEntity, map[string]string{"listing_id": "no parser for listing source"}, nil
			default:
				return http.StatusInternalServerError, struct{}{}, err
			}
		}

		head := make([]byte, 1)
		n, readErr := req.File.Read(head)
		if readErr == io.EOF {
			return http.StatusBadRequest, map[string]string{"file": "file is empty"}, nil
		}
		if readErr != nil {
			return http.StatusInternalServerError, struct{}{}, readErr
		}

		reader := io.MultiReader(bytes.NewReader(head[:n]), req.File)
		storedFilename, err := fileWriter.WriteCsv(reader)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		upload, err := marketdata.NewDailyUpload(listingID, listing.Source, storedFilename, req.Filename)
		if err != nil {
			if removeErr := fileRemover.Remove(storedFilename); removeErr != nil {
				log.Warn(ctx, "failed removing stored upload file after invalid upload", "path", storedFilename, "error", removeErr.Error())
			}
			return http.StatusBadRequest, map[string]string{"upload": err.Error()}, nil
		}
		if err := uploadStore.Create(ctx, upload); err != nil {
			if removeErr := fileRemover.Remove(storedFilename); removeErr != nil {
				log.Warn(ctx, "failed removing stored upload file after persist failure", "path", storedFilename, "error", removeErr.Error())
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		if enqueuer != nil {
			if err := enqueuer.Enqueue(ctx, upload.ID); err != nil {
				log.Info(ctx, "daily upload persisted but enqueue failed", "upload_id", upload.ID, "error", err.Error())
			}
		}

		return http.StatusAccepted, api.DailyUploadAcceptedResponse{
			UploadID: upload.ID,
			Status:   string(upload.Status),
		}, nil
	}

	decodeFn := httpx.DecoderFunc[uploadDailiesMultipartRequest](func(r *http.Request) (uploadDailiesMultipartRequest, error) {
		return httpx.DecodeMultipartFile[uploadDailiesMultipartRequest](r, httpx.MultipartFileDecoderOptions{
			FieldName: "file",
			MaxBytes:  10 * 1024 * 1024,
			MaxMemory: 20 * 1024 * 1024,
		})
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}

type getDailyUploadStatusRequest struct {
	UploadID string
}

// GetDailyUploadStatus returns async processing status for a daily upload.
//
// @Summary Get daily upload status
// @Description Returns counters and row-level errors for a daily upload processing job.
// @Tags dailies
// @Accept json
// @Produce json
// @Param upload_id path string true "Daily upload UUID"
// @Success 200 {object} api.DailyUploadStatusResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/dailies/uploads/{upload_id} [get]
func GetDailyUploadStatus(
	log logging.Logger,
	uploadStore dailyUploadStore,
) http.Handler {
	endpoint := func(ctx context.Context, req getDailyUploadStatusRequest) (status int, res any, err error) {
		if strings.TrimSpace(req.UploadID) == "" {
			return http.StatusBadRequest, map[string]string{"upload_id": "upload_id is required"}, nil
		}
		uploadID, err := uuid.Parse(req.UploadID)
		if err != nil {
			return http.StatusBadRequest, map[string]string{"upload_id": "upload_id must be a valid UUID"}, nil
		}
		upload, err := uploadStore.FetchByID(ctx, uploadID)
		if err != nil {
			if errors.Is(err, marketdata.ErrDailyUploadNotFound) {
				return http.StatusNotFound, map[string]string{"upload_id": "daily upload not found"}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		if upload == nil {
			return http.StatusNotFound, map[string]string{"upload_id": "daily upload not found"}, nil
		}

		return http.StatusOK, api.DailyUploadStatusResponse{
			ID:            upload.ID,
			ListingID:     upload.ListingID,
			Source:        string(upload.Source),
			Status:        string(upload.Status),
			StatusMessage: upload.StatusMessage,
			TotalRows:     upload.TotalRows,
			InsertedRows:  upload.InsertedRows,
			DuplicateRows: upload.DuplicateRows,
			ErrorRows:     upload.ErrorRows,
			RowErrors:     toDailyUploadRowErrors(upload.RowErrors),
			CreatedAt:     upload.CreatedAt,
			StartedAt:     upload.StartedAt,
			FinishedAt:    upload.FinishedAt,
			UpdatedAt:     upload.UpdatedAt,
		}, nil
	}

	decodeFn := httpx.DecoderFunc[getDailyUploadStatusRequest](func(r *http.Request) (getDailyUploadStatusRequest, error) {
		return getDailyUploadStatusRequest{UploadID: r.PathValue("upload_id")}, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}

func isAllowedDailyUploadFilename(name string) bool {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(name)))
	return ext == ".csv" || ext == ".txt"
}

func toDailyUploadRowErrors(rows []marketdata.DailyUploadRowError) []api.DailyUploadRowErrorResponse {
	out := make([]api.DailyUploadRowErrorResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, api.DailyUploadRowErrorResponse{
			RowNumber: row.RowNumber,
			Reason:    row.Reason,
		})
	}
	return out
}
