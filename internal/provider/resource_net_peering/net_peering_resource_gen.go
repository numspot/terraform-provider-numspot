// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_net_peering

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func NetPeeringResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"accepter_net": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						Computed:            true,
						Description:         "The account ID of the owner of the accepter Net.",
						MarkdownDescription: "The account ID of the owner of the accepter Net.",
					},
					"ip_range": schema.StringAttribute{
						Computed:            true,
						Description:         "The IP range for the accepter Net, in CIDR notation (for example, `10.0.0.0/16`).",
						MarkdownDescription: "The IP range for the accepter Net, in CIDR notation (for example, `10.0.0.0/16`).",
					},
					"net_id": schema.StringAttribute{
						Computed:            true,
						Description:         "The ID of the accepter Net.",
						MarkdownDescription: "The ID of the accepter Net.",
					},
				},
				CustomType: AccepterNetType{
					ObjectType: types.ObjectType{
						AttrTypes: AccepterNetValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Information about the accepter Net.",
				MarkdownDescription: "Information about the accepter Net.",
			},
			"accepter_net_id": schema.StringAttribute{
				Required:            true,
				Description:         "The ID of the Net you want to connect with.",
				MarkdownDescription: "The ID of the Net you want to connect with.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the Net peering.",
				MarkdownDescription: "The ID of the Net peering.",
			},
			"source_net": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						Computed:            true,
						Description:         "The account ID of the owner of the source Net.",
						MarkdownDescription: "The account ID of the owner of the source Net.",
					},
					"ip_range": schema.StringAttribute{
						Computed:            true,
						Description:         "The IP range for the source Net, in CIDR notation (for example, `10.0.0.0/16`).",
						MarkdownDescription: "The IP range for the source Net, in CIDR notation (for example, `10.0.0.0/16`).",
					},
					"net_id": schema.StringAttribute{
						Computed:            true,
						Description:         "The ID of the source Net.",
						MarkdownDescription: "The ID of the source Net.",
					},
				},
				CustomType: SourceNetType{
					ObjectType: types.ObjectType{
						AttrTypes: SourceNetValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Information about the source Net.",
				MarkdownDescription: "Information about the source Net.",
			},
			"source_net_id": schema.StringAttribute{
				Required:            true,
				Description:         "The ID of the Net you send the peering request from.",
				MarkdownDescription: "The ID of the Net you send the peering request from.",
			},
			"state": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"message": schema.StringAttribute{
						Computed:            true,
						Description:         "Additional information about the state of the Net peering.",
						MarkdownDescription: "Additional information about the state of the Net peering.",
					},
					"name": schema.StringAttribute{
						Computed:            true,
						Description:         "The state of the Net peering (`pending-acceptance` \\| `active` \\| `rejected` \\| `failed` \\| `expired` \\| `deleted`).",
						MarkdownDescription: "The state of the Net peering (`pending-acceptance` \\| `active` \\| `rejected` \\| `failed` \\| `expired` \\| `deleted`).",
					},
				},
				CustomType: StateType{
					ObjectType: types.ObjectType{
						AttrTypes: StateValue{}.AttributeTypes(ctx),
					},
				},
				Computed:            true,
				Description:         "Information about the state of the Net peering.",
				MarkdownDescription: "Information about the state of the Net peering.",
			},
		},
	}
}

type NetPeeringModel struct {
	AccepterNet   AccepterNetValue `tfsdk:"accepter_net"`
	AccepterNetId types.String     `tfsdk:"accepter_net_id"`
	Id            types.String     `tfsdk:"id"`
	SourceNet     SourceNetValue   `tfsdk:"source_net"`
	SourceNetId   types.String     `tfsdk:"source_net_id"`
	State         StateValue       `tfsdk:"state"`
}

var _ basetypes.ObjectTypable = AccepterNetType{}

type AccepterNetType struct {
	basetypes.ObjectType
}

func (t AccepterNetType) Equal(o attr.Type) bool {
	other, ok := o.(AccepterNetType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t AccepterNetType) String() string {
	return "AccepterNetType"
}

func (t AccepterNetType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	accountIdAttribute, ok := attributes["account_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`account_id is missing from object`)

		return nil, diags
	}

	accountIdVal, ok := accountIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`account_id expected to be basetypes.StringValue, was: %T`, accountIdAttribute))
	}

	ipRangeAttribute, ok := attributes["ip_range"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`ip_range is missing from object`)

		return nil, diags
	}

	ipRangeVal, ok := ipRangeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`ip_range expected to be basetypes.StringValue, was: %T`, ipRangeAttribute))
	}

	netIdAttribute, ok := attributes["net_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`net_id is missing from object`)

		return nil, diags
	}

	netIdVal, ok := netIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`net_id expected to be basetypes.StringValue, was: %T`, netIdAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return AccepterNetValue{
		AccountId: accountIdVal,
		IpRange:   ipRangeVal,
		NetId:     netIdVal,
		state:     attr.ValueStateKnown,
	}, diags
}

func NewAccepterNetValueNull() AccepterNetValue {
	return AccepterNetValue{
		state: attr.ValueStateNull,
	}
}

func NewAccepterNetValueUnknown() AccepterNetValue {
	return AccepterNetValue{
		state: attr.ValueStateUnknown,
	}
}

func NewAccepterNetValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (AccepterNetValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing AccepterNetValue Attribute Value",
				"While creating a AccepterNetValue value, a missing attribute value was detected. "+
					"A AccepterNetValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AccepterNetValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid AccepterNetValue Attribute Type",
				"While creating a AccepterNetValue value, an invalid attribute value was detected. "+
					"A AccepterNetValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AccepterNetValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("AccepterNetValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra AccepterNetValue Attribute Value",
				"While creating a AccepterNetValue value, an extra attribute value was detected. "+
					"A AccepterNetValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra AccepterNetValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewAccepterNetValueUnknown(), diags
	}

	accountIdAttribute, ok := attributes["account_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`account_id is missing from object`)

		return NewAccepterNetValueUnknown(), diags
	}

	accountIdVal, ok := accountIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`account_id expected to be basetypes.StringValue, was: %T`, accountIdAttribute))
	}

	ipRangeAttribute, ok := attributes["ip_range"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`ip_range is missing from object`)

		return NewAccepterNetValueUnknown(), diags
	}

	ipRangeVal, ok := ipRangeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`ip_range expected to be basetypes.StringValue, was: %T`, ipRangeAttribute))
	}

	netIdAttribute, ok := attributes["net_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`net_id is missing from object`)

		return NewAccepterNetValueUnknown(), diags
	}

	netIdVal, ok := netIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`net_id expected to be basetypes.StringValue, was: %T`, netIdAttribute))
	}

	if diags.HasError() {
		return NewAccepterNetValueUnknown(), diags
	}

	return AccepterNetValue{
		AccountId: accountIdVal,
		IpRange:   ipRangeVal,
		NetId:     netIdVal,
		state:     attr.ValueStateKnown,
	}, diags
}

func NewAccepterNetValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) AccepterNetValue {
	object, diags := NewAccepterNetValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewAccepterNetValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t AccepterNetType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewAccepterNetValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewAccepterNetValueUnknown(), nil
	}

	if in.IsNull() {
		return NewAccepterNetValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewAccepterNetValueMust(AccepterNetValue{}.AttributeTypes(ctx), attributes), nil
}

func (t AccepterNetType) ValueType(ctx context.Context) attr.Value {
	return AccepterNetValue{}
}

var _ basetypes.ObjectValuable = AccepterNetValue{}

type AccepterNetValue struct {
	AccountId basetypes.StringValue `tfsdk:"account_id"`
	IpRange   basetypes.StringValue `tfsdk:"ip_range"`
	NetId     basetypes.StringValue `tfsdk:"net_id"`
	state     attr.ValueState
}

func (v AccepterNetValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)

	var val tftypes.Value
	var err error

	attrTypes["account_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["ip_range"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["net_id"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 3)

		val, err = v.AccountId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["account_id"] = val

		val, err = v.IpRange.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["ip_range"] = val

		val, err = v.NetId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["net_id"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v AccepterNetValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v AccepterNetValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v AccepterNetValue) String() string {
	return "AccepterNetValue"
}

func (v AccepterNetValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"account_id": basetypes.StringType{},
			"ip_range":   basetypes.StringType{},
			"net_id":     basetypes.StringType{},
		},
		map[string]attr.Value{
			"account_id": v.AccountId,
			"ip_range":   v.IpRange,
			"net_id":     v.NetId,
		})

	return objVal, diags
}

func (v AccepterNetValue) Equal(o attr.Value) bool {
	other, ok := o.(AccepterNetValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.AccountId.Equal(other.AccountId) {
		return false
	}

	if !v.IpRange.Equal(other.IpRange) {
		return false
	}

	if !v.NetId.Equal(other.NetId) {
		return false
	}

	return true
}

func (v AccepterNetValue) Type(ctx context.Context) attr.Type {
	return AccepterNetType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v AccepterNetValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"account_id": basetypes.StringType{},
		"ip_range":   basetypes.StringType{},
		"net_id":     basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = SourceNetType{}

type SourceNetType struct {
	basetypes.ObjectType
}

func (t SourceNetType) Equal(o attr.Type) bool {
	other, ok := o.(SourceNetType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t SourceNetType) String() string {
	return "SourceNetType"
}

func (t SourceNetType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	accountIdAttribute, ok := attributes["account_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`account_id is missing from object`)

		return nil, diags
	}

	accountIdVal, ok := accountIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`account_id expected to be basetypes.StringValue, was: %T`, accountIdAttribute))
	}

	ipRangeAttribute, ok := attributes["ip_range"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`ip_range is missing from object`)

		return nil, diags
	}

	ipRangeVal, ok := ipRangeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`ip_range expected to be basetypes.StringValue, was: %T`, ipRangeAttribute))
	}

	netIdAttribute, ok := attributes["net_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`net_id is missing from object`)

		return nil, diags
	}

	netIdVal, ok := netIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`net_id expected to be basetypes.StringValue, was: %T`, netIdAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return SourceNetValue{
		AccountId: accountIdVal,
		IpRange:   ipRangeVal,
		NetId:     netIdVal,
		state:     attr.ValueStateKnown,
	}, diags
}

func NewSourceNetValueNull() SourceNetValue {
	return SourceNetValue{
		state: attr.ValueStateNull,
	}
}

func NewSourceNetValueUnknown() SourceNetValue {
	return SourceNetValue{
		state: attr.ValueStateUnknown,
	}
}

func NewSourceNetValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (SourceNetValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing SourceNetValue Attribute Value",
				"While creating a SourceNetValue value, a missing attribute value was detected. "+
					"A SourceNetValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("SourceNetValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid SourceNetValue Attribute Type",
				"While creating a SourceNetValue value, an invalid attribute value was detected. "+
					"A SourceNetValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("SourceNetValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("SourceNetValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra SourceNetValue Attribute Value",
				"While creating a SourceNetValue value, an extra attribute value was detected. "+
					"A SourceNetValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra SourceNetValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewSourceNetValueUnknown(), diags
	}

	accountIdAttribute, ok := attributes["account_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`account_id is missing from object`)

		return NewSourceNetValueUnknown(), diags
	}

	accountIdVal, ok := accountIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`account_id expected to be basetypes.StringValue, was: %T`, accountIdAttribute))
	}

	ipRangeAttribute, ok := attributes["ip_range"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`ip_range is missing from object`)

		return NewSourceNetValueUnknown(), diags
	}

	ipRangeVal, ok := ipRangeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`ip_range expected to be basetypes.StringValue, was: %T`, ipRangeAttribute))
	}

	netIdAttribute, ok := attributes["net_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`net_id is missing from object`)

		return NewSourceNetValueUnknown(), diags
	}

	netIdVal, ok := netIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`net_id expected to be basetypes.StringValue, was: %T`, netIdAttribute))
	}

	if diags.HasError() {
		return NewSourceNetValueUnknown(), diags
	}

	return SourceNetValue{
		AccountId: accountIdVal,
		IpRange:   ipRangeVal,
		NetId:     netIdVal,
		state:     attr.ValueStateKnown,
	}, diags
}

func NewSourceNetValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) SourceNetValue {
	object, diags := NewSourceNetValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewSourceNetValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t SourceNetType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewSourceNetValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewSourceNetValueUnknown(), nil
	}

	if in.IsNull() {
		return NewSourceNetValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewSourceNetValueMust(SourceNetValue{}.AttributeTypes(ctx), attributes), nil
}

func (t SourceNetType) ValueType(ctx context.Context) attr.Value {
	return SourceNetValue{}
}

var _ basetypes.ObjectValuable = SourceNetValue{}

type SourceNetValue struct {
	AccountId basetypes.StringValue `tfsdk:"account_id"`
	IpRange   basetypes.StringValue `tfsdk:"ip_range"`
	NetId     basetypes.StringValue `tfsdk:"net_id"`
	state     attr.ValueState
}

func (v SourceNetValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)

	var val tftypes.Value
	var err error

	attrTypes["account_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["ip_range"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["net_id"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 3)

		val, err = v.AccountId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["account_id"] = val

		val, err = v.IpRange.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["ip_range"] = val

		val, err = v.NetId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["net_id"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v SourceNetValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v SourceNetValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v SourceNetValue) String() string {
	return "SourceNetValue"
}

func (v SourceNetValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"account_id": basetypes.StringType{},
			"ip_range":   basetypes.StringType{},
			"net_id":     basetypes.StringType{},
		},
		map[string]attr.Value{
			"account_id": v.AccountId,
			"ip_range":   v.IpRange,
			"net_id":     v.NetId,
		})

	return objVal, diags
}

func (v SourceNetValue) Equal(o attr.Value) bool {
	other, ok := o.(SourceNetValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.AccountId.Equal(other.AccountId) {
		return false
	}

	if !v.IpRange.Equal(other.IpRange) {
		return false
	}

	if !v.NetId.Equal(other.NetId) {
		return false
	}

	return true
}

func (v SourceNetValue) Type(ctx context.Context) attr.Type {
	return SourceNetType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v SourceNetValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"account_id": basetypes.StringType{},
		"ip_range":   basetypes.StringType{},
		"net_id":     basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = StateType{}

type StateType struct {
	basetypes.ObjectType
}

func (t StateType) Equal(o attr.Type) bool {
	other, ok := o.(StateType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t StateType) String() string {
	return "StateType"
}

func (t StateType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	messageAttribute, ok := attributes["message"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`message is missing from object`)

		return nil, diags
	}

	messageVal, ok := messageAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`message expected to be basetypes.StringValue, was: %T`, messageAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return nil, diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return StateValue{
		Message: messageVal,
		Name:    nameVal,
		state:   attr.ValueStateKnown,
	}, diags
}

func NewStateValueNull() StateValue {
	return StateValue{
		state: attr.ValueStateNull,
	}
}

func NewStateValueUnknown() StateValue {
	return StateValue{
		state: attr.ValueStateUnknown,
	}
}

func NewStateValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (StateValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing StateValue Attribute Value",
				"While creating a StateValue value, a missing attribute value was detected. "+
					"A StateValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("StateValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid StateValue Attribute Type",
				"While creating a StateValue value, an invalid attribute value was detected. "+
					"A StateValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("StateValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("StateValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra StateValue Attribute Value",
				"While creating a StateValue value, an extra attribute value was detected. "+
					"A StateValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra StateValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewStateValueUnknown(), diags
	}

	messageAttribute, ok := attributes["message"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`message is missing from object`)

		return NewStateValueUnknown(), diags
	}

	messageVal, ok := messageAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`message expected to be basetypes.StringValue, was: %T`, messageAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewStateValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	if diags.HasError() {
		return NewStateValueUnknown(), diags
	}

	return StateValue{
		Message: messageVal,
		Name:    nameVal,
		state:   attr.ValueStateKnown,
	}, diags
}

func NewStateValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) StateValue {
	object, diags := NewStateValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewStateValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t StateType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewStateValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewStateValueUnknown(), nil
	}

	if in.IsNull() {
		return NewStateValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewStateValueMust(StateValue{}.AttributeTypes(ctx), attributes), nil
}

func (t StateType) ValueType(ctx context.Context) attr.Value {
	return StateValue{}
}

var _ basetypes.ObjectValuable = StateValue{}

type StateValue struct {
	Message basetypes.StringValue `tfsdk:"message"`
	Name    basetypes.StringValue `tfsdk:"name"`
	state   attr.ValueState
}

func (v StateValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["message"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.Message.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["message"] = val

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v StateValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v StateValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v StateValue) String() string {
	return "StateValue"
}

func (v StateValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"message": basetypes.StringType{},
			"name":    basetypes.StringType{},
		},
		map[string]attr.Value{
			"message": v.Message,
			"name":    v.Name,
		})

	return objVal, diags
}

func (v StateValue) Equal(o attr.Value) bool {
	other, ok := o.(StateValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Message.Equal(other.Message) {
		return false
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	return true
}

func (v StateValue) Type(ctx context.Context) attr.Type {
	return StateType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v StateValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"message": basetypes.StringType{},
		"name":    basetypes.StringType{},
	}
}
