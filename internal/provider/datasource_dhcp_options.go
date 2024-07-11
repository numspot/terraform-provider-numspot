package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_dhcp_options"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type DHCPOptionsDataSourceModel struct {
	Items             []datasource_dhcp_options.DhcpOptionsModel `tfsdk:"items"`
	IDs               types.List                                 `tfsdk:"ids"`
	Default           types.Bool                                 `tfsdk:"default"`
	DomainNameServers types.List                                 `tfsdk:"domain_name_servers"`
	DomainNames       types.List                                 `tfsdk:"domain_names"`
	LogServers        types.List                                 `tfsdk:"log_servers"`
	NTPServers        types.List                                 `tfsdk:"ntp_servers"`
	TagKeys           types.List                                 `tfsdk:"tag_keys"`
	TagValues         types.List                                 `tfsdk:"tag_values"`
	Tags              types.List                                 `tfsdk:"tags"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &dhcpOptionsDataSource{}
)

func (d *dhcpOptionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewDHCPOptionsDataSource() datasource.DataSource {
	return &dhcpOptionsDataSource{}
}

type dhcpOptionsDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *dhcpOptionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_options"
}

// Schema defines the schema for the data source.
func (d *dhcpOptionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_dhcp_options.DhcpOptionsDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *dhcpOptionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan DHCPOptionsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := DhcpOptionsFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*numspot.ReadDhcpOptionsResponse, error) {
		return d.provider.NumspotClient.ReadDhcpOptionsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty DHCP options list")
	}

	objectItems, diags := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, DHCPOptionsFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
