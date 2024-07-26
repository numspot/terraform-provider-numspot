package vpcpeering

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

type VpcPeeringsDataSourceModel struct {
	Items               []VpcPeeringDatasourceItemModel `tfsdk:"items"`
	AccepterVpcIpRanges types.List                      `tfsdk:"accepter_vpc_ip_ranges"`
	AccepterVpcVpcIds   types.List                      `tfsdk:"accepter_vpc_vpc_ids"`
	ExpirationDates     types.List                      `tfsdk:"expiration_dates"`
	Ids                 types.List                      `tfsdk:"ids"`
	SourceVpcIpRanges   types.List                      `tfsdk:"source_vpc_ip_ranges"`
	SourceVpcVpcIds     types.List                      `tfsdk:"source_vpc_vpc_ids"`
	StateMessages       types.List                      `tfsdk:"state_messages"`
	StateNames          types.List                      `tfsdk:"state_names"`
	TagKeys             types.List                      `tfsdk:"tag_keys"`
	TagValues           types.List                      `tfsdk:"tag_values"`
	Tags                types.List                      `tfsdk:"tags"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &vpcPeeringsDataSource{}
)

func (d *vpcPeeringsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVpcPeeringsDataSource() datasource.DataSource {
	return &vpcPeeringsDataSource{}
}

type vpcPeeringsDataSource struct {
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *vpcPeeringsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_peerings"
}

// Schema defines the schema for the data source.
func (d *vpcPeeringsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = VpcPeeringDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vpcPeeringsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VpcPeeringsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := VpcPeeringsFromTfToAPIReadParams(ctx, plan)
	res := utils2.ExecuteRequest(func() (*numspot.ReadVpcPeeringsResponse, error) {
		return d.provider.GetNumspotClient().ReadVpcPeeringsWithResponse(ctx, d.provider.GetSpaceID(), &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty VPC Peering list")
	}

	objectItems, diags := utils2.FromHttpGenericListToTfList(ctx, res.JSON200.Items, VpcPeeringsFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
