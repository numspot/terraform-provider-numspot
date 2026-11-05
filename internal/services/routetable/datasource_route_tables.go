package routetable

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/routetable/datasource_route_table"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &routetablesDataSource{}

// routetablesDataSource is the Terraform data source implementation for numspot_route_tables.
type routetablesDataSource struct {
	services.BaseDataSource
}

// NewRouteTablesDataSource returns a new routetablesDataSource instance.
func NewRouteTablesDataSource() datasource.DataSource {
	return &routetablesDataSource{}
}

// Metadata sets the data source type name.
func (d *routetablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_tables"
}

// Schema defines the schema for the data source.
func (d *routetablesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_route_table.RouteTableDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *routetablesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_route_table.RouteTableModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadRouteTables(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadRouteTables(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read route_tables", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.RouteTable) (datasource_route_table.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	tagsList := types.ListNull(datasource_route_table.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_route_table.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_route_table.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_route_table.NewItemsValue(datasource_route_table.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":                                 types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"link_route_tables":                  mappingLinkRouteTables(ctx, item, &diags),
		"route_propagating_virtual_gateways": mappingRoutePropagatingVirtualGateways(ctx, item, &diags),
		"routes":                             mappingRoutes(ctx, item, &diags),
		"vpc_id":                             types.StringValue(utils.ConvertStringPtrToString(item.VpcId)),
		"tags":                               tagsList,
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_route_table.TagsValue, diag.Diagnostics) {
	return datasource_route_table.NewTagsValue(datasource_route_table.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingLinkRouteTable(ctx context.Context, item *api.LinkRouteTable) (datasource_route_table.LinkRouteTablesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_route_table.NewLinkRouteTablesValue(
		datasource_route_table.LinkRouteTablesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":             types.StringPointerValue(item.Id),
			"main":           types.BoolPointerValue(item.Main),
			"route_table_id": types.StringPointerValue(item.RouteTableId),
			"subnet_id":      types.StringPointerValue(item.SubnetId),
			"vpc_id":         types.StringPointerValue(item.VpcId),
		},
	)
	diags.Append(d...)
	return value, diags
}

func mappingLinkRouteTables(ctx context.Context, item *api.RouteTable, diags *diag.Diagnostics) types.List {
	if item.LinkRouteTables == nil {
		return types.ListNull(datasource_route_table.LinkRouteTablesValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.LinkRouteTables, mappingLinkRouteTable)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_route_table.LinkRouteTablesValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingRoutePropagatingVirtualGateway(
	ctx context.Context,
	item *api.RoutePropagatingVirtualGateway,
) (datasource_route_table.RoutePropagatingVirtualGatewaysValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_route_table.NewRoutePropagatingVirtualGatewaysValue(
		datasource_route_table.RoutePropagatingVirtualGatewaysValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"virtual_gateway_id": types.StringPointerValue(item.VirtualGatewayId),
		},
	)
	diags.Append(d...)
	return value, diags
}

func mappingRoutePropagatingVirtualGateways(ctx context.Context, item *api.RouteTable, diags *diag.Diagnostics) types.List {
	if item.RoutePropagatingVirtualGateways == nil {
		return types.ListNull(datasource_route_table.RoutePropagatingVirtualGatewaysValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.RoutePropagatingVirtualGateways, mappingRoutePropagatingVirtualGateway)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_route_table.RoutePropagatingVirtualGatewaysValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingRoute(ctx context.Context, item *api.Route) (datasource_route_table.RoutesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_route_table.NewRoutesValue(datasource_route_table.RoutesValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"creation_method":        types.StringPointerValue(item.CreationMethod),
		"destination_ip_range":   types.StringPointerValue(item.DestinationIpRange),
		"destination_service_id": types.StringPointerValue(item.DestinationServiceId),
		"gateway_id":             types.StringPointerValue(item.GatewayId),
		"nat_gateway_id":         types.StringPointerValue(item.NatGatewayId),
		"nic_id":                 types.StringPointerValue(item.NicId),
		"state":                  types.StringPointerValue(item.State),
		"vm_id":                  types.StringPointerValue(item.VmId),
		"vpc_peering_id":         types.StringPointerValue(item.VpcPeeringId),
	})
	diags.Append(d...)
	return value, diags
}

func mappingRoutes(ctx context.Context, item *api.RouteTable, diags *diag.Diagnostics) types.List {
	if item.Routes == nil {
		return types.ListNull(datasource_route_table.RoutesValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Routes, mappingRoute)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_route_table.RoutesValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func deserializeReadRouteTables(ctx context.Context, tf *datasource_route_table.RouteTableModel, diags *diag.Diagnostics) *api.ReadRouteTablesParams {
	return &api.ReadRouteTablesParams{
		LinkRouteTableIds:               utils.ConvertTfListToArrayOfString(ctx, tf.LinkRouteTableIds, diags),
		LinkRouteTableLinkRouteTableIds: utils.ConvertTfListToArrayOfString(ctx, tf.LinkRouteTableLinkRouteTableIds, diags),
		LinkRouteTableMain:              tf.LinkRouteTableMain.ValueBoolPointer(),
		LinkSubnetIds:                   utils.ConvertTfListToArrayOfString(ctx, tf.LinkSubnetIds, diags),
		RouteCreationMethods:            utils.ConvertTfListToArrayOfString(ctx, tf.RouteCreationMethods, diags),
		RouteDestinationIpRanges:        utils.ConvertTfListToArrayOfString(ctx, tf.RouteDestinationIpRanges, diags),
		RouteDestinationServiceIds:      utils.ConvertTfListToArrayOfString(ctx, tf.RouteDestinationServiceIds, diags),
		RouteGatewayIds:                 utils.ConvertTfListToArrayOfString(ctx, tf.RouteGatewayIds, diags),
		RouteStates:                     utils.ConvertTfListToArrayOfString(ctx, tf.RouteStates, diags),
		RouteVmIds:                      utils.ConvertTfListToArrayOfString(ctx, tf.RouteVmIds, diags),
		TagKeys:                         utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:                       utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:                            utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		VpcIds:                          utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		RouteNatGatewayIds:              utils.ConvertTfListToArrayOfString(ctx, tf.RouteNatGatewayIds, diags),
		RouteVpcPeeringIds:              utils.ConvertTfListToArrayOfString(ctx, tf.RouteVpcPeeringIds, diags),
		Ids:                             utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
