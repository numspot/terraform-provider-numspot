package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_vpc_peering"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VpcPeeringsDataSourceModel struct {
	Items                 []datasource_vpc_peering.VpcPeeringModel `tfsdk:"items"`
	ExpirationDates       types.List                               `tfsdk:"expiration_dates"`
	StateMessages         types.List                               `tfsdk:"state_messages"`
	StateNames            types.List                               `tfsdk:"state_names"`
	AccepterVpcAccountIds types.List                               `tfsdk:"accepter_vpc_account_ids"`
	AccepterVpcIpRanges   types.List                               `tfsdk:"accepter_vpc_ip_ranges"`
	AccepterVpcVpcIds     types.List                               `tfsdk:"accepter_vpc_vpc_ids"`
	IDs                   types.List                               `tfsdk:"ids"`
	SourceVpcAccountIds   types.List                               `tfsdk:"source_vpc_account_ids"`
	SourceVpcIpRanges     types.List                               `tfsdk:"source_vpc_ip_ranges"`
	SourceVpcVpcIds       types.List                               `tfsdk:"source_vpc_vpc_ids"`
	TagKeys               types.List                               `tfsdk:"tag_keys"`
	TagValues             types.List                               `tfsdk:"tag_values"`
	Tags                  types.List                               `tfsdk:"tags"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &vpcPeeringsDataSource{}
)

func (d *vpcPeeringsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
func NewVpcPeeringsDataSource() datasource.DataSource {
	return &vpcPeeringsDataSource{}
}

// coffeesDataSource is the data source implementation.
type vpcPeeringsDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *vpcPeeringsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_peerings"
}

// Schema defines the schema for the data source.
func (d *vpcPeeringsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_vpc_peering.VpcPeeringDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vpcPeeringsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VpcPeeringsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := VpcPeeringsFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadVpcPeeringsResponse, error) {
		return d.provider.IaasClient.ReadVpcPeeringsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty VPC Peering list")
	}

	objectItems, diags := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, VpcPeeringsFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
