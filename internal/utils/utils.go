package utils

import (
	"context"
	"reflect"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-numspot/internal/sdk/api"
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

func FromTfStringToAzNamePtr(str types.String) *api.AvailabilityZoneName {
	if str.IsUnknown() || str.IsNull() {
		return nil
	}

	ret := api.AvailabilityZoneName(str.ValueString())
	return &ret
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

func FromStringListToTfStringList(ctx context.Context, arr []string, diags *diag.Diagnostics) types.List {
	if arr == nil {
		listValue, diagnostics := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(diagnostics...)
		return listValue
	}
	listValue, diagnostics := types.ListValueFrom(ctx, types.StringType, arr)
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

func FromIntListToTfInt64List(ctx context.Context, arr []int, diags *diag.Diagnostics) types.List {
	if arr == nil {
		listValue, diagnostics := types.ListValueFrom(ctx, types.Int64Type, []int{})
		diags.Append(diagnostics...)
		return listValue
	}
	listValue, diagnostics := types.ListValueFrom(ctx, types.Int64Type, arr)
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

func GenericListToTfListValue[A ITFValue, B any](ctx context.Context, fn func(ctx context.Context, from B, diags *diag.Diagnostics) A, from []B, diags *diag.Diagnostics) basetypes.ListValue {
	emptyA := new(A)

	to := make([]A, 0, len(from))
	for i := range from {
		res := fn(ctx, from[i], diags)
		if diags.HasError() {
			return types.List{}
		}

		to = append(to, res)
	}

	listValue, diagnostics := types.ListValueFrom(ctx, (*emptyA).Type(ctx), to)
	diags.Append(diagnostics...)
	return listValue
}

func GenericSetToTfSetValue[A ITFValue, B any](ctx context.Context, fn func(ctx context.Context, from B, diags *diag.Diagnostics) A, from []B, diags *diag.Diagnostics) basetypes.SetValue {
	emptyA := new(A)

	to := make([]A, 0, len(from))
	for i := range from {
		res := fn(ctx, from[i], diags)
		if diags.HasError() {
			return types.Set{}
		}

		to = append(to, res)
	}

	setValue, diagnostics := types.SetValueFrom(ctx, (*emptyA).Type(ctx), to)
	diags.Append(diagnostics...)
	return setValue
}

func StringListToTfListValue(ctx context.Context, from []string, diags *diag.Diagnostics) types.List {
	return GenericListToTfListValue(
		ctx,
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

func FromTfStringSetToStringList(ctx context.Context, set types.Set, diags *diag.Diagnostics) []string {
	return TfSetToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, set, diags)
}

func FromHttpGenericListToTfList[httpType any, tfType any](
	ctx context.Context,
	httpItems *[]httpType,
	httpToTfParser func(context.Context, *httpType, *diag.Diagnostics) *tfType,
	diags *diag.Diagnostics,
) []tfType {
	itemList := make([]tfType, 0, len(*httpItems))

	for _, item := range *httpItems {
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
	return toCreate, toDelete
}

func ConvertTfListToArrayOfString(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	slice := tfListToArrayOfString(ctx, list, diags)
	return slice
}

func ConvertTfListToArrayOfAzName(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]api.AvailabilityZoneName {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	slice := tfListToArrayOfAzName(ctx, list, diags)
	return slice
}

func tfListToArrayOfString(ctx context.Context, tfList types.List, diags *diag.Diagnostics) *[]string {
	elements := tfList.Elements()
	res := make([]string, 0, len(elements))

	if len(elements) == 0 {
		res = []string{}
	} else {
		tempList := make([]types.String, 0, len(elements))
		diags.Append(tfList.ElementsAs(ctx, &tempList, false)...)

		for _, e := range tempList {
			res = append(res, e.ValueString())
		}
	}

	return &res
}

func tfListToArrayOfAzName(ctx context.Context, tfList types.List, diags *diag.Diagnostics) *[]api.AvailabilityZoneName {
	elements := tfList.Elements()
	res := make([]api.AvailabilityZoneName, 0, len(elements))

	if len(elements) == 0 {
		res = []api.AvailabilityZoneName{}
	} else {
		tempList := make([]types.String, 0, len(elements))
		diags.Append(tfList.ElementsAs(ctx, &tempList, false)...)

		for _, e := range tempList {
			res = append(res, api.AvailabilityZoneName(e.ValueString()))
		}
	}

	return &res
}

func ConvertTfListToArrayOfInt(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]int {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	slice := tfListToArrayOfInt(ctx, list, diags)
	return slice
}

func tfListToArrayOfInt(ctx context.Context, tfList types.List, diags *diag.Diagnostics) *[]int {
	elements := tfList.Elements()
	res := make([]int, 0, len(elements))

	if len(elements) == 0 {
		res = []int{}
	} else {
		tempList := make([]types.Int64, 0, len(elements))
		diags.Append(tfList.ElementsAs(ctx, &tempList, false)...)

		for _, e := range tempList {
			res = append(res, int(e.ValueInt64()))
		}
	}

	return &res
}

func ConvertTfListToArrayOfTime(ctx context.Context, list types.List, format string, diags *diag.Diagnostics) *[]time.Time {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	slice := tfListToArrayOfTime(ctx, list, format, diags)
	return slice
}

func tfListToArrayOfTime(ctx context.Context, tfList types.List, format string, diags *diag.Diagnostics) *[]time.Time {
	elements := tfList.Elements()
	res := make([]time.Time, 0, len(elements))

	if len(elements) == 0 {
		res = []time.Time{}
	} else {
		tempList := make([]types.String, 0, len(elements))
		diags.Append(tfList.ElementsAs(ctx, &tempList, false)...)

		for _, e := range tempList {
			t, err := time.Parse(format, e.ValueString())
			if err != nil {
				return &[]time.Time{}
			}
			res = append(res, t)
		}
	}

	return &res
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
	return toCreate, toDelete
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

// ConvertStringPtrToString Convert string ptr to string
func ConvertStringPtrToString(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}

// ConvertIntPtrToInt64 Convert int ptr to int
func ConvertIntPtrToInt64(value *int) int64 {
	if value != nil {
		return int64(*value)
	}
	return 0
}

// CreateListValueItems Convert list with any items type of datasource
func CreateListValueItems[T ITFValue](ctx context.Context, items []T, diags *diag.Diagnostics) basetypes.ListValue {
	var serializeDiags diag.Diagnostics
	emptyT := new(T)

	list := types.ListNull((*emptyT).Type(ctx))

	if len(items) > 0 {
		list, serializeDiags = types.ListValueFrom(ctx, (*emptyT).Type(ctx), items)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	return list
}

// SerializeDatasourceItemsWithDiags Serialize Any Datasource Items With mapping fonction
func SerializeDatasourceItemsWithDiags[In any, Out any](ctx context.Context, http []In, diags *diag.Diagnostics, mapping func(context.Context, In, *diag.Diagnostics) (Out, diag.Diagnostics)) []Out {
	var serializeDiags diag.Diagnostics
	itemsValue := make([]Out, len(http))

	if len(http) > 0 {
		ll := len(http)

		for i := 0; ll > i; i++ {
			itemsValue[i], serializeDiags = mapping(ctx, http[i], diags)
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}
	}

	return itemsValue
}

// SerializeDatasourceItems Serialize Any Datasource Items With mapping fonction
func SerializeDatasourceItems[In any, Out any](ctx context.Context, http []In, mapping func(context.Context, In) (Out, diag.Diagnostics)) ([]Out, diag.Diagnostics) {
	serializeDiags := diag.Diagnostics{}
	itemsValue := make([]Out, len(http))

	if len(http) > 0 {
		ll := len(http)

		for i := 0; ll > i; i++ {
			var tmpDiags diag.Diagnostics
			itemsValue[i], tmpDiags = mapping(ctx, http[i])
			if tmpDiags.HasError() {
				serializeDiags.Append(tmpDiags...)
				continue
			}
		}
	}

	return itemsValue, serializeDiags
}

// ConvertAzNamePtrToString Convert string ptr to AzName
func ConvertAzNamePtrToString(value *api.AvailabilityZoneName) string {
	if value != nil {
		return string(*value)
	}
	return ""
}
