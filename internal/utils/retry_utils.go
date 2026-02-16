package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"terraform-provider-numspot/internal/sdk/api"
)

type TfRequestResp interface {
	StatusCode() int
}

const (
	TfRequestRetryTimeout      = 5 * time.Minute
	TfRequestStateRetryTimeout = 15 * time.Minute
	TfRequestRetryDelay        = 5 * time.Second
)

var (
	StatusCodeRetryOnDelete     = []int{http.StatusConflict, http.StatusFailedDependency}
	StatusCodeStopRetryOnDelete = []int{http.StatusNoContent, http.StatusCreated, http.StatusBadRequest}
	StatusCodeRetryOnCreate     = []int{http.StatusConflict, http.StatusFailedDependency, http.StatusAccepted}
	StatusCodeStopRetryOnCreate = []int{http.StatusNoContent, http.StatusCreated, http.StatusOK, http.StatusForbidden}

	StateStopRetryOnCreate = []string{"RUNNING", "FAILED"}
	StateRetryOnCreate     = []string{"PENDING", "CREATING", "DELETING"}
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

func GetErrorMessage(res TfRequestResp) (string, error) {
	errorResponse, err := getFieldFromReflectStructPtr(reflect.ValueOf(res), "Body")
	if err != nil {
		return "", err
	}
	concreteErrorResponse, ok := errorResponse.Interface().([]byte)

	if !ok {
		return "", fmt.Errorf("failed to parse %v to byte array", errorResponse)
	}

	return HandleError(concreteErrorResponse).Error(), err
}

func checkRetryCondition(res TfRequestResp, err error, stopRetryCodes []int, retryCodes []int) *retry.RetryError {
	if err != nil {
		return retry.NonRetryableError(err)
	}

	if slices.Contains(stopRetryCodes, res.StatusCode()) {
		return nil
	} else {
		errorMessage, err := GetErrorMessage(res)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("error : got http status code %v but failed to parse error message. Reason : %v", res.StatusCode(), err))
		}

		if slices.Contains(retryCodes, res.StatusCode()) {
			time.Sleep(ParseRetryBackoff()) // Delay not handled in RetryContext. Might find a better solution later
			return retry.RetryableError(fmt.Errorf("error : retry timeout reached (%v). Error message : %v", TfRequestRetryTimeout, errorMessage))
		} else {
			return retry.NonRetryableError(errors.New(errorMessage))
		}
	}
}

func RetryDeleteUntilResourceAvailable[R TfRequestResp, id string | api.ResourceIdentifier](
	ctx context.Context,
	spaceID api.SpaceId,
	deleteId id,
	fun func(context.Context, api.SpaceId, id, ...api.RequestEditorFn) (R, error),
) error {
	var res R
	return retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		// tflog.Debug(ctx, fmt.Sprintf("Retry delete on resource: %s", id))
		res, err = fun(ctx, spaceID, deleteId)
		tflog.Debug(ctx, fmt.Sprintf("Retry delete got response: %d", res.StatusCode()))

		return checkRetryCondition(res, err, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate)
	})
}

func RetryCreateUntilResourceAvailable[R TfRequestResp](
	ctx context.Context,
	spaceID api.SpaceId,
	fun func(context.Context, api.SpaceId, ...api.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID)

		return checkRetryCondition(res, err, []int{http.StatusCreated}, []int{http.StatusConflict, http.StatusFailedDependency})
	})

	return res, retryError
}

func RetryCreateUntilResourceAvailableWithBody[R TfRequestResp, BodyType any](
	ctx context.Context,
	spaceID api.SpaceId,
	body BodyType,
	fun func(context.Context, api.SpaceId, BodyType, ...api.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID, body)

		return checkRetryCondition(res, err, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate)
	})

	return res, retryError
}

func RetryUntilResourceAvailableWithBody[R TfRequestResp, BodyType any](
	ctx context.Context,
	spaceID api.SpaceId,
	resourceID string,
	body BodyType,
	fun func(context.Context, api.SpaceId, string, BodyType, ...api.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID, resourceID, body)

		return checkRetryCondition(res, err, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate)
	})

	return res, retryError
}

func RetryDeleteUntilWithBody[R TfRequestResp, BodyType any](
	ctx context.Context,
	spaceID api.SpaceId,
	resourceID string,
	body BodyType,
	fun func(context.Context, api.SpaceId, string, BodyType, ...api.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID, resourceID, body)

		return checkRetryCondition(res, err, StatusCodeStopRetryOnDelete, StatusCodeRetryOnDelete)
	})

	return res, retryError
}

func getFieldFromReflectStructPtr(structPtr reflect.Value, fieldName string) (reflect.Value, error) {
	if structPtr.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("expected a pointer but found %v", structPtr)
	}

	structValue := structPtr.Elem()

	if structValue.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("expected a struct but found %v", structValue)
	}

	fieldValue := structValue.FieldByName(fieldName)

	if !fieldValue.IsValid() {
		return reflect.Value{}, fmt.Errorf("expected field '%s' in struct but found %v", fieldName, fieldValue)
	}

	return fieldValue, nil
}

func getFieldFromReflectStruct(structValue reflect.Value, fieldName string) (reflect.Value, error) {
	if structValue.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("expected a struct but found %v", structValue)
	}

	fieldValue := structValue.FieldByName(fieldName)

	if !fieldValue.IsValid() {
		return reflect.Value{}, fmt.Errorf("expected field '%s' in struct but found %v", fieldName, fieldValue)
	}

	return fieldValue, nil
}

func ReadResourceUtils[R TfRequestResp, id string | api.ResourceIdentifier](
	ctx context.Context,
	createdId id,
	spaceID api.SpaceId,
	readFunction func(context.Context, api.SpaceId, id, ...api.RequestEditorFn) (*R, error),
) (interface{}, string, error) {
	readRes, err := readFunction(ctx, spaceID, createdId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read resource : %v", err.Error())
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
	if stateValuePtr.Kind() == reflect.Ptr {
		stateValue = stateValuePtr.Elem()

		if stateValue.Type().String() != "string" {
			return nil, "", fmt.Errorf("field 'State' was expected to be a string but %v found", stateValue.Type())
		}

	} else if stateValuePtr.Kind() == reflect.String {
		stateValue = reflect.ValueOf(stateValuePtr.Interface())
	} else {
		return nil, "", fmt.Errorf("expected a pointer or a string but found %v", stateValuePtr)
	}

	data := json200ValuePtr.Interface()
	stateValueStr := stateValue.String()

	return data, stateValueStr, nil
}

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
	}

	return createStateConf.WaitForStateContext(ctx)
}

func RetryReadUntilStatusStateValid[R TfRequestResp, ID string | api.ResourceIdentifier](
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
			return ReadStatusResourceUtils(ctx, createdId, spaceID, readFunction)
		},
		Timeout: TfRequestStateRetryTimeout,
		Delay:   ParseRetryBackoff(),
	}

	return createStateConf.WaitForStateContext(ctx)
}

func ReadStatusResourceUtils[R TfRequestResp, id string | api.ResourceIdentifier](
	ctx context.Context,
	createdId id,
	spaceID api.SpaceId,
	readFunction func(context.Context, api.SpaceId, id, ...api.RequestEditorFn) (*R, error),
) (interface{}, string, error) {
	readRes, err := readFunction(ctx, spaceID, createdId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read resource : %v", err.Error())
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
	if stateValuePtr.Kind() == reflect.Ptr {
		stateValue = stateValuePtr.Elem()

		if stateValue.Type().String() != "string" {
			return nil, "", fmt.Errorf("field 'State' was expected to be a string but %v found", stateValue.Type())
		}

	} else if stateValuePtr.Kind() == reflect.String {
		stateValue = reflect.ValueOf(stateValuePtr.Interface())
	} else {
		return nil, "", fmt.Errorf("expected a pointer or a string but found %v", stateValuePtr)
	}

	data := json200ValuePtr.Interface()
	stateValueStr := stateValue.String()

	return data, stateValueStr, nil
}

func RetryReadUntilStatusStateValidWith2ID[R TfRequestResp, ID1, ID2 string | api.ResourceIdentifier](
	ctx context.Context,
	parentID ID1,
	childID ID2,
	spaceID api.SpaceId,
	pendingStates []string,
	targetStates []string,
	readFunction func(context.Context, api.SpaceId, ID1, ID2, ...api.RequestEditorFn) (*R, error),
) (interface{}, error) {
	createStateConf := &retry.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			return ReadStatusResourceUtilsWith2ID(ctx, parentID, childID, spaceID, readFunction)
		},
		Timeout: TfRequestRetryTimeout,
		Delay:   ParseRetryBackoff(),
	}

	return createStateConf.WaitForStateContext(ctx)
}

func ReadStatusResourceUtilsWith2ID[R TfRequestResp, ID1, ID2 string | api.ResourceIdentifier](
	ctx context.Context,
	parentID ID1,
	childID ID2,
	spaceID api.SpaceId,
	readFunction func(context.Context, api.SpaceId, ID1, ID2, ...api.RequestEditorFn) (*R, error),
) (interface{}, string, error) {
	readRes, err := readFunction(ctx, spaceID, parentID, childID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read resource : %v", err.Error())
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
	if stateValuePtr.Kind() == reflect.Ptr {
		stateValue = stateValuePtr.Elem()

		if stateValue.Type().String() != "string" {
			return nil, "", fmt.Errorf("field 'State' was expected to be a string but %v found", stateValue.Type())
		}

	} else if stateValuePtr.Kind() == reflect.String {
		stateValue = reflect.ValueOf(stateValuePtr.Interface())
	} else {
		return nil, "", fmt.Errorf("expected a pointer or a string but found %v", stateValuePtr)
	}

	data := json200ValuePtr.Interface()
	stateValueStr := stateValue.String()

	return data, stateValueStr, nil
}
