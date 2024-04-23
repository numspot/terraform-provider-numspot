package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_virtual_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VirtualGatewaysDataSourceModel struct {
	VirtualGateways []datasource_virtual_gateway.VirtualGatewayModel `tfsdk:"virtual_gateways"`
	ConnectionTypes types.List                                       `tfsdk:"connection_types"`
	LinkStates      types.List                                       `tfsdk:"link_states"`
	States          types.List                                       `tfsdk:"states"`
	TagKeys         types.List                                       `tfsdk:"tag_keys"`
	TagValues       types.List                                       `tfsdk:"tag_values"`
	Tags            types.List                                       `tfsdk:"tags"`
	LinkVpcIds      types.List                                       `tfsdk:"link_vpc_ids"`
	IDs             types.List                                       `tfsdk:"ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &virtualGatewaysDataSource{}
)

func (d *virtualGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVirtualGatewaysDataSource() datasource.DataSource {
	return &virtualGatewaysDataSource{}
}

type virtualGatewaysDataSource struct {
	provider Provider
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
	var state, plan VirtualGatewaysDataSourceModel
	request.Config.Get(ctx, &plan)

	params := VirtualGatewaysFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadVirtualGatewaysResponse, error) {
		return d.provider.ApiClient.ReadVirtualGatewaysWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Virtual Gateways list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := VirtualGatewaysFromHttpToTfDatasource(ctx, &item)
		if diags != nil {
			response.Diagnostics.AddError("Error while converting Virtual Gateway HTTP object to Terraform object", diags.Errors()[0].Detail())
		}
		state.VirtualGateways = append(state.VirtualGateways, *tf)
	}
	state.IDs = plan.IDs
	state.States = plan.States
	state.Tags = plan.Tags
	state.TagKeys = plan.TagKeys
	state.TagValues = plan.TagValues
	state.ConnectionTypes = plan.ConnectionTypes
	state.LinkStates = plan.LinkStates
	state.LinkVpcIds = plan.LinkVpcIds

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
