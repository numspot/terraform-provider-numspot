package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_service"
)

func InternetServiceFromTfToHttp(tf resource_internet_service.InternetServiceModel) *api.InternetServiceSchema {
	return &api.InternetServiceSchema{
		Id:    tf.Id.ValueStringPointer(),
		NetId: tf.NetId.ValueStringPointer(),
		State: tf.State.ValueStringPointer(),
	}
}

func InternetServiceFromHttpToTf(http *api.InternetServiceSchema) resource_internet_service.InternetServiceModel {
	return resource_internet_service.InternetServiceModel{
		Id:    types.StringPointerValue(http.Id),
		NetId: types.StringPointerValue(http.NetId),
		State: types.StringPointerValue(http.State),
	}
}

func InternetServiceFromTfToCreateRequest(_ resource_internet_service.InternetServiceModel) api.CreateInternetServiceJSONRequestBody {
	return api.CreateInternetServiceJSONRequestBody{}
}
