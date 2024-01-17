package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_image"
)

func ImageFromTfToHttp(tf resource_image.ImageModel) *api.ImageSchema {
	return &api.ImageSchema{}
}

func ImageFromHttpToTf(http *api.ImageSchema) resource_image.ImageModel {
	return resource_image.ImageModel{}
}

func ImageFromTfToCreateRequest(tf resource_image.ImageModel) api.CreateImageJSONRequestBody {
	return api.CreateImageJSONRequestBody{}
}
