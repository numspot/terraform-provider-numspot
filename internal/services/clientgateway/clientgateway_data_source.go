package clientgateway

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
	"terraform-provider-numspot/internal/services/clientgateway/datasource_client_gateway"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &clientGatewaysDataSource{}
)

func (d *clientGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewClientGatewaysDataSource() datasource.DataSource {
	return &clientGatewaysDataSource{}
}

type clientGatewaysDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *clientGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_gateways"
}

// Schema defines the schema for the data source.
func (d *clientGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_client_gateway.ClientGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *clientGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_client_gateway.ClientGatewayModel

	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	clientGateways, err := core.ReadClientGateways(ctx, d.provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read client gateways", err.Error())
		return
	}

	clientGatewayItems := serializeClientGatewaysDatasource(ctx, clientGateways, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = clientGatewayItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func serializeClientGatewaysDatasource(ctx context.Context, clientGateways []api.ClientGateway, diags *diag.Diagnostics) datasource_client_gateway.ClientGatewayModel {
	clientGatewayList := types.ListNull(new(datasource_client_gateway.ItemsValue).Type(ctx))
	var serializeDiags diag.Diagnostics

	if len(clientGateways) > 0 {
		ll := len(clientGateways)
		itemsValue := make([]datasource_client_gateway.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			itemsValue[i], serializeDiags = datasource_client_gateway.NewItemsValue(datasource_client_gateway.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"bgp_asn":         types.Int64Value(int64(clientGateways[i].BgpAsn)),
				"connection_type": types.StringValue(clientGateways[i].ConnectionType),
				"id":              types.StringValue(clientGateways[i].Id.String()),
				"public_ip":       types.StringValue(clientGateways[i].PublicIp),
				"state":           types.StringValue(clientGateways[i].State),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		clientGatewayList, serializeDiags = types.ListValueFrom(ctx, new(datasource_client_gateway.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}

	}

	return datasource_client_gateway.ClientGatewayModel{
		Items: clientGatewayList,
	}
}
