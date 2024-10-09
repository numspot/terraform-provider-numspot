package utils

import (
	"context"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	if tfInt.IsUnknown() || tfInt.IsNull() {
		return nil
	}

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

func FromUUIDListToTfStringSet(ctx context.Context, arr []uuid.UUID, diags *diag.Diagnostics) types.Set {
	if arr == nil {
		arr = make([]uuid.UUID, 0)
	}

	nArr := make([]string, 0, len(arr))
	for _, id := range arr {
		nArr = append(nArr, id.String())
	}

	setValue, diagnostics := types.SetValueFrom(ctx, types.StringType, nArr)

	diags.Append(diagnostics...)
	return setValue
}

func FromStringListToTfStringList(ctx context.Context, arr []string, diags *diag.Diagnostics) types.List {
	if arr == nil {
		arr = make([]string, 0)
	}
	listValue, diagnostics := types.ListValueFrom(ctx, types.StringType, arr)
	diags.Append(diagnostics...)

	return listValue
}

func FromStringListPointerToTfStringSet(ctx context.Context, arr *[]string, diags *diag.Diagnostics) types.Set {
	if arr == nil {
		setValue, diagnostics := types.SetValueFrom(ctx, types.StringType, []string{})
		diags.Append(diagnostics...)
		return setValue
	}
	setValue, diagnostics := types.SetValueFrom(ctx, types.StringType, *arr)
	diags.Append(diagnostics...)
	return setValue
}

func FromStringListPointerToTfStringList(ctx context.Context, arr *[]string, diags *diag.Diagnostics) types.List {
	if arr == nil {
		listValue, diagnostics := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(diagnostics...)
		return listValue
	}
	listValue, diagnostics := types.ListValueFrom(ctx, types.StringType, *arr)
	diags.Append(diagnostics...)
	return listValue
}

func FromIntListPointerToTfInt64List(ctx context.Context, arr *[]int, diags *diag.Diagnostics) types.List {
	if arr == nil {
		listValue, diagnostics := types.ListValueFrom(ctx, types.Int64Type, []int{})
		diags.Append(diagnostics...)
		return listValue
	}
	listValue, diagnostics := types.ListValueFrom(ctx, types.Int64Type, *arr)
	diags.Append(diagnostics...)
	return listValue
}

func TFInt64ListToIntList(ctx context.Context, list types.List, diags *diag.Diagnostics) []int {
	return TfListToGenericList(func(a types.Int64) int {
		return int(a.ValueInt64())
	}, ctx, list, diags)
}

func TFInt64ListToIntListPointer(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]int {
	if list.IsNull() {
		return nil
	}
	arr := TfListToGenericList(func(a types.Int64) int {
		return int(a.ValueInt64())
	}, ctx, list, diags)

	return &arr
}

func TfSetToGenericSet[A, B any](fun func(A) B, ctx context.Context, set types.Set, diags *diag.Diagnostics) []B {
	if len(set.Elements()) == 0 {
		return nil
	}

	tfList := make([]A, 0, len(set.Elements()))
	res := make([]B, 0, len(set.Elements()))

	diags.Append(set.ElementsAs(ctx, &tfList, false)...)

	for _, e := range tfList {
		res = append(res, fun(e))
	}

	return res
}

func TfListToGenericList[A, B any](fun func(A) B, ctx context.Context, list types.List, diags *diag.Diagnostics) []B {
	if len(list.Elements()) == 0 {
		return nil
	}

	tfList := make([]A, 0, len(list.Elements()))
	res := make([]B, 0, len(list.Elements()))

	diags.Append(list.ElementsAs(ctx, &tfList, false)...)
	for _, e := range tfList {
		res = append(res, fun(e))
	}

	return res
}

func TfSetToGenericList[A, B any](fun func(A) B, ctx context.Context, set types.Set, diags *diag.Diagnostics) []B {
	if len(set.Elements()) == 0 {
		return nil
	}

	tfList := make([]A, 0, len(set.Elements()))
	res := make([]B, 0, len(set.Elements()))

	diags.Append(set.ElementsAs(ctx, &tfList, false)...)
	if diags.HasError() {
		return nil
	}
	for _, e := range tfList {
		res = append(res, fun(e))
	}

	return res
}

func TfStringListToStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	return TfListToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, list, diags)
}

func TfStringSetToStringPtrSet(ctx context.Context, set types.Set, diags *diag.Diagnostics) *[]string {
	if set.IsNull() {
		return nil
	}
	slice := TfSetToGenericSet(func(a types.String) string {
		return a.ValueString()
	}, ctx, set, diags)
	return &slice
}

func TfStringListToStringPtrList(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	slice := TfListToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, list, diags)
	return &slice
}

func TfStringListToTimeList(ctx context.Context, list types.List, format string, diags *diag.Diagnostics) []time.Time {
	if list.IsNull() {
		return nil
	}
	return TfListToGenericList(func(a types.String) time.Time {
		t, err := time.Parse(format, a.ValueString())
		if err != nil {
			return time.Time{}
		}
		return t
	}, ctx, list, diags)
}

func GenericListToTfListValue[A ITFValue, B any](ctx context.Context, tfListInnerObjType A, fn func(ctx context.Context, from B, diags *diag.Diagnostics) A, from []B, diags *diag.Diagnostics) basetypes.ListValue {
	var emptyA A

	to := make([]A, 0, len(from))
	for i := range from {
		res := fn(ctx, from[i], diags)
		if diags.HasError() {
			return types.List{}
		}

		to = append(to, res)
	}

	listValue, diagnostics := types.ListValueFrom(ctx, emptyA.Type(ctx), to)
	diags.Append(diagnostics...)
	return listValue
}

func GenericSetToTfSetValue[A ITFValue, B any](ctx context.Context, tfListInnerObjType A, fn func(ctx context.Context, from B, diags *diag.Diagnostics) A, from []B, diags *diag.Diagnostics) basetypes.SetValue {
	var emptyA A

	to := make([]A, 0, len(from))
	for i := range from {
		res := fn(ctx, from[i], diags)
		if diags.HasError() {
			return types.Set{}
		}

		to = append(to, res)
	}

	setValue, diagnostics := types.SetValueFrom(ctx, emptyA.Type(ctx), to)
	diags.Append(diagnostics...)
	return setValue
}

func StringSetToTfSetValue(ctx context.Context, from []string, diags *diag.Diagnostics) types.Set {
	return GenericSetToTfSetValue(
		ctx,
		basetypes.StringValue{},
		func(_ context.Context, from string, diags *diag.Diagnostics) types.String {
			return types.StringValue(from)
		},
		from,
		diags,
	)
}

func StringListToTfListValue(ctx context.Context, from []string, diags *diag.Diagnostics) types.List {
	return GenericListToTfListValue(
		ctx,
		basetypes.StringValue{},
		func(_ context.Context, from string, diags *diag.Diagnostics) types.String {
			return types.StringValue(from)
		},
		from,
		diags,
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

func FromTfStringListToStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	return TfListToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, list, diags)
}

func FromTfStringSetToStringList(ctx context.Context, set types.Set, diags *diag.Diagnostics) []string {
	return TfSetToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, set, diags)
}

func ParseUUID(id string, diags *diag.Diagnostics) uuid.UUID {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		diags.AddError("Failed to parse id", err.Error())
	}
	return parsedUUID
}

func FromHttpGenericListToTfList[httpType any, tfType any](
	ctx context.Context,
	http_items *[]httpType,
	httpToTfParser func(context.Context, *httpType, *diag.Diagnostics) *tfType,
	diags *diag.Diagnostics,
) []tfType {
	itemList := make([]tfType, 0, len(*http_items))

	for _, item := range *http_items {
		tf := httpToTfParser(ctx, &item, diags)
		if diags.HasError() {
			return nil
		}

		itemList = append(itemList, *tf)
	}

	return itemList
}

func Diff[A basetypes.ObjectValuable](current, desired []A) (toCreate, toDelete []A) {
	toCreate = diff(desired, current)
	toDelete = diff(current, desired)
	return
}

// Return the subset of slice1 elements that are not in slice2
func diff[A basetypes.ObjectValuable](slice1, slice2 []A) []A {
	diff := make([]A, 0)

	for _, ea := range slice1 {
		found := false
		for _, eb := range slice2 {
			if reflect.DeepEqual(ea, eb) {
				found = true
			}
		}

		if !found {
			diff = append(diff, ea)
		}
	}
	return diff
}

func DiffComparable[A comparable](current, desired []A) (toCreate, toDelete []A) {
	toCreate = diffComparable(desired, current)
	toDelete = diffComparable(current, desired)
	return
}

// Return the subset of slice1 elements that are not in slice2
func diffComparable[A comparable](slice1, slice2 []A) []A {
	diff := make([]A, 0)
	for _, ea := range slice1 {
		if !slices.Contains(slice2, ea) {
			diff = append(diff, ea)
		}
	}
	return diff
}

func EmptyTrueBoolPointer() *bool {
	value := true
	return &value
}

func EmptyStrPointer() *string {
	value := ""
	return &value
}
