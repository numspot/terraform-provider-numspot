package utils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func FromTfInt64ToIntPtr(tfInt types.Int64) *int {
	val := int(tfInt.ValueInt64())
	return &val
}

func FromTfInt64ToInt(tfInt types.Int64) int {
	return int(tfInt.ValueInt64())
}

func FromIntToTfInt64(val int) types.Int64 {
	val64 := int64(val)
	return types.Int64Value(val64)
}

func FromIntPtrToTfInt64(val *int) types.Int64 {
	return FromIntToTfInt64(*val)
}

func IsTfValueNull(val attr.Value) bool {
	if val.IsNull() || val.IsUnknown() {
		return true
	}
	return false
}

func FromStringListToTfStringList(ctx context.Context, arr []string) (types.List, diag.Diagnostics) {
	return types.ListValueFrom(ctx, types.StringType, arr)
}

func TfListToGenericList[A, B any](fun func(A) B, ctx context.Context, list types.List) []B {
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
