package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_gateway"
)

func InternetServiceFromTfToHttp(tf resource_internet_gateway.InternetGatewayModel) *api.InternetGateway {
	return &api.InternetGateway{
		Id:    tf.Id.ValueStringPointer(),
		VpcId: tf.VpcIp.ValueStringPointer(),
		State: tf.State.ValueStringPointer(),
	}
}

func InternetServiceFromHttpToTf(http *api.InternetGateway) resource_internet_gateway.InternetGatewayModel {
	return resource_internet_gateway.InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		VpcIp: types.StringPointerValue(http.VpcId),
		State: types.StringPointerValue(http.State),
	}
}
