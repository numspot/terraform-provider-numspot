package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nat_service"
)

func NatServiceFromHttpToTf(http *api.NatServiceSchema) resource_nat_service.NatServiceModel {
	return resource_nat_service.NatServiceModel{
		Id:         types.StringPointerValue(http.Id),
		NetId:      types.StringPointerValue(http.NetId),
		PublicIpId: types.String{}, // FIXME
		PublicIps:  types.List{},   // FIXME
		State:      types.StringPointerValue(http.State),
		SubnetId:   types.StringPointerValue(http.SubnetId),
	}
}

func NatServiceFromTfToCreateRequest(tf *resource_nat_service.NatServiceModel) api.CreateNatServiceJSONRequestBody {
	return api.CreateNatServiceJSONRequestBody{
		PublicIpId: tf.PublicIpId.ValueString(),
		SubnetId:   tf.SubnetId.ValueString(),
	}
}
