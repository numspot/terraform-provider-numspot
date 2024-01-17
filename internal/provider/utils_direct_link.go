package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link"
)

func DirectLinkFromTfToHttp(tf resource_direct_link.DirectLinkModel) *api.DirectLinkSchema {
	return &api.DirectLinkSchema{}
}

func DirectLinkFromHttpToTf(http *api.DirectLinkSchema) resource_direct_link.DirectLinkModel {
	return resource_direct_link.DirectLinkModel{}
}

func DirectLinkFromTfToCreateRequest(tf resource_direct_link.DirectLinkModel) api.CreateDirectLinkJSONRequestBody {
	return api.CreateDirectLinkJSONRequestBody{}
}
