package nic

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/nic/datasource_nic"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &nicsDataSource{}

// nicsDataSource is the Terraform data source implementation for numspot_nics.
type nicsDataSource struct {
	services.BaseDataSource
}

// NewNicsDataSource returns a new nicsDataSource instance.
func NewNicsDataSource() datasource.DataSource {
	return &nicsDataSource{}
}

// Metadata sets the data source type name.
func (d *nicsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nics"
}

// Schema defines the schema for the data source.
func (d *nicsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_nic.NicDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *nicsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_nic.NicModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadNics(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadNicsWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read nics", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.Nic) (datasource_nic.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	tagsList := types.ListNull(datasource_nic.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_nic.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_nic.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_nic.NewItemsValue(datasource_nic.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(item.AvailabilityZoneName)),
		"description":            types.StringValue(utils.ConvertStringPtrToString(item.Description)),
		"id":                     types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"is_source_dest_checked": types.BoolValue(*item.IsSourceDestChecked),
		"link_nic":               mappingLinkNic(ctx, item.LinkNic, &diags),
		"link_public_ip":         mappingLinkPublicIp(ctx, item.LinkPublicIp, &diags),
		"mac_address":            types.StringValue(utils.ConvertStringPtrToString(item.MacAddress)),
		"private_dns_name":       types.StringValue(utils.ConvertStringPtrToString(item.PrivateDnsName)),
		"private_ips":            mappingPrivateIps(ctx, item, &diags),
		"security_groups":        mappingSecurityGroups(ctx, item, &diags),
		"state":                  types.StringValue(utils.ConvertStringPtrToString(item.State)),
		"subnet_id":              types.StringValue(utils.ConvertStringPtrToString(item.SubnetId)),
		"vpc_id":                 types.StringValue(utils.ConvertStringPtrToString(item.VpcId)),
		"tags":                   tagsList,
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_nic.TagsValue, diag.Diagnostics) {
	return datasource_nic.NewTagsValue(datasource_nic.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingLinkNic(ctx context.Context, item *api.LinkNic, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_nic.LinkNicValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_nic.NewLinkNicValue(datasource_nic.LinkNicValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(item.DeleteOnVmDeletion),
		"device_number":         utils.FromIntPtrToTfInt64(item.DeviceNumber),
		"id":                    types.StringPointerValue(item.Id),
		"state":                 types.StringPointerValue(item.State),
		"vm_id":                 types.StringPointerValue(item.VmId),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingLinkPublicIp(ctx context.Context, item *api.LinkPublicIp, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_nic.LinkPublicIpValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_nic.NewLinkPublicIpValue(datasource_nic.LinkPublicIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":              types.StringPointerValue(item.Id),
		"public_dns_name": types.StringPointerValue(item.PublicDnsName),
		"public_ip":       types.StringPointerValue(item.PublicIp),
		"public_ip_id":    types.StringPointerValue(item.PublicIpId),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingLinkPublicIpPrivateIp(ctx context.Context, item *api.LinkPublicIp, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_nic.LinkPublicIpPrivateIpValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_nic.NewLinkPublicIpPrivateIpValue(datasource_nic.LinkPublicIpPrivateIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":              types.StringPointerValue(item.Id),
		"public_dns_name": types.StringPointerValue(item.PublicDnsName),
		"public_ip":       types.StringPointerValue(item.PublicIp),
		"public_ip_id":    types.StringPointerValue(item.PublicIpId),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingPrivateIp(ctx context.Context, item *api.PrivateIp) (datasource_nic.PrivateIpsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_nic.NewPrivateIpsValue(datasource_nic.PrivateIpsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"is_primary":                types.BoolPointerValue(item.IsPrimary),
		"link_public_ip_private_ip": mappingLinkPublicIpPrivateIp(ctx, item.LinkPublicIp, &diags),
		"private_dns_name":          types.StringPointerValue(item.PrivateDnsName),
		"private_ip":                types.StringPointerValue(item.PrivateIp),
	})
	diags.Append(d...)
	return value, diags
}

func mappingPrivateIps(ctx context.Context, item *api.Nic, diags *diag.Diagnostics) types.List {
	if item.PrivateIps == nil {
		return types.ListNull(datasource_nic.PrivateIpsValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.PrivateIps, mappingPrivateIp)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_nic.PrivateIpsValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingSecurityGroup(ctx context.Context, item *api.SecurityGroupLight) (datasource_nic.SecurityGroupsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_nic.NewSecurityGroupsValue(datasource_nic.SecurityGroupsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"security_group_id":   types.StringPointerValue(item.SecurityGroupId),
		"security_group_name": types.StringPointerValue(item.SecurityGroupName),
	})
	diags.Append(d...)
	return value, diags
}

func mappingSecurityGroups(ctx context.Context, item *api.Nic, diags *diag.Diagnostics) types.List {
	if item.SecurityGroups == nil {
		return types.ListNull(datasource_nic.SecurityGroupsValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.SecurityGroups, mappingSecurityGroup)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_nic.SecurityGroupsValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func deserializeReadNics(ctx context.Context, tf *datasource_nic.NicModel, diags *diag.Diagnostics) *api.ReadNicsParams {
	return &api.ReadNicsParams{
		Descriptions:                    utils.ConvertTfListToArrayOfString(ctx, tf.Descriptions, diags),
		IsSourceDestCheck:               tf.IsSourceDestCheck.ValueBoolPointer(),
		LinkNicDeleteOnVmDeletion:       tf.LinkNicDeleteOnVmDeletion.ValueBoolPointer(),
		LinkNicDeviceNumbers:            utils.ConvertTfListToArrayOfInt(ctx, tf.LinkNicDeviceNumbers, diags),
		LinkNicLinkNicIds:               utils.ConvertTfListToArrayOfString(ctx, tf.LinkNicLinkNicIds, diags),
		LinkNicStates:                   utils.ConvertTfListToArrayOfString(ctx, tf.LinkNicStates, diags),
		LinkNicVmIds:                    utils.ConvertTfListToArrayOfString(ctx, tf.LinkNicVmIds, diags),
		LinkPublicIpLinkPublicIpIds:     utils.ConvertTfListToArrayOfString(ctx, tf.LinkPublicIpLinkPublicIpIds, diags),
		LinkPublicIpPublicIpIds:         utils.ConvertTfListToArrayOfString(ctx, tf.LinkPublicIpPublicIpIds, diags),
		LinkPublicIpPublicIps:           utils.ConvertTfListToArrayOfString(ctx, tf.LinkPublicIpPublicIps, diags),
		MacAddresses:                    utils.ConvertTfListToArrayOfString(ctx, tf.MacAddresses, diags),
		PrivateDnsNames:                 utils.ConvertTfListToArrayOfString(ctx, tf.PrivateDnsNames, diags),
		PrivateIpsLinkPublicIpPublicIps: utils.ConvertTfListToArrayOfString(ctx, tf.PrivateIpsLinkPublicIpPublicIps, diags),
		PrivateIpsPrimaryIp:             tf.PrivateIpsPrimaryIp.ValueBoolPointer(),
		PrivateIpsPrivateIps:            utils.ConvertTfListToArrayOfString(ctx, tf.PrivateIpsPrivateIps, diags),
		SecurityGroupIds:                utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:              utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupNames, diags),
		States:                          utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		SubnetIds:                       utils.ConvertTfListToArrayOfString(ctx, tf.SubnetIds, diags),
		TagKeys:                         utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:                       utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:                            utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		VpcIds:                          utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		Ids:                             utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		AvailabilityZoneNames:           utils.ConvertTfListToArrayOfAzName(ctx, tf.AvailabilityZoneNames, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
