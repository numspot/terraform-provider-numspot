package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_volume"
)

func VolumeFromTfToHttp(tf resource_volume.VolumeModel) *api.VolumeSchema {
	return &api.VolumeSchema{}
}

func VolumeFromHttpToTf(http *api.VolumeSchema) resource_volume.VolumeModel {
	return resource_volume.VolumeModel{}
}

func VolumeFromTfToCreateRequest(tf resource_volume.VolumeModel) api.CreateVolumeJSONRequestBody {
	return api.CreateVolumeJSONRequestBody{}
}
