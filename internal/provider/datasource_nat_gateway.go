package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_nat_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type NatGatewaysDataSourceModel struct {
	NatGateways []datasource_nat_gateway.NatGatewayModel `tfsdk:"nat_gateways"`
	IDs         types.List                               `tfsdk:"ids"`
	States      types.List                               `tfsdk:"states"`
	TagKeys     types.List                               `tfsdk:"tag_keys"`
	TagValues   types.List                               `tfsdk:"tag_values"`
	Tags        types.List                               `tfsdk:"tags"`
	SubnetIds   types.List                               `tfsdk:"subnet_ids"`
	VpcIds      types.List                               `tfsdk:"vpc_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &natGatewaysDataSource{}
)

func (d *natGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
func NewNatGatewaysDataSource() datasource.DataSource {
	return &natGatewaysDataSource{}
}

// coffeesDataSource is the data source implementation.
type natGatewaysDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *natGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_gateways"
}

// Schema defines the schema for the data source.
func (d *natGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_nat_gateway.NatGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *natGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan NatGatewaysDataSourceModel
	request.Config.Get(ctx, &plan)

	params := NatGatewaysFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadNatGatewayResponse, error) {
		return d.provider.ApiClient.ReadNatGatewayWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Nat Gateways list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := NatGatewaysFromHttpToTfDatasource(ctx, &item)
		if diags != nil {
			response.Diagnostics.AddError("Error while converting Nat Gateway HTTP object to Terraform object", diags.Errors()[0].Detail())
		}
		state.NatGateways = append(state.NatGateways, *tf)
	}
	state.IDs = plan.IDs
	state.States = plan.States
	state.Tags = plan.Tags
	state.TagKeys = plan.TagKeys
	state.TagValues = plan.TagValues
	state.SubnetIds = plan.SubnetIds
	state.VpcIds = plan.VpcIds

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
