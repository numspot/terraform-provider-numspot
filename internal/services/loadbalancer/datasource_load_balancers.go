package loadbalancer

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

type loadBalancersDataSourceModel struct {
	Items             []LoadBalancerModelDatasource `tfsdk:"items"`
	LoadBalancerNames types.List                    `tfsdk:"load_balancer_names"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &loadBalancersDataSource{}
)

func (d *loadBalancersDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewLoadBalancersDataSource() datasource.DataSource {
	return &loadBalancersDataSource{}
}

type loadBalancersDataSource struct {
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *loadBalancersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancers"
}

// Schema defines the schema for the data source.
func (d *loadBalancersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = LoadBalancerDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *loadBalancersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan loadBalancersDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := numspot.ReadLoadBalancersParams{}
	if !plan.LoadBalancerNames.IsNull() {
		lbNames := utils2.TfStringListToStringList(ctx, plan.LoadBalancerNames)
		params.LoadBalancerNames = &lbNames
	}
	res := utils2.ExecuteRequest(func() (*numspot.ReadLoadBalancersResponse, error) {
		return d.provider.GetNumspotClient().ReadLoadBalancersWithResponse(ctx, d.provider.GetSpaceID(), &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty load balancers list")
	}

	objectItems, diags := utils2.FromHttpGenericListToTfList(ctx, res.JSON200.Items, LoadBalancerFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
