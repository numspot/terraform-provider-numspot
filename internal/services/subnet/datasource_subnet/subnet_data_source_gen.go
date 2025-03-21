// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package datasource_subnet

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func SubnetDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone_names": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The names of the Subregions in which the Subnets are located.",
				MarkdownDescription: "The names of the Subregions in which the Subnets are located.",
			},
			"available_ips_counts": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
				Description:         "The number of available IPs.",
				MarkdownDescription: "The number of available IPs.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the Subnets.",
				MarkdownDescription: "The IDs of the Subnets.",
			},
			"ip_ranges": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IP ranges in the Subnets, in CIDR notation (for example, `10.0.0.0/16`).",
				MarkdownDescription: "The IP ranges in the Subnets, in CIDR notation (for example, `10.0.0.0/16`).",
			},
			"items": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"availability_zone_name": schema.StringAttribute{
							Computed:            true,
							Description:         "The name of the Subregion in which the Subnet is located.",
							MarkdownDescription: "The name of the Subregion in which the Subnet is located.",
						},
						"available_ips_count": schema.Int64Attribute{
							Computed:            true,
							Description:         "The number of available IPs in the Subnets.",
							MarkdownDescription: "The number of available IPs in the Subnets.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Subnet.",
							MarkdownDescription: "The ID of the Subnet.",
						},
						"ip_range": schema.StringAttribute{
							Computed:            true,
							Description:         "The IP range in the Subnet, in CIDR notation (for example, `10.0.0.0/16`).",
							MarkdownDescription: "The IP range in the Subnet, in CIDR notation (for example, `10.0.0.0/16`).",
						},
						"map_public_ip_on_launch": schema.BoolAttribute{
							Computed:            true,
							Description:         "If true, a public IP is assigned to the network interface cards (NICs) created in the specified Subnet.",
							MarkdownDescription: "If true, a public IP is assigned to the network interface cards (NICs) created in the specified Subnet.",
						},
						"state": schema.StringAttribute{
							Computed:            true,
							Description:         "The state of the Subnet (`pending` \\| `available` \\| `deleted`).",
							MarkdownDescription: "The state of the Subnet (`pending` \\| `available` \\| `deleted`).",
						},
						"tags": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Computed:            true,
										Description:         "The key of the tag, with a minimum of 1 character.",
										MarkdownDescription: "The key of the tag, with a minimum of 1 character.",
									},
									"value": schema.StringAttribute{
										Computed:            true,
										Description:         "The value of the tag, between 0 and 255 characters.",
										MarkdownDescription: "The value of the tag, between 0 and 255 characters.",
									},
								},
								CustomType: TagsType{
									ObjectType: types.ObjectType{
										AttrTypes: TagsValue{}.AttributeTypes(ctx),
									},
								},
							},
							Computed:            true,
							Description:         "One or more tags associated with the Subnet.",
							MarkdownDescription: "One or more tags associated with the Subnet.",
						},
						"vpc_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The ID of the Vpc in which the Subnet is.",
							MarkdownDescription: "The ID of the Vpc in which the Subnet is.",
						},
					},
					CustomType: ItemsType{
						ObjectType: types.ObjectType{
							AttrTypes: ItemsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed:            true,
				Description:         "Information about one or more Subnets.",
				MarkdownDescription: "Information about one or more Subnets.",
			},
			"states": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The states of the Subnets (`pending` \\| `available` \\| `deleted`).",
				MarkdownDescription: "The states of the Subnets (`pending` \\| `available` \\| `deleted`).",
			},
			"tag_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The keys of the tags associated with the Subnets.",
				MarkdownDescription: "The keys of the tags associated with the Subnets.",
			},
			"tag_values": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The values of the tags associated with the Subnets.",
				MarkdownDescription: "The values of the tags associated with the Subnets.",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The key/value combination of the tags associated with the Subnets, in the following format: &quot;Filters&quot;:{&quot;Tags&quot;:[&quot;TAGKEY=TAGVALUE&quot;]}.",
				MarkdownDescription: "The key/value combination of the tags associated with the Subnets, in the following format: &quot;Filters&quot;:{&quot;Tags&quot;:[&quot;TAGKEY=TAGVALUE&quot;]}.",
			},
			"vpc_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The IDs of the Vpcs in which the Subnets are.",
				MarkdownDescription: "The IDs of the Vpcs in which the Subnets are.",
			},
		},
	}
}

type SubnetModel struct {
	AvailabilityZoneNames types.List `tfsdk:"availability_zone_names"`
	AvailableIpsCounts    types.List `tfsdk:"available_ips_counts"`
	Ids                   types.List `tfsdk:"ids"`
	IpRanges              types.List `tfsdk:"ip_ranges"`
	Items                 types.List `tfsdk:"items"`
	States                types.List `tfsdk:"states"`
	TagKeys               types.List `tfsdk:"tag_keys"`
	TagValues             types.List `tfsdk:"tag_values"`
	Tags                  types.List `tfsdk:"tags"`
	VpcIds                types.List `tfsdk:"vpc_ids"`
}

var _ basetypes.ObjectTypable = ItemsType{}

type ItemsType struct {
	basetypes.ObjectType
}

func (t ItemsType) Equal(o attr.Type) bool {
	other, ok := o.(ItemsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ItemsType) String() string {
	return "ItemsType"
}

func (t ItemsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	availabilityZoneNameAttribute, ok := attributes["availability_zone_name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`availability_zone_name is missing from object`)

		return nil, diags
	}

	availabilityZoneNameVal, ok := availabilityZoneNameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`availability_zone_name expected to be basetypes.StringValue, was: %T`, availabilityZoneNameAttribute))
	}

	availableIpsCountAttribute, ok := attributes["available_ips_count"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`available_ips_count is missing from object`)

		return nil, diags
	}

	availableIpsCountVal, ok := availableIpsCountAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`available_ips_count expected to be basetypes.Int64Value, was: %T`, availableIpsCountAttribute))
	}

	idAttribute, ok := attributes["id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`id is missing from object`)

		return nil, diags
	}

	idVal, ok := idAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`id expected to be basetypes.StringValue, was: %T`, idAttribute))
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

	mapPublicIpOnLaunchAttribute, ok := attributes["map_public_ip_on_launch"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`map_public_ip_on_launch is missing from object`)

		return nil, diags
	}

	mapPublicIpOnLaunchVal, ok := mapPublicIpOnLaunchAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`map_public_ip_on_launch expected to be basetypes.BoolValue, was: %T`, mapPublicIpOnLaunchAttribute))
	}

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

	tagsAttribute, ok := attributes["tags"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`tags is missing from object`)

		return nil, diags
	}

	tagsVal, ok := tagsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`tags expected to be basetypes.ListValue, was: %T`, tagsAttribute))
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

	return ItemsValue{
		AvailabilityZoneName: availabilityZoneNameVal,
		AvailableIpsCount:    availableIpsCountVal,
		Id:                   idVal,
		IpRange:              ipRangeVal,
		MapPublicIpOnLaunch:  mapPublicIpOnLaunchVal,
		State:                stateVal,
		Tags:                 tagsVal,
		VpcId:                vpcIdVal,
		state:                attr.ValueStateKnown,
	}, diags
}

func NewItemsValueNull() ItemsValue {
	return ItemsValue{
		state: attr.ValueStateNull,
	}
}

func NewItemsValueUnknown() ItemsValue {
	return ItemsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewItemsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing ItemsValue Attribute Value",
				"While creating a ItemsValue value, a missing attribute value was detected. "+
					"A ItemsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ItemsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid ItemsValue Attribute Type",
				"While creating a ItemsValue value, an invalid attribute value was detected. "+
					"A ItemsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ItemsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("ItemsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra ItemsValue Attribute Value",
				"While creating a ItemsValue value, an extra attribute value was detected. "+
					"A ItemsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra ItemsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewItemsValueUnknown(), diags
	}

	availabilityZoneNameAttribute, ok := attributes["availability_zone_name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`availability_zone_name is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	availabilityZoneNameVal, ok := availabilityZoneNameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`availability_zone_name expected to be basetypes.StringValue, was: %T`, availabilityZoneNameAttribute))
	}

	availableIpsCountAttribute, ok := attributes["available_ips_count"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`available_ips_count is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	availableIpsCountVal, ok := availableIpsCountAttribute.(basetypes.Int64Value)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`available_ips_count expected to be basetypes.Int64Value, was: %T`, availableIpsCountAttribute))
	}

	idAttribute, ok := attributes["id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`id is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	idVal, ok := idAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`id expected to be basetypes.StringValue, was: %T`, idAttribute))
	}

	ipRangeAttribute, ok := attributes["ip_range"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`ip_range is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	ipRangeVal, ok := ipRangeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`ip_range expected to be basetypes.StringValue, was: %T`, ipRangeAttribute))
	}

	mapPublicIpOnLaunchAttribute, ok := attributes["map_public_ip_on_launch"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`map_public_ip_on_launch is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	mapPublicIpOnLaunchVal, ok := mapPublicIpOnLaunchAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`map_public_ip_on_launch expected to be basetypes.BoolValue, was: %T`, mapPublicIpOnLaunchAttribute))
	}

	stateAttribute, ok := attributes["state"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`state is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	stateVal, ok := stateAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`state expected to be basetypes.StringValue, was: %T`, stateAttribute))
	}

	tagsAttribute, ok := attributes["tags"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`tags is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	tagsVal, ok := tagsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`tags expected to be basetypes.ListValue, was: %T`, tagsAttribute))
	}

	vpcIdAttribute, ok := attributes["vpc_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`vpc_id is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	vpcIdVal, ok := vpcIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`vpc_id expected to be basetypes.StringValue, was: %T`, vpcIdAttribute))
	}

	if diags.HasError() {
		return NewItemsValueUnknown(), diags
	}

	return ItemsValue{
		AvailabilityZoneName: availabilityZoneNameVal,
		AvailableIpsCount:    availableIpsCountVal,
		Id:                   idVal,
		IpRange:              ipRangeVal,
		MapPublicIpOnLaunch:  mapPublicIpOnLaunchVal,
		State:                stateVal,
		Tags:                 tagsVal,
		VpcId:                vpcIdVal,
		state:                attr.ValueStateKnown,
	}, diags
}

func NewItemsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) ItemsValue {
	object, diags := NewItemsValue(attributeTypes, attributes)

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

		panic("NewItemsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t ItemsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewItemsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewItemsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewItemsValueNull(), nil
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

	return NewItemsValueMust(ItemsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t ItemsType) ValueType(ctx context.Context) attr.Value {
	return ItemsValue{}
}

var _ basetypes.ObjectValuable = ItemsValue{}

type ItemsValue struct {
	AvailabilityZoneName basetypes.StringValue `tfsdk:"availability_zone_name"`
	AvailableIpsCount    basetypes.Int64Value  `tfsdk:"available_ips_count"`
	Id                   basetypes.StringValue `tfsdk:"id"`
	IpRange              basetypes.StringValue `tfsdk:"ip_range"`
	MapPublicIpOnLaunch  basetypes.BoolValue   `tfsdk:"map_public_ip_on_launch"`
	State                basetypes.StringValue `tfsdk:"state"`
	Tags                 basetypes.ListValue   `tfsdk:"tags"`
	VpcId                basetypes.StringValue `tfsdk:"vpc_id"`
	state                attr.ValueState
}

func (v ItemsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 8)

	var val tftypes.Value
	var err error

	attrTypes["availability_zone_name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["available_ips_count"] = basetypes.Int64Type{}.TerraformType(ctx)
	attrTypes["id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["ip_range"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["map_public_ip_on_launch"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["state"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["tags"] = basetypes.ListType{
		ElemType: TagsValue{}.Type(ctx),
	}.TerraformType(ctx)
	attrTypes["vpc_id"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 8)

		val, err = v.AvailabilityZoneName.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["availability_zone_name"] = val

		val, err = v.AvailableIpsCount.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["available_ips_count"] = val

		val, err = v.Id.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["id"] = val

		val, err = v.IpRange.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["ip_range"] = val

		val, err = v.MapPublicIpOnLaunch.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["map_public_ip_on_launch"] = val

		val, err = v.State.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["state"] = val

		val, err = v.Tags.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["tags"] = val

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

func (v ItemsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ItemsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ItemsValue) String() string {
	return "ItemsValue"
}

func (v ItemsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	tags := types.ListValueMust(
		TagsType{
			basetypes.ObjectType{
				AttrTypes: TagsValue{}.AttributeTypes(ctx),
			},
		},
		v.Tags.Elements(),
	)

	if v.Tags.IsNull() {
		tags = types.ListNull(
			TagsType{
				basetypes.ObjectType{
					AttrTypes: TagsValue{}.AttributeTypes(ctx),
				},
			},
		)
	}

	if v.Tags.IsUnknown() {
		tags = types.ListUnknown(
			TagsType{
				basetypes.ObjectType{
					AttrTypes: TagsValue{}.AttributeTypes(ctx),
				},
			},
		)
	}

	attributeTypes := map[string]attr.Type{
		"availability_zone_name":  basetypes.StringType{},
		"available_ips_count":     basetypes.Int64Type{},
		"id":                      basetypes.StringType{},
		"ip_range":                basetypes.StringType{},
		"map_public_ip_on_launch": basetypes.BoolType{},
		"state":                   basetypes.StringType{},
		"tags": basetypes.ListType{
			ElemType: TagsValue{}.Type(ctx),
		},
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
			"availability_zone_name":  v.AvailabilityZoneName,
			"available_ips_count":     v.AvailableIpsCount,
			"id":                      v.Id,
			"ip_range":                v.IpRange,
			"map_public_ip_on_launch": v.MapPublicIpOnLaunch,
			"state":                   v.State,
			"tags":                    tags,
			"vpc_id":                  v.VpcId,
		})

	return objVal, diags
}

func (v ItemsValue) Equal(o attr.Value) bool {
	other, ok := o.(ItemsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.AvailabilityZoneName.Equal(other.AvailabilityZoneName) {
		return false
	}

	if !v.AvailableIpsCount.Equal(other.AvailableIpsCount) {
		return false
	}

	if !v.Id.Equal(other.Id) {
		return false
	}

	if !v.IpRange.Equal(other.IpRange) {
		return false
	}

	if !v.MapPublicIpOnLaunch.Equal(other.MapPublicIpOnLaunch) {
		return false
	}

	if !v.State.Equal(other.State) {
		return false
	}

	if !v.Tags.Equal(other.Tags) {
		return false
	}

	if !v.VpcId.Equal(other.VpcId) {
		return false
	}

	return true
}

func (v ItemsValue) Type(ctx context.Context) attr.Type {
	return ItemsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v ItemsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"availability_zone_name":  basetypes.StringType{},
		"available_ips_count":     basetypes.Int64Type{},
		"id":                      basetypes.StringType{},
		"ip_range":                basetypes.StringType{},
		"map_public_ip_on_launch": basetypes.BoolType{},
		"state":                   basetypes.StringType{},
		"tags": basetypes.ListType{
			ElemType: TagsValue{}.Type(ctx),
		},
		"vpc_id": basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = TagsType{}

type TagsType struct {
	basetypes.ObjectType
}

func (t TagsType) Equal(o attr.Type) bool {
	other, ok := o.(TagsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t TagsType) String() string {
	return "TagsType"
}

func (t TagsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	keyAttribute, ok := attributes["key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`key is missing from object`)

		return nil, diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`key expected to be basetypes.StringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`value is missing from object`)

		return nil, diags
	}

	valueVal, ok := valueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`value expected to be basetypes.StringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return TagsValue{
		Key:   keyVal,
		Value: valueVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewTagsValueNull() TagsValue {
	return TagsValue{
		state: attr.ValueStateNull,
	}
}

func NewTagsValueUnknown() TagsValue {
	return TagsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewTagsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (TagsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing TagsValue Attribute Value",
				"While creating a TagsValue value, a missing attribute value was detected. "+
					"A TagsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("TagsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid TagsValue Attribute Type",
				"While creating a TagsValue value, an invalid attribute value was detected. "+
					"A TagsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("TagsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("TagsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra TagsValue Attribute Value",
				"While creating a TagsValue value, an extra attribute value was detected. "+
					"A TagsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra TagsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewTagsValueUnknown(), diags
	}

	keyAttribute, ok := attributes["key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`key is missing from object`)

		return NewTagsValueUnknown(), diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`key expected to be basetypes.StringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`value is missing from object`)

		return NewTagsValueUnknown(), diags
	}

	valueVal, ok := valueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`value expected to be basetypes.StringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return NewTagsValueUnknown(), diags
	}

	return TagsValue{
		Key:   keyVal,
		Value: valueVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewTagsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) TagsValue {
	object, diags := NewTagsValue(attributeTypes, attributes)

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

		panic("NewTagsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t TagsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewTagsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewTagsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewTagsValueNull(), nil
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

	return NewTagsValueMust(TagsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t TagsType) ValueType(ctx context.Context) attr.Value {
	return TagsValue{}
}

var _ basetypes.ObjectValuable = TagsValue{}

type TagsValue struct {
	Key   basetypes.StringValue `tfsdk:"key"`
	Value basetypes.StringValue `tfsdk:"value"`
	state attr.ValueState
}

func (v TagsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["key"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["value"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.Key.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["key"] = val

		val, err = v.Value.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["value"] = val

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

func (v TagsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v TagsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v TagsValue) String() string {
	return "TagsValue"
}

func (v TagsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"key":   basetypes.StringType{},
		"value": basetypes.StringType{},
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
			"key":   v.Key,
			"value": v.Value,
		})

	return objVal, diags
}

func (v TagsValue) Equal(o attr.Value) bool {
	other, ok := o.(TagsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Key.Equal(other.Key) {
		return false
	}

	if !v.Value.Equal(other.Value) {
		return false
	}

	return true
}

func (v TagsValue) Type(ctx context.Context) attr.Type {
	return TagsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v TagsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"key":   basetypes.StringType{},
		"value": basetypes.StringType{},
	}
}
