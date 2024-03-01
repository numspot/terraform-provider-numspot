package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_gateway"
)

func InternetServiceFromTfToHttp(tf resource_internet_gateway.InternetGatewayModel) *api.InternetService {
	return &api.InternetService{
		Id:    tf.Id.ValueStringPointer(),
		VpcId: tf.VpcIp.ValueStringPointer(),
		State: tf.State.ValueStringPointer(),
	}
}

func InternetServiceFromHttpToTf(http *api.InternetService) resource_internet_gateway.InternetGatewayModel {
	return resource_internet_gateway.InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		VpcIp: types.StringPointerValue(http.VpcId),
		State: types.StringPointerValue(http.State),
	}
}
