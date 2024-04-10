package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_direct_link"
)

func DirectLinkFromHttpToTf(http *iaas.DirectLink) resource_direct_link.DirectLinkModel {
	return resource_direct_link.DirectLinkModel{
		Bandwidth:  types.StringPointerValue(http.Bandwidth),
		Name:       types.StringPointerValue(http.Name),
		Id:         types.StringPointerValue(http.Id),
		Location:   types.StringPointerValue(http.Location),
		RegionName: types.StringPointerValue(http.RegionName),
		State:      types.StringPointerValue(http.State),
	}
}

func DirectLinkFromTfToCreateRequest(tf *resource_direct_link.DirectLinkModel) iaas.CreateDirectLinkJSONRequestBody {
	return iaas.CreateDirectLinkJSONRequestBody{
		Bandwidth: tf.Bandwidth.ValueString(),
		Name:      tf.Name.ValueString(),
		Location:  tf.Location.ValueString(),
	}
}
