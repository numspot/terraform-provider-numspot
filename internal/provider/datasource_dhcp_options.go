package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_dhcp_options"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
	"net/http"
)

type DHCPOptionsDataSourceModel struct {
	DHCPOptions       []datasource_dhcp_options.DhcpOptionsModel `tfsdk:"dhcp_options"`
	IDs               types.List                                 `tfsdk:"ids"`
	Default           types.Bool                                 `tfsdk:"default"`
	DomainNameServers types.List                                 `tfsdk:"domain_name_servers"`
	DomainNames       types.List                                 `tfsdk:"domain_names"`
	LogServers        types.List                                 `tfsdk:"log_servers"`
	NTPServers        types.List                                 `tfsdk:"ntp_servers"`
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

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewDHCPOptionsDataSource() datasource.DataSource {
	return &dhcpOptionsDataSource{}
}

// coffeesDataSource is the data source implementation.
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
	request.Config.Get(ctx, &plan)

	params := DhcpOptionsFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*api.ReadDhcpOptionsResponse, error) {
		return d.provider.ApiClient.ReadDhcpOptionsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty DHCP options list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := DHCPOptionsFromHttpToTfDatasource(ctx, &item)
		if diags != nil {
			response.Diagnostics.AddError("Error while converting DHCPOptions HTTP object to Terraform object", diags.Errors()[0].Detail())
		}
		state.DHCPOptions = append(state.DHCPOptions, *tf)
	}
	state.IDs = plan.IDs
	state.NTPServers = plan.NTPServers
	state.DomainNames = plan.DomainNames
	state.DomainNameServers = plan.DomainNameServers
	state.LogServers = plan.LogServers
	state.Default = plan.Default
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
