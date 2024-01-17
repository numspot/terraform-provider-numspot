package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link"
)

func DirectLinkFromTfToHttp(tf resource_direct_link.DirectLinkModel) *api.DirectLinkSchema {
	return &api.DirectLinkSchema{
		Bandwidth:  tf.Bandwidth.ValueStringPointer(),
		Id:         tf.Id.ValueStringPointer(),
		Location:   tf.Location.ValueStringPointer(),
		Name:       tf.Name.ValueStringPointer(),
		RegionName: tf.RegionName.ValueStringPointer(),
		State:      tf.Bandwidth.ValueStringPointer(),
	}
}

func DirectLinkFromHttpToTf(http *api.DirectLinkSchema) resource_direct_link.DirectLinkModel {
	return resource_direct_link.DirectLinkModel{
		Bandwidth:      types.StringPointerValue(http.Bandwidth),
		DirectLinkName: types.StringPointerValue(http.Name),
		Id:             types.StringPointerValue(http.Id),
		Location:       types.StringPointerValue(http.Location),
		Name:           types.StringPointerValue(http.Name),
		RegionName:     types.StringPointerValue(http.RegionName),
		State:          types.StringPointerValue(http.State),
	}
}

func DirectLinkFromTfToCreateRequest(tf resource_direct_link.DirectLinkModel) api.CreateDirectLinkJSONRequestBody {
	return api.CreateDirectLinkJSONRequestBody{
		Bandwidth:      tf.Bandwidth.ValueStringPointer(),
		DirectLinkName: tf.Name.ValueStringPointer(),
		Location:       tf.Location.ValueStringPointer(),
	}
}
