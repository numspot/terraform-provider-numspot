package routetable

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type RouteTablesDataSourceModel struct {
	Items                           []RouteTableModelDatasource `tfsdk:"items"`
	Ids                             types.List                  `tfsdk:"ids"`
	LinkRouteTableIds               types.List                  `tfsdk:"link_route_table_ids"`
	LinkRouteTableLinkRouteTableIds types.List                  `tfsdk:"link_route_table_link_route_table_ids"`
	LinkRouteTableMain              types.Bool                  `tfsdk:"link_route_table_main"`
	LinkSubnetIds                   types.List                  `tfsdk:"link_subnet_ids"`
	RouteCreationMethods            types.List                  `tfsdk:"route_creation_methods"`
	RouteDestinationIpRanges        types.List                  `tfsdk:"route_destination_ip_ranges"`
	RouteDestinationServiceIds      types.List                  `tfsdk:"route_destination_service_ids"`
	RouteGatewayIds                 types.List                  `tfsdk:"route_gateway_ids"`
	RouteNatGatewayIds              types.List                  `tfsdk:"route_nat_gateway_ids"`
	RouteStates                     types.List                  `tfsdk:"route_states"`
	RouteVmIds                      types.List                  `tfsdk:"route_vm_ids"`
	RouteVpcPeeringIds              types.List                  `tfsdk:"route_vpc_peering_ids"`
	TagKeys                         types.List                  `tfsdk:"tag_keys"`
	TagValues                       types.List                  `tfsdk:"tag_values"`
	Tags                            types.List                  `tfsdk:"tags"`
	VpcIds                          types.List                  `tfsdk:"vpc_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &routeTablesDataSource{}
)

func (d *routeTablesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
}

func NewRouteTablesDataSource() datasource.DataSource {
	return &routeTablesDataSource{}
}

type routeTablesDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *routeTablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_tables"
}

// Schema defines the schema for the data source.
func (d *routeTablesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = RouteTableDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *routeTablesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan RouteTablesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeRouteTableDatasourceParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	routeTables, err := core.ReadRouteTables(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("failed to read route tables", err.Error())
		return
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, routeTables, serializeRouteTableDatasource, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeRouteTableDatasourceParams(ctx context.Context, tf RouteTablesDataSourceModel, diags *diag.Diagnostics) numspot.ReadRouteTablesParams {
	return numspot.ReadRouteTablesParams{
		TagKeys:                         utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:                       utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:                            utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:                             utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		RouteVpcPeeringIds:              utils.TfStringListToStringPtrList(ctx, tf.RouteVpcPeeringIds, diags),
		RouteNatGatewayIds:              utils.TfStringListToStringPtrList(ctx, tf.RouteNatGatewayIds, diags),
		RouteVmIds:                      utils.TfStringListToStringPtrList(ctx, tf.RouteVmIds, diags),
		RouteCreationMethods:            utils.TfStringListToStringPtrList(ctx, tf.RouteCreationMethods, diags),
		RouteDestinationIpRanges:        utils.TfStringListToStringPtrList(ctx, tf.RouteDestinationIpRanges, diags),
		RouteDestinationServiceIds:      utils.TfStringListToStringPtrList(ctx, tf.RouteDestinationServiceIds, diags),
		RouteGatewayIds:                 utils.TfStringListToStringPtrList(ctx, tf.RouteGatewayIds, diags),
		RouteStates:                     utils.TfStringListToStringPtrList(ctx, tf.RouteStates, diags),
		VpcIds:                          utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		LinkRouteTableIds:               utils.TfStringListToStringPtrList(ctx, tf.LinkRouteTableIds, diags),
		LinkRouteTableMain:              utils.FromTfBoolToBoolPtr(tf.LinkRouteTableMain),
		LinkRouteTableLinkRouteTableIds: utils.TfStringListToStringPtrList(ctx, tf.LinkRouteTableLinkRouteTableIds, diags),
		LinkSubnetIds:                   utils.TfStringListToStringPtrList(ctx, tf.LinkSubnetIds, diags),
	}
}

func serializeRouteTableDatasource(ctx context.Context, http *numspot.RouteTable, diags *diag.Diagnostics) *RouteTableModelDatasource {
	var (
		tagsList                            = types.ListNull(tags.TagsValue{}.Type(ctx))
		linkRouteTablesList                 = types.ListNull(LinkRouteTablesValue{}.Type(ctx))
		routes                              = types.ListNull(RoutesValue{}.Type(ctx))
		routePropagatingVirtualGatewaysList = types.ListNull(RoutePropagatingVirtualGatewaysValue{}.Type(ctx))
	)

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.LinkRouteTables != nil {
		linkRouteTablesList = utils.GenericListToTfListValue(ctx, serializeRouteTableLink, *http.LinkRouteTables, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.RoutePropagatingVirtualGateways != nil {
		routePropagatingVirtualGatewaysList = utils.GenericListToTfListValue(
			ctx,
			serializeRouteTableRoutePropagatingVirtualGateways,
			*http.RoutePropagatingVirtualGateways, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.Routes != nil {
		routes = utils.GenericListToTfListValue(ctx, serializeRoute, *http.Routes, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &RouteTableModelDatasource{
		Id:                              types.StringPointerValue(http.Id),
		Tags:                            tagsList,
		LinkRouteTables:                 linkRouteTablesList,
		RoutePropagatingVirtualGateways: routePropagatingVirtualGatewaysList,
		VpcId:                           types.StringPointerValue(http.VpcId),
		Routes:                          routes,
	}
}

func serializeRouteTableRoutePropagatingVirtualGateways(ctx context.Context, route numspot.RoutePropagatingVirtualGateway, diags *diag.Diagnostics) RoutePropagatingVirtualGatewaysValue {
	value, diagnostics := NewRoutePropagatingVirtualGatewaysValue(
		RoutePropagatingVirtualGatewaysValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"virtual_gateway_id": types.StringPointerValue(route.VirtualGatewayId),
		},
	)
	diags.Append(diagnostics...)
	return value
}
