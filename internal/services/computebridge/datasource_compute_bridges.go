package computebridge

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/computebridge/datasource_compute_bridge"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &computebridgesDataSource{}

// computebridgesDataSource is the Terraform data source implementation for numspot_compute_bridges.
type computebridgesDataSource struct {
	services.BaseDataSource
}

// NewComputeBridgesDataSource returns a new computebridgesDataSource instance.
func NewComputeBridgesDataSource() datasource.DataSource {
	return &computebridgesDataSource{}
}

// Metadata sets the data source type name.
func (d *computebridgesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_bridges"
}

// Schema defines the schema for the data source.
func (d *computebridgesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_compute_bridge.ComputeBridgeDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *computebridgesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_compute_bridge.ComputeBridgeModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	result, err := core.ReadComputeBridges(ctx, d.Provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read compute_bridges", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.ComputeBridge) (datasource_compute_bridge.ItemsValue, diag.Diagnostics) {
	value, valueDiags := datasource_compute_bridge.NewItemsValue(datasource_compute_bridge.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"destination":          types.StringValue(item.Destination),
		"destination_id":       types.StringValue(item.DestinationId),
		"destination_ip_range": types.StringValue(item.DestinationIpRange),
		"gateway_id":           types.StringValue(item.GatewayId),
		"id":                   types.StringValue(item.Id.String()),
		"source":               types.StringValue(item.Source),
		"source_id":            types.StringValue(item.SourceId),
		"source_ip_range":      types.StringValue(item.SourceIpRange),
	})
	return value, valueDiags
}
