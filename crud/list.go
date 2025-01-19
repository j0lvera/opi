package crud

import (
	"context"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"time"
)

// PaginatedQuery represents the standard pagination query parameters.
// It requires page number and page size through form validation.
type PaginatedQuery struct {
	Page     int `form:"page" validate:"required,min=1"`              // Page number, starting from 1
	PageSize int `form:"page_size" validate:"required,min=1,max=100"` // Number of items per page, max 100
}

// Listable defines the types that support listing and counting operations.
// T is the type of the items being listed.
// Q is the type of the query parameters.
type Listable[T any, Q any] interface {
	// List retrieves a slice of items based on the query parameters and pagination.
	List(ctx context.Context, query Q, offset int, limit int) ([]T, error)

	// Count returns the total number of items matching the query parameters.
	Count(ctx context.Context, query Q) (int64, error)
}

// ListHandler is a generic HTTP handler for listing items.
// It provides pagination, query parameter decoding, and validation.
type ListHandler[T any, Q any] struct {
	querier  Listable[T, Q]      // The querier to fetch data from the database
	decoder  *form.Decoder       // Decoder for query parameters
	validate *validator.Validate // Validator for query parameters
	writer   ResponseWriter      // Response writer interface
}

// NewListHandler creates a new ListHandler with the provided querier and response writer.
// It initializes the decoder and validator with default settings.
func NewListHandler[T any, Q any](querier Listable[T, Q], writer ResponseWriter) *ListHandler[T, Q] {
	return &ListHandler[T, Q]{
		querier:  querier,
		decoder:  form.NewDecoder(),
		validate: validator.New(validator.WithRequiredStructEnabled()),
		writer:   writer,
	}
}

// Handle processes HTTP requests for listing operations.
// It performs the following steps:
// 1. Decodes query parameters from the request.
// 2. Validates the query parameters.
// 3. Handles pagination if the query implements GetPagination().
// 4. Fetches the data from the querier.
// 5. Counts the total number of items.
// 6. Writes the response with the paginated data.
//
// Teh response format is:
//
//	{
//		"items": [...],
//		"pagination": {
//		"total": n,
//		"page": x
//		"page_size": y
//	}
func (h *ListHandler[T, Q]) Handle(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	slog.Debug("list.start",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_add", r.RemoteAddr,
	)

	// set timeout for the entire request
	// TODO: make this configurable via env var
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// decode the query params
	var query Q
	if err := h.decoder.Decode(&query, r.URL.Query()); err != nil {
		slog.Error("unable to decode query params", "error", err)
		slog.Debug("query params decoding", "query_params", r.URL.Query())
		h.writer.Error(w, ErrBadRequest, http.StatusBadRequest)
		return
	}

	// validate the query params
	if err := h.validate.Struct(query); err != nil {
		//validationErrors := ParseValidationErrors(err)
		//res := map[string][]ValidationError{
		//	"errors": validationErrors,
		//}
		slog.Error("unable to validate query params", "error", err)
		slog.Debug("query params decoding", "query_params", r.URL.Query())
		h.writer.Error(w, ErrBadRequest, http.StatusBadRequest)
		return
	}

	var offset, limit int
	if pg, ok := any(query).(interface{ GetPagination() PaginatedQuery }); ok {
		pagination := pg.GetPagination()
		offset = (pagination.Page - 1) * pagination.PageSize
		limit = pagination.PageSize
	}

	// list the items with pagination
	items, err := h.querier.List(ctx, query, offset, limit)
	if err != nil {
		slog.Error("unable to list items", "error", err)
		slog.Debug("listing", "query", query, "offset", offset, "limit", limit)
		h.writer.Error(w, ErrInternal, http.StatusInternalServerError)
		return
	}

	// count the total number of items
	total, err := h.querier.Count(ctx, query)
	if err != nil {
		slog.Error("unable to count items", "error", err)
		slog.Debug("counting", "query", query)
		h.writer.Error(w, ErrInternal, http.StatusInternalServerError)
		return
	}

	// write the response
	res := map[string]interface{}{
		"items": items,
		"pagination": map[string]interface{}{
			"total":     total,
			"page":      offset/limit + 1,
			"page_size": limit,
		},
	}
	err = h.writer.Response(w, res, http.StatusOK)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("writing response", "response", res)
	}

	slog.Debug("list.complete", "duration", time.Since(startTime))
}
