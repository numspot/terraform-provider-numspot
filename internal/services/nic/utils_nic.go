package nic

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func privatesIpFromApi(ctx context.Context, elt numspot.PrivateIp) (PrivateIpsValue, diag.Diagnostics) {
	var (
		linkPublicIpTf  LinkPublicIpValue
		linkPublicIpObj basetypes.ObjectValue
		diagnostics     diag.Diagnostics
	)

	if elt.LinkPublicIp != nil {
		linkPublicIpTf, diagnostics = linkPublicIpFromApi(ctx, *elt.LinkPublicIp)
		if diagnostics.HasError() {
			return PrivateIpsValue{}, diagnostics
		}

		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		if diagnostics.HasError() {
			return PrivateIpsValue{}, diagnostics
		}
	} else {
		linkPublicIpTf = NewLinkPublicIpValueNull()
		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		if diagnostics.HasError() {
			return PrivateIpsValue{}, diagnostics
		}
	}

	return NewPrivateIpsValue(
		PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(elt.IsPrimary),
			"link_public_ip":   linkPublicIpObj,
			"private_dns_name": types.StringPointerValue(elt.PrivateDnsName),
			"private_ip":       types.StringPointerValue(elt.PrivateIp),
		},
	)
}

func securityGroupLightFromApi(ctx context.Context, elt numspot.SecurityGroupLight) (SecurityGroupsValue, diag.Diagnostics) {
	return NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		},
	)
}

func linkPublicIpFromApi(ctx context.Context, elt numspot.LinkPublicIp) (LinkPublicIpValue, diag.Diagnostics) {
	return NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":              types.StringPointerValue(elt.Id),
			"public_dns_name": types.StringPointerValue(elt.PublicDnsName),
			"public_ip":       types.StringPointerValue(elt.PublicIp),
			"public_ip_id":    types.StringPointerValue(elt.PublicIpId),
		},
	)
}

func NicFromHttpToTf(ctx context.Context, http *numspot.Nic) (*NicModel, diag.Diagnostics) {
	var (
		linkPublicIpTf LinkPublicIpValue
		tagsTf         types.List
		diags          diag.Diagnostics
	)
	// Private IPs
	privateIps, diags := utils.GenericListToTfListValue(ctx, PrivateIpsValue{}, privatesIpFromApi, utils.GetPtrValue(http.PrivateIps))
	if diags.HasError() {
		return nil, diags
	}

	if http.SecurityGroups == nil {
		return nil, diags
	}
	// Retrieve security groups id
	securityGroupIds := make([]string, 0, len(*http.SecurityGroups))
	for _, e := range *http.SecurityGroups {
		if e.SecurityGroupId != nil {
			securityGroupIds = append(securityGroupIds, *e.SecurityGroupId)
		}
	}

	// Security Group Ids
	securityGroupsIdTf, diags := utils.StringListToTfListValue(ctx, securityGroupIds)
	if diags.HasError() {
		return nil, diags
	}

	// Security Groups
	securityGroupsTf, diags := utils.GenericListToTfListValue(ctx, SecurityGroupsValue{}, securityGroupLightFromApi, utils.GetPtrValue(http.SecurityGroups))
	if diags.HasError() {
		return nil, diags
	}

	// Link Public Ip
	if http.LinkPublicIp != nil {
		linkPublicIpTf, diags = linkPublicIpFromApi(ctx, *http.LinkPublicIp)
		if diags.HasError() {
			return nil, diags
		}
	} else {
		linkPublicIpTf = NewLinkPublicIpValueNull()
	}

	var macAddress *string
	if http.MacAddress != nil {
		lowerMacAddr := strings.ToLower(*http.MacAddress)
		macAddress = &lowerMacAddr
	}

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &NicModel{
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
		Tags:                 tagsTf,
	}, diags
}

func NicFromTfToCreateRequest(ctx context.Context, tf *NicModel) numspot.CreateNicJSONRequestBody {
	var privateIpsPtr *[]numspot.PrivateIpLight
	var securityGroupIdsPtr *[]string

	if !(tf.PrivateIps.IsNull() || tf.PrivateIps.IsUnknown()) {
		privateIps := utils.TfListToGenericList(func(a PrivateIpsValue) numspot.PrivateIpLight {
			return numspot.PrivateIpLight{
				IsPrimary: a.IsPrimary.ValueBoolPointer(),
				PrivateIp: a.PrivateIp.ValueStringPointer(),
			}
		}, ctx, tf.PrivateIps)
		privateIpsPtr = &privateIps
	}

	if !(tf.SecurityGroupIds.IsNull() || tf.SecurityGroupIds.IsUnknown()) {
		securityGroupIds := utils.TfStringListToStringList(ctx, tf.SecurityGroupIds)
		securityGroupIdsPtr = &securityGroupIds
	}
	return numspot.CreateNicJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		PrivateIps:       privateIpsPtr,
		SecurityGroupIds: securityGroupIdsPtr,
		SubnetId:         tf.SubnetId.ValueString(),
	}
}

func NicsFromTfToAPIReadParams(ctx context.Context, tf NicsDataSourceModel) numspot.ReadNicsParams {
	return numspot.ReadNicsParams{
		Descriptions: utils.TfStringListToStringPtrList(ctx, tf.Descriptions),

		IsSourceDestCheck:               utils.FromTfBoolToBoolPtr(tf.IsSourceDestCheck),
		LinkNicDeleteOnVmDeletion:       utils.FromTfBoolToBoolPtr(tf.LinkNicDeleteOnVmDeletion),
		LinkNicDeviceNumbers:            utils.TFInt64ListToIntListPointer(ctx, tf.LinkNicDeviceNumbers),
		LinkNicLinkNicIds:               utils.TfStringListToStringPtrList(ctx, tf.LinkNicLinkNicIds),
		LinkNicStates:                   utils.TfStringListToStringPtrList(ctx, tf.LinkNicStates),
		LinkNicVmIds:                    utils.TfStringListToStringPtrList(ctx, tf.LinkNicVmIds),
		MacAddresses:                    utils.TfStringListToStringPtrList(ctx, tf.MacAddresses),
		LinkPublicIpLinkPublicIpIds:     utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpLinkPublicIpIds),
		LinkPublicIpPublicIpIds:         utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpPublicIpIds),
		LinkPublicIpPublicIps:           utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpPublicIps),
		PrivateDnsNames:                 utils.TfStringListToStringPtrList(ctx, tf.PrivateDnsNames),
		PrivateIpsPrimaryIp:             utils.FromTfBoolToBoolPtr(tf.PrivateIpsPrimaryIp),
		PrivateIpsLinkPublicIpPublicIps: utils.TfStringListToStringPtrList(ctx, tf.PrivateIpsLinkPublicIpPublicIps),
		PrivateIpsPrivateIps:            utils.TfStringListToStringPtrList(ctx, tf.PrivateIpsPrivateIps),
		SecurityGroupIds:                utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		SecurityGroupNames:              utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames),
		States:                          utils.TfStringListToStringPtrList(ctx, tf.States),
		SubnetIds:                       utils.TfStringListToStringPtrList(ctx, tf.SubnetIds),
		VpcIds:                          utils.TfStringListToStringPtrList(ctx, tf.VpcIds),
		Ids:                             utils.TfStringListToStringPtrList(ctx, tf.Ids),
		AvailabilityZoneNames:           utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames),
		Tags:                            utils.TfStringListToStringPtrList(ctx, tf.Tags),
		TagKeys:                         utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:                       utils.TfStringListToStringPtrList(ctx, tf.TagValues),
	}
}

func linkPublicIpFromHTTP(ctx context.Context, http numspot.LinkPublicIp) (LinkPublicIpValue, diag.Diagnostics) {
	return NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":              types.StringPointerValue(http.Id),
			"public_dns_name": types.StringPointerValue(http.PublicDnsName),
			"public_ip":       types.StringPointerValue(http.PublicIp),
			"public_ip_id":    types.StringPointerValue(http.PublicIpId),
		})
}

func privateIpsFromHTTP(ctx context.Context, http numspot.PrivateIp) (PrivateIpsValue, diag.Diagnostics) {
	linkPublicIp, diags := linkPublicIpFromHTTP(ctx, utils.GetPtrValue(http.LinkPublicIp))
	if diags.HasError() {
		return PrivateIpsValue{}, diags
	}

	linkPublicIpObjectValue, diags := linkPublicIp.ToObjectValue(ctx)
	if diags.HasError() {
		return PrivateIpsValue{}, diags
	}

	return NewPrivateIpsValue(
		PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(http.IsPrimary),
			"link_public_ip":   linkPublicIpObjectValue,
			"private_dns_name": types.StringPointerValue(http.PrivateDnsName),
			"private_ip":       types.StringPointerValue(http.PrivateIp),
		})
}

func securityGroupsFromHTTP(ctx context.Context, http numspot.SecurityGroupLight) (SecurityGroupsValue, diag.Diagnostics) {
	return NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(http.SecurityGroupId),
			"security_group_name": types.StringPointerValue(http.SecurityGroupName),
		})
}

func NicsFromHttpToTfDatasource(ctx context.Context, http *numspot.Nic) (*NicModelDatasource, diag.Diagnostics) {
	if http == nil {
		return nil, nil
	}

	var (
		tagsList     types.List
		linkNic      LinkNicValue
		linkPublicIp LinkPublicIpValue
		diag         diag.Diagnostics
	)

	if http.Tags != nil {
		tagsList, diag = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diag.HasError() {
			return nil, diag
		}
	}

	if http.LinkNic != nil {
		linkNic, diag = NewLinkNicValue(LinkNicValue{}.AttributeTypes(ctx),
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
		linkPublicIp, diag = NewLinkPublicIpValue(LinkPublicIpValue{}.AttributeTypes(ctx),
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

	privateIps, diags := utils.GenericListToTfListValue(ctx, PrivateIpsValue{}, privateIpsFromHTTP, utils.GetPtrValue(http.PrivateIps))
	if diags.HasError() {
		return nil, diags
	}

	securityGroups, diags := utils.GenericListToTfListValue(ctx, SecurityGroupsValue{}, securityGroupsFromHTTP, utils.GetPtrValue(http.SecurityGroups))
	if diags.HasError() {
		return nil, diags
	}

	return &NicModelDatasource{
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

func NicFromTfToUpdaterequest(ctx context.Context, tf *NicModel, diagnostics *diag.Diagnostics) numspot.UpdateNicJSONRequestBody {
	linkNic := numspot.LinkNicToUpdate{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.LinkNic.DeleteOnVmDeletion),
		LinkNicId:          utils.FromTfStringToStringPtr(tf.LinkNic.Id),
	}

	return numspot.UpdateNicJSONRequestBody{
		SecurityGroupIds: utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		Description:      utils.FromTfStringToStringPtr(tf.Description),
		LinkNic:          &linkNic,
	}
}
