package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_virtual_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func vpcToVGLinkFromApi(ctx context.Context, from api.VpcToVirtualGatewayLink) (resource_virtual_gateway.NetToVirtualGatewayLinksValue, diag.Diagnostics) {
	return resource_virtual_gateway.NewNetToVirtualGatewayLinksValue(
		resource_virtual_gateway.NetToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
		map[string]attr.Value{},
	)
}

func VirtualGatewayFromHttpToTf(ctx context.Context, http *api.VirtualGateway) (*resource_virtual_gateway.VirtualGatewayModel, diag.Diagnostics) {
	var netToVirtualGatewaysLinkTd types.List
	var diagnostics diag.Diagnostics

	if http.NetToVirtualGatewayLinks != nil {
		netToVirtualGatewaysLinkTd, diagnostics = utils.GenericListToTfListValue(
			ctx,
			resource_virtual_gateway.NetToVirtualGatewayLinksValue{},
			vpcToVGLinkFromApi,
			*http.NetToVirtualGatewayLinks,
		)
		if diagnostics.HasError() {
			return nil, diagnostics
		}
	} else {
		netToVirtualGatewaysLinkTd = types.ListNull(resource_virtual_gateway.NetToVirtualGatewayLinksValue{}.Type(ctx))
	}

	return &resource_virtual_gateway.VirtualGatewayModel{
		ConnectionType:           types.StringPointerValue(http.ConnectionType),
		Id:                       types.StringPointerValue(http.Id),
		NetToVirtualGatewayLinks: netToVirtualGatewaysLinkTd,
		State:                    types.StringPointerValue(http.State),
	}, diagnostics
}

func VirtualGatewayFromTfToCreateRequest(tf resource_virtual_gateway.VirtualGatewayModel) api.CreateVirtualGatewayJSONRequestBody {
	return api.CreateVirtualGatewayJSONRequestBody{
		ConnectionType: tf.ConnectionType.ValueString(),
	}
}
