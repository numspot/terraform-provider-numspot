package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_vpc"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VPCsDataSourceModel struct {
	VPCs              []datasource_vpc.VpcModel `tfsdk:"vpcs"`
	IDs               types.List                `tfsdk:"ids"`
	DHCPOptionsSetIds types.List                `tfsdk:"dhcp_options_set_ids"`
	IPRanges          types.List                `tfsdk:"ip_ranges"`
	IsDefault         types.Bool                `tfsdk:"is_default"`
	States            types.List                `tfsdk:"states"`
	TagKeys           types.List                `tfsdk:"tag_keys"`
	TagValues         types.List                `tfsdk:"tag_values"`
	Tags              types.List                `tfsdk:"tags"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &vpcsDataSource{}
)

func (d *vpcsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
func NewVPCsDataSource() datasource.DataSource {
	return &vpcsDataSource{}
}

// coffeesDataSource is the data source implementation.
type vpcsDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *vpcsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcs"
}

// Schema defines the schema for the data source.
func (d *vpcsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_vpc.VpcDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vpcsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VPCsDataSourceModel
	request.Config.Get(ctx, &plan)

	params := VPCsFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadVpcsResponse, error) {
		return d.provider.ApiClient.ReadVpcsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty VPCs list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := VPCsFromHttpToTfDatasource(ctx, &item)
		if diags != nil {
			response.Diagnostics.AddError("Error while converting VPC HTTP object to Terraform object", diags.Errors()[0].Detail())
		}
		state.VPCs = append(state.VPCs, *tf)
	}
	state.IDs = plan.IDs
	state.States = plan.States
	state.IPRanges = plan.IPRanges
	state.IsDefault = plan.IsDefault
	state.DHCPOptionsSetIds = plan.DHCPOptionsSetIds
	state.Tags = plan.Tags
	state.TagKeys = plan.TagKeys
	state.TagValues = plan.TagValues

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
