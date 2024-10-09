// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package virtualgateway

import (
	"context"
	"fmt"
	"strings"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func VirtualGatewayResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"connection_type": schema.StringAttribute{
				Required:            true,
				Description:         "The type of VPN connection supported by the virtual gateway (only `ipsec.1` is supported).",
				MarkdownDescription: "The type of VPN connection supported by the virtual gateway (only `ipsec.1` is supported).",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the virtual gateway.",
				MarkdownDescription: "The ID of the virtual gateway.",
			},

			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the virtual gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
				MarkdownDescription: "The state of the virtual gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
			},
			"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
			"vpc_id": schema.StringAttribute{ // MANUALLY EDITED : add vpc_id attribute
				Computed:            true,
				Optional:            true,
				Description:         "The ID of the Vpc to which the virtual gateway is attached.",
				MarkdownDescription: "The ID of the Vpc to which the virtual gateway is attached.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // MANUALLY EDITED : Adds RequireReplace
				},
			},
			"vpc_to_virtual_gateway_links": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the attachment (`attaching` \\| `attached` \\| `detaching` \\| `detached`).",
							MarkdownDescription: "The state of the attachment (`attaching` \\| `attached` \\| `detaching` \\| `detached`).",
						},
						"vpc_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Vpc to which the virtual gateway is attached.",
							MarkdownDescription: "The ID of the Vpc to which the virtual gateway is attached.",
						},
					},
					CustomType: VpcToVirtualGatewayLinksType{
						ObjectType: types.ObjectType{
							AttrTypes: VpcToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed:            true,
				Description:         "the Vpc to which the virtual gateway is attached.",
				MarkdownDescription: "the Vpc to which the virtual gateway is attached.",
			},
			// MANUALLY EDITED : SpaceId Removed
		},
	}
}

type VirtualGatewayModel struct {
	ConnectionType           types.String `tfsdk:"connection_type"`
	Id                       types.String `tfsdk:"id"`
	State                    types.String `tfsdk:"state"`
	Tags                     types.List   `tfsdk:"tags"`
	VpcToVirtualGatewayLinks types.List   `tfsdk:"vpc_to_virtual_gateway_links"`
	VpcId                    types.String `tfsdk:"vpc_id"` // MANUALLY EDITED : add vpc_id attribute
	// MANUALLY EDITED : SpaceId Removed
}

var _ basetypes.ObjectTypable = VpcToVirtualGatewayLinksType{}

type VpcToVirtualGatewayLinksType struct {
	basetypes.ObjectType
}

func (t VpcToVirtualGatewayLinksType) Equal(o attr.Type) bool {
	other, ok := o.(VpcToVirtualGatewayLinksType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t VpcToVirtualGatewayLinksType) String() string {
	return "VpcToVirtualGatewayLinksType"
}

func (t VpcToVirtualGatewayLinksType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	stateAttribute, ok := attributes["state"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`state is missing from object`)

		return nil, diags
	}

	stateVal, ok := stateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`state expected to be basetypes.StringValue, was: %T`, stateAttribute))
	}

	vpcIdAttribute, ok := attributes["vpc_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`vpc_id is missing from object`)

		return nil, diags
	}

	vpcIdVal, ok := vpcIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`vpc_id expected to be basetypes.StringValue, was: %T`, vpcIdAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return VpcToVirtualGatewayLinksValue{
		State: stateVal,
		VpcId: vpcIdVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewVpcToVirtualGatewayLinksValueNull() VpcToVirtualGatewayLinksValue {
	return VpcToVirtualGatewayLinksValue{
		state: attr.ValueStateNull,
	}
}

func NewVpcToVirtualGatewayLinksValueUnknown() VpcToVirtualGatewayLinksValue {
	return VpcToVirtualGatewayLinksValue{
		state: attr.ValueStateUnknown,
	}
}

func NewVpcToVirtualGatewayLinksValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (VpcToVirtualGatewayLinksValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing VpcToVirtualGatewayLinksValue Attribute Value",
				"While creating a VpcToVirtualGatewayLinksValue value, a missing attribute value was detected. "+
					"A VpcToVirtualGatewayLinksValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("VpcToVirtualGatewayLinksValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid VpcToVirtualGatewayLinksValue Attribute Type",
				"While creating a VpcToVirtualGatewayLinksValue value, an invalid attribute value was detected. "+
					"A VpcToVirtualGatewayLinksValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("VpcToVirtualGatewayLinksValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("VpcToVirtualGatewayLinksValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra VpcToVirtualGatewayLinksValue Attribute Value",
				"While creating a VpcToVirtualGatewayLinksValue value, an extra attribute value was detected. "+
					"A VpcToVirtualGatewayLinksValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra VpcToVirtualGatewayLinksValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewVpcToVirtualGatewayLinksValueUnknown(), diags
	}

	stateAttribute, ok := attributes["state"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`state is missing from object`)

		return NewVpcToVirtualGatewayLinksValueUnknown(), diags
	}

	stateVal, ok := stateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`state expected to be basetypes.StringValue, was: %T`, stateAttribute))
	}

	vpcIdAttribute, ok := attributes["vpc_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`vpc_id is missing from object`)

		return NewVpcToVirtualGatewayLinksValueUnknown(), diags
	}

	vpcIdVal, ok := vpcIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`vpc_id expected to be basetypes.StringValue, was: %T`, vpcIdAttribute))
	}

	if diags.HasError() {
		return NewVpcToVirtualGatewayLinksValueUnknown(), diags
	}

	return VpcToVirtualGatewayLinksValue{
		State: stateVal,
		VpcId: vpcIdVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewVpcToVirtualGatewayLinksValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) VpcToVirtualGatewayLinksValue {
	object, diags := NewVpcToVirtualGatewayLinksValue(attributeTypes, attributes)

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

		panic("NewVpcToVirtualGatewayLinksValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t VpcToVirtualGatewayLinksType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewVpcToVirtualGatewayLinksValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewVpcToVirtualGatewayLinksValueUnknown(), nil
	}

	if in.IsNull() {
		return NewVpcToVirtualGatewayLinksValueNull(), nil
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

	return NewVpcToVirtualGatewayLinksValueMust(VpcToVirtualGatewayLinksValue{}.AttributeTypes(ctx), attributes), nil
}

func (t VpcToVirtualGatewayLinksType) ValueType(ctx context.Context) attr.Value {
	return VpcToVirtualGatewayLinksValue{}
}

var _ basetypes.ObjectValuable = VpcToVirtualGatewayLinksValue{}

type VpcToVirtualGatewayLinksValue struct {
	State basetypes.StringValue `tfsdk:"state"`
	VpcId basetypes.StringValue `tfsdk:"vpc_id"`
	state attr.ValueState
}

func (v VpcToVirtualGatewayLinksValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["state"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["vpc_id"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.State.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["state"] = val

		val, err = v.VpcId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["vpc_id"] = val

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

func (v VpcToVirtualGatewayLinksValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v VpcToVirtualGatewayLinksValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v VpcToVirtualGatewayLinksValue) String() string {
	return "VpcToVirtualGatewayLinksValue"
}

func (v VpcToVirtualGatewayLinksValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"state":  basetypes.StringType{},
		"vpc_id": basetypes.StringType{},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"state":  v.State,
			"vpc_id": v.VpcId,
		})

	return objVal, diags
}

func (v VpcToVirtualGatewayLinksValue) Equal(o attr.Value) bool {
	other, ok := o.(VpcToVirtualGatewayLinksValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.State.Equal(other.State) {
		return false
	}

	if !v.VpcId.Equal(other.VpcId) {
		return false
	}

	return true
}

func (v VpcToVirtualGatewayLinksValue) Type(ctx context.Context) attr.Type {
	return VpcToVirtualGatewayLinksType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v VpcToVirtualGatewayLinksValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"state":  basetypes.StringType{},
		"vpc_id": basetypes.StringType{},
	}
}