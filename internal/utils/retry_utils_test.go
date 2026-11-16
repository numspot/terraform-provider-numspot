package utils

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTfResponse is a minimal TfRequestResp double carrying the "Body" field
// that GetErrorMessage reflects on.
type fakeTfResponse struct {
	Body       []byte
	statusCode int
}

func (f fakeTfResponse) StatusCode() int { return f.statusCode }

// errFakeNetworkFailure simulates a network-level failure (connection reset,
// EOF): checkRetryCondition must treat it as retryable.
var errFakeNetworkFailure = errors.New("connection reset by peer")

const forbiddenBody = `{"title":"Forbidden Access","type":"urn:numspot:errors#forbidden_access"}`

func TestCheckRetryConditionCreate(t *testing.T) {
	// Contract after the nil-JSON201 panic fix: only true 2xx statuses may
	// stop the retry loop as a success. Every other status must produce a
	// retry.RetryError (never a silent success), because create callers
	// dereference the parsed JSON201 payload right after the loop.
	//
	// Not parallel: t.Setenv below must not race with parallel subtests.
	t.Setenv("RETRY_BACKOFF", "1ms")
	tests := []struct {
		name            string
		res             TfRequestResp
		err             error
		wantNil         bool
		wantRetryable   bool
		wantErrContains string
	}{
		{
			name:    "201 created is a success",
			res:     &fakeTfResponse{statusCode: http.StatusCreated},
			wantNil: true,
		},
		{
			name:    "200 ok is a success",
			res:     &fakeTfResponse{statusCode: http.StatusOK},
			wantNil: true,
		},
		{
			name:    "204 no content is a success",
			res:     &fakeTfResponse{statusCode: http.StatusNoContent},
			wantNil: true,
		},
		{
			name: "403 forbidden is a non-retryable error, not a success",
			res: &fakeTfResponse{
				statusCode: http.StatusForbidden,
				Body:       []byte(forbiddenBody),
			},
			wantRetryable:   false,
			wantErrContains: "Forbidden Access",
		},
		{
			name: "500 without stop/retry status is a non-retryable error",
			res: &fakeTfResponse{
				statusCode: http.StatusInternalServerError,
				Body:       []byte(`{"title":"Internal Server Error"}`),
			},
			wantRetryable:   false,
			wantErrContains: "Internal Server Error",
		},
		{
			name: "409 conflict is retryable",
			res: &fakeTfResponse{
				statusCode: http.StatusConflict,
				Body:       []byte(`{"title":"Conflict"}`),
			},
			wantRetryable: true,
		},
		{
			name: "429 too many requests is retryable",
			res: &fakeTfResponse{
				statusCode: http.StatusTooManyRequests,
				Body:       []byte(`{"title":"Too Many Requests","type":"urn:numspot:errors#too_many_requests"}`),
			},
			wantRetryable: true,
		},
		{
			name:          "network error is retryable",
			err:           errFakeNetworkFailure,
			wantRetryable: true,
		},
		{
			name:            "wrapped API error is non-retryable",
			res:             &fakeTfResponse{statusCode: http.StatusInternalServerError},
			err:             fmt.Errorf("wrapped: %w", &APIError{StatusCode: 500, Title: "boom"}),
			wantRetryable:   false,
			wantErrContains: "wrapped",
		},
		{
			name:            "nil response is retryable",
			err:             nil,
			wantRetryable:   true,
			wantErrContains: "nil response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Not parallel: relies on the RETRY_BACKOFF set by the parent.
			got := checkRetryCondition(
				tt.res, tt.err, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate,
			)

			if tt.wantNil {
				require.Nil(t, got, "expected retry loop success, got %v", got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.wantRetryable, got.Retryable)
			assert.Contains(t, got.Err.Error(), tt.wantErrContains)
		})
	}
}

// TestCheckRetryConditionForbiddenIsNotSuccess is the focused regression test
// for the provider crash on numspot_vpc create: a 403 used to stop the retry
// loop as a success, then the caller panicked dereferencing a nil JSON201.
func TestCheckRetryConditionForbiddenIsNotSuccess(t *testing.T) {
	t.Parallel()

	res := &fakeTfResponse{statusCode: http.StatusForbidden, Body: []byte(forbiddenBody)}
	got := checkRetryCondition(res, nil, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate)

	require.NotNil(t, got, "403 must not be reported as a retry-loop success")
	assert.False(t, got.Retryable, "403 is a permanent error and must not be retried")
	assert.ErrorContains(t, got.Err, "Forbidden Access")
}
