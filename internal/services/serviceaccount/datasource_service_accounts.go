package serviceaccount

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type ServiceAccountsDataSourceModel struct {
	Items              []ServiceAccountModel `tfsdk:"items"`
	SpaceID            types.String          `tfsdk:"space_id"`
	ServiceAccountName types.String          `tfsdk:"service_account_name"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &serviceAccountsDataSource{}
)

func (d *serviceAccountsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewServiceAccountsDataSource() datasource.DataSource {
	return &serviceAccountsDataSource{}
}

type serviceAccountsDataSource struct {
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *serviceAccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_accounts"
}

// Schema defines the schema for the data source.
func (d *serviceAccountsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ServiceAccountDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *serviceAccountsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan ServiceAccountsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	spaceId, err := uuid.Parse(plan.SpaceID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space id", fmt.Sprintf("space id should be in UUID format but was '%s'", plan.SpaceID.ValueString()))
		return
	}

	serviceAccounts := []ServiceAccountModel{}
	params := ServiceAccountsFromTfToAPIReadParams(plan)
	d.fetchPaginatedServiceAccounts(ctx, spaceId, &params, &serviceAccounts, response)

	state.Items = serviceAccounts
	state.SpaceID = plan.SpaceID
	state.ServiceAccountName = plan.ServiceAccountName

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (d *serviceAccountsDataSource) fetchPaginatedServiceAccounts(
	ctx context.Context,
	spaceID uuid.UUID,
	requestParams *numspot.ListServiceAccountSpaceParams,
	svcAccountsHolder *[]ServiceAccountModel,
	response *datasource.ReadResponse,
) {
	body := numspot.ListServiceAccountSpaceJSONRequestBody{}
	res := utils.ExecuteRequest(func() (*numspot.ListServiceAccountSpaceResponse, error) {
		return d.provider.GetNumspotClient().ListServiceAccountSpaceWithResponse(ctx, spaceID, requestParams, body)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Subnets list")
	}

	for _, item := range res.JSON200.Items {
		tf := ServiceAccountEditedResponseFromHTTPToTFDataSource(item)
		*svcAccountsHolder = append(*svcAccountsHolder, tf)
	}

	if res.JSON200.NextPageToken != nil {
		requestParams.Page = new(numspot.ListServiceAccounts)
		requestParams.Page.NextToken = res.JSON200.NextPageToken
		d.fetchPaginatedServiceAccounts(ctx, spaceID, requestParams, svcAccountsHolder, response)
	}
}
