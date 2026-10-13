package clientgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/clientgateway/datasource_client_gateway"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &clientgatewaysDataSource{}

// clientgatewaysDataSource is the Terraform data source implementation for numspot_client_gateways.
type clientgatewaysDataSource struct {
	services.BaseDataSource
}

// NewClientGatewaysDataSource returns a new clientgatewaysDataSource instance.
func NewClientGatewaysDataSource() datasource.DataSource {
	return &clientgatewaysDataSource{}
}

// Metadata sets the data source type name.
func (d *clientgatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_gateways"
}

// Schema defines the schema for the data source.
func (d *clientgatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_client_gateway.ClientGatewayDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *clientgatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	result, err := core.ReadClientGateways(ctx, d.Provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read client_gateways", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.ClientGateway) (datasource_client_gateway.ItemsValue, diag.Diagnostics) {
	value, valueDiags := datasource_client_gateway.NewItemsValue(datasource_client_gateway.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"bgp_asn":         utils.FromIntToTfInt64(item.BgpAsn),
		"connection_type": types.StringValue(item.ConnectionType),
		"id":              types.StringValue(item.Id.String()),
		"public_ip":       types.StringValue(item.PublicIp),
		"state":           types.StringValue(item.State),
	})
	return value, valueDiags
}
