package subnet

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type SubnetsDataSourceModel struct {
	Items                 []SubnetModel `tfsdk:"items"`
	AvailabilityZoneNames types.List    `tfsdk:"availability_zone_names"`
	AvailableIpsCounts    types.List    `tfsdk:"available_ips_counts"`
	Ids                   types.List    `tfsdk:"ids"`
	IpRanges              types.List    `tfsdk:"ip_ranges"`
	States                types.List    `tfsdk:"states"`
	TagKeys               types.List    `tfsdk:"tag_keys"`
	TagValues             types.List    `tfsdk:"tag_values"`
	Tags                  types.List    `tfsdk:"tags"`
	VpcIds                types.List    `tfsdk:"vpc_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &subnetsDataSource{}
)

func (d *subnetsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewSubnetsDataSource() datasource.DataSource {
	return &subnetsDataSource{}
}

type subnetsDataSource struct {
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *subnetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnets"
}

// Schema defines the schema for the data source.
func (d *subnetsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SubnetDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *subnetsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan SubnetsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := SubnetsFromTfToAPIReadParams(ctx, plan, &response.Diagnostics)
	res := utils.ExecuteRequest(func() (*numspot.ReadSubnetsResponse, error) {
		return d.provider.GetNumspotClient().ReadSubnetsWithResponse(ctx, d.provider.GetSpaceID(), &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Subnets list")
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, SubnetsFromHttpToTfDatasource, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
