package clientgateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type ClientGatewaysDataSourceModel struct {
	Items           []ClientGatewayModel `tfsdk:"items"`
	BgpAsns         types.List           `tfsdk:"bgp_asns"`
	ConnectionTypes types.List           `tfsdk:"connection_types"`
	IDs             types.List           `tfsdk:"ids"`
	PublicIps       types.List           `tfsdk:"public_ips"`
	States          types.List           `tfsdk:"states"`
	TagKeys         types.List           `tfsdk:"tag_keys"`
	TagValues       types.List           `tfsdk:"tag_values"`
	Tags            types.List           `tfsdk:"tags"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &clientGatewaysDataSource{}
)

func (d *clientGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(services.IProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
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
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *clientGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_gateways"
}

// Schema defines the schema for the data source.
func (d *clientGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ClientGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *clientGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan ClientGatewaysDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := ClientGatewaysFromTfToAPIReadParams(ctx, plan)
	res := utils2.ExecuteRequest(func() (*numspot.ReadClientGatewaysResponse, error) {
		return d.provider.GetNumspotClient().ReadClientGatewaysWithResponse(ctx, d.provider.GetSpaceID(), &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Client Gateways list")
	}

	objectItems, diags := utils2.FromHttpGenericListToTfList(ctx, res.JSON200.Items, ClientGatewaysFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
