package utils

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errFakeEOF simulates a generic network failure: never an errAPIError.
// (errFakeNetworkFailure lives in retry_utils_test.go, same package.)
var errFakeEOF = errors.New("EOF")

func TestParseHTTPError(t *testing.T) {
	t.Parallel()

	successStatuses := []int{http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent}

	t.Run("parses problem-details bodies and keeps the status code", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name       string
			body       string
			statusCode int
			want       *APIError
		}{
			{
				name:       "403 forbidden access",
				body:       `{"title":"Forbidden Access","type":"urn:numspot:errors#forbidden_access"}`,
				statusCode: http.StatusForbidden,
				want: &APIError{
					StatusCode: http.StatusForbidden,
					Type:       "urn:numspot:errors#forbidden_access",
					Title:      "Forbidden Access",
				},
			},
			{
				name:       "404 not_found",
				body:       `{"title":"Not Found","type":"urn:numspot:errors#not_found"}`,
				statusCode: http.StatusNotFound,
				want: &APIError{
					StatusCode: http.StatusNotFound,
					Type:       "urn:numspot:errors#not_found",
					Title:      "Not Found",
				},
			},
			{
				name:       "404 resource_not_found with detail",
				body:       `{"detail":"Vpc vpc-123 does not exist","title":"Resource Not Found","type":"urn:numspot:errors#resource_not_found"}`,
				statusCode: http.StatusNotFound,
				want: &APIError{
					StatusCode: http.StatusNotFound,
					Type:       "urn:numspot:errors#resource_not_found",
					Title:      "Resource Not Found",
					Detail:     "Vpc vpc-123 does not exist",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				err := ParseHTTPError([]byte(tt.body), tt.statusCode)
				require.Error(t, err)

				apiErr := &APIError{}
				require.True(t, errors.As(err, &apiErr))
				assert.Equal(t, tt.want, apiErr)
			})
		}
	})

	t.Run("success statuses return nil", func(t *testing.T) {
		t.Parallel()

		for _, statusCode := range successStatuses {
			assert.NoError(t, ParseHTTPError([]byte(`{"title":"unused"}`), statusCode))
		}
	})

	t.Run("non-JSON and empty bodies keep the legacy message", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			body string
			want string
		}{
			{name: "non-JSON body", body: "504 Gateway Timeout", want: "API error: 504 Gateway Timeout"},
			{name: "empty body", body: "", want: "API error: "},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				err := ParseHTTPError([]byte(tt.body), http.StatusBadRequest)
				require.Error(t, err)
				assert.Equal(t, tt.want, err.Error())

				apiErr := &APIError{}
				require.True(t, errors.As(err, &apiErr))
				assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
			})
		}
	})
}

func TestAPIErrorMessageRendering(t *testing.T) {
	t.Parallel()

	// The rendered message must stay identical to the historical
	// fmt.Errorf("%w: %s", errAPIError, ...) output.
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "title only",
			err:  HandleError([]byte(`{"title":"Forbidden Access","type":"urn:numspot:errors#forbidden_access"}`)),
			want: "API error: Forbidden Access",
		},
		{
			name: "title and detail",
			err:  HandleError([]byte(`{"title":"Title","detail":"Detail","type":"urn:numspot:errors#some_error"}`)),
			want: "API error: Title: Detail",
		},
		{
			name: "empty detail is not appended",
			err:  HandleError([]byte(`{"title":"Title","detail":""}`)),
			want: "API error: Title",
		},
		{
			name: "non-JSON body",
			err:  HandleError([]byte("not json")),
			want: "API error: not json",
		},
		{
			name: "empty body",
			err:  HandleError(nil),
			want: "API error: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestAPIErrorWrapsErrAPIError(t *testing.T) {
	t.Parallel()

	// Contract required by checkRetryCondition (retry_utils.go): API errors
	// are classified via errors.Is(err, errAPIError), network errors are not.
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "parsed API error",
			err:  HandleError([]byte(`{"title":"Forbidden Access"}`)),
			want: true,
		},
		{
			name: "wrapped once with WrapError",
			err:  WrapError(HandleError([]byte(`{"title":"Forbidden Access"}`)), "failed to read vpcs by id"),
			want: true,
		},
		{
			name: "wrapped with fmt.Errorf",
			err:  fmt.Errorf("ctx: %w", HandleError([]byte(`{"title":"Forbidden Access"}`))),
			want: true,
		},
		{
			name: "network error",
			err:  errFakeNetworkFailure,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, errors.Is(tt.err, errAPIError))
		})
	}
}

func TestIsResourceGone(t *testing.T) {
	t.Parallel()

	forbidden := ParseHTTPError(
		[]byte(`{"title":"Forbidden Access","type":"urn:numspot:errors#forbidden_access"}`),
		http.StatusForbidden,
	)
	notFound := ParseHTTPError(
		[]byte(`{"title":"Not Found","type":"urn:numspot:errors#not_found"}`),
		http.StatusNotFound,
	)
	notFoundUnparsable := ParseHTTPError([]byte("503 backend unavailable"), http.StatusNotFound)
	badRequest := ParseHTTPError(
		[]byte(`{"title":"Bad Request","type":"urn:numspot:errors#bad_request"}`),
		http.StatusBadRequest,
	)
	serverError := ParseHTTPError(
		[]byte(`{"title":"Internal Server Error","type":"urn:numspot:errors#internal_error"}`),
		http.StatusInternalServerError,
	)

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "403 forbidden", err: forbidden, want: true},
		{name: "404 not found", err: notFound, want: true},
		{name: "404 with unparsable body", err: notFoundUnparsable, want: true},
		{name: "400 bad request", err: badRequest, want: false},
		{name: "500 server error", err: serverError, want: false},
		{name: "generic network error", err: errFakeEOF, want: false},
		{
			name: "wrapped API error is traversed",
			err:  WrapError(fmt.Errorf("failed to read vpcs by id: %w", forbidden), "read"),
			want: true,
		},
		{
			name: "wrapped non-gone error stays false",
			err:  WrapError(fmt.Errorf("failed to read vpcs by id: %w", badRequest), "read"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsResourceGone(tt.err))
		})
	}
}

func TestIsResourceGoneViaHandleErrorCallers(t *testing.T) {
	t.Parallel()

	// HandleError call sites (e.g. core/nic.go, core/tags.go) do not carry a
	// status code: the error must not be classified as gone without one.
	err := WrapError(HandleError([]byte(`{"title":"Forbidden Access"}`)), "failed to parse API error response")

	assert.False(t, IsResourceGone(err))
	assert.Equal(t, "failed to parse API error response: API error: Forbidden Access", err.Error())
}
