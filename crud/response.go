package crud

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ResponseWriter defines standardized HTTP responses.
// It provides methods for writing successful and error responses.
type ResponseWriter interface {
	// Error writes an error with a given status code and error message.
	// Sets the appropriate headers and encodes the details in JSON format.
	//
	// Parameters:
	// - w: The http.ResponseWriter to write the response to.
	// - err: The error to be written as a response.
	// - status: The HTTP status code to be included in the response.
	Error(w http.ResponseWriter, err error, status int)
	// Response writes a successful response with a given status code and data.
	// Sets the appropriate headers and encodes the data in JSON format.
	//
	// Type Parameters:
	// - T: The type of the data to be written in the response
	//
	// Parameters:
	// - w: The http.ResponseWriter to write the response to.
	// - v: The data
	// - status: The HTTP status code to be included in the response.
	Response(w http.ResponseWriter, v any, status int) error
}

// ErrorResponse represents the structure of an error response.
type ErrorResponse struct {
	Status  int         `json:"status"`            // HTTP status code
	Message string      `json:"message"`           // human-readable error message
	Details interface{} `json:"details,omitempty"` // additional error details
}

type DefaultResponseWriter struct{}

func (w *DefaultResponseWriter) Response(writer http.ResponseWriter, v any, status int) error {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	if v == nil {
		return nil
	}

	return json.NewEncoder(writer).Encode(v)
}

func (w *DefaultResponseWriter) Error(writer http.ResponseWriter, err error, status int) {
	res := ErrorResponse{
		Status:  status,
		Message: err.Error(),
	}

	// If the error implements the Details() method, include the details in the response
	if detailed, ok := err.(interface{ Details() interface{} }); ok {
		res.Details = detailed.Details()
	}

	err = w.Response(writer, res, status)
	if err != nil {
		slog.Error("unable to write error response",
			"status", status,
			"error", err,
		)
		http.Error(writer, ErrInternal.Error(), http.StatusInternalServerError)
	}
}
