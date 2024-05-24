package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_public_ip"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type PublicIpsDataSourceModel struct {
	Items           []datasource_public_ip.PublicIpModel `tfsdk:"items"`
	LinkPublicIpIds types.List                           `tfsdk:"link_public_ip_ids"`
	NicIds          types.List                           `tfsdk:"nic_ids"`
	TagKeys         types.List                           `tfsdk:"tag_keys"`
	TagValues       types.List                           `tfsdk:"tag_values"`
	Tags            types.List                           `tfsdk:"tags"`
	PrivateIps      types.List                           `tfsdk:"private_ips"`
	VmIds           types.List                           `tfsdk:"vm_ids"`
	IDs             types.List                           `tfsdk:"ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &publicIpsDataSource{}
)

func (d *publicIpsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewPublicIpsDataSource() datasource.DataSource {
	return &publicIpsDataSource{}
}

type publicIpsDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *publicIpsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ips"
}

// Schema defines the schema for the data source.
func (d *publicIpsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_public_ip.PublicIpDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *publicIpsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan PublicIpsDataSourceModel
	request.Config.Get(ctx, &plan)

	params := PublicIpsFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadPublicIpsResponse, error) {
		return d.provider.ApiClient.ReadPublicIpsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Public Ips list")
	}

	objectItems, diags := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, PublicIpsFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
