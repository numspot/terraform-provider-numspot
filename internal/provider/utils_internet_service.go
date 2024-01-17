package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_internet_service"
)

func InternetServiceFromTfToHttp(tf resource_internet_service.InternetServiceModel) *api.InternetServiceSchema {
	return &api.InternetServiceSchema{}
}

func InternetServiceFromHttpToTf(http *api.InternetServiceSchema) resource_internet_service.InternetServiceModel {
	return resource_internet_service.InternetServiceModel{}
}

func InternetServiceFromTfToCreateRequest(tf resource_internet_service.InternetServiceModel) api.CreateInternetServiceJSONRequestBody {
	return api.CreateInternetServiceJSONRequestBody{}
}
