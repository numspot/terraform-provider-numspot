package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_route_table"
)

func RouteTableFromTfToHttp(tf resource_route_table.RouteTableModel) *api.RouteTableSchema {
	return &api.RouteTableSchema{
		Id:                              tf.Id.ValueStringPointer(),
		NetId:                           tf.NetId.ValueStringPointer(),
		RoutePropagatingVirtualGateways: nil,
		Routes:                          nil,
	}
}

type RRModel struct {
	CreationMethod       basetypes.StringValue `tfsdk:"creation_method"`
	DestinationIpRange   basetypes.StringValue `tfsdk:"destination_ip_range"`
	DestinationServiceId basetypes.StringValue `tfsdk:"destination_service_id"`
	GatewayId            basetypes.StringValue `tfsdk:"gateway_id"`
	NatServiceId         basetypes.StringValue `tfsdk:"nat_service_id"`
	NetAccessPointId     basetypes.StringValue `tfsdk:"net_access_point_id"`
	NetPeeringId         basetypes.StringValue `tfsdk:"net_peering_id"`
	NicId                basetypes.StringValue `tfsdk:"nic_id"`
	State                basetypes.StringValue `tfsdk:"state"`
	VmAccountId          basetypes.StringValue `tfsdk:"vm_account_id"`
	VmId                 basetypes.StringValue `tfsdk:"vm_id"`
	state                attr.ValueState
}

func RouteTableFromHttpToTf(ctx context.Context, http *api.RouteTableSchema) (*resource_route_table.RouteTableModel, diag.Diagnostics) {
	routes := make([]resource_route_table.RoutesValue, 0, len(*http.Routes))
	for _, route := range *http.Routes {
		nroutev, _ := resource_route_table.NewRoutesValue(
			resource_route_table.RoutesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"creation_method":        types.StringPointerValue(route.CreationMethod),
				"destination_ip_range":   types.StringPointerValue(route.DestinationIpRange),
				"destination_service_id": types.StringPointerValue(route.DestinationServiceId),
				"gateway_id":             types.StringPointerValue(route.GatewayId),
				"nat_service_id":         types.StringPointerValue(route.NatServiceId),
				"net_access_point_id":    types.StringPointerValue(route.NetAccessPointId),
				"net_peering_id":         types.StringPointerValue(route.NetPeeringId),
				"nic_id":                 types.StringPointerValue(route.NicId),
				"state":                  types.StringPointerValue(route.State),
				"vm_account_id":          types.StringPointerValue(route.VmAccountId),
				"vm_id":                  types.StringPointerValue(route.VmId),
			},
		)

		routes = append(routes, nroutev)
	}

	tfRoutes, diag := types.ListValueFrom(
		ctx,
		resource_route_table.RoutesValue{}.Type(ctx),
		routes,
	)
	if diag.HasError() {
		return nil, diag
	}

	res := resource_route_table.RouteTableModel{
		Id:                              types.StringPointerValue(http.Id),
		NetId:                           types.StringPointerValue(http.NetId),
		RoutePropagatingVirtualGateways: types.ListNull(resource_route_table.RoutePropagatingVirtualGatewaysValue{}.Type(context.Background())),
		Routes:                          tfRoutes,
	}

	return &res, nil
}

func RouteTableFromTfToCreateRequest(tf resource_route_table.RouteTableModel) api.CreateRouteTableJSONRequestBody {
	return api.CreateRouteTableJSONRequestBody{
		NetId: tf.NetId.ValueString(),
	}
}
