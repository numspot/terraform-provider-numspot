package utils

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-numspot/internal/sdk/api"
)

type (
	// ITFValue is a Terraform framework value that can report its attribute type.
	ITFValue interface {
		Type(ctx context.Context) attr.Type
	}

	// TFType is a Terraform framework type that supports null/unknown checks.
	TFType interface {
		IsNull() bool
		IsUnknown() bool
	}
)

// WrapError wraps err with msg using %w so callers can return it directly.
// If err is nil, it returns nil.
func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// WrapErrorf wraps err with a formatted message using %w so callers can return it directly.
// If err is nil, it returns nil.
func WrapErrorf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// GetPtrValue returns pointer value.
func GetPtrValue[R any](val *R) R {
	var value R
	if val != nil {
		value = *val
	}
	return value
}

// FromTfStringToStringPtr handles from TF string to string pointer.
func FromTfStringToStringPtr(str types.String) *string {
	if str.IsUnknown() || str.IsNull() {
		return nil
	}

	return str.ValueStringPointer()
}

// FromTfStringToStringPtrExcludeEmpty handles from TF string to string pointer exclude empty.

// FromTfStringToStringPtrExcludeEmpty handles from TF string to string pointer exclude empty.
func FromTfStringToStringPtrExcludeEmpty(str types.String) *string {
	if str.IsUnknown() || str.IsNull() || str.ValueString() == "" {
		return nil
	}
	return str.ValueStringPointer()
}

// FromTfStringToAzNamePtr handles from TF string to AZ name pointer.
func FromTfStringToAzNamePtr(str types.String) *api.AvailabilityZoneName {
	if str.IsUnknown() || str.IsNull() {
		return nil
	}

	ret := api.AvailabilityZoneName(str.ValueString())
	return &ret
}

// ConvertUuidPtrToString handles convert UUID pointer to string.
func ConvertUuidPtrToString(u *api.Uuid) string {
	if u == nil {
		return ""
	}
	return u.String()
}

// FromTfStringToUuid handles from TF string to UUID.
func FromTfStringToUuid(str types.String) api.Uuid {
	parsed, _ := uuid.Parse(str.ValueString())
	return parsed
}

// StringToUuid parses a plain string ID into an api.Uuid (api.ResourceIdentifier).
// It is used by core functions that accept the generator's string IDs but call
// SDK methods typed on api.ResourceIdentifier.
func StringToUuid(s string) api.Uuid {
	parsed, _ := uuid.Parse(s)
	return parsed
}

// TimePtrToTfString formats a *time.Time as an RFC3339 string value, or returns
// a null string when the pointer is nil. Used to map SDK *time.Time fields
// (e.g. creationDate) onto Terraform types.String attributes.
func TimePtrToTfString(t *time.Time) types.String {
	if t == nil {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}

// PreserveFromState resolves a create-only input attribute when copying it from
// the plan into the post-apply state. Attributes not present in the API read
// response are copied from the plan, but unconfigured computed_optional
// attributes are UNKNOWN in the plan and must not leak into the state as
// unknown (Terraform requires all values to be known after apply).
func PreserveFromState[T attr.Value](plan T) T {
	if plan.IsUnknown() {
		var zero T
		return zero
	}
	return plan
}

// TfStringToTimePtr parses an RFC3339 string value into a *time.Time, or returns
// nil when the value is null/unknown/unparseable. Reverse of TimePtrToTfString.
func TfStringToTimePtr(str types.String) *time.Time {
	if str.IsNull() || str.IsUnknown() {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, str.ValueString())
	if err != nil {
		return nil
	}
	return &parsed
}

// TimeToTfString formats a value time.Time as an RFC3339 string, or null when
// the time is the zero value (mirrors TimePtrToTfString for non-pointer fields).
func TimeToTfString(t time.Time) types.String {
	if t.IsZero() {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}

// TfStringToTime parses an RFC3339 string into a value time.Time, returning the
// zero time when null/unknown/unparseable. Reverse of TimeToTfString.
func TfStringToTime(str types.String) time.Time {
	if t := TfStringToTimePtr(str); t != nil {
		return *t
	}
	return time.Time{}
}

// FromTfStringToUuidPtr handles from TF string to UUID pointer.
func FromTfStringToUuidPtr(str types.String) *api.Uuid {
	if str.IsUnknown() || str.IsNull() {
		return nil
	}
	parsed, err := uuid.Parse(str.ValueString())
	if err != nil {
		return nil
	}
	return &parsed
}

// FromTfBoolToBoolPtr handles from TF bool to bool pointer.
func FromTfBoolToBoolPtr(bl types.Bool) *bool {
	if bl.IsUnknown() || bl.IsNull() {
		return nil
	}

	return bl.ValueBoolPointer()
}

// FromTfInt64ToIntPtr handles from TF int64 to int pointer.
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

// FromTfInt64ToInt handles from TF int64 to int.
func FromTfInt64ToInt(tfInt types.Int64) int {
	return int(tfInt.ValueInt64())
}

// FromIntToTfInt64 handles from int to TF int64.
func FromIntToTfInt64(val int) types.Int64 {
	val64 := int64(val)
	return types.Int64Value(val64)
}

// FromIntPtrToTfInt64 handles from int pointer to TF int64.
func FromIntPtrToTfInt64(val *int) types.Int64 {
	if val == nil {
		return types.Int64Null()
	}

	return FromIntToTfInt64(*val)
}

// FromInt32PtrToTfInt64 handles from int32 pointer to TF int64.
func FromInt32PtrToTfInt64(val *int32) types.Int64 {
	if val == nil {
		return types.Int64Null()
	}

	return types.Int64Value(int64(*val))
}

// FromTfInt64ToInt32Ptr handles from TF int64 to int32 pointer.
func FromTfInt64ToInt32Ptr(val types.Int64) *int32 {
	if val.IsNull() || val.IsUnknown() {
		return nil
	}

	v := int32(val.ValueInt64())
	return &v
}

// PointerOf handles pointer of.
func PointerOf[T any](v T) *T {
	return &v
}

// IsTfValueNull handles is TF value null.
func IsTfValueNull(val attr.Value) bool {
	return val.IsNull() || val.IsUnknown()
}

// FromStringListPointerToTfStringSet handles from string list pointer to TF string set.
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

// FromStringListPointerToTfStringList handles from string list pointer to TF string list.
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

// FromStringListToTfStringList handles from string list to TF string list.
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

// FromIntListPointerToTfInt64List handles from int list pointer to TF int64 list.
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

// FromIntListToTfInt64List handles from int list to TF int64 list.
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

// TFInt64ListToIntList handles TF int64 list to int list.
func TFInt64ListToIntList(ctx context.Context, list types.List, diags *diag.Diagnostics) []int {
	return TfListToGenericList(func(a types.Int64) int {
		return int(a.ValueInt64())
	}, ctx, list, diags)
}

// TFInt64ListToIntListPointer handles TF int64 list to int list pointer.
func TFInt64ListToIntListPointer(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]int {
	if list.IsNull() {
		return nil
	}
	arr := TfListToGenericList(func(a types.Int64) int {
		return int(a.ValueInt64())
	}, ctx, list, diags)

	return &arr
}

// TfSetToGenericSet handles TF set to generic set.
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

// TfListToGenericList handles TF list to generic list.
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

// TfSetToGenericList handles TF set to generic list.
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

// TfStringListToStringList handles TF string list to string list.
func TfStringListToStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	return TfListToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, list, diags)
}

// TfStringListToStringPtrList handles TF string list to string pointer list.
func TfStringListToStringPtrList(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	slice := TfListToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, list, diags)
	return &slice
}

// TfStringListToTimeList handles TF string list to time list.
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

// GenericListToTfListValue handles generic list to TF list value.
func GenericListToTfListValue[A ITFValue, B any](
	ctx context.Context,
	fn func(ctx context.Context, from *B, diags *diag.Diagnostics) A,
	from []B,
	diags *diag.Diagnostics,
) basetypes.ListValue {
	emptyA := new(A)

	to := make([]A, 0, len(from))
	for i := range from {
		res := fn(ctx, &from[i], diags)
		if diags.HasError() {
			return types.List{}
		}

		to = append(to, res)
	}

	listValue, diagnostics := types.ListValueFrom(ctx, (*emptyA).Type(ctx), to)
	diags.Append(diagnostics...)
	return listValue
}

// GenericSetToTfSetValue handles generic set to TF set value.
func GenericSetToTfSetValue[A ITFValue, B any](
	ctx context.Context,
	fn func(ctx context.Context, from *B, diags *diag.Diagnostics) A,
	from []B,
	diags *diag.Diagnostics,
) basetypes.SetValue {
	emptyA := new(A)

	to := make([]A, 0, len(from))
	for i := range from {
		res := fn(ctx, &from[i], diags)
		if diags.HasError() {
			return types.Set{}
		}

		to = append(to, res)
	}

	setValue, diagnostics := types.SetValueFrom(ctx, (*emptyA).Type(ctx), to)
	diags.Append(diagnostics...)
	return setValue
}

// StringListToTfListValue handles string list to TF list value.
func StringListToTfListValue(ctx context.Context, from []string, diags *diag.Diagnostics) types.List {
	return GenericListToTfListValue(
		ctx,
		func(_ context.Context, from *string, diags *diag.Diagnostics) types.String {
			return types.StringValue(*from)
		},
		from,
		diags,
	)
}

// FromTfStringValueToTfOrNull handles from TF string value to TF or null.
func FromTfStringValueToTfOrNull(element basetypes.StringValue) basetypes.StringValue {
	if element.IsNull() || element.IsUnknown() {
		return types.StringNull()
	}

	return element
}

// FromTfBoolValueToTfOrNull handles from TF bool value to TF or null.
func FromTfBoolValueToTfOrNull(element basetypes.BoolValue) basetypes.BoolValue {
	if element.IsNull() || element.IsUnknown() {
		return types.BoolNull()
	}

	return element
}

// FromTfStringSetToStringList handles from TF string set to string list.
func FromTfStringSetToStringList(ctx context.Context, set types.Set, diags *diag.Diagnostics) []string {
	return TfSetToGenericList(func(a types.String) string {
		return a.ValueString()
	}, ctx, set, diags)
}

// FromTfStringSetToStringListPtr converts a TF string set into *[]string,
// returning nil when the set is null or unknown (so unset lists are omitted from
// the request body rather than serialized as an empty array).
func FromTfStringSetToStringListPtr(ctx context.Context, set types.Set, diags *diag.Diagnostics) *[]string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	l := FromTfStringSetToStringList(ctx, set, diags)
	return &l
}

// FromHttpGenericListToTfList handles from HTTP generic list to TF list.
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

// Diff handles diff.
func Diff[A basetypes.ObjectValuable](current, desired []A) (toCreate, toDelete []A) {
	toCreate = diff(desired, current)
	toDelete = diff(current, desired)
	return toCreate, toDelete
}

// ConvertTfListToArrayOfString handles convert TF list to array of string.
func ConvertTfListToArrayOfString(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	slice := tfListToArrayOfString(ctx, list, diags)
	return slice
}

// ConvertTfListToArrayOfAzName handles convert TF list to array of AZ name.
func ConvertTfListToArrayOfAzName(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]api.AvailabilityZoneName {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	slice := tfListToArrayOfAzName(ctx, list, diags)
	return slice
}

// ConvertTfListToArrayOfStringEnum converts a Terraform string list into a
// pointer to a slice of SDK string-backed enum types (e.g. ReadVmsParamsTypes).
// It returns nil when the list is null or unknown.
func ConvertTfListToArrayOfStringEnum[T ~string](ctx context.Context, list types.List, diags *diag.Diagnostics) *[]T {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	res := make([]T, 0, len(elements))

	if len(elements) == 0 {
		res = []T{}
	} else {
		tempList := make([]types.String, 0, len(elements))
		diags.Append(list.ElementsAs(ctx, &tempList, false)...)

		for _, e := range tempList {
			res = append(res, T(e.ValueString()))
		}
	}

	return &res
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

// ConvertTfListToArrayOfInt handles convert TF list to array of int.
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

// ConvertTfListToArrayOfTime handles convert TF list to array of time.
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

// DiffComparable handles diff comparable.
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

// EmptyTrueBoolPointer handles empty true bool pointer.
func EmptyTrueBoolPointer() *bool {
	value := true
	return &value
}

// ConvertStringPtrToString Convert string pointer to string
func ConvertStringPtrToString(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}

// ConvertIntPtrToInt64 Convert int pointer to int
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
func SerializeDatasourceItemsWithDiags[In any, Out any](
	ctx context.Context,
	http []In,
	diags *diag.Diagnostics,
	mapping func(context.Context, *In, *diag.Diagnostics) (Out, diag.Diagnostics),
) []Out {
	var serializeDiags diag.Diagnostics
	itemsValue := make([]Out, len(http))

	if len(http) > 0 {
		ll := len(http)

		for i := range ll {
			itemsValue[i], serializeDiags = mapping(ctx, &http[i], diags)
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}
	}

	return itemsValue
}

// SerializeDatasourceItems Serialize Any Datasource Items With mapping fonction
func SerializeDatasourceItems[In any, Out any](
	ctx context.Context,
	http []In,
	mapping func(context.Context, *In) (Out, diag.Diagnostics),
) ([]Out, diag.Diagnostics) {
	serializeDiags := diag.Diagnostics{}
	itemsValue := make([]Out, len(http))

	if len(http) > 0 {
		ll := len(http)

		for i := range ll {
			var tmpDiags diag.Diagnostics
			itemsValue[i], tmpDiags = mapping(ctx, &http[i])
			if tmpDiags.HasError() {
				serializeDiags.Append(tmpDiags...)
				continue
			}
		}
	}

	return itemsValue, serializeDiags
}

// ConvertAzNamePtrToString Convert string pointer to AzName
func ConvertAzNamePtrToString(value *api.AvailabilityZoneName) string {
	if value != nil {
		return string(*value)
	}
	return ""
}
