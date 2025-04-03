package routetable

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/routetable/datasource_route_table"
	"terraform-provider-numspot/internal/utils"
)

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
			"Unexpected Datasource Configure Type",
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
	resp.Schema = datasource_route_table.RouteTableDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *routeTablesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_route_table.RouteTableModel
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

	objectItems := utils.SerializeDatasourceItemsWithDiags(ctx, *routeTables, &response.Diagnostics, mappingItemsValue)
	if response.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeRouteTableDatasourceParams(ctx context.Context, tf datasource_route_table.RouteTableModel, diags *diag.Diagnostics) api.ReadRouteTablesParams {
	return api.ReadRouteTablesParams{
		TagKeys:                         utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:                       utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:                            utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:                             utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		RouteVpcPeeringIds:              utils.ConvertTfListToArrayOfString(ctx, tf.RouteVpcPeeringIds, diags),
		RouteNatGatewayIds:              utils.ConvertTfListToArrayOfString(ctx, tf.RouteNatGatewayIds, diags),
		RouteVmIds:                      utils.ConvertTfListToArrayOfString(ctx, tf.RouteVmIds, diags),
		RouteCreationMethods:            utils.ConvertTfListToArrayOfString(ctx, tf.RouteCreationMethods, diags),
		RouteDestinationIpRanges:        utils.ConvertTfListToArrayOfString(ctx, tf.RouteDestinationIpRanges, diags),
		RouteDestinationServiceIds:      utils.ConvertTfListToArrayOfString(ctx, tf.RouteDestinationServiceIds, diags),
		RouteGatewayIds:                 utils.ConvertTfListToArrayOfString(ctx, tf.RouteGatewayIds, diags),
		RouteStates:                     utils.ConvertTfListToArrayOfString(ctx, tf.RouteStates, diags),
		VpcIds:                          utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		LinkRouteTableIds:               utils.ConvertTfListToArrayOfString(ctx, tf.LinkRouteTableIds, diags),
		LinkRouteTableMain:              tf.LinkRouteTableMain.ValueBoolPointer(),
		LinkRouteTableLinkRouteTableIds: utils.ConvertTfListToArrayOfString(ctx, tf.LinkRouteTableLinkRouteTableIds, diags),
		LinkSubnetIds:                   utils.ConvertTfListToArrayOfString(ctx, tf.LinkSubnetIds, diags),
	}
}

func mappingItemsValue(ctx context.Context, routeTable api.RouteTable, diags *diag.Diagnostics) (datasource_route_table.ItemsValue, diag.Diagnostics) {
	var serializeDiags diag.Diagnostics

	tagsList := types.ListNull(datasource_route_table.ItemsValue{}.Type(ctx))
	linkRouteTablesList := types.List{}
	propagatingRouteTablesList := types.List{}
	routesList := types.List{}

	if routeTable.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *routeTable.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_route_table.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_route_table.ItemsValue{}, serializeDiags
		}
	}

	if routeTable.LinkRouteTables != nil {
		linkRouteTablesList, serializeDiags = mappingRouteTableLink(ctx, routeTable, diags)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	if routeTable.RoutePropagatingVirtualGateways != nil {
		propagatingRouteTablesList, serializeDiags = mappingRouteTablePropagating(ctx, routeTable, diags)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	if routeTable.Routes != nil {
		routesList, serializeDiags = mappingRoutes(ctx, routeTable, diags)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	return datasource_route_table.NewItemsValue(datasource_route_table.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":                                 types.StringValue(utils.ConvertStringPtrToString(routeTable.Id)),
		"link_route_tables":                  linkRouteTablesList,
		"route_propagating_virtual_gateways": propagatingRouteTablesList,
		"routes":                             routesList,
		"tags":                               tagsList,
		"vpc_id":                             types.StringValue(utils.ConvertStringPtrToString(routeTable.VpcId)),
	})
}

func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_route_table.TagsValue, diag.Diagnostics) {
	return datasource_route_table.NewTagsValue(datasource_route_table.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingRouteTableLink(ctx context.Context, routeTable api.RouteTable, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	ll := len(*routeTable.LinkRouteTables)
	elementValue := make([]datasource_route_table.LinkRouteTablesValue, ll)
	for y, link := range *routeTable.LinkRouteTables {
		elementValue[y], *diags = datasource_route_table.NewLinkRouteTablesValue(datasource_route_table.LinkRouteTablesValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"id":             types.StringValue(utils.ConvertStringPtrToString(link.Id)),
			"main":           types.BoolPointerValue(link.Main),
			"route_table_id": types.StringValue(utils.ConvertStringPtrToString(link.RouteTableId)),
			"subnet_id":      types.StringValue(utils.ConvertStringPtrToString(link.SubnetId)),
			"vpc_id":         types.StringValue(utils.ConvertStringPtrToString(link.VpcId)),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_route_table.LinkRouteTablesValue).Type(ctx), elementValue)
}

func mappingRouteTablePropagating(ctx context.Context, routeTable api.RouteTable, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	lp := len(*routeTable.RoutePropagatingVirtualGateways)
	elementValue := make([]datasource_route_table.RoutePropagatingVirtualGatewaysValue, lp)
	for y, propagating := range *routeTable.RoutePropagatingVirtualGateways {
		elementValue[y], *diags = datasource_route_table.NewRoutePropagatingVirtualGatewaysValue(datasource_route_table.RoutePropagatingVirtualGatewaysValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"virtual_gateway_id": types.StringValue(utils.ConvertStringPtrToString(propagating.VirtualGatewayId)),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_route_table.RoutePropagatingVirtualGatewaysValue).Type(ctx), elementValue)
}

func mappingRoutes(ctx context.Context, routeTable api.RouteTable, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	lr := len(*routeTable.Routes)
	elementValue := make([]datasource_route_table.RoutesValue, lr)
	for y, route := range *routeTable.Routes {
		elementValue[y], *diags = datasource_route_table.NewRoutesValue(datasource_route_table.RoutesValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"creation_method":        types.StringValue(utils.ConvertStringPtrToString(route.CreationMethod)),
			"destination_ip_range":   types.StringValue(utils.ConvertStringPtrToString(route.DestinationIpRange)),
			"destination_service_id": types.StringValue(utils.ConvertStringPtrToString(route.DestinationServiceId)),
			"gateway_id":             types.StringValue(utils.ConvertStringPtrToString(route.GatewayId)),
			"nat_gateway_id":         types.StringValue(utils.ConvertStringPtrToString(route.NatGatewayId)),
			"nic_id":                 types.StringValue(utils.ConvertStringPtrToString(route.NicId)),
			"state":                  types.StringValue(utils.ConvertStringPtrToString(route.State)),
			"vm_id":                  types.StringValue(utils.ConvertStringPtrToString(route.VmId)),
			"vpc_peering_id":         types.StringValue(utils.ConvertStringPtrToString(route.VpcPeeringId)),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_route_table.RoutesValue).Type(ctx), elementValue)
}
