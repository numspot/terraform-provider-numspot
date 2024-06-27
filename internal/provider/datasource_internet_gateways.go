package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_internet_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type InternetGatewaysDataSourceModel struct {
	Items      []datasource_internet_gateway.InternetGatewayModel `tfsdk:"items"`
	IDs        types.List                                         `tfsdk:"ids"`
	LinkStates types.List                                         `tfsdk:"link_states"`
	TagKeys    types.List                                         `tfsdk:"tag_keys"`
	TagValues  types.List                                         `tfsdk:"tag_values"`
	Tags       types.List                                         `tfsdk:"tags"`
	LinkVpcIds types.List                                         `tfsdk:"link_vpc_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &internetGatewaysDataSource{}
)

func (d *internetGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(Provider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
}

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewInternetGatewaysDataSource() datasource.DataSource {
	return &internetGatewaysDataSource{}
}

// coffeesDataSource is the data source implementation.
type internetGatewaysDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *internetGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateways"
}

// Schema defines the schema for the data source.
func (d *internetGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_internet_gateway.InternetGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *internetGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan InternetGatewaysDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := InternetGatewaysFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadInternetGatewaysResponse, error) {
		return d.provider.IaasClient.ReadInternetGatewaysWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Internet Gateways list")
	}

	objectItems, diags := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, InternetGatewaysFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
