package serviceaccount

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type ServiceAccountsDataSourceModel struct {
	Items             []ServiceAccountDataSourceModel `tfsdk:"items"`
	SpaceID           types.String                    `tfsdk:"space_id"`
	ServiceAccountIDs types.Set                       `tfsdk:"service_account_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &serviceAccountsDataSource{}
)

func (d *serviceAccountsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
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
	provider *client.NumSpotSDK
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

	serviceAccounts := []ServiceAccountDataSourceModel{}
	d.fetchPaginatedServiceAccounts(ctx, spaceId, &plan, &serviceAccounts, response)

	state.Items = serviceAccounts
	state.SpaceID = plan.SpaceID
	state.ServiceAccountIDs = plan.ServiceAccountIDs

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func getServiceAccountIDs(ctx context.Context, tfIDs types.Set) ([]openapi_types.UUID, error) {
	strIDs := make([]string, len(tfIDs.Elements()))
	uuids := make([]openapi_types.UUID, len(tfIDs.Elements()))
	tfIDs.ElementsAs(ctx, &strIDs, false)
	for i := range strIDs {
		svcAccUUID, err := uuid.Parse(strIDs[i])
		if err != nil {
			return nil, err
		}
		uuids[i] = svcAccUUID
	}
	return uuids, nil
}

func (d *serviceAccountsDataSource) fetchPaginatedServiceAccounts(
	ctx context.Context,
	spaceID uuid.UUID,
	plan *ServiceAccountsDataSourceModel,
	svcAccountsHolder *[]ServiceAccountDataSourceModel,
	response *datasource.ReadResponse,
) {
	numspotClient, err := d.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	svcAccountIDs, err := getServiceAccountIDs(ctx, plan.ServiceAccountIDs)
	if err != nil {
		response.Diagnostics.AddError("failed to deserialize service accounts IDs", err.Error())
		return
	}
	body := numspot.ListServiceAccountSpaceJSONRequestBody{Items: svcAccountIDs}
	params := ServiceAccountsFromTfToAPIReadParams(*plan)

	res := utils.ExecuteRequest(func() (*numspot.ListServiceAccountSpaceResponse, error) {
		return numspotClient.ListServiceAccountSpaceWithResponse(ctx, spaceID, &params, body)
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
		params.Page = new(numspot.ListServiceAccounts)
		params.Page.NextToken = res.JSON200.NextPageToken
		d.fetchPaginatedServiceAccounts(ctx, spaceID, plan, svcAccountsHolder, response)
	}
}
