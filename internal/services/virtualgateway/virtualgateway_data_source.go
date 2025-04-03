package virtualgateway

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
	"terraform-provider-numspot/internal/services/virtualgateway/datasource_virtual_gateway"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &virtualGatewaysDataSource{}
)

func (d *virtualGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVirtualGatewaysDataSource() datasource.DataSource {
	return &virtualGatewaysDataSource{}
}

type virtualGatewaysDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *virtualGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_gateways"
}

// Schema defines the schema for the data source.
func (d *virtualGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_virtual_gateway.VirtualGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *virtualGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotVirtualGateway, err := core.ReadVirtualGatewaysWithParams(ctx, d.provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read virtual gateways", err.Error())
		return
	}

	objectItems := serializeVirtualGatewayDatasource(ctx, numspotVirtualGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func serializeVirtualGatewayDatasource(ctx context.Context, virtualGateways []api.VirtualGateway, diags *diag.Diagnostics) datasource_virtual_gateway.VirtualGatewayModel {
	virtualGatewaysList := types.ListNull(new(datasource_virtual_gateway.ItemsValue).Type(ctx))
	var serializeDiags diag.Diagnostics

	if len(virtualGateways) > 0 {
		ll := len(virtualGateways)
		itemsValue := make([]datasource_virtual_gateway.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			links := serializeVirtualGatewayLinksValue(ctx, virtualGateways[i].VpcToVirtualGatewayLinks, diags)

			itemsValue[i], serializeDiags = datasource_virtual_gateway.NewItemsValue(datasource_virtual_gateway.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"connection_type":              types.StringValue(virtualGateways[i].ConnectionType),
				"id":                           types.StringValue(virtualGateways[i].Id.String()),
				"state":                        types.StringValue(virtualGateways[i].State),
				"vpc_to_virtual_gateway_links": links,
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		virtualGatewaysList, serializeDiags = types.ListValueFrom(ctx, new(datasource_virtual_gateway.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}

	}

	return datasource_virtual_gateway.VirtualGatewayModel{
		Items: virtualGatewaysList,
	}
}

func serializeVirtualGatewayLinksValue(ctx context.Context, virtualGatewayLinks []api.VpcToVirtualGatewayLink, diags *diag.Diagnostics) types.List {
	linkList := types.ListNull(new(datasource_virtual_gateway.VpcToVirtualGatewayLinksValue).Type(ctx))
	var serializeDiags diag.Diagnostics

	if len(virtualGatewayLinks) > 0 {
		ll := len(virtualGatewayLinks)
		linksValue := make([]datasource_virtual_gateway.VpcToVirtualGatewayLinksValue, ll)

		for i := 0; ll > i; i++ {
			linksValue[i], serializeDiags = datasource_virtual_gateway.NewVpcToVirtualGatewayLinksValue(datasource_virtual_gateway.VpcToVirtualGatewayLinksValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"state":  types.StringValue(virtualGatewayLinks[i].State),
				"vpc_id": types.StringValue(virtualGatewayLinks[i].VpcId),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}

		}

		linkList, serializeDiags = types.ListValueFrom(ctx, new(datasource_virtual_gateway.VpcToVirtualGatewayLinksValue).Type(ctx), linksValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}

	}

	return linkList
}
