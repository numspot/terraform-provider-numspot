package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpc"
)

func NetFromTfToHttp(tf *resource_vpc.VpcModel) *api.Vpc {
	return &api.Vpc{
		DhcpOptionsSetId: tf.DhcpOptionsSetId.ValueStringPointer(),
		Id:               tf.Id.ValueStringPointer(),
		IpRange:          tf.IpRange.ValueStringPointer(),
		State:            tf.State.ValueStringPointer(),
		Tenancy:          tf.Tenancy.ValueStringPointer(),
	}
}

func NetFromHttpToTf(http *api.Vpc) resource_vpc.VpcModel {
	return resource_vpc.VpcModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
	}
}

func NetFromTfToCreateRequest(tf *resource_vpc.VpcModel) api.CreateVpcJSONRequestBody {
	return api.CreateVpcJSONRequestBody{
		IpRange: tf.IpRange.ValueString(),
		Tenancy: tf.Tenancy.ValueStringPointer(),
	}
}
