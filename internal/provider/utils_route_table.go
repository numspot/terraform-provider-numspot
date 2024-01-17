package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_route_table"
)

func RouteTableFromTfToHttp(tf resource_route_table.RouteTableModel) *api.RouteTableSchema {
	return &api.RouteTableSchema{}
}

func RouteTableFromHttpToTf(http *api.RouteTableSchema) resource_route_table.RouteTableModel {
	return resource_route_table.RouteTableModel{}
}

func RouteTableFromTfToCreateRequest(tf resource_route_table.RouteTableModel) api.CreateRouteTableJSONRequestBody {
	return api.CreateRouteTableJSONRequestBody{}
}
