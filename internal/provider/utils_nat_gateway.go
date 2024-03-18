package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_nat_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func publicIpFromApi(ctx context.Context, elt api.PublicIpLight) (resource_nat_gateway.PublicIpsValue, diag.Diagnostics) {
	return resource_nat_gateway.NewPublicIpsValue(
		resource_nat_gateway.PublicIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_ip":    types.StringPointerValue(elt.PublicIp),
			"public_ip_id": types.StringPointerValue(elt.PublicIp),
		},
	)
}

func NatGatewayFromHttpToTf(ctx context.Context, http *api.NatGateway) (*resource_nat_gateway.NatGatewayModel, diag.Diagnostics) {
	// Public Ips
	publicIpsTf, diagnostics := utils.GenericListToTfListValue(
		ctx,
		resource_nat_gateway.PublicIpsValue{},
		publicIpFromApi,
		*http.PublicIps,
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// PublicIpId must be the id of the first public io
	publicIpId := (*http.PublicIps)[0].PublicIpId

	return &resource_nat_gateway.NatGatewayModel{
		Id:         types.StringPointerValue(http.Id),
		PublicIpId: types.StringPointerValue(publicIpId),
		PublicIps:  publicIpsTf,
		State:      types.StringPointerValue(http.State),
		SubnetId:   types.StringPointerValue(http.SubnetId),
		VpcId:      types.StringPointerValue(http.VpcId),
	}, nil
}

func NatGatewayFromTfToCreateRequest(tf resource_nat_gateway.NatGatewayModel) api.CreateNatGatewayJSONRequestBody {
	return api.CreateNatGatewayJSONRequestBody{
		PublicIpId: tf.PublicIpId.ValueString(),
		SubnetId:   tf.SubnetId.ValueString(),
	}
}
