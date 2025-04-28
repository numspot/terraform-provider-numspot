package nic

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/nic/datasource_nic"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &nicsDataSource{}

func (d *nicsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

func NewNicsDataSource() datasource.DataSource {
	return &nicsDataSource{}
}

type nicsDataSource struct {
	provider *client.NumSpotSDK
}

func (d *nicsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nics"
}

func (d *nicsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_nic.NicDataSourceSchema(ctx)
}

func (d *nicsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var plan, state datasource_nic.NicModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeReadParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	nics, err := core.ReadNicsWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read nic with params", err.Error())
		return
	}

	objectItems := utils.SerializeDatasourceItemsWithDiags(ctx, *nics, &response.Diagnostics, mappingItemsValue)
	if response.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeReadParams(ctx context.Context, tf datasource_nic.NicModel, diags *diag.Diagnostics) api.ReadNicsParams {
	return api.ReadNicsParams{
		Descriptions:                    utils.ConvertTfListToArrayOfString(ctx, tf.Descriptions, diags),
		IsSourceDestCheck:               tf.IsSourceDestCheck.ValueBoolPointer(),
		LinkNicDeleteOnVmDeletion:       tf.LinkNicDeleteOnVmDeletion.ValueBoolPointer(),
		LinkNicDeviceNumbers:            utils.ConvertTfListToArrayOfInt(ctx, tf.LinkNicDeviceNumbers, diags),
		LinkNicLinkNicIds:               utils.ConvertTfListToArrayOfString(ctx, tf.LinkNicLinkNicIds, diags),
		LinkNicStates:                   utils.ConvertTfListToArrayOfString(ctx, tf.LinkNicStates, diags),
		LinkNicVmIds:                    utils.ConvertTfListToArrayOfString(ctx, tf.LinkNicVmIds, diags),
		MacAddresses:                    utils.ConvertTfListToArrayOfString(ctx, tf.MacAddresses, diags),
		LinkPublicIpLinkPublicIpIds:     utils.ConvertTfListToArrayOfString(ctx, tf.LinkPublicIpLinkPublicIpIds, diags),
		LinkPublicIpPublicIpIds:         utils.ConvertTfListToArrayOfString(ctx, tf.LinkPublicIpPublicIpIds, diags),
		LinkPublicIpPublicIps:           utils.ConvertTfListToArrayOfString(ctx, tf.LinkPublicIpPublicIps, diags),
		PrivateDnsNames:                 utils.ConvertTfListToArrayOfString(ctx, tf.PrivateDnsNames, diags),
		PrivateIpsPrimaryIp:             tf.PrivateIpsPrimaryIp.ValueBoolPointer(),
		PrivateIpsLinkPublicIpPublicIps: utils.ConvertTfListToArrayOfString(ctx, tf.PrivateIpsLinkPublicIpPublicIps, diags),
		PrivateIpsPrivateIps:            utils.ConvertTfListToArrayOfString(ctx, tf.PrivateIpsPrivateIps, diags),
		SecurityGroupIds:                utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:              utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupNames, diags),
		States:                          utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		SubnetIds:                       utils.ConvertTfListToArrayOfString(ctx, tf.SubnetIds, diags),
		VpcIds:                          utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		Ids:                             utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		AvailabilityZoneNames:           utils.ConvertTfListToArrayOfAzName(ctx, tf.AvailabilityZoneNames, diags),
		Tags:                            utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		TagKeys:                         utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:                       utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
	}
}

func mappingItemsValue(ctx context.Context, nic api.Nic, diags *diag.Diagnostics) (datasource_nic.ItemsValue, diag.Diagnostics) {
	var serializeDiags diag.Diagnostics
	var linkNic basetypes.ObjectValue
	var linkPublicIp basetypes.ObjectValue

	tagsList := types.ListNull(datasource_nic.ItemsValue{}.Type(ctx))
	securityGroupsList := types.List{}
	privateIpsList := types.List{}

	if nic.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *nic.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_nic.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_nic.ItemsValue{}, serializeDiags
		}
	}

	if nic.SecurityGroups != nil {
		securityGroupsList, serializeDiags = mappingSecurityGroups(ctx, nic, diags)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	if nic.PrivateIps != nil {
		privateIpsList, serializeDiags = mappingPrivateIps(ctx, nic, diags)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	if nic.LinkNic != nil {
		linkNicValue, linkNicDiags := mappingLinkNic(ctx, nic.LinkNic, diags)
		if linkNicDiags.HasError() {
			diags.Append(linkNicDiags...)
		}
		linkNic, serializeDiags = linkNicValue.ToObjectValue(ctx)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		linkNic, serializeDiags = datasource_nic.NewLinkNicValueNull().ToObjectValue(ctx)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	if nic.LinkPublicIp != nil {
		linkPublicIpValue, linkPublicIpDiags := mappingLinkPublicIp(ctx, nic.LinkPublicIp, diags)
		if linkPublicIpDiags.HasError() {
			diags.Append(linkPublicIpDiags...)
		}
		linkPublicIp, serializeDiags = linkPublicIpValue.ToObjectValue(ctx)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		linkPublicIp, serializeDiags = datasource_nic.NewLinkPublicIpValueNull().ToObjectValue(ctx)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	return datasource_nic.NewItemsValue(datasource_nic.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(nic.AvailabilityZoneName)),
		"description":            types.StringValue(utils.ConvertStringPtrToString(nic.Description)),
		"id":                     types.StringValue(utils.ConvertStringPtrToString(nic.Id)),
		"is_source_dest_checked": types.BoolPointerValue(nic.IsSourceDestChecked),
		"link_nic":               linkNic,
		"link_public_ip":         linkPublicIp,
		"mac_address":            types.StringValue(utils.ConvertStringPtrToString(nic.MacAddress)),
		"private_dns_name":       types.StringValue(utils.ConvertStringPtrToString(nic.PrivateDnsName)),
		"private_ips":            privateIpsList,
		"security_groups":        securityGroupsList,
		"state":                  types.StringValue(utils.ConvertStringPtrToString(nic.State)),
		"subnet_id":              types.StringValue(utils.ConvertStringPtrToString(nic.SubnetId)),
		"tags":                   tagsList,
		"vpc_id":                 types.StringValue(utils.ConvertStringPtrToString(nic.VpcId)),
	})
}

func mappingLinkNic(ctx context.Context, linkNic *api.LinkNic, diags *diag.Diagnostics) (datasource_nic.LinkNicValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_nic.NewLinkNicValue(datasource_nic.LinkNicValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(linkNic.DeleteOnVmDeletion),
		"device_number":         types.Int64Value(utils.ConvertIntPtrToInt64(linkNic.DeviceNumber)),
		"id":                    types.StringPointerValue(linkNic.Id),
		"state":                 types.StringPointerValue(linkNic.State),
		"vm_id":                 types.StringPointerValue(linkNic.VmId),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingPrivateIps(ctx context.Context, nic api.Nic, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	var mappingDiags diag.Diagnostics
	var linkPublicIpPrivateIp basetypes.ObjectValue

	lp := len(*nic.PrivateIps)
	elementValue := make([]datasource_nic.PrivateIpsValue, lp)

	for y, privateIp := range *nic.PrivateIps {

		if privateIp.LinkPublicIp != nil {
			linkPublicValue, serializeLinkPublicDiags := mappingLinkPublicIp(ctx, privateIp.LinkPublicIp, diags)
			if serializeLinkPublicDiags.HasError() {
				diags.Append(serializeLinkPublicDiags...)
			}
			linkPublicIpPrivateIp, mappingDiags = linkPublicValue.ToObjectValue(ctx)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		} else {
			linkPublicIpPrivateIp, mappingDiags = datasource_nic.NewLinkPublicIpPrivateIpValueNull().ToObjectValue(ctx)
			if mappingDiags.HasError() {
				diags.Append(mappingDiags...)
			}
		}

		elementValue[y], *diags = datasource_nic.NewPrivateIpsValue(datasource_nic.PrivateIpsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"is_primary":                types.BoolPointerValue(privateIp.IsPrimary),
			"private_dns_name":          types.StringPointerValue(privateIp.PrivateDnsName),
			"private_ip":                types.StringPointerValue(privateIp.PrivateIp),
			"link_public_ip_private_ip": linkPublicIpPrivateIp,
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_nic.PrivateIpsValue).Type(ctx), elementValue)
}

func mappingLinkPublicIp(ctx context.Context, linkPublicIp *api.LinkPublicIp, diags *diag.Diagnostics) (datasource_nic.LinkPublicIpPrivateIpValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_nic.NewLinkPublicIpPrivateIpValue(datasource_nic.LinkPublicIpPrivateIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":              types.StringPointerValue(linkPublicIp.PublicIpId),
		"public_dns_name": types.StringPointerValue(linkPublicIp.PublicDnsName),
		"public_ip":       types.StringPointerValue(linkPublicIp.PublicIp),
		"public_ip_id":    types.StringPointerValue(linkPublicIp.PublicIpId),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingSecurityGroups(ctx context.Context, nic api.Nic, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	ls := len(*nic.SecurityGroups)
	elementValue := make([]datasource_nic.SecurityGroupsValue, ls)
	for y, securityGroup := range *nic.SecurityGroups {
		elementValue[y], *diags = datasource_nic.NewSecurityGroupsValue(datasource_nic.SecurityGroupsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(securityGroup.SecurityGroupId),
			"security_group_name": types.StringPointerValue(securityGroup.SecurityGroupName),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_nic.SecurityGroupsValue).Type(ctx), elementValue)
}

func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_nic.TagsValue, diag.Diagnostics) {
	return datasource_nic.NewTagsValue(datasource_nic.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}
