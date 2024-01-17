package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link_interface"
)

func DirectLinkInterfaceFromTfToHttp(tf resource_direct_link_interface.DirectLinkInterfaceModel) *api.DirectLinkInterfacesSchema {
	return &api.DirectLinkInterfacesSchema{}
}

func DirectLinkInterfaceFromHttpToTf(http *api.DirectLinkInterfacesSchema) resource_direct_link_interface.DirectLinkInterfaceModel {
	return resource_direct_link_interface.DirectLinkInterfaceModel{}
}

func DirectLinkInterfaceFromTfToCreateRequest(tf resource_direct_link_interface.DirectLinkInterfaceModel) api.CreateDirectLinkInterfaceJSONRequestBody {
	return api.CreateDirectLinkInterfaceJSONRequestBody{}
}
