package retry_utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type TfRequestResp interface {
	StatusCode() int
}

const (
	TfRequestRetryTimeout = 5 * time.Minute
	TfRequestRetryDelay   = 5 * time.Second
)

var (
	StatusCodeRetryOnDelete     = []int{http.StatusConflict, http.StatusFailedDependency}
	StatusCodeStopRetryOnDelete = []int{http.StatusNoContent, http.StatusCreated}
	StatusCodeRetryOnCreate     = []int{http.StatusConflict, http.StatusFailedDependency}
	StatusCodeStopRetryOnCreate = []int{http.StatusNoContent, http.StatusCreated}
)

func getErrorMessage(res TfRequestResp) (string, error) {
	errorResponse, err := getFieldFromReflectStructPtr(reflect.ValueOf(res), "Body")
	if err != nil {
		return "", err
	}
	concreteErrorResponse, ok := errorResponse.Interface().([]byte)

	if !ok {
		return "", fmt.Errorf("failed to parse %v to byte array", errorResponse)
	}

	return utils.HandleError(concreteErrorResponse).Error(), err
}

func checkRetryCondition(res TfRequestResp, err error, stopRetryCodes []int, retryCodes []int) *retry.RetryError {
	if err != nil {
		return retry.NonRetryableError(err)
	}

	if slices.Contains(stopRetryCodes, res.StatusCode()) {
		return nil
	} else {
		errorMessage, err := getErrorMessage(res)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("error : got http status code %v but failed to parse error message. Reason : %v", res.StatusCode(), err))
		}

		if slices.Contains(retryCodes, res.StatusCode()) {
			time.Sleep(TfRequestRetryDelay) // Delay not handled in RetryContext. Might find a better solution later
			return retry.RetryableError(fmt.Errorf("error : retry timeout reached (%v). Error message : %v", TfRequestRetryTimeout, errorMessage))
		} else {
			return retry.NonRetryableError(errors.New(errorMessage))
		}
	}
}

func RetryUnlinkUntilSuccess[A any, R TfRequestResp](
	ctx context.Context,
	spaceId iaas.SpaceId,
	resourceId string,
	unlinkBody A,
	fun func(context.Context, iaas.SpaceId, string, A, ...iaas.RequestEditorFn) (R, error),
) error {
	return retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		tflog.Debug(ctx, fmt.Sprintf("Retry delete on resource: %s", resourceId))
		res, err := fun(ctx, spaceId, resourceId, unlinkBody)
		tflog.Debug(ctx, fmt.Sprintf("Retry delete got response: %d", res.StatusCode()))

		return checkRetryCondition(res, err, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate)
	})
}

func RetryDeleteUntilResourceAvailable[R TfRequestResp](
	ctx context.Context,
	spaceID iaas.SpaceId,
	id string,
	fun func(context.Context, iaas.SpaceId, string, ...iaas.RequestEditorFn) (R, error),
) error {
	return retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		tflog.Debug(ctx, fmt.Sprintf("Retry delete on resource: %s", id))
		res, err := fun(ctx, spaceID, id)
		tflog.Debug(ctx, fmt.Sprintf("Retry delete got response: %d", res.StatusCode()))

		return checkRetryCondition(res, err, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate)
	})
}

func RetryCreateUntilResourceAvailable[R TfRequestResp](
	ctx context.Context,
	spaceID iaas.SpaceId,
	fun func(context.Context, iaas.SpaceId, ...iaas.RequestEditorFn) (R, error),
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
	spaceID iaas.SpaceId,
	body BodyType,
	fun func(context.Context, iaas.SpaceId, BodyType, ...iaas.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID, body)

		return checkRetryCondition(res, err, StatusCodeStopRetryOnCreate, StatusCodeRetryOnCreate)
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

func ReadResourceUtils[R TfRequestResp](
	ctx context.Context,
	createdId string,
	spaceID iaas.SpaceId,
	readFunction func(context.Context, iaas.SpaceId, string, ...iaas.RequestEditorFn) (*R, error),
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

	if stateValuePtr.Kind() != reflect.Ptr {
		return nil, "", fmt.Errorf("expected a pointer but found %v", stateValuePtr)
	}

	stateValue := stateValuePtr.Elem()

	if stateValue.Type().String() != "string" {
		return nil, "", fmt.Errorf("field 'State' was expected to be a string but %v found", stateValue.Type())
	}

	data := json200ValuePtr.Interface()
	stateValueStr := stateValue.String()

	return data, stateValueStr, nil
}

func RetryReadUntilStateValid[R TfRequestResp](
	ctx context.Context,
	createdId string,
	spaceID iaas.SpaceId,
	pendingStates []string,
	targetStates []string,
	readFunction func(context.Context, iaas.SpaceId, string, ...iaas.RequestEditorFn) (*R, error),
) (interface{}, error) {
	createStateConf := &retry.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			return ReadResourceUtils(ctx, createdId, spaceID, readFunction)
		},
		Timeout: TfRequestRetryTimeout,
		Delay:   TfRequestRetryDelay,
	}

	return createStateConf.WaitForStateContext(ctx)
}
