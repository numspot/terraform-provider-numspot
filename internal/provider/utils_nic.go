package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nic"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func NicFromHttpToTf(ctx context.Context, http *api.NicSchema) resource_nic.NicModel {
	return resource_nic.NicModel{
		AccountId:           types.StringPointerValue(http.AccountId),
		Description:         types.StringPointerValue(http.AccountId),
		Id:                  types.StringPointerValue(http.Id),
		IsSourceDestChecked: types.BoolPointerValue(http.IsSourceDestChecked),
		LinkPublicIp:        resource_nic.LinkPublicIpValue{}, // FIXME Handle this
		MacAddress:          types.StringPointerValue(http.MacAddress),
		NetId:               types.StringPointerValue(http.NetId),
		PrivateDnsName:      types.StringPointerValue(http.PrivateDnsName),
		PrivateIps:          types.List{}, // FIXME Handle this
		SecurityGroupIds:    types.List{}, // FIXME Handle this
		SecurityGroups:      types.List{}, // FIXME Handle this
		State:               types.StringPointerValue(http.State),
		SubnetId:            types.StringPointerValue(http.SubnetId),
		SubregionName:       types.StringPointerValue(http.SubregionName),
	}
}

func NicFromTfToCreateRequest(ctx context.Context, tf *resource_nic.NicModel) api.CreateNicJSONRequestBody {
	privateIps := utils.TfListToGenericList(func(a resource_nic.PrivateIpsValue) api.PrivateIpLightSchema {
		return api.PrivateIpLightSchema{
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
