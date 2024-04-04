package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_load_balancer"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type loadBalancersDataSourceModel struct {
	LoadBalancers     []datasource_load_balancer.LoadBalancerModel `tfsdk:"load_balancers"`
	LoadBalancerNames types.List                                   `tfsdk:"load_balancer_names"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &loadBalancersDataSource{}
)

func (d *loadBalancersDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
func NewLoadBalancersDataSource() datasource.DataSource {
	return &loadBalancersDataSource{}
}

// coffeesDataSource is the data source implementation.
type loadBalancersDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *loadBalancersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancers"
}

// Schema defines the schema for the data source.
func (d *loadBalancersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_load_balancer.LoadBalancerDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *loadBalancersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan loadBalancersDataSourceModel
	request.Config.Get(ctx, &plan)

	params := iaas.ReadLoadBalancersParams{}
	if !plan.LoadBalancerNames.IsNull() {
		lbNames := utils.TfStringListToStringList(ctx, plan.LoadBalancerNames)
		params.LoadBalancerNames = &lbNames
	}
	res := utils.ExecuteRequest(func() (*iaas.ReadLoadBalancersResponse, error) {
		return d.provider.ApiClient.ReadLoadBalancersWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty load balancers list")
	}

	for _, item := range *res.JSON200.Items {
		tf := LoadBalancerFromHttpToTfDatasource(ctx, &item)
		state.LoadBalancers = append(state.LoadBalancers, tf)
	}
	state.LoadBalancerNames = plan.LoadBalancerNames
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
