package crud

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockQuerier implements Listable interface for testing
type MockQuerier[T any, Q any] struct {
	mock.Mock
}

func (m *MockQuerier[T, Q]) List(ctx context.Context, query Q, offset int, limit int) ([]T, error) {
	args := m.Called(ctx, query, offset, limit)
	return args.Get(0).([]T), args.Error(1)
}

func (m *MockQuerier[T, Q]) Count(ctx context.Context, query Q) (int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(int64), args.Error(1)
}

type MockResponseWriter struct {
	mock.Mock
}

func (m *MockResponseWriter) Response(w http.ResponseWriter, data interface{}, status int) error {
	args := m.Called(w, data, status)
	return args.Error(0)
}

func (m *MockResponseWriter) Error(w http.ResponseWriter, err error, status int) {
	m.Called(w, err, status)
}

type TestQuery struct {
	PaginatedQuery
	Name string `form:"name"`
}

func (q TestQuery) GetPagination() PaginatedQuery {
	return q.PaginatedQuery
}

func TestNewListHandler(t *testing.T) {
	querier := &MockQuerier[string, TestQuery]{}
	writer := &MockResponseWriter{}

	handler := NewListHandler[string, TestQuery](querier, writer)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.decoder)
	assert.NotNil(t, handler.validate)
	assert.Equal(t, querier, handler.querier)
	assert.Equal(t, writer, handler.writer)
}

func TestListHandler_Handle_ValidRequest(t *testing.T) {
	querier := &MockQuerier[string, TestQuery]{}
	writer := &MockResponseWriter{}
	handler := NewListHandler[string, TestQuery](querier, writer)

	items := []string{"item1"}
	querier.On("List", mock.Anything, mock.Anything, 0, 10).Return(items, nil)
	querier.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)

	writer.On("Response", mock.Anything, mock.Anything, http.StatusOK).Return(nil)

	req := httptest.NewRequest("GET", "/?page=1&page_size=10&name=test", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	querier.AssertExpectations(t)
	writer.AssertExpectations(t)
}

func TestListHandler_Handle_InvalidQuery(t *testing.T) {
	querier := &MockQuerier[string, TestQuery]{}
	writer := &MockResponseWriter{}
	handler := NewListHandler[string, TestQuery](querier, writer)

	writer.On("Error", mock.Anything, ErrBadRequest, http.StatusBadRequest).Return()

	req := httptest.NewRequest("GET", "/?page=0&page_size=0", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	writer.AssertExpectations(t)
}

func TestListHandler_Handle_QuerierError(t *testing.T) {
	querier := &MockQuerier[string, TestQuery]{}
	writer := &MockResponseWriter{}
	handler := NewListHandler[string, TestQuery](querier, writer)

	querier.On("List", mock.Anything, mock.Anything, 0, 10).Return([]string{}, assert.AnError)
	writer.On("Error", mock.Anything, ErrInternal, http.StatusInternalServerError).Return()

	req := httptest.NewRequest("GET", "/?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	querier.AssertExpectations(t)
	writer.AssertExpectations(t)
}

//func TestListHandler_Handle_Timeout(t *testing.T) {
//	querier := &MockQuerier[string, TestQuery]{}
//	writer := &MockResponseWriter{}
//	handler := NewListHandler[string, TestQuery](querier, writer)
//
//	querier.On("List", mock.Anything, mock.Anything, 0, 10).After(31*time.Second).Return([]string{}, nil)
//	querier.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)
//	writer.On("Response", mock.Anything, mock.Anything, http.StatusOK).Return(nil)
//	writer.On("Error", mock.Anything, ErrInternal, http.StatusInternalServerError).Return()
//
//	req := httptest.NewRequest("GET", "/?page=1&page_size=10", nil)
//	w := httptest.NewRecorder()
//
//	handler.Handle(w, req)
//
//	writer.AssertExpectations(t)
//}
