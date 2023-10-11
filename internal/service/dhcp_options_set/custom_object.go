package dhcp_options_set

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.ObjectTypable = CustomObjectType{}

type CustomObjectType struct {
	basetypes.ObjectType
}

func (t CustomObjectType) Equal(o attr.Type) bool {
	other, ok := o.(CustomObjectType)
	if !ok {
		return false
	}
	return t.ObjectType.Equal(other.ObjectType)
}

func (t CustomObjectType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ObjectType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	objValue, ok := attrValue.(basetypes.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	// Convert to your Custom Object Value (defined in the next section)
	return CustomObjectValue{
		ObjectValue: objValue,
	}, nil
}

func (t CustomObjectType) String() string {
	return "CustomObjectType"
}

func (t CustomObjectType) ValueType(ctx context.Context) attr.Value {
	return CustomObjectValue{}
}

func (t CustomObjectType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	return CustomObjectValue{
		ObjectValue: in,
	}, nil
}

var _ basetypes.ObjectValuable = CustomObjectValue{}

type CustomObjectValue struct {
	basetypes.ObjectValue
	A types.String
	B types.Int64
}

func (v CustomObjectValue) Equal(o attr.Value) bool {
	other, ok := o.(CustomObjectValue)
	if !ok {
		return false
	}
	return v.ObjectValue.Equal(other.ObjectValue)
}

func (v CustomObjectValue) Type(ctx context.Context) attr.Type {
	return CustomObjectType{}
}
