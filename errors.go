// Package coinglass provides a Go SDK for the Coinglass API v4.
package coinglass

import (
	"errors"
	"fmt"
)

// APIError represents an error response returned by the Coinglass API.
// It is returned by every service method when the API responds with a
// non-2xx HTTP status code or a non-zero API code.
type APIError struct {
	// StatusCode is the HTTP status code returned by the API.
	StatusCode int
	// Code is the Coinglass API-level code, if present in the response.
	Code string
	// Message is a human-readable error message.
	Message string
	// RawBody contains the raw response body for debugging purposes.
	RawBody []byte
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	return fmt.Sprintf("coinglass: api error: status=%d code=%s message=%s", e.StatusCode, e.Code, e.Message)
}

// Sentinel errors that can be checked with errors.Is against errors
// returned from service methods. The underlying *APIError is always
// available via errors.As for accessing StatusCode/RawBody.
var (
	// ErrUnauthorized is returned when the API responds with HTTP 401.
	ErrUnauthorized = errors.New("coinglass: unauthorized — check your API key")
	// ErrNotFound is returned when the API responds with HTTP 404.
	ErrNotFound = errors.New("coinglass: resource not found")
	// ErrRateLimited is returned when the API responds with HTTP 429 and
	// all retry attempts have been exhausted.
	ErrRateLimited = errors.New("coinglass: rate limit exceeded")
)

// wrappedAPIError wraps an *APIError together with a sentinel error so
// that both errors.As(&APIError{}) and errors.Is(sentinel) work on the
// value returned to the caller.
type wrappedAPIError struct {
	*APIError
	sentinel error
}

// Unwrap allows errors.Is/errors.As to traverse to the sentinel error.
func (w *wrappedAPIError) Unwrap() error {
	return w.sentinel
}

// Is reports whether this error matches target, supporting comparison
// against the wrapped sentinel error.
func (w *wrappedAPIError) Is(target error) bool {
	return errors.Is(w.sentinel, target)
}

// As allows errors.As(err, &apiErr) to retrieve the underlying
// *APIError from a wrapped error value.
func (w *wrappedAPIError) As(target interface{}) bool {
	if ptr, ok := target.(**APIError); ok {
		*ptr = w.APIError
		return true
	}
	return false
}

// newAPIError builds an error value for the given status code, message
// and raw body, attaching the appropriate sentinel error when the status
// code is a well-known one (401, 404, 429).
func newAPIError(statusCode int, code, message string, rawBody []byte) error {
	apiErr := &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		RawBody:    rawBody,
	}

	switch statusCode {
	case 401:
		return &wrappedAPIError{APIError: apiErr, sentinel: ErrUnauthorized}
	case 404:
		return &wrappedAPIError{APIError: apiErr, sentinel: ErrNotFound}
	case 429:
		return &wrappedAPIError{APIError: apiErr, sentinel: ErrRateLimited}
	default:
		return apiErr
	}
}
