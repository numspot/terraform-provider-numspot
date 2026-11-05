package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
)

// TfRequestResp handles TF request response.
type TfRequestResp interface {
	StatusCode() int
}

const (
	// TfRequestRetryTimeout is the timeout for retrying Terraform requests.
	TfRequestRetryTimeout = 5 * time.Minute
	// TfRequestStateRetryTimeout is the timeout for retrying until a resource reaches its target state.
	TfRequestStateRetryTimeout = 15 * time.Minute
	// TfRequestRetryDelay is the delay between retries.
	TfRequestRetryDelay = 5 * time.Second
)

// errNilResponse is returned when a retry callback yields a nil typed
// response (deletes on 4xx responses may return a nil pointer).
var errNilResponse = errors.New("nil response received")

var (
	// StatusCodeRetryOnDelete handles status code retry on delete.
	StatusCodeRetryOnDelete = []int{http.StatusConflict, http.StatusFailedDependency}
	// StatusCodeStopRetryOnDelete lists the HTTP status codes that stop retry on delete.
	// 202 Accepted: async deletes (e.g. kubernetes cluster teardown) answer
	// 202 immediately; same success contract as utils.ParseHTTPError.
	StatusCodeStopRetryOnDelete = []int{http.StatusNoContent, http.StatusCreated, http.StatusAccepted, http.StatusBadRequest}
	// StatusCodeRetryOnCreate handles status code retry on create.
	StatusCodeRetryOnCreate = []int{http.StatusConflict, http.StatusFailedDependency, http.StatusAccepted}
	// StatusCodeStopRetryOnCreate handles status code stop retry on create.
	StatusCodeStopRetryOnCreate = []int{http.StatusNoContent, http.StatusCreated, http.StatusOK, http.StatusForbidden}

	// StateStopRetryOnCreate handles state stop retry on create.
	StateStopRetryOnCreate = []string{"RUNNING", "FAILED"}
	// StateRetryOnCreate handles state retry on create.
	StateRetryOnCreate = []string{"PENDING", "CREATING", "DELETING"}

	// errFailedToParseToByteArray is returned when the response body cannot be
	// cast to a byte slice.
	errFailedToParseToByteArray = errors.New("failed to parse value to byte array")
	// errHTTPErrorMessageParseFailed is returned when the error body of a failed
	// HTTP response cannot be parsed.
	errHTTPErrorMessageParseFailed = errors.New("error : got http status code")
	// errRetryTimeoutReached is returned when a retry loop exhausts its timeout.
	errRetryTimeoutReached = errors.New("error : retry timeout reached")
	// errUnexpectedHTTPStatus is the static sentinel for unexpected HTTP status
	// codes that do not carry a parsable error body.
	errUnexpectedHTTPStatus = errors.New("unexpected HTTP status")
	// errExpectedPointer is returned when a reflect.Value is not a pointer.
	errExpectedPointer = errors.New("expected a pointer but found")
	// errExpectedStruct is returned when a reflect.Value is not a struct.
	errExpectedStruct = errors.New("expected a struct but found")
	// errExpectedField is returned when a struct field is missing.
	errExpectedField = errors.New("expected field in struct but found")
	// errStateExpectedString is returned when a State field is not a string.
	errStateExpectedString = errors.New("field 'State' was expected to be a string but found")
	// errStateExpectedPointerOrString is returned when a State field has an
	// unexpected kind.
	errStateExpectedPointerOrString = errors.New("expected a pointer or a string but found")
)

// ParseRetryBackoff retrieves the retry backoff duration from the
// "RETRY_BACKOFF" environment variable. If the environment variable is not set
// or contains an invalid value, it defaults to 5 seconds.
//
// The environment variable "RETRY_BACKOFF" should be a valid string representing
// a duration, such as "500ms", "1s", "2m", etc., as accepted by time.ParseDuration.
//
// Returns:
//   - time.Duration: The duration parsed from the environment variable, or 5 seconds if
//     the environment variable is not set or contains an invalid value.
//
// Example environment variable:
//
//	RETRY_BACKOFF="2s"  // This sets the retry backoff to 2 seconds.
//
// TODO:
//
// A refactoring should be made later to this,
// All this logic should be encapsulated into an object.
//
// env variable parsing should be done provider bootstrapping phase.
func ParseRetryBackoff() time.Duration {
	// Default retry backoff set to 5s
	retryBackoff := TfRequestRetryDelay

	retryBackoffStr := os.Getenv("RETRY_BACKOFF")
	// Parse the string into time.Duration
	val, err := time.ParseDuration(retryBackoffStr)
	if err == nil {
		retryBackoff = val
	}

	return retryBackoff
}

// retryBackoffPollInterval returns the fixed interval to wait between state
// polls when RETRY_BACKOFF is explicitly set (offline acceptance tests replay
// recorded HTTP responses, so the only real cost is the wait itself). When the
// variable is unset — the production case — it returns 0, which keeps the
// helper's built-in exponential backoff: no behavior change outside tests.
func retryBackoffPollInterval() time.Duration {
	if os.Getenv("RETRY_BACKOFF") == "" {
		return 0
	}
	return ParseRetryBackoff()
}

// retryContext wraps helper/retry.RetryContext to honor RETRY_BACKOFF. The SDK
// implementation hardcodes a MinTimeout of 500ms and an exponential inter-poll
// backoff (up to 10s) with no way to override it, which made replayed retry
// sequences (e.g. repeated 409 on create) re-live the real API pacing. When
// RETRY_BACKOFF is unset — the production case — it delegates to the SDK
// function unchanged; the replica below only runs when the variable is set,
// replacing the exponential ramp with a fixed poll interval.
func retryContext(ctx context.Context, timeout time.Duration, f retry.RetryFunc) error {
	if os.Getenv("RETRY_BACKOFF") == "" {
		return retry.RetryContext(ctx, timeout, f)
	}

	// Replica of helper/retry.RetryContext (terraform-plugin-sdk/v2
	// helper/retry/wait.go) with a fixed PollInterval and no MinTimeout floor.
	var resultErr error
	var resultErrMu sync.Mutex

	conf := &retry.StateChangeConf{
		Pending:      []string{"retryableerror"},
		Target:       []string{"success"},
		Timeout:      timeout,
		PollInterval: retryBackoffPollInterval(),
		Refresh: func() (interface{}, string, error) {
			rerr := f()

			resultErrMu.Lock()
			defer resultErrMu.Unlock()

			if rerr == nil {
				resultErr = nil
				return 42, "success", nil
			}

			resultErr = rerr.Err

			if rerr.Retryable {
				return 42, "retryableerror", nil
			}
			return nil, "quit", rerr.Err
		},
	}

	_, waitErr := conf.WaitForStateContext(ctx)

	// resultErr may be nil because the wait timed out and resultErr was never
	// set; this is still an error.
	resultErrMu.Lock()
	defer resultErrMu.Unlock()
	if resultErr == nil {
		return waitErr
	}
	// resultErr takes precedence over waitErr if both are set because it is
	// more likely to be useful.
	return resultErr
}

// GetErrorMessage returns error message.
func GetErrorMessage(res TfRequestResp) (string, error) {
	errorResponse, err := getFieldFromReflectStructPtr(reflect.ValueOf(res), "Body")
	if err != nil {
		return "", err
	}
	concreteErrorResponse, ok := errorResponse.Interface().([]byte)

	if !ok {
		return "", fmt.Errorf("%w: %v", errFailedToParseToByteArray, errorResponse)
	}

	return HandleError(concreteErrorResponse).Error(), err
}

func checkRetryCondition(res TfRequestResp, err error, stopRetryCodes, retryCodes []int) *retry.RetryError {
	if err != nil {
		// Network-level failures (connection reset, timeout, EOF) are transient:
		// retry them. Real API errors are wrapped with errAPIError and carry a
		// status code, so they fall through to the status-code logic below.
		if !errors.Is(err, errAPIError) {
			return retry.RetryableError(err)
		}
		return retry.NonRetryableError(err)
	}
	if isNilResponse(res) {
		return retry.RetryableError(errNilResponse)
	}

	if slices.Contains(stopRetryCodes, res.StatusCode()) {
		return nil
	} else {
		errorMessage, err := GetErrorMessage(res)
		if err != nil {
			return retry.NonRetryableError(
				fmt.Errorf("%w %v but failed to parse error message. Reason : %v", errHTTPErrorMessageParseFailed, res.StatusCode(), err),
			)
		}

		if slices.Contains(retryCodes, res.StatusCode()) {
			time.Sleep(ParseRetryBackoff()) // Delay not handled in RetryContext. Might find a better solution later
			return retry.RetryableError(fmt.Errorf("%w (%v). Error message : %v", errRetryTimeoutReached, TfRequestRetryTimeout, errorMessage))
		} else {
			if errorMessage == "" {
				// Non-error bodies (e.g. a 2xx shape without `title`) parse
				// into an empty api.Error; never surface a blank message.
				errorMessage = fmt.Sprintf("%d with no parsable error body", res.StatusCode())
			}
			return retry.NonRetryableError(fmt.Errorf("%w %s", errUnexpectedHTTPStatus, errorMessage))
		}
	}
}

// RetryDeleteUntilResourceAvailable retries delete until resource available.
func RetryDeleteUntilResourceAvailable[R TfRequestResp, id string | api.ResourceIdentifier](
	ctx context.Context,
	spaceID api.SpaceId,
	deleteId id,
	fun func(context.Context, api.SpaceId, id, ...api.RequestEditorFn) (R, error),
) error {
	var res R
	retryError := retryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID, deleteId)
		if !isNilResponse(res) {
			tflog.Debug(ctx, fmt.Sprintf("Retry delete got response: %d", res.StatusCode()))
		}

		return checkRetryCondition(res, err, StatusCodeStopRetryOnDelete, StatusCodeRetryOnDelete)
	})

	return WrapError(retryError, "failed to retry delete operation")
}

// RetryCreateUntilResourceAvailable retries create until resource available.
func RetryCreateUntilResourceAvailable[R TfRequestResp](
	ctx context.Context,
	spaceID api.SpaceId,
	fun func(context.Context, api.SpaceId, ...api.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID)

		return checkRetryCondition(res, err, []int{http.StatusCreated}, []int{http.StatusConflict, http.StatusFailedDependency})
	})

	return res, WrapError(retryError, "failed to retry create operation")
}

// RetryCreateUntilResourceAvailableWithBody retries create until resource available with body.
func RetryCreateUntilResourceAvailableWithBody[R TfRequestResp, BodyType any](
	ctx context.Context,
	spaceID api.SpaceId,
	body BodyType,
	fun func(context.Context, api.SpaceId, BodyType, ...api.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID, body)

		return checkRetryCondition(res, err, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate)
	})

	return res, WrapError(retryError, "failed to retry create operation")
}

// RetryUntilResourceAvailableWithBody retries until resource available with body.
func RetryUntilResourceAvailableWithBody[R TfRequestResp, BodyType any](
	ctx context.Context,
	spaceID api.SpaceId,
	resourceID string,
	body BodyType,
	fun func(context.Context, api.SpaceId, string, BodyType, ...api.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID, resourceID, body)

		return checkRetryCondition(res, err, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate)
	})

	return res, WrapError(retryError, "failed to retry update operation")
}

// RetryDeleteUntilWithBody retries delete until with body.
func RetryDeleteUntilWithBody[R TfRequestResp, BodyType any](
	ctx context.Context,
	spaceID api.SpaceId,
	resourceID string,
	body BodyType,
	fun func(context.Context, api.SpaceId, string, BodyType, ...api.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID, resourceID, body)

		return checkRetryCondition(res, err, StatusCodeStopRetryOnDelete, StatusCodeRetryOnDelete)
	})

	return res, WrapError(retryError, "failed to retry delete operation")
}

func getFieldFromReflectStructPtr(structPtr reflect.Value, fieldName string) (reflect.Value, error) {
	if structPtr.Kind() != reflect.Pointer {
		return reflect.Value{}, fmt.Errorf("%w %v", errExpectedPointer, structPtr)
	}

	structValue := structPtr.Elem()

	if structValue.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("%w %v", errExpectedStruct, structValue)
	}

	fieldValue := structValue.FieldByName(fieldName)

	if !fieldValue.IsValid() {
		return reflect.Value{}, fmt.Errorf("%w '%s': %v", errExpectedField, fieldName, fieldValue)
	}

	return fieldValue, nil
}

func getFieldFromReflectStruct(structValue reflect.Value, fieldName string) (reflect.Value, error) {
	if structValue.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("%w %v", errExpectedStruct, structValue)
	}

	fieldValue := structValue.FieldByName(fieldName)

	if !fieldValue.IsValid() {
		return reflect.Value{}, fmt.Errorf("%w '%s': %v", errExpectedField, fieldName, fieldValue)
	}

	return fieldValue, nil
}

// ReadResourceUtils reads resource utils.
func ReadResourceUtils[R TfRequestResp, id string | api.ResourceIdentifier](
	ctx context.Context,
	createdId id,
	spaceID api.SpaceId,
	readFunction func(context.Context, api.SpaceId, id, ...api.RequestEditorFn) (*R, error),
) (data interface{}, state string, err error) {
	readRes, err := readFunction(ctx, spaceID, createdId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read resource: %w", err)
	}

	// Use reflection to access State attribute inside of interface object
	json200ValuePtr, err := getFieldFromReflectStructPtr(reflect.ValueOf(readRes), "JSON200")
	if err != nil {
		return nil, "", err
	}
	stateValuePtr, err := getFieldFromReflectStructPtr(json200ValuePtr, "State")
	if err != nil {
		return nil, "", err
	}

	var stateValue reflect.Value
	switch stateValuePtr.Kind() {
	case reflect.Pointer:
		stateValue = stateValuePtr.Elem()

		if stateValue.Type().String() != "string" {
			return nil, "", fmt.Errorf("%w %v", errStateExpectedString, stateValue.Type())
		}

	case reflect.String:
		stateValue = reflect.ValueOf(stateValuePtr.Interface())
	default:
		return nil, "", fmt.Errorf("%w %v", errStateExpectedPointerOrString, stateValuePtr)
	}

	return json200ValuePtr.Interface(), stateValue.String(), nil
}

// RetryReadUntilStateValid retries read until state valid.
func RetryReadUntilStateValid[R TfRequestResp, ID string | api.ResourceIdentifier](
	ctx context.Context,
	createdId ID,
	spaceID api.SpaceId,
	pendingStates []string,
	targetStates []string,
	readFunction func(context.Context, api.SpaceId, ID, ...api.RequestEditorFn) (*R, error),
) (interface{}, error) {
	createStateConf := &retry.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			return ReadResourceUtils(ctx, createdId, spaceID, readFunction)
		},
		Timeout: TfRequestRetryTimeout,
		Delay:   ParseRetryBackoff(),
		// When RETRY_BACKOFF is explicitly set (offline tests replaying
		// recorded responses), poll at a fixed interval instead of the
		// helper's built-in exponential backoff (100ms doubling, 10s cap),
		// which ignores the backoff knob and made replayed status-polling
		// sequences re-live the real provisioning pace (~53 pending polls ≈
		// 7 min). Unset (production) returns 0, restoring the SDK's original
		// exponential backoff: no behavior change.
		PollInterval: retryBackoffPollInterval(),
	}

	read, err := createStateConf.WaitForStateContext(ctx)
	return read, WrapError(err, "failed to wait for resource state")
}

// RetryReadUntilStatusStateValid retries read until status state valid.
func RetryReadUntilStatusStateValid[R TfRequestResp, ID string | api.ResourceIdentifier](
	ctx context.Context,
	createdId ID,
	spaceID api.SpaceId,
	pendingStates []string,
	targetStates []string,
	readFunction func(context.Context, api.SpaceId, ID, ...api.RequestEditorFn) (*R, error),
) (interface{}, error) {
	return RetryReadUntilStatusStateValidWithTimeout(ctx, createdId, spaceID, pendingStates, targetStates, readFunction, TfRequestStateRetryTimeout)
}

// RetryReadUntilStatusStateValidWithTimeout retries read until status state valid with timeout.
func RetryReadUntilStatusStateValidWithTimeout[R TfRequestResp, ID string | api.ResourceIdentifier](
	ctx context.Context,
	createdId ID,
	spaceID api.SpaceId,
	pendingStates []string,
	targetStates []string,
	readFunction func(context.Context, api.SpaceId, ID, ...api.RequestEditorFn) (*R, error),
	timeout time.Duration,
) (interface{}, error) {
	createStateConf := &retry.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			return ReadStatusResourceUtils(ctx, createdId, spaceID, readFunction)
		},
		Timeout:      timeout,
		Delay:        ParseRetryBackoff(),
		PollInterval: retryBackoffPollInterval(), // unset RETRY_BACKOFF → SDK's exponential backoff (0)
	}

	read, err := createStateConf.WaitForStateContext(ctx)
	return read, WrapError(err, "failed to wait for resource state")
}

// ReadStatusResourceUtils reads status resource utils.
func ReadStatusResourceUtils[R TfRequestResp, id string | api.ResourceIdentifier](
	ctx context.Context,
	createdId id,
	spaceID api.SpaceId,
	readFunction func(context.Context, api.SpaceId, id, ...api.RequestEditorFn) (*R, error),
) (data interface{}, state string, err error) {
	readRes, err := readFunction(ctx, spaceID, createdId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read resource: %w", err)
	}

	// Use reflection to access State attribute inside of interface object
	json200ValuePtr, err := getFieldFromReflectStructPtr(reflect.ValueOf(readRes), "JSON200")
	if err != nil {
		return nil, "", err
	}
	statusValuePtr, err := getFieldFromReflectStructPtr(json200ValuePtr, "Status")
	if err != nil {
		return nil, "", err
	}
	stateValuePtr, err := getFieldFromReflectStruct(statusValuePtr, "State")
	if err != nil {
		return nil, "", err
	}

	var stateValue reflect.Value
	switch stateValuePtr.Kind() {
	case reflect.Pointer:
		stateValue = stateValuePtr.Elem()

		if stateValue.Type().String() != "string" {
			return nil, "", fmt.Errorf("%w %v", errStateExpectedString, stateValue.Type())
		}

	case reflect.String:
		stateValue = reflect.ValueOf(stateValuePtr.Interface())
	default:
		return nil, "", fmt.Errorf("%w %v", errStateExpectedPointerOrString, stateValuePtr)
	}

	return json200ValuePtr.Interface(), stateValue.String(), nil
}

// isNilResponse reports whether the generic response value is a nil pointer
// (deletes on 4xx responses may return a nil typed response).
func isNilResponse(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}
