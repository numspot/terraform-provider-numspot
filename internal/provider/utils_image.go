package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_image"
)

func ImageFromTfToHttp(tf *resource_image.ImageModel) *api.Image {
	return &api.Image{}
}

func ImageFromHttpToTf(http *api.Image) resource_image.ImageModel {
	return resource_image.ImageModel{}
}

func ImageFromTfToCreateRequest(tf *resource_image.ImageModel) api.CreateImageJSONRequestBody {
	return api.CreateImageJSONRequestBody{}
}
