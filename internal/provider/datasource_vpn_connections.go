package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_vpn_connection"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VpnConnectionsDataSourceModel struct {
	Items                    []datasource_vpn_connection.VpnConnectionModel `tfsdk:"items"`
	BgpAsns                  types.List                                     `tfsdk:"bgp_asns"`
	ClientGatewayIds         types.List                                     `tfsdk:"client_gateway_ids"`
	ConnectionTypes          types.List                                     `tfsdk:"connection_types"`
	RouteDestinationIpRanges types.List                                     `tfsdk:"routes_destination_ip_ranges"`
	StaticRouteOnly          types.Bool                                     `tfsdk:"static_routes_only"`
	VirtualGatewayIds        types.List                                     `tfsdk:"virtual_gateway_ids"`
	States                   types.List                                     `tfsdk:"states"`
	TagKeys                  types.List                                     `tfsdk:"tag_keys"`
	TagValues                types.List                                     `tfsdk:"tag_values"`
	Tags                     types.List                                     `tfsdk:"tags"`
	Ids                      types.List                                     `tfsdk:"ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &vpnConnectionsDataSource{}
)

func (d *vpnConnectionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVpnConnectionsDataSource() datasource.DataSource {
	return &vpnConnectionsDataSource{}
}

type vpnConnectionsDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *vpnConnectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_connections"
}

// Schema defines the schema for the data source.
func (d *vpnConnectionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_vpn_connection.VpnConnectionDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vpnConnectionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VpnConnectionsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := VpnConnectionsFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*numspot.ReadVpnConnectionsResponse, error) {
		return d.provider.NumspotClient.ReadVpnConnectionsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Vpn Connection list")
	}

	objectItems, diags := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, VpnConnectionsFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
