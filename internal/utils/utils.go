package utils

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
	return FromIntToTfInt64(*val)
}

func PointerOf[T any](v T) *T {
	return &v
}
func IsTfValueNull(val attr.Value) bool {
	return val.IsNull() || val.IsUnknown()
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

type ITFValue interface {
	Type(ctx context.Context) attr.Type
}

func GenericListToTfListValue[A ITFValue, B any](ctx context.Context, fn func(ctx context.Context, from B) (A, diag.Diagnostics), from []B) (basetypes.ListValue, diag.Diagnostics) {
	diagnostics := diag.Diagnostics{}
	if len(from) == 0 {
		return types.List{}, diagnostics
	}

	to := make([]A, 0, len(from))
	for i := range from {
		res, diags := fn(ctx, from[i])
		if diags.HasError() {
			diagnostics.Append(diags...)
			return types.List{}, diagnostics
		}

		to = append(to, res)
	}

	return types.ListValueFrom(ctx, to[0].Type(ctx), to)
}
