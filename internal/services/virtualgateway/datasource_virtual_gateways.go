package virtualgateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VirtualGatewaysDataSourceModel struct {
	Items           []VirtualGatewayModelItemDataSource `tfsdk:"items"`
	ConnectionTypes types.List                          `tfsdk:"connection_types"`
	LinkStates      types.List                          `tfsdk:"link_states"`
	States          types.List                          `tfsdk:"states"`
	TagKeys         types.List                          `tfsdk:"tag_keys"`
	TagValues       types.List                          `tfsdk:"tag_values"`
	Tags            types.List                          `tfsdk:"tags"`
	LinkVpcIds      types.List                          `tfsdk:"link_vpc_ids"`
	IDs             types.List                          `tfsdk:"ids"`
}

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
			"Unexpected Resource Configure Type",
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
	resp.Schema = VirtualGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *virtualGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VirtualGatewaysDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := d.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	params := VirtualGatewaysFromTfToAPIReadParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.ReadVirtualGatewaysResponse, error) {
		return numspotClient.ReadVirtualGatewaysWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Virtual Gateways list")
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, VirtualGatewayDataSourceFromHttpToTf, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
