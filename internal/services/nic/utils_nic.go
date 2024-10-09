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

func privatesIpFromApi(ctx context.Context, elt numspot.PrivateIp, diags *diag.Diagnostics) PrivateIpsValue {
	var (
		linkPublicIpTf  LinkPublicIpValue
		linkPublicIpObj basetypes.ObjectValue
	)

	if elt.LinkPublicIp != nil {
		linkPublicIpTf = linkPublicIpFromApi(ctx, *elt.LinkPublicIp, diags)
		if diags.HasError() {
			return PrivateIpsValue{}
		}

		var diagnostics diag.Diagnostics
		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		diags.Append(diagnostics...)
		if diags.HasError() {
			return PrivateIpsValue{}
		}
	} else {
		linkPublicIpTf = NewLinkPublicIpValueNull()
		var diagnostics diag.Diagnostics
		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		diags.Append(diagnostics...)
		if diagnostics.HasError() {
			return PrivateIpsValue{}
		}
	}

	value, diagnostics := NewPrivateIpsValue(
		PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(elt.IsPrimary),
			"link_public_ip":   linkPublicIpObj,
			"private_dns_name": types.StringPointerValue(elt.PrivateDnsName),
			"private_ip":       types.StringPointerValue(elt.PrivateIp),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func securityGroupLightFromApi(ctx context.Context, elt numspot.SecurityGroupLight, diags *diag.Diagnostics) SecurityGroupsValue {
	value, diagnostics := NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func linkPublicIpFromApi(ctx context.Context, elt numspot.LinkPublicIp, diags *diag.Diagnostics) LinkPublicIpValue {
	value, diagnostics := NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":              types.StringPointerValue(elt.Id),
			"public_dns_name": types.StringPointerValue(elt.PublicDnsName),
			"public_ip":       types.StringPointerValue(elt.PublicIp),
			"public_ip_id":    types.StringPointerValue(elt.PublicIpId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func NicFromHttpToTf(ctx context.Context, http *numspot.Nic, diags *diag.Diagnostics) *NicModel {
	var (
		linkPublicIpTf LinkPublicIpValue
		linkNicTf      LinkNicValue
		tagsTf         types.List
	)
	// Private IPs
	privateIps := utils.GenericListToTfListValue(ctx, PrivateIpsValue{}, privatesIpFromApi, utils.GetPtrValue(http.PrivateIps), diags)
	if diags.HasError() {
		return nil
	}

	if http.SecurityGroups == nil {
		return nil
	}
	// Retrieve security groups id
	securityGroupIds := make([]string, 0, len(*http.SecurityGroups))
	for _, e := range *http.SecurityGroups {
		if e.SecurityGroupId != nil {
			securityGroupIds = append(securityGroupIds, *e.SecurityGroupId)
		}
	}

	// Security Group Ids
	securityGroupsIdTf := utils.StringListToTfListValue(ctx, securityGroupIds, diags)
	if diags.HasError() {
		return nil
	}

	// Security Groups
	securityGroupsTf := utils.GenericListToTfListValue(ctx, SecurityGroupsValue{}, securityGroupLightFromApi, utils.GetPtrValue(http.SecurityGroups), diags)
	if diags.HasError() {
		return nil
	}

	// Link Public Ip
	if http.LinkPublicIp != nil {
		linkPublicIpTf = linkPublicIpFromApi(ctx, *http.LinkPublicIp, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		linkPublicIpTf = NewLinkPublicIpValueNull()
	}

	// Link NIC
	if http.LinkNic != nil {
		linkNicTf = linkNICFromHTTP(ctx, *http.LinkNic, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		linkNicTf = NewLinkNicValueNull()
	}

	var macAddress *string
	if http.MacAddress != nil {
		lowerMacAddr := strings.ToLower(*http.MacAddress)
		macAddress = &lowerMacAddr
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &NicModel{
		LinkNic:              linkNicTf,
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
	}
}

func NicFromTfToCreateRequest(ctx context.Context, tf *NicModel, diags *diag.Diagnostics) numspot.CreateNicJSONRequestBody {
	var privateIpsPtr *[]numspot.PrivateIpLight
	var securityGroupIdsPtr *[]string

	if !(tf.PrivateIps.IsNull() || tf.PrivateIps.IsUnknown()) {
		privateIps := utils.TfListToGenericList(func(a PrivateIpsValue) numspot.PrivateIpLight {
			return numspot.PrivateIpLight{
				IsPrimary: a.IsPrimary.ValueBoolPointer(),
				PrivateIp: a.PrivateIp.ValueStringPointer(),
			}
		}, ctx, tf.PrivateIps, diags)
		privateIpsPtr = &privateIps
	}

	if !(tf.SecurityGroupIds.IsNull() || tf.SecurityGroupIds.IsUnknown()) {
		securityGroupIds := utils.TfStringListToStringList(ctx, tf.SecurityGroupIds, diags)
		securityGroupIdsPtr = &securityGroupIds
	}
	return numspot.CreateNicJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		PrivateIps:       privateIpsPtr,
		SecurityGroupIds: securityGroupIdsPtr,
		SubnetId:         tf.SubnetId.ValueString(),
	}
}

func NicsFromTfToAPIReadParams(ctx context.Context, tf NicsDataSourceModel, diags *diag.Diagnostics) numspot.ReadNicsParams {
	return numspot.ReadNicsParams{
		Descriptions:                    utils.TfStringListToStringPtrList(ctx, tf.Descriptions, diags),
		IsSourceDestCheck:               utils.FromTfBoolToBoolPtr(tf.IsSourceDestCheck),
		LinkNicDeleteOnVmDeletion:       utils.FromTfBoolToBoolPtr(tf.LinkNicDeleteOnVmDeletion),
		LinkNicDeviceNumbers:            utils.TFInt64ListToIntListPointer(ctx, tf.LinkNicDeviceNumbers, diags),
		LinkNicLinkNicIds:               utils.TfStringListToStringPtrList(ctx, tf.LinkNicLinkNicIds, diags),
		LinkNicStates:                   utils.TfStringListToStringPtrList(ctx, tf.LinkNicStates, diags),
		LinkNicVmIds:                    utils.TfStringListToStringPtrList(ctx, tf.LinkNicVmIds, diags),
		MacAddresses:                    utils.TfStringListToStringPtrList(ctx, tf.MacAddresses, diags),
		LinkPublicIpLinkPublicIpIds:     utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpLinkPublicIpIds, diags),
		LinkPublicIpPublicIpIds:         utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpPublicIpIds, diags),
		LinkPublicIpPublicIps:           utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpPublicIps, diags),
		PrivateDnsNames:                 utils.TfStringListToStringPtrList(ctx, tf.PrivateDnsNames, diags),
		PrivateIpsPrimaryIp:             utils.FromTfBoolToBoolPtr(tf.PrivateIpsPrimaryIp),
		PrivateIpsLinkPublicIpPublicIps: utils.TfStringListToStringPtrList(ctx, tf.PrivateIpsLinkPublicIpPublicIps, diags),
		PrivateIpsPrivateIps:            utils.TfStringListToStringPtrList(ctx, tf.PrivateIpsPrivateIps, diags),
		SecurityGroupIds:                utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:              utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames, diags),
		States:                          utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		SubnetIds:                       utils.TfStringListToStringPtrList(ctx, tf.SubnetIds, diags),
		VpcIds:                          utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		Ids:                             utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		AvailabilityZoneNames:           utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames, diags),
		Tags:                            utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		TagKeys:                         utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:                       utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
	}
}

func linkPublicIpFromHTTP(ctx context.Context, http numspot.LinkPublicIp, diags *diag.Diagnostics) LinkPublicIpValue {
	value, diagnostics := NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":              types.StringPointerValue(http.Id),
			"public_dns_name": types.StringPointerValue(http.PublicDnsName),
			"public_ip":       types.StringPointerValue(http.PublicIp),
			"public_ip_id":    types.StringPointerValue(http.PublicIpId),
		})
	diags.Append(diagnostics...)
	return value
}

func linkNICFromHTTP(ctx context.Context, http numspot.LinkNic, diags *diag.Diagnostics) LinkNicValue {
	value, diagnostics := NewLinkNicValue(
		LinkNicValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":                    types.StringPointerValue(http.Id),
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_number":         utils.FromIntPtrToTfInt64(http.DeviceNumber),
			"state":                 types.StringPointerValue(http.State),
			"vm_id":                 types.StringPointerValue(http.VmId),
		})
	diags.Append(diagnostics...)
	return value
}

func privateIpsFromHTTP(ctx context.Context, http numspot.PrivateIp, diags *diag.Diagnostics) PrivateIpsValue {
	linkPublicIp := linkPublicIpFromHTTP(ctx, utils.GetPtrValue(http.LinkPublicIp), diags)
	if diags.HasError() {
		return PrivateIpsValue{}
	}

	linkPublicIpObjectValue, diagnostics := linkPublicIp.ToObjectValue(ctx)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return PrivateIpsValue{}
	}

	value, diagnostics := NewPrivateIpsValue(
		PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(http.IsPrimary),
			"link_public_ip":   linkPublicIpObjectValue,
			"private_dns_name": types.StringPointerValue(http.PrivateDnsName),
			"private_ip":       types.StringPointerValue(http.PrivateIp),
		})
	diags.Append(diagnostics...)
	return value
}

func securityGroupsFromHTTP(ctx context.Context, http numspot.SecurityGroupLight, diags *diag.Diagnostics) SecurityGroupsValue {
	value, diagnostics := NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(http.SecurityGroupId),
			"security_group_name": types.StringPointerValue(http.SecurityGroupName),
		})
	diags.Append(diagnostics...)
	return value
}

func NicsFromHttpToTfDatasource(ctx context.Context, http *numspot.Nic, diags *diag.Diagnostics) *NicModelDatasource {
	if http == nil {
		return nil
	}

	var (
		tagsList     types.List
		linkNic      LinkNicValue
		linkPublicIp LinkPublicIpValue
	)

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	if http.LinkNic != nil {
		var diagnostics diag.Diagnostics
		linkNic, diagnostics = NewLinkNicValue(LinkNicValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"delete_on_vm_deletion": types.BoolPointerValue(http.LinkNic.DeleteOnVmDeletion),
				"device_number":         utils.FromIntPtrToTfInt64(http.LinkNic.DeviceNumber),
				"id":                    types.StringPointerValue(http.LinkNic.Id),
				"state":                 types.StringPointerValue(http.LinkNic.State),
				"vm_id":                 types.StringPointerValue(http.LinkNic.VmId),
			})
		diags.Append(diagnostics...)
		if diags.HasError() {
			return nil
		}
	}

	if http.LinkPublicIp != nil {
		var diagnostics diag.Diagnostics
		linkPublicIp, diagnostics = NewLinkPublicIpValue(LinkPublicIpValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"id":              types.StringPointerValue(http.LinkPublicIp.Id),
				"public_dns_name": types.StringPointerValue(http.LinkPublicIp.PublicDnsName),
				"public_ip":       types.StringPointerValue(http.LinkPublicIp.PublicIp),
				"public_ip_id":    types.StringPointerValue(http.LinkPublicIp.PublicIpId),
			})
		diags.Append(diagnostics...)
		if diags.HasError() {
			return nil
		}
	}

	privateIps := utils.GenericListToTfListValue(ctx, PrivateIpsValue{}, privateIpsFromHTTP, utils.GetPtrValue(http.PrivateIps), diags)
	if diags.HasError() {
		return nil
	}

	securityGroups := utils.GenericListToTfListValue(ctx, SecurityGroupsValue{}, securityGroupsFromHTTP, utils.GetPtrValue(http.SecurityGroups), diags)
	if diags.HasError() {
		return nil
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
	}
}

func NicFromTfToUpdaterequest(ctx context.Context, tf *NicModel, diags *diag.Diagnostics) numspot.UpdateNicJSONRequestBody {
	linkNic := numspot.LinkNicToUpdate{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.LinkNic.DeleteOnVmDeletion),
		LinkNicId:          utils.FromTfStringToStringPtr(tf.LinkNic.Id),
	}

	return numspot.UpdateNicJSONRequestBody{
		SecurityGroupIds: utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		Description:      utils.FromTfStringToStringPtr(tf.Description),
		LinkNic:          &linkNic,
	}
}

func NicFromTfToLinkNICRequest(tf *NicModel) numspot.LinkNicJSONRequestBody {
	return numspot.LinkNicJSONRequestBody{
		DeviceNumber: utils.FromTfInt64ToInt(tf.LinkNic.DeviceNumber),
		VmId:         tf.LinkNic.VmId.ValueString(),
	}
}
