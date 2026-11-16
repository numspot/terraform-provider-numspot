package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
)

// errAPIError is the static sentinel for API error responses.
var errAPIError = errors.New("API error")

// APIError represents a Numspot API error response (RFC 9457 problem details).
// It keeps the HTTP status code and the problem URN alongside the historical
// message so callers can classify errors without re-parsing the body.
type APIError struct {
	// StatusCode is the HTTP status code of the response (0 when unknown,
	// e.g. when the error is built without an HTTP context).
	StatusCode int
	// Type is the RFC 9457 URN of the error, e.g. urn:numspot:errors#forbidden_access.
	Type string
	// Title is the human-readable summary of the problem type.
	Title string
	// Detail is the human-readable explanation specific to this occurrence.
	Detail string
}

// Error renders the error message. The output is identical to the historical
// fmt.Errorf("%w: %s", errAPIError, ...) rendering, so user-visible messages
// do not change.
func (e *APIError) Error() string {
	errorString := e.Title
	if e.Detail != "" {
		errorString = errorString + ": " + e.Detail
	}

	return errAPIError.Error() + ": " + errorString
}

// Unwrap exposes the static errAPIError sentinel. checkRetryCondition
// (retry_utils.go) relies on errors.Is(err, errAPIError) to distinguish
// network failures (retryable) from API errors (non-retryable): this contract
// must be preserved.
func (e *APIError) Unwrap() error {
	return errAPIError
}

// ParseHTTPError handles parse HTTP error.
func ParseHTTPError(httpResponseBody []byte, statusCode int) error {
	if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusAccepted || statusCode == http.StatusNoContent {
		return nil
	}

	apiErr := HandleError(httpResponseBody)
	apiErr.StatusCode = statusCode

	return apiErr
}

// HandleError handles handle error.
func HandleError(httpResponseBody []byte) *APIError {
	var apiError api.Error

	err := json.Unmarshal(httpResponseBody, &apiError)
	if err != nil && len(httpResponseBody) != 0 {
		return &APIError{Title: string(httpResponseBody)}
	}

	var detail string
	if apiError.Detail != nil {
		detail = *apiError.Detail
	}

	return &APIError{
		Type:   apiError.Type,
		Title:  apiError.Title,
		Detail: detail,
	}
}

// IsResourceGone reports whether err is an API error meaning the resource
// cannot be seen: 404 (introuvable) ou 403 (l'IAM masque l'existence).
func IsResourceGone(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound || apiErr.StatusCode == http.StatusForbidden
	}

	return false
}
