// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package natgateway

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func NatGatewayResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the NAT gateway.",
				MarkdownDescription: "The ID of the NAT gateway.",
			},
			"public_ip_id": schema.StringAttribute{
				Required:            true,
				Description:         "The allocation ID of the public IP to associate with the NAT gateway.<br />\nIf the public IP is already associated with another resource, you must first disassociate it.",
				MarkdownDescription: "The allocation ID of the public IP to associate with the NAT gateway.<br />\nIf the public IP is already associated with another resource, you must first disassociate it.",
				PlanModifiers: []planmodifier.String{ // MANUALLY EDITED : Add requires replace
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_ips": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"public_ip": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(), // MANUALLY EDITED : Adds RequireReplace
							},
							Description:         "The public IP associated with the NAT gateway.",
							MarkdownDescription: "The public IP associated with the NAT gateway.",
						},
						"public_ip_id": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(), // MANUALLY EDITED : Adds RequireReplace
							},
							Description:         "The allocation ID of the public IP associated with the NAT gateway.",
							MarkdownDescription: "The allocation ID of the public IP associated with the NAT gateway.",
						},
					},
					CustomType: PublicIpsType{
						ObjectType: types.ObjectType{
							AttrTypes: PublicIpsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed:            true,
				Description:         "Information about the public IP or IPs associated with the NAT gateway.",
				MarkdownDescription: "Information about the public IP or IPs associated with the NAT gateway.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the NAT gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
				MarkdownDescription: "The state of the NAT gateway (`pending` \\| `available` \\| `deleting` \\| `deleted`).",
			},
			"subnet_id": schema.StringAttribute{
				Required:            true,
				Description:         "The ID of the Subnet in which you want to create the NAT gateway.",
				MarkdownDescription: "The ID of the Subnet in which you want to create the NAT gateway.",
				PlanModifiers: []planmodifier.String{ // MANUALLY EDITED : Add requires replace
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": tags.TagsSchema(ctx), // MANUALLY EDITED : Use shared tags
			"vpc_id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the Vpc in which the NAT gateway is.",
				MarkdownDescription: "The ID of the Vpc in which the NAT gateway is.",
			},
			// MANUALLY EDITED : SpaceId Removed
		},
		DeprecationMessage: "Managing IAAS services with Terraform is deprecated", // MANUALLY EDITED : Add Deprecation message
	}
}

type NatGatewayModel struct {
	Id         types.String `tfsdk:"id"`
	PublicIpId types.String `tfsdk:"public_ip_id"`
	PublicIps  types.List   `tfsdk:"public_ips"`
	State      types.String `tfsdk:"state"`
	SubnetId   types.String `tfsdk:"subnet_id"`
	VpcId      types.String `tfsdk:"vpc_id"`
	Tags       types.List   `tfsdk:"tags"`
	// MANUALLY EDITED : SpaceId Removed
}

// MANUALLY EDITED : All functions associated with Tags removed

var _ basetypes.ObjectTypable = PublicIpsType{}

type PublicIpsType struct {
	basetypes.ObjectType
}

func (t PublicIpsType) Equal(o attr.Type) bool {
	other, ok := o.(PublicIpsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t PublicIpsType) String() string {
	return "PublicIpsType"
}

func (t PublicIpsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	publicIpAttribute, ok := attributes["public_ip"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`public_ip is missing from object`)

		return nil, diags
	}

	publicIpVal, ok := publicIpAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`public_ip expected to be basetypes.StringValue, was: %T`, publicIpAttribute))
	}

	publicIpIdAttribute, ok := attributes["public_ip_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`public_ip_id is missing from object`)

		return nil, diags
	}

	publicIpIdVal, ok := publicIpIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`public_ip_id expected to be basetypes.StringValue, was: %T`, publicIpIdAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return PublicIpsValue{
		PublicIp:   publicIpVal,
		PublicIpId: publicIpIdVal,
		state:      attr.ValueStateKnown,
	}, diags
}

func NewPublicIpsValueNull() PublicIpsValue {
	return PublicIpsValue{
		state: attr.ValueStateNull,
	}
}

func NewPublicIpsValueUnknown() PublicIpsValue {
	return PublicIpsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewPublicIpsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (PublicIpsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing PublicIpsValue Attribute Value",
				"While creating a PublicIpsValue value, a missing attribute value was detected. "+
					"A PublicIpsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("PublicIpsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid PublicIpsValue Attribute Type",
				"While creating a PublicIpsValue value, an invalid attribute value was detected. "+
					"A PublicIpsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("PublicIpsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("PublicIpsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra PublicIpsValue Attribute Value",
				"While creating a PublicIpsValue value, an extra attribute value was detected. "+
					"A PublicIpsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra PublicIpsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewPublicIpsValueUnknown(), diags
	}

	publicIpAttribute, ok := attributes["public_ip"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`public_ip is missing from object`)

		return NewPublicIpsValueUnknown(), diags
	}

	publicIpVal, ok := publicIpAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`public_ip expected to be basetypes.StringValue, was: %T`, publicIpAttribute))
	}

	publicIpIdAttribute, ok := attributes["public_ip_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`public_ip_id is missing from object`)

		return NewPublicIpsValueUnknown(), diags
	}

	publicIpIdVal, ok := publicIpIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`public_ip_id expected to be basetypes.StringValue, was: %T`, publicIpIdAttribute))
	}

	if diags.HasError() {
		return NewPublicIpsValueUnknown(), diags
	}

	return PublicIpsValue{
		PublicIp:   publicIpVal,
		PublicIpId: publicIpIdVal,
		state:      attr.ValueStateKnown,
	}, diags
}

func NewPublicIpsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) PublicIpsValue {
	object, diags := NewPublicIpsValue(attributeTypes, attributes)

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

		panic("NewPublicIpsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t PublicIpsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewPublicIpsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewPublicIpsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewPublicIpsValueNull(), nil
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

	return NewPublicIpsValueMust(PublicIpsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t PublicIpsType) ValueType(ctx context.Context) attr.Value {
	return PublicIpsValue{}
}

var _ basetypes.ObjectValuable = PublicIpsValue{}

type PublicIpsValue struct {
	PublicIp   basetypes.StringValue `tfsdk:"public_ip"`
	PublicIpId basetypes.StringValue `tfsdk:"public_ip_id"`
	state      attr.ValueState
}

func (v PublicIpsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["public_ip"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["public_ip_id"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.PublicIp.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["public_ip"] = val

		val, err = v.PublicIpId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["public_ip_id"] = val

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

func (v PublicIpsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v PublicIpsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v PublicIpsValue) String() string {
	return "PublicIpsValue"
}

func (v PublicIpsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"public_ip":    basetypes.StringType{},
			"public_ip_id": basetypes.StringType{},
		},
		map[string]attr.Value{
			"public_ip":    v.PublicIp,
			"public_ip_id": v.PublicIpId,
		})

	return objVal, diags
}

func (v PublicIpsValue) Equal(o attr.Value) bool {
	other, ok := o.(PublicIpsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.PublicIp.Equal(other.PublicIp) {
		return false
	}

	if !v.PublicIpId.Equal(other.PublicIpId) {
		return false
	}

	return true
}

func (v PublicIpsValue) Type(ctx context.Context) attr.Type {
	return PublicIpsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v PublicIpsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"public_ip":    basetypes.StringType{},
		"public_ip_id": basetypes.StringType{},
	}
}
