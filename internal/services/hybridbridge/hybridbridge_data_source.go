package hybridbridge

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/hybridbridge/datasource_hybrid_bridge"
)

var _ datasource.DataSource = &hybridBridgeDataSource{}

func NewHybridBridgeDataSource() datasource.DataSource {
	return &hybridBridgeDataSource{}
}

func (d *hybridBridgeDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

type hybridBridgeDataSource struct {
	provider *client.NumSpotSDK
}

func (d *hybridBridgeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hybrid_bridges"
}

func (d *hybridBridgeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_hybrid_bridge.HybridBridgeDataSourceSchema(ctx)
}

func (d *hybridBridgeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state, plan datasource_hybrid_bridge.HybridBridgeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := core.ReadHybridBridges(ctx, d.provider)
	if err != nil {
		resp.Diagnostics.AddError("Error reading hybrid bridges", err.Error())
		return
	}

	serverCertificateItems := serializeHybridBridges(ctx, read.Items, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = serverCertificateItems.Items

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func serializeHybridBridges(ctx context.Context, hybridBridges []api.HybridBridge, diags *diag.Diagnostics) datasource_hybrid_bridge.HybridBridgeModel {
	hybridBridgeList := types.List{}

	if len(hybridBridges) != 0 {
		ll := len(hybridBridges)
		itemsValue := make([]datasource_hybrid_bridge.ItemsValue, ll)

		for i := 0; ll > i; i++ {

			routeObj, serializeDiags := datasource_hybrid_bridge.NewRouteValue(datasource_hybrid_bridge.RouteValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"destination_ip_range": types.StringValue(hybridBridges[i].Route.DestinationIpRange),
				"gateway_id":           types.StringValue(hybridBridges[i].Route.GatewayId),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}

			itemsValue[i], serializeDiags = datasource_hybrid_bridge.NewItemsValue(datasource_hybrid_bridge.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"id":    types.StringValue(hybridBridges[i].Id.String()),
				"route": routeObj,
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}

			hybridBridgeList, serializeDiags = types.ListValueFrom(ctx, new(datasource_hybrid_bridge.ItemsValue).Type(ctx), itemsValue)
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
			}
		}
	} else {
		hybridBridgeList = types.ListNull(new(datasource_hybrid_bridge.ItemsValue).Type(ctx))
	}

	return datasource_hybrid_bridge.HybridBridgeModel{
		Items: hybridBridgeList,
	}
}
