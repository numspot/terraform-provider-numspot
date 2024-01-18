package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

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

func RouteTableFromHttpToTf(http *api.RouteTableSchema) resource_route_table.RouteTableModel {
	return resource_route_table.RouteTableModel{
		Id:                              types.StringPointerValue(http.Id),
		NetId:                           types.StringPointerValue(http.NetId),
		RoutePropagatingVirtualGateways: types.List{},
		Routes:                          types.List{},
	}
}

func RouteTableFromTfToCreateRequest(tf resource_route_table.RouteTableModel) api.CreateRouteTableJSONRequestBody {
	return api.CreateRouteTableJSONRequestBody{
		NetId: tf.NetId.ValueString(),
	}
}
