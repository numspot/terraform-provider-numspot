package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_service"
)

func InternetServiceFromTfToHttp(tf resource_internet_service.InternetServiceModel) *api.InternetService {
	return &api.InternetService{
		InternetGatewayId: tf.Id.ValueStringPointer(),
		VpcId:             tf.NetId.ValueStringPointer(),
		State:             tf.State.ValueStringPointer(),
	}
}

func InternetServiceFromHttpToTf(http *api.InternetService) resource_internet_service.InternetServiceModel {
	return resource_internet_service.InternetServiceModel{
		Id:    types.StringPointerValue(http.InternetGatewayId),
		NetId: types.StringPointerValue(http.VpcId),
		State: types.StringPointerValue(http.State),
	}
}
