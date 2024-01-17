// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_direct_link_interface

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

func DirectLinkInterfaceResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bgp_asn": schema.Int64Attribute{
				Computed:            true,
				Description:         "The BGP (Border Gateway Protocol) ASN (Autonomous System Number) on the customer's side of the DirectLink interface.",
				MarkdownDescription: "The BGP (Border Gateway Protocol) ASN (Autonomous System Number) on the customer's side of the DirectLink interface.",
			},
			"bgp_key": schema.StringAttribute{
				Computed:            true,
				Description:         "The BGP authentication key.",
				MarkdownDescription: "The BGP authentication key.",
			},
			"client_private_ip": schema.StringAttribute{
				Computed:            true,
				Description:         "The IP on the customer's side of the DirectLink interface.",
				MarkdownDescription: "The IP on the customer's side of the DirectLink interface.",
			},
			"direct_link_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The ID of the existing DirectLink for which you want to create the DirectLink interface.",
				MarkdownDescription: "The ID of the existing DirectLink for which you want to create the DirectLink interface.",
			},
			"direct_link_interface": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"bgp_asn": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Description:         "The BGP (Border Gateway Protocol) ASN (Autonomous System Number) on the customer's side of the DirectLink interface. This number must be between `64512` and `65534`.",
						MarkdownDescription: "The BGP (Border Gateway Protocol) ASN (Autonomous System Number) on the customer's side of the DirectLink interface. This number must be between `64512` and `65534`.",
					},
					"bgp_key": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "The BGP authentication key.",
						MarkdownDescription: "The BGP authentication key.",
					},
					"client_private_ip": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "The IP on the customer's side of the DirectLink interface.",
						MarkdownDescription: "The IP on the customer's side of the DirectLink interface.",
					},
					"name": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "The name of the DirectLink interface.",
						MarkdownDescription: "The name of the DirectLink interface.",
					},
					"numspot_private_ip": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "The IP on the OUTSCALE side of the DirectLink interface.",
						MarkdownDescription: "The IP on the OUTSCALE side of the DirectLink interface.",
					},
					"virtual_gateway_id": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "The ID of the target virtual gateway.",
						MarkdownDescription: "The ID of the target virtual gateway.",
					},
					"vlan": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Description:         "The VLAN number associated with the DirectLink interface.",
						MarkdownDescription: "The VLAN number associated with the DirectLink interface.",
					},
				},
				CustomType: DirectLinkInterfaceType{
					ObjectType: types.ObjectType{
						AttrTypes: DirectLinkInterfaceValue{}.AttributeTypes(ctx),
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "Information about the DirectLink interface.",
				MarkdownDescription: "Information about the DirectLink interface.",
			},
			"direct_link_interface_id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the DirectLink interface.",
				MarkdownDescription: "The ID of the DirectLink interface.",
			},
			"direct_link_interface_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The name of the DirectLink interface.",
				MarkdownDescription: "The name of the DirectLink interface.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "ID for /directLinkInterfaces",
				MarkdownDescription: "ID for /directLinkInterfaces",
			},
			"interface_type": schema.StringAttribute{
				Computed:            true,
				Description:         "The type of the DirectLink interface (always `private`).",
				MarkdownDescription: "The type of the DirectLink interface (always `private`).",
			},
			"location": schema.StringAttribute{
				Computed:            true,
				Description:         "The datacenter where the DirectLink interface is located.",
				MarkdownDescription: "The datacenter where the DirectLink interface is located.",
			},
			"mtu": schema.Int64Attribute{
				Computed:            true,
				Description:         "The maximum transmission unit (MTU) of the DirectLink interface, in bytes (always `1500`).",
				MarkdownDescription: "The maximum transmission unit (MTU) of the DirectLink interface, in bytes (always `1500`).",
			},
			"outscale_private_ip": schema.StringAttribute{
				Computed:            true,
				Description:         "The IP on the OUTSCALE side of the DirectLink interface.",
				MarkdownDescription: "The IP on the OUTSCALE side of the DirectLink interface.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				Description:         "The state of the DirectLink interface (`pending` \\| `available` \\| `deleting` \\| `deleted` \\| `confirming` \\| `rejected` \\| `expired`).",
				MarkdownDescription: "The state of the DirectLink interface (`pending` \\| `available` \\| `deleting` \\| `deleted` \\| `confirming` \\| `rejected` \\| `expired`).",
			},
			"virtual_gateway_id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the target virtual gateway.",
				MarkdownDescription: "The ID of the target virtual gateway.",
			},
			"vlan": schema.Int64Attribute{
				Computed:            true,
				Description:         "The VLAN number associated with the DirectLink interface.",
				MarkdownDescription: "The VLAN number associated with the DirectLink interface.",
			},
		},
	}
}

type DirectLinkInterfaceModel struct {
	BgpAsn                  types.Int64              `tfsdk:"bgp_asn"`
	BgpKey                  types.String             `tfsdk:"bgp_key"`
	ClientPrivateIp         types.String             `tfsdk:"client_private_ip"`
	DirectLinkId            types.String             `tfsdk:"direct_link_id"`
	DirectLinkInterface     DirectLinkInterfaceValue `tfsdk:"direct_link_interface"`
	DirectLinkInterfaceId   types.String             `tfsdk:"direct_link_interface_id"`
	DirectLinkInterfaceName types.String             `tfsdk:"direct_link_interface_name"`
	Id                      types.String             `tfsdk:"id"`
	InterfaceType           types.String             `tfsdk:"interface_type"`
	Location                types.String             `tfsdk:"location"`
	Mtu                     types.Int64              `tfsdk:"mtu"`
	OutscalePrivateIp       types.String             `tfsdk:"outscale_private_ip"`
	State                   types.String             `tfsdk:"state"`
	VirtualGatewayId        types.String             `tfsdk:"virtual_gateway_id"`
	Vlan                    types.Int64              `tfsdk:"vlan"`
}

var _ basetypes.ObjectTypable = DirectLinkInterfaceType{}

type DirectLinkInterfaceType struct {
	basetypes.ObjectType
}

func (t DirectLinkInterfaceType) Equal(o attr.Type) bool {
	other, ok := o.(DirectLinkInterfaceType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t DirectLinkInterfaceType) String() string {
	return "DirectLinkInterfaceType"
}

func (t DirectLinkInterfaceType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	bgpAsnAttribute, ok := attributes["bgp_asn"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`bgp_asn is missing from object`)

		return nil, diags
	}

	bgpAsnVal, ok := bgpAsnAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`bgp_asn expected to be basetypes.Int64Value, was: %T`, bgpAsnAttribute))
	}

	bgpKeyAttribute, ok := attributes["bgp_key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`bgp_key is missing from object`)

		return nil, diags
	}

	bgpKeyVal, ok := bgpKeyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`bgp_key expected to be basetypes.StringValue, was: %T`, bgpKeyAttribute))
	}

	clientPrivateIpAttribute, ok := attributes["client_private_ip"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`client_private_ip is missing from object`)

		return nil, diags
	}

	clientPrivateIpVal, ok := clientPrivateIpAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`client_private_ip expected to be basetypes.StringValue, was: %T`, clientPrivateIpAttribute))
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

	numspotPrivateIpAttribute, ok := attributes["numspot_private_ip"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`numspot_private_ip is missing from object`)

		return nil, diags
	}

	numspotPrivateIpVal, ok := numspotPrivateIpAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`numspot_private_ip expected to be basetypes.StringValue, was: %T`, numspotPrivateIpAttribute))
	}

	virtualGatewayIdAttribute, ok := attributes["virtual_gateway_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`virtual_gateway_id is missing from object`)

		return nil, diags
	}

	virtualGatewayIdVal, ok := virtualGatewayIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`virtual_gateway_id expected to be basetypes.StringValue, was: %T`, virtualGatewayIdAttribute))
	}

	vlanAttribute, ok := attributes["vlan"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`vlan is missing from object`)

		return nil, diags
	}

	vlanVal, ok := vlanAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`vlan expected to be basetypes.Int64Value, was: %T`, vlanAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return DirectLinkInterfaceValue{
		BgpAsn:           bgpAsnVal,
		BgpKey:           bgpKeyVal,
		ClientPrivateIp:  clientPrivateIpVal,
		Name:             nameVal,
		NumspotPrivateIp: numspotPrivateIpVal,
		VirtualGatewayId: virtualGatewayIdVal,
		Vlan:             vlanVal,
		state:            attr.ValueStateKnown,
	}, diags
}

func NewDirectLinkInterfaceValueNull() DirectLinkInterfaceValue {
	return DirectLinkInterfaceValue{
		state: attr.ValueStateNull,
	}
}

func NewDirectLinkInterfaceValueUnknown() DirectLinkInterfaceValue {
	return DirectLinkInterfaceValue{
		state: attr.ValueStateUnknown,
	}
}

func NewDirectLinkInterfaceValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (DirectLinkInterfaceValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing DirectLinkInterfaceValue Attribute Value",
				"While creating a DirectLinkInterfaceValue value, a missing attribute value was detected. "+
					"A DirectLinkInterfaceValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("DirectLinkInterfaceValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid DirectLinkInterfaceValue Attribute Type",
				"While creating a DirectLinkInterfaceValue value, an invalid attribute value was detected. "+
					"A DirectLinkInterfaceValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("DirectLinkInterfaceValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("DirectLinkInterfaceValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra DirectLinkInterfaceValue Attribute Value",
				"While creating a DirectLinkInterfaceValue value, an extra attribute value was detected. "+
					"A DirectLinkInterfaceValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra DirectLinkInterfaceValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewDirectLinkInterfaceValueUnknown(), diags
	}

	bgpAsnAttribute, ok := attributes["bgp_asn"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`bgp_asn is missing from object`)

		return NewDirectLinkInterfaceValueUnknown(), diags
	}

	bgpAsnVal, ok := bgpAsnAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`bgp_asn expected to be basetypes.Int64Value, was: %T`, bgpAsnAttribute))
	}

	bgpKeyAttribute, ok := attributes["bgp_key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`bgp_key is missing from object`)

		return NewDirectLinkInterfaceValueUnknown(), diags
	}

	bgpKeyVal, ok := bgpKeyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`bgp_key expected to be basetypes.StringValue, was: %T`, bgpKeyAttribute))
	}

	clientPrivateIpAttribute, ok := attributes["client_private_ip"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`client_private_ip is missing from object`)

		return NewDirectLinkInterfaceValueUnknown(), diags
	}

	clientPrivateIpVal, ok := clientPrivateIpAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`client_private_ip expected to be basetypes.StringValue, was: %T`, clientPrivateIpAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewDirectLinkInterfaceValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	numspotPrivateIpAttribute, ok := attributes["numspot_private_ip"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`numspot_private_ip is missing from object`)

		return NewDirectLinkInterfaceValueUnknown(), diags
	}

	numspotPrivateIpVal, ok := numspotPrivateIpAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`numspot_private_ip expected to be basetypes.StringValue, was: %T`, numspotPrivateIpAttribute))
	}

	virtualGatewayIdAttribute, ok := attributes["virtual_gateway_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`virtual_gateway_id is missing from object`)

		return NewDirectLinkInterfaceValueUnknown(), diags
	}

	virtualGatewayIdVal, ok := virtualGatewayIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`virtual_gateway_id expected to be basetypes.StringValue, was: %T`, virtualGatewayIdAttribute))
	}

	vlanAttribute, ok := attributes["vlan"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`vlan is missing from object`)

		return NewDirectLinkInterfaceValueUnknown(), diags
	}

	vlanVal, ok := vlanAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`vlan expected to be basetypes.Int64Value, was: %T`, vlanAttribute))
	}

	if diags.HasError() {
		return NewDirectLinkInterfaceValueUnknown(), diags
	}

	return DirectLinkInterfaceValue{
		BgpAsn:           bgpAsnVal,
		BgpKey:           bgpKeyVal,
		ClientPrivateIp:  clientPrivateIpVal,
		Name:             nameVal,
		NumspotPrivateIp: numspotPrivateIpVal,
		VirtualGatewayId: virtualGatewayIdVal,
		Vlan:             vlanVal,
		state:            attr.ValueStateKnown,
	}, diags
}

func NewDirectLinkInterfaceValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) DirectLinkInterfaceValue {
	object, diags := NewDirectLinkInterfaceValue(attributeTypes, attributes)

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

		panic("NewDirectLinkInterfaceValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t DirectLinkInterfaceType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewDirectLinkInterfaceValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewDirectLinkInterfaceValueUnknown(), nil
	}

	if in.IsNull() {
		return NewDirectLinkInterfaceValueNull(), nil
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

	return NewDirectLinkInterfaceValueMust(DirectLinkInterfaceValue{}.AttributeTypes(ctx), attributes), nil
}

func (t DirectLinkInterfaceType) ValueType(ctx context.Context) attr.Value {
	return DirectLinkInterfaceValue{}
}

var _ basetypes.ObjectValuable = DirectLinkInterfaceValue{}

type DirectLinkInterfaceValue struct {
	BgpAsn           basetypes.Int64Value  `tfsdk:"bgp_asn"`
	BgpKey           basetypes.StringValue `tfsdk:"bgp_key"`
	ClientPrivateIp  basetypes.StringValue `tfsdk:"client_private_ip"`
	Name             basetypes.StringValue `tfsdk:"name"`
	NumspotPrivateIp basetypes.StringValue `tfsdk:"numspot_private_ip"`
	VirtualGatewayId basetypes.StringValue `tfsdk:"virtual_gateway_id"`
	Vlan             basetypes.Int64Value  `tfsdk:"vlan"`
	state            attr.ValueState
}

func (v DirectLinkInterfaceValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 7)

	var val tftypes.Value
	var err error

	attrTypes["bgp_asn"] = basetypes.Int64Type{}.TerraformType(ctx)
	attrTypes["bgp_key"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["client_private_ip"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["numspot_private_ip"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["virtual_gateway_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["vlan"] = basetypes.Int64Type{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 7)

		val, err = v.BgpAsn.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["bgp_asn"] = val

		val, err = v.BgpKey.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["bgp_key"] = val

		val, err = v.ClientPrivateIp.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["client_private_ip"] = val

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		val, err = v.NumspotPrivateIp.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["numspot_private_ip"] = val

		val, err = v.VirtualGatewayId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["virtual_gateway_id"] = val

		val, err = v.Vlan.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["vlan"] = val

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

func (v DirectLinkInterfaceValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v DirectLinkInterfaceValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v DirectLinkInterfaceValue) String() string {
	return "DirectLinkInterfaceValue"
}

func (v DirectLinkInterfaceValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"bgp_asn":            basetypes.Int64Type{},
			"bgp_key":            basetypes.StringType{},
			"client_private_ip":  basetypes.StringType{},
			"name":               basetypes.StringType{},
			"numspot_private_ip": basetypes.StringType{},
			"virtual_gateway_id": basetypes.StringType{},
			"vlan":               basetypes.Int64Type{},
		},
		map[string]attr.Value{
			"bgp_asn":            v.BgpAsn,
			"bgp_key":            v.BgpKey,
			"client_private_ip":  v.ClientPrivateIp,
			"name":               v.Name,
			"numspot_private_ip": v.NumspotPrivateIp,
			"virtual_gateway_id": v.VirtualGatewayId,
			"vlan":               v.Vlan,
		})

	return objVal, diags
}

func (v DirectLinkInterfaceValue) Equal(o attr.Value) bool {
	other, ok := o.(DirectLinkInterfaceValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.BgpAsn.Equal(other.BgpAsn) {
		return false
	}

	if !v.BgpKey.Equal(other.BgpKey) {
		return false
	}

	if !v.ClientPrivateIp.Equal(other.ClientPrivateIp) {
		return false
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	if !v.NumspotPrivateIp.Equal(other.NumspotPrivateIp) {
		return false
	}

	if !v.VirtualGatewayId.Equal(other.VirtualGatewayId) {
		return false
	}

	if !v.Vlan.Equal(other.Vlan) {
		return false
	}

	return true
}

func (v DirectLinkInterfaceValue) Type(ctx context.Context) attr.Type {
	return DirectLinkInterfaceType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v DirectLinkInterfaceValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"bgp_asn":            basetypes.Int64Type{},
		"bgp_key":            basetypes.StringType{},
		"client_private_ip":  basetypes.StringType{},
		"name":               basetypes.StringType{},
		"numspot_private_ip": basetypes.StringType{},
		"virtual_gateway_id": basetypes.StringType{},
		"vlan":               basetypes.Int64Type{},
	}
}
