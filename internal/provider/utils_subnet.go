package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_subnet"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SubnetFromTfToHttp(tf resource_subnet.SubnetModel) *api.SubnetSchema {
	return &api.SubnetSchema{
		AvailableIpsCount:   utils.FromTfInt64ToIntPtr(tf.AvailableIpsCount),
		Id:                  tf.Id.ValueStringPointer(),
		IpRange:             tf.IpRange.ValueStringPointer(),
		MapPublicIpOnLaunch: tf.MapPublicIpOnLaunch.ValueBoolPointer(),
		NetId:               tf.NetId.ValueStringPointer(),
		State:               tf.State.ValueStringPointer(),
		SubregionName:       tf.SubregionName.ValueStringPointer(),
	}
}

func SubnetFromHttpToTf(http *api.SubnetSchema) resource_subnet.SubnetModel {
	return resource_subnet.SubnetModel{
		AvailableIpsCount:   utils.FromIntPtrToTfInt64(http.AvailableIpsCount),
		Id:                  types.StringPointerValue(http.Id),
		IpRange:             types.StringPointerValue(http.IpRange),
		MapPublicIpOnLaunch: types.BoolPointerValue(http.MapPublicIpOnLaunch),
		NetId:               types.StringPointerValue(http.NetId),
		State:               types.StringPointerValue(http.State),
		SubregionName:       types.StringPointerValue(http.SubregionName),
	}
}

func SubnetFromTfToCreateRequest(tf resource_subnet.SubnetModel) api.CreateSubnetJSONRequestBody {
	return api.CreateSubnetJSONRequestBody{
		IpRange:       tf.IpRange.ValueStringPointer(),
		NetId:         tf.NetId.ValueStringPointer(),
		SubregionName: tf.SubregionName.ValueStringPointer(),
	}
}
