package routetable

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

type RouteTablesDataSourceModel struct {
	Items                           []RouteTableModelDatasource `tfsdk:"items"`
	Ids                             types.List                  `tfsdk:"ids"`
	LinkRouteTableIds               types.List                  `tfsdk:"link_route_table_ids"`
	LinkRouteTableLinkRouteTableIds types.List                  `tfsdk:"link_route_table_link_route_table_ids"`
	LinkRouteTableMain              types.Bool                  `tfsdk:"link_route_table_main"`
	LinkSubnetIds                   types.List                  `tfsdk:"link_subnet_ids"`
	RouteCreationMethods            types.List                  `tfsdk:"route_creation_methods"`
	RouteDestinationIpRanges        types.List                  `tfsdk:"route_destination_ip_ranges"`
	RouteDestinationServiceIds      types.List                  `tfsdk:"route_destination_service_ids"`
	RouteGatewayIds                 types.List                  `tfsdk:"route_gateway_ids"`
	RouteNatGatewayIds              types.List                  `tfsdk:"route_nat_gateway_ids"`
	RouteStates                     types.List                  `tfsdk:"route_states"`
	RouteVmIds                      types.List                  `tfsdk:"route_vm_ids"`
	RouteVpcPeeringIds              types.List                  `tfsdk:"route_vpc_peering_ids"`
	TagKeys                         types.List                  `tfsdk:"tag_keys"`
	TagValues                       types.List                  `tfsdk:"tag_values"`
	Tags                            types.List                  `tfsdk:"tags"`
	VpcIds                          types.List                  `tfsdk:"vpc_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &routeTablesDataSource{}
)

func (d *routeTablesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewRouteTablesDataSource() datasource.DataSource {
	return &routeTablesDataSource{}
}

type routeTablesDataSource struct {
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *routeTablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_tables"
}

// Schema defines the schema for the data source.
func (d *routeTablesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = RouteTableDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *routeTablesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan RouteTablesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := RouteTablesFromTfToAPIReadParams(ctx, plan)
	res := utils2.ExecuteRequest(func() (*numspot.ReadRouteTablesResponse, error) {
		return d.provider.GetNumspotClient().ReadRouteTablesWithResponse(ctx, d.provider.GetSpaceID(), &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Route Table list")
	}

	objectItems, diags := utils2.FromHttpGenericListToTfList(ctx, res.JSON200.Items, RouteTablesFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
