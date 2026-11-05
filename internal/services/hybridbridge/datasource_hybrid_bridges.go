package hybridbridge

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/hybridbridge/datasource_hybrid_bridge"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &hybridbridgesDataSource{}

// hybridbridgesDataSource is the Terraform data source implementation for numspot_hybrid_bridges.
type hybridbridgesDataSource struct {
	services.BaseDataSource
}

// NewHybridBridgesDataSource returns a new hybridbridgesDataSource instance.
func NewHybridBridgesDataSource() datasource.DataSource {
	return &hybridbridgesDataSource{}
}

// Metadata sets the data source type name.
func (d *hybridbridgesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hybrid_bridges"
}

// Schema defines the schema for the data source.
func (d *hybridbridgesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_hybrid_bridge.HybridBridgeDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *hybridbridgesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_hybrid_bridge.HybridBridgeModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	result, err := core.ReadHybridBridges(ctx, d.Provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read hybrid_bridges", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.HybridBridge) (datasource_hybrid_bridge.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, valueDiags := datasource_hybrid_bridge.NewItemsValue(datasource_hybrid_bridge.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"destination":    types.StringValue(item.Destination),
		"destination_id": types.StringValue(item.DestinationId),
		"id":             types.StringValue(item.Id.String()),
		"route":          mappingRoute(ctx, &item.Route, &diags),
		"source":         types.StringValue(item.Source),
		"source_id":      types.StringValue(item.SourceId.String()),
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingRoute(ctx context.Context, item *api.RouteConnectivity, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_hybrid_bridge.RouteValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_hybrid_bridge.NewRouteValue(datasource_hybrid_bridge.RouteValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"destination_ip_range": types.StringValue(item.DestinationIpRange),
		"gateway_id":           types.StringValue(item.GatewayId),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}
