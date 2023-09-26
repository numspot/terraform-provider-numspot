package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_net"
)

func NetFromTfToHttp(tf resource_net.NetModel) *api.NetSchema {
	return &api.NetSchema{
		DhcpOptionsSetId: tf.DhcpOptionsSetId.ValueStringPointer(),
		Id:               tf.Id.ValueStringPointer(),
		IpRange:          tf.IpRange.ValueStringPointer(),
		State:            tf.State.ValueStringPointer(),
		Tenancy:          tf.Tenancy.ValueStringPointer(),
	}
}

func NetFromHttpToTf(http *api.NetSchema) resource_net.NetModel {
	return resource_net.NetModel{
		DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
		Id:               types.StringPointerValue(http.Id),
		IpRange:          types.StringPointerValue(http.IpRange),
		State:            types.StringPointerValue(http.State),
		Tenancy:          types.StringPointerValue(http.Tenancy),
	}
}

func NetFromTfToCreateRequest(tf resource_net.NetModel) api.CreateNetJSONRequestBody {
	return api.CreateNetJSONRequestBody{
		IpRange: tf.IpRange.ValueString(),
		Tenancy: tf.Tenancy.ValueStringPointer(),
	}
}
