package utils

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
)

type (
	TfRequestResp interface {
		StatusCode() int
	}

	ITFValue interface {
		Type(ctx context.Context) attr.Type
	}

	TFType interface {
		IsNull() bool
		IsUnknown() bool
	}
)

const (
	TfRequestRetryTimeout = 5 * time.Minute
	TfRequestRetryDelay   = 5 * time.Second
)

func FromTfStringToStringPtr(str types.String) *string {
	if str.IsUnknown() || str.IsNull() {
		return nil
	}

	return str.ValueStringPointer()
}

func FromTfBoolToBoolPtr(bl types.Bool) *bool {
	if bl.IsUnknown() || bl.IsNull() {
		return nil
	}

	return bl.ValueBoolPointer()
}

func FromTfInt64ToIntPtr(tfInt types.Int64) *int {
	if tfInt.ValueInt64Pointer() != nil {
		val := int(tfInt.ValueInt64())
		return &val
	}
	return nil
}

func FromTfInt64ToInt(tfInt types.Int64) int {
	return int(tfInt.ValueInt64())
}

func FromIntToTfInt64(val int) types.Int64 {
	val64 := int64(val)
	return types.Int64Value(val64)
}

func FromIntPtrToTfInt64(val *int) types.Int64 {
	if val == nil {
		return types.Int64Null()
	}

	return FromIntToTfInt64(*val)
}

func PointerOf[T any](v T) *T {
	return &v
}

func IsTfValueNull(val attr.Value) bool {
	return val.IsNull() || val.IsUnknown()
}

func FromStringListToTfStringList(ctx context.Context, arr []string) (types.List, diag.Diagnostics) {
	if arr == nil {
		arr = make([]string, 0)
	}
	return types.ListValueFrom(ctx, types.StringType, arr)
}

func FromStringListPointerToTfStringList(ctx context.Context, arr *[]string) (types.List, diag.Diagnostics) {
	if arr == nil {
		return types.ListValueFrom(ctx, types.StringType, []string{})
	}
	return types.ListValueFrom(ctx, types.StringType, *arr)
}

func FromIntListPointerToTfInt64List(ctx context.Context, arr *[]int) (types.List, diag.Diagnostics) {
	if arr == nil {
		return types.ListValueFrom(ctx, types.Int64Type, []int{})
	}
	return types.ListValueFrom(ctx, types.Int64Type, *arr)
}

func TFInt64ListToIntList(ctx context.Context, list types.List) []int {
	return TfListToGenericList(func(a types.Int64) int {
		return int(a.ValueInt64())
	}, ctx, list)
}

func TFInt64ListToIntListPointer(ctx context.Context, list types.List) *[]int {
	arr := TfListToGenericList(func(a types.Int64) int {
		return int(a.ValueInt64())
	}, ctx, list)

	return &arr
}

func TfListToGenericList[A, B any](fun func(A) B, ctx context.Context, list types.List) []B {
	if len(list.Elements()) == 0 {
		return nil
	}

	tfList := make([]A, 0, len(list.Elements()))
	res := make([]B, 0, len(list.Elements()))

	list.ElementsAs(ctx, &tfList, false)
	for _, e := range tfList {
		res = append(res, fun(e))
	}

	return res
}

func TfStringListToStringList(ctx context.Context, list types.List) []string {
	return TfListToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, list)
}

func TfStringListToStringPtrList(ctx context.Context, list types.List) *[]string {
	slice := TfListToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, list)
	return &slice
}

func TfStringListToTimeList(ctx context.Context, list types.List, format string) []time.Time {
	slice := TfListToGenericList(func(a types.String) time.Time {
		t, err := time.Parse(format, a.ValueString())
		if err != nil {
			return time.Time{}
		}
		return t
	}, ctx, list)
	return slice
}

func GenericListToTfListValue[A ITFValue, B any](ctx context.Context, tfListInnerObjType A, fn func(ctx context.Context, from B) (A, diag.Diagnostics), from []B) (basetypes.ListValue, diag.Diagnostics) {
	if len(from) == 0 {
		return types.ListNull(tfListInnerObjType.Type(ctx)), diag.Diagnostics{}
	}

	to := make([]A, 0, len(from))
	for i := range from {
		res, diags := fn(ctx, from[i])
		if diags.HasError() {
			return types.List{}, diags
		}

		to = append(to, res)
	}

	return types.ListValueFrom(ctx, to[0].Type(ctx), to)
}

func StringListToTfListValue(ctx context.Context, from []string) (types.List, diag.Diagnostics) {
	return GenericListToTfListValue(
		ctx,
		basetypes.StringValue{},
		func(_ context.Context, from string) (types.String, diag.Diagnostics) {
			return types.StringValue(from), nil
		},
		from,
	)
}

func FromTfStringValueToTfOrNull(element basetypes.StringValue) basetypes.StringValue {
	if element.IsNull() || element.IsUnknown() {
		return types.StringNull()
	}

	return element
}

func FromTfBoolValueToTfOrNull(element basetypes.BoolValue) basetypes.BoolValue {
	if element.IsNull() || element.IsUnknown() {
		return types.BoolNull()
	}

	return element
}

func checkRetryCondition(res TfRequestResp, err error, stopRetryCodes []int, retryCodes []int) *retry.RetryError {
	if err != nil {
		return retry.NonRetryableError(err)
	}

	if slices.Contains(stopRetryCodes, res.StatusCode()) {
		return nil
	} else if slices.Contains(retryCodes, res.StatusCode()) {
		time.Sleep(TfRequestRetryDelay) // Delay not handled in RetryContext. Might find a better solution later
		return retry.RetryableError(fmt.Errorf("got status code %v. Must retry request", res.StatusCode()))
	} else {
		return retry.NonRetryableError(fmt.Errorf("got %d status code that is not in stop status codes (%v)"+
			" or retry status codes (%v)", res.StatusCode(), stopRetryCodes, retryCodes))
	}
}

func RetryDeleteUntilResourceAvailable[R TfRequestResp](
	ctx context.Context,
	spaceID iaas.SpaceId,
	id string,
	fun func(context.Context, iaas.SpaceId, string, ...iaas.RequestEditorFn) (R, error),
) error {
	return retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		res, err := fun(ctx, spaceID, id)

		return checkRetryCondition(res, err, []int{http.StatusNoContent}, []int{http.StatusConflict, http.StatusFailedDependency})
	})
}

func RetryCreateUntilResourceAvailable[R TfRequestResp](
	ctx context.Context,
	spaceID iaas.SpaceId,
	body iaas.CreateVpnConnectionJSONRequestBody,
	fun func(context.Context, iaas.SpaceId, iaas.CreateVpnConnectionJSONRequestBody, ...iaas.RequestEditorFn) (R, error),
) (R, error) {
	var res R
	retryError := retry.RetryContext(ctx, TfRequestRetryTimeout, func() *retry.RetryError {
		var err error
		res, err = fun(ctx, spaceID, body)

		return checkRetryCondition(res, err, []int{http.StatusCreated}, []int{http.StatusConflict, http.StatusFailedDependency})
	})

	return res, retryError
}
