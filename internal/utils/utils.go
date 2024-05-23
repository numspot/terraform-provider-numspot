package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type EntityType string

var (
	EntityTypePermission     EntityType = "permission"
	EntityTypeRole           EntityType = "role"
	EntityTypeSpace          EntityType = "space"
	EntityTypeServiceAccount EntityType = "service account"
)

type (
	ITFValue interface {
		Type(ctx context.Context) attr.Type
	}

	TFType interface {
		IsNull() bool
		IsUnknown() bool
	}
)

func GetPtrValue[R any](val *R) R {
	var value R
	if val != nil {
		value = *val
	}
	return value
}

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

func FromUUIDListToTfStringSet(ctx context.Context, arr []uuid.UUID) (types.Set, diag.Diagnostics) {
	if arr == nil {
		arr = make([]uuid.UUID, 0)
	}

	nArr := make([]string, 0, len(arr))
	for _, id := range arr {
		nArr = append(nArr, id.String())
	}

	return types.SetValueFrom(ctx, types.StringType, nArr)
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

func TfSetToGenericList[A, B any](fun func(A) B, ctx context.Context, set types.Set) []B {
	if len(set.Elements()) == 0 {
		return nil
	}

	tfList := make([]A, 0, len(set.Elements()))
	res := make([]B, 0, len(set.Elements()))

	set.ElementsAs(ctx, &tfList, false)
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

func GenericSetToTfSetValue[A ITFValue, B any](ctx context.Context, tfListInnerObjType A, fn func(ctx context.Context, from B) (A, diag.Diagnostics), from []B) (basetypes.SetValue, diag.Diagnostics) {
	if len(from) == 0 {
		return types.SetNull(tfListInnerObjType.Type(ctx)), diag.Diagnostics{}
	}

	to := make([]A, 0, len(from))
	for i := range from {
		res, diags := fn(ctx, from[i])
		if diags.HasError() {
			return types.Set{}, diags
		}

		to = append(to, res)
	}

	return types.SetValueFrom(ctx, to[0].Type(ctx), to)
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

func FromTfStringListToStringList(ctx context.Context, list types.List) []string {
	return TfListToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, list)
}

func FromTfStringSetToStringList(ctx context.Context, set types.Set) []string {
	return TfSetToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, set)
}

func ParseUUID(id string, entityType EntityType) (uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		diags.AddError(fmt.Sprintf("Failed to parse %s id", entityType), err.Error())
	}
	return parsedUUID, diags
}
