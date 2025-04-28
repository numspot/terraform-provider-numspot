package computebridge

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
	"terraform-provider-numspot/internal/services/computebridge/datasource_compute_bridge"
)

var _ datasource.DataSource = &computeBridgeDataSource{}

type computeBridgeDataSource struct {
	provider *client.NumSpotSDK
}

func NewComputeBridgeDataSource() datasource.DataSource {
	return &computeBridgeDataSource{}
}

func (d *computeBridgeDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

func (d *computeBridgeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_bridges"
}

func (d *computeBridgeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_compute_bridge.ComputeBridgeDataSourceSchema(ctx)
}

func (d *computeBridgeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state, plan datasource_compute_bridge.ComputeBridgeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := core.ReadComputeBridges(ctx, d.provider)
	if err != nil {
		resp.Diagnostics.AddError("Error reading compute bridges", err.Error())
		return
	}

	computeBridgeItems := serializeComputeBridges(ctx, read, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = computeBridgeItems.Items

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func serializeComputeBridges(ctx context.Context, computeBridges *api.ComputeBridges, diags *diag.Diagnostics) datasource_compute_bridge.ComputeBridgeModel {
	computeBridgeList := types.ListNull(new(datasource_compute_bridge.ItemsValue).Type(ctx))
	var serializeDiags diag.Diagnostics

	if len((*computeBridges).Items) != 0 {
		ll := len((*computeBridges).Items)
		itemsValue := make([]datasource_compute_bridge.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			itemsValue[i], serializeDiags = datasource_compute_bridge.NewItemsValue(datasource_compute_bridge.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"destination_ip_range": types.StringValue((*computeBridges).Items[i].DestinationIpRange),
				"gateway_id":           types.StringValue((*computeBridges).Items[i].GatewayId),
				"id":                   types.StringValue((*computeBridges).Items[i].Id.String()),
				"source_ip_range":      types.StringValue((*computeBridges).Items[i].SourceIpRange),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		computeBridgeList, serializeDiags = types.ListValueFrom(ctx, new(datasource_compute_bridge.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	return datasource_compute_bridge.ComputeBridgeModel{
		Items: computeBridgeList,
	}
}
