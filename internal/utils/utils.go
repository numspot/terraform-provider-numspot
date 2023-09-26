package utils

import "github.com/hashicorp/terraform-plugin-framework/types"

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
