package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nat_service"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func publicIpFromApi(ctx context.Context, elt api.PublicIpLightSchema) (resource_nat_service.PublicIpsValue, diag.Diagnostics) {
	return resource_nat_service.NewPublicIpsValue(
		resource_nat_service.PublicIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_ip":    types.StringPointerValue(elt.PublicIp),
			"public_ip_id": types.StringPointerValue(elt.PublicIpId),
		},
	)
}

func NatServiceFromHttpToTf(ctx context.Context, http *api.NatServiceSchema) (*resource_nat_service.NatServiceModel, diag.Diagnostics) {
	// Public Ips
	publicIpsTf, diagnostics := utils.GenericListToTfListValue(
		ctx,
		resource_nat_service.PublicIpsValue{},
		publicIpFromApi,
		*http.PublicIps,
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	return &resource_nat_service.NatServiceModel{
		Id:        types.StringPointerValue(http.Id),
		NetId:     types.StringPointerValue(http.NetId),
		PublicIps: publicIpsTf,
		State:     types.StringPointerValue(http.State),
		SubnetId:  types.StringPointerValue(http.SubnetId),
	}, nil
}

func NatServiceFromTfToCreateRequest(tf *resource_nat_service.NatServiceModel) api.CreateNatServiceJSONRequestBody {
	return api.CreateNatServiceJSONRequestBody{
		PublicIpId: tf.PublicIpId.ValueString(),
		SubnetId:   tf.SubnetId.ValueString(),
	}
}
