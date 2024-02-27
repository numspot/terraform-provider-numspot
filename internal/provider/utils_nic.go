package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nic"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func privatesIpFromApi(ctx context.Context, elt api.PrivateIp) (resource_nic.PrivateIpsValue, diag.Diagnostics) {
	linkPublicIpTf, diagnostics := linkPublicIpFromApi(ctx, *elt.LinkPublicIp)
	if diagnostics.HasError() {
		return resource_nic.PrivateIpsValue{}, diagnostics
	}

	return resource_nic.NewPrivateIpsValue(
		resource_nic.PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(elt.IsPrimary),
			"link_public_ip":   linkPublicIpTf,
			"private_dns_name": types.StringPointerValue(elt.PrivateDnsName),
			"private_ip":       types.StringPointerValue(elt.PrivateIp),
		},
	)
}

func securityGroupLightFromApi(ctx context.Context, elt api.SecurityGroupLight) (resource_nic.SecurityGroupsValue, diag.Diagnostics) {
	return resource_nic.NewSecurityGroupsValue(
		resource_nic.SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		},
	)
}

func linkPublicIpFromApi(ctx context.Context, elt api.LinkPublicIp) (resource_nic.LinkPublicIpValue, diag.Diagnostics) {
	return resource_nic.NewLinkPublicIpValue(
		resource_nic.LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":                   types.StringPointerValue(elt.Id),
			"public_dns_name":      types.StringPointerValue(elt.PublicDnsName),
			"public_ip":            types.StringPointerValue(elt.PublicIp),
			"public_ip_account_id": types.StringPointerValue(elt.PublicIpAccountId),
			"public_ip_id":         types.StringPointerValue(elt.PublicIpId),
		},
	)
}

func NicFromHttpToTf(ctx context.Context, http *api.Nic) (*resource_nic.NicModel, diag.Diagnostics) {
	// Private IPs
	privateIps, diagnostics := utils.GenericListToTfListValue(ctx, resource_nic.PrivateIpsValue{}, privatesIpFromApi, *http.PrivateIps)
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
	securityGroupsTf, diagnostics := utils.GenericListToTfListValue(ctx, resource_nic.SecurityGroupsValue{}, securityGroupLightFromApi, *http.SecurityGroups)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Link Public Ip
	linkPublicIpTf, diagnostics := linkPublicIpFromApi(ctx, *http.LinkPublicIp)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	return &resource_nic.NicModel{
		AccountId:           types.StringPointerValue(http.AccountId),
		Description:         types.StringPointerValue(http.AccountId),
		Id:                  types.StringPointerValue(http.Id),
		IsSourceDestChecked: types.BoolPointerValue(http.IsSourceDestChecked),
		LinkPublicIp:        linkPublicIpTf,
		MacAddress:          types.StringPointerValue(http.MacAddress),
		NetId:               types.StringPointerValue(http.VpcId),
		PrivateDnsName:      types.StringPointerValue(http.PrivateDnsName),
		PrivateIps:          privateIps,
		SecurityGroupIds:    securityGroupsIdTf,
		SecurityGroups:      securityGroupsTf,
		State:               types.StringPointerValue(http.State),
		SubnetId:            types.StringPointerValue(http.SubnetId),
		SubregionName:       types.StringPointerValue(http.AvailabilityZoneName),
	}, nil
}

func NicFromTfToCreateRequest(ctx context.Context, tf *resource_nic.NicModel) api.CreateNicJSONRequestBody {
	privateIps := utils.TfListToGenericList(func(a resource_nic.PrivateIpsValue) api.PrivateIpLight {
		return api.PrivateIpLight{
			IsPrimary: a.IsPrimary.ValueBoolPointer(),
			PrivateIp: a.PrivateIp.ValueStringPointer(),
		}
	}, ctx, tf.PrivateIps)
	securityGroupIds := utils.TfStringListToStringList(ctx, tf.SecurityGroupIds)

	return api.CreateNicJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		PrivateIps:       &privateIps,
		SecurityGroupIds: &securityGroupIds,
		SubnetId:         tf.SubnetId.ValueString(),
	}
}
