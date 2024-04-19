package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_nic"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nic"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func privatesIpFromApi(ctx context.Context, elt iaas.PrivateIp) (resource_nic.PrivateIpsValue, diag.Diagnostics) {
	var (
		linkPublicIpTf  resource_nic.LinkPublicIpValue
		linkPublicIpObj basetypes.ObjectValue
		diagnostics     diag.Diagnostics
	)

	if elt.LinkPublicIp != nil {
		linkPublicIpTf, diagnostics = linkPublicIpFromApi(ctx, *elt.LinkPublicIp)
		if diagnostics.HasError() {
			return resource_nic.PrivateIpsValue{}, diagnostics
		}

		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		if diagnostics.HasError() {
			return resource_nic.PrivateIpsValue{}, diagnostics
		}
	} else {
		linkPublicIpTf = resource_nic.NewLinkPublicIpValueNull()
		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		if diagnostics.HasError() {
			return resource_nic.PrivateIpsValue{}, diagnostics
		}
	}

	return resource_nic.NewPrivateIpsValue(
		resource_nic.PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(elt.IsPrimary),
			"link_public_ip":   linkPublicIpObj,
			"private_dns_name": types.StringPointerValue(elt.PrivateDnsName),
			"private_ip":       types.StringPointerValue(elt.PrivateIp),
		},
	)
}

func securityGroupLightFromApi(ctx context.Context, elt iaas.SecurityGroupLight) (resource_nic.SecurityGroupsValue, diag.Diagnostics) {
	return resource_nic.NewSecurityGroupsValue(
		resource_nic.SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		},
	)
}

func linkPublicIpFromApi(ctx context.Context, elt iaas.LinkPublicIp) (resource_nic.LinkPublicIpValue, diag.Diagnostics) {
	return resource_nic.NewLinkPublicIpValue(
		resource_nic.LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":              types.StringPointerValue(elt.Id),
			"public_dns_name": types.StringPointerValue(elt.PublicDnsName),
			"public_ip":       types.StringPointerValue(elt.PublicIp),
			"public_ip_id":    types.StringPointerValue(elt.PublicIpId),
		},
	)
}

func NicFromHttpToTf(ctx context.Context, http *iaas.Nic) (*resource_nic.NicModel, diag.Diagnostics) {
	var linkPublicIpTf resource_nic.LinkPublicIpValue
	// Private IPs
	privateIps, diagnostics := utils.GenericListToTfListValue(ctx, resource_nic.PrivateIpsValue{}, privatesIpFromApi, utils.GetPtrValue(http.PrivateIps))
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Retrieve security groups id
	securityGroupIds := make([]string, 0, len(*http.SecurityGroups))
	for _, e := range *http.SecurityGroups {
		securityGroupIds = append(securityGroupIds, *e.SecurityGroupId)
	}

	// Security Group Ids
	securityGroupsIdTf, diagnostics := utils.StringListToTfListValue(ctx, securityGroupIds)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Security Groups
	securityGroupsTf, diagnostics := utils.GenericListToTfListValue(ctx, resource_nic.SecurityGroupsValue{}, securityGroupLightFromApi, utils.GetPtrValue(http.SecurityGroups))
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Link Public Ip
	if http.LinkPublicIp != nil {
		linkPublicIpTf, diagnostics = linkPublicIpFromApi(ctx, *http.LinkPublicIp)
		if diagnostics.HasError() {
			return nil, diagnostics
		}
	} else {
		linkPublicIpTf = resource_nic.NewLinkPublicIpValueNull()
	}

	var macAddress *string
	if http.MacAddress != nil {
		lowerMacAddr := strings.ToLower(*http.MacAddress)
		macAddress = &lowerMacAddr
	}

	return &resource_nic.NicModel{
		Description:          types.StringPointerValue(http.Description),
		Id:                   types.StringPointerValue(http.Id),
		IsSourceDestChecked:  types.BoolPointerValue(http.IsSourceDestChecked),
		LinkPublicIp:         linkPublicIpTf,
		MacAddress:           types.StringPointerValue(macAddress),
		VpcId:                types.StringPointerValue(http.VpcId),
		PrivateDnsName:       types.StringPointerValue(http.PrivateDnsName),
		PrivateIps:           privateIps,
		SecurityGroupIds:     securityGroupsIdTf,
		SecurityGroups:       securityGroupsTf,
		State:                types.StringPointerValue(http.State),
		SubnetId:             types.StringPointerValue(http.SubnetId),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
	}, diagnostics
}

func NicFromTfToCreateRequest(ctx context.Context, tf *resource_nic.NicModel) iaas.CreateNicJSONRequestBody {
	privateIps := utils.TfListToGenericList(func(a resource_nic.PrivateIpsValue) iaas.PrivateIpLight {
		return iaas.PrivateIpLight{
			IsPrimary: a.IsPrimary.ValueBoolPointer(),
			PrivateIp: a.PrivateIp.ValueStringPointer(),
		}
	}, ctx, tf.PrivateIps)
	securityGroupIds := utils.TfStringListToStringList(ctx, tf.SecurityGroupIds)

	return iaas.CreateNicJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		PrivateIps:       &privateIps,
		SecurityGroupIds: &securityGroupIds,
		SubnetId:         tf.SubnetId.ValueString(),
	}
}

func NicsFromTfToAPIReadParams(ctx context.Context, tf NicsDataSourceModel) iaas.ReadNicsParams {
	return iaas.ReadNicsParams{
		Descriptions: utils.TfStringListToStringPtrList(ctx, tf.Descriptions),

		IsSourceDestCheck:               utils.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
		LinkNicDeleteOnVmDeletion:       utils.FromTfBoolToBoolPtr(tf.LinkNicDeleteOnVMDeletion),
		LinkNicDeviceNumbers:            utils.TFInt64ListToIntListPointer(ctx, tf.LinkNicDeviceNumbers),
		LinkNicLinkNicIds:               utils.TfStringListToStringPtrList(ctx, tf.LinkNicIds),
		LinkNicStates:                   utils.TfStringListToStringPtrList(ctx, tf.LinkNicStates),
		LinkNicVmIds:                    utils.TfStringListToStringPtrList(ctx, tf.LinkNicVMIds),
		MacAddresses:                    utils.TfStringListToStringPtrList(ctx, tf.MacAddresses),
		LinkPublicIpLinkPublicIpIds:     utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpLinkPublicIpIds),
		LinkPublicIpPublicIpIds:         utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpPublicIpIds),
		LinkPublicIpPublicIps:           utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpPublicIps),
		PrivateDnsNames:                 utils.TfStringListToStringPtrList(ctx, tf.PrivateDnsNames),
		PrivateIpsPrimaryIp:             utils.FromTfBoolToBoolPtr(tf.PrivateIpIsPrimary),
		PrivateIpsLinkPublicIpPublicIps: utils.TfStringListToStringPtrList(ctx, tf.PrivateIpLinkPublicIpPublicIps),
		PrivateIpsPrivateIps:            utils.TfStringListToStringPtrList(ctx, tf.PrivateIpPrivateIps),
		SecurityGroupIds:                utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		SecurityGroupNames:              utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames),
		States:                          utils.TfStringListToStringPtrList(ctx, tf.States),
		SubnetIds:                       utils.TfStringListToStringPtrList(ctx, tf.SubnetIds),
		VpcIds:                          utils.TfStringListToStringPtrList(ctx, tf.VpcIds),
		Ids:                             utils.TfStringListToStringPtrList(ctx, tf.IDs),
		AvailabilityZoneNames:           utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames),
		Tags:                            utils.TfStringListToStringPtrList(ctx, tf.Tags),
		TagKeys:                         utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:                       utils.TfStringListToStringPtrList(ctx, tf.TagValues),
	}
}

func linkPublicIpFromHTTP(ctx context.Context, http iaas.LinkPublicIp) (datasource_nic.LinkPublicIpValue, diag.Diagnostics) {
	return datasource_nic.NewLinkPublicIpValue(
		resource_nic.LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":              types.StringPointerValue(http.Id),
			"public_dns_name": types.StringPointerValue(http.PublicDnsName),
			"public_ip":       types.StringPointerValue(http.PublicIp),
			"public_ip_id":    types.StringPointerValue(http.PublicIpId),
		})
}

func privateIpsFromHTTP(ctx context.Context, http iaas.PrivateIp) (datasource_nic.PrivateIpsValue, diag.Diagnostics) {
	linkPublicIp, diags := linkPublicIpFromHTTP(ctx, utils.GetPtrValue(http.LinkPublicIp))
	if diags.HasError() {
		return datasource_nic.PrivateIpsValue{}, diags
	}

	linkPublicIpObjectValue, diags := linkPublicIp.ToObjectValue(ctx)
	if diags.HasError() {
		return datasource_nic.PrivateIpsValue{}, diags
	}

	return datasource_nic.NewPrivateIpsValue(
		datasource_nic.PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(http.IsPrimary),
			"link_public_ip":   linkPublicIpObjectValue,
			"private_dns_name": types.StringPointerValue(http.PrivateDnsName),
			"private_ip":       types.StringPointerValue(http.PrivateIp),
		})
}

func securityGroupsFromHTTP(ctx context.Context, http iaas.SecurityGroupLight) (datasource_nic.SecurityGroupsValue, diag.Diagnostics) {
	return datasource_nic.NewSecurityGroupsValue(
		datasource_nic.SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(http.SecurityGroupId),
			"security_group_name": types.StringPointerValue(http.SecurityGroupName),
		})
}

func NicsFromHttpToTfDatasource(ctx context.Context, http *iaas.Nic) (*datasource_nic.NicModel, diag.Diagnostics) {
	if http == nil {
		return nil, nil
	}

	var (
		tagsList     types.List
		linkNic      datasource_nic.LinkNicValue
		linkPublicIp datasource_nic.LinkPublicIpValue
		diag         diag.Diagnostics
	)

	if http.Tags != nil {
		tagsList, diag = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diag.HasError() {
			return nil, diag
		}
	}

	if http.LinkNic != nil {
		linkNic, diag = datasource_nic.NewLinkNicValue(datasource_nic.LinkNicValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"delete_on_vm_deletion": types.BoolPointerValue(http.LinkNic.DeleteOnVmDeletion),
				"device_number":         utils.FromIntPtrToTfInt64(http.LinkNic.DeviceNumber),
				"id":                    types.StringPointerValue(http.LinkNic.Id),
				"state":                 types.StringPointerValue(http.LinkNic.State),
				"vm_id":                 types.StringPointerValue(http.LinkNic.VmId),
			})
		if diag.HasError() {
			return nil, diag
		}
	}

	if http.LinkPublicIp != nil {
		linkPublicIp, diag = datasource_nic.NewLinkPublicIpValue(datasource_nic.LinkPublicIpValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"id":              types.StringPointerValue(http.LinkPublicIp.Id),
				"public_dns_name": types.StringPointerValue(http.LinkPublicIp.PublicDnsName),
				"public_ip":       types.StringPointerValue(http.LinkPublicIp.PublicIp),
				"public_ip_id":    types.StringPointerValue(http.LinkPublicIp.PublicIpId),
			})
		if diag.HasError() {
			return nil, diag
		}
	}

	privateIps, diags := utils.GenericListToTfListValue(ctx, datasource_nic.PrivateIpsValue{}, privateIpsFromHTTP, utils.GetPtrValue(http.PrivateIps))
	if diags.HasError() {
		return nil, diags
	}

	securityGroups, diags := utils.GenericListToTfListValue(ctx, datasource_nic.SecurityGroupsValue{}, securityGroupsFromHTTP, utils.GetPtrValue(http.SecurityGroups))
	if diags.HasError() {
		return nil, diags
	}

	return &datasource_nic.NicModel{
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Description:          types.StringPointerValue(http.Description),
		Id:                   types.StringPointerValue(http.Id),
		IsSourceDestChecked:  types.BoolPointerValue(http.IsSourceDestChecked),
		LinkNic:              linkNic,
		LinkPublicIp:         linkPublicIp,
		MacAddress:           types.StringPointerValue(http.MacAddress),
		PrivateDnsName:       types.StringPointerValue(http.PrivateDnsName),
		PrivateIps:           privateIps,
		SecurityGroups:       securityGroups,
		State:                types.StringPointerValue(http.State),
		SubnetId:             types.StringPointerValue(http.SubnetId),
		VpcId:                types.StringPointerValue(http.VpcId),
		Tags:                 tagsList,
	}, nil
}
