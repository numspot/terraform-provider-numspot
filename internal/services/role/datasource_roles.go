package role

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type RolesDataSourceModel struct {
	Items   []RolesModel `tfsdk:"items"`
	SpaceID types.String `tfsdk:"space_id"`
	Name    types.String `tfsdk:"name"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &rolesDataSource{}
)

func (d *rolesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewRolesDatasource() datasource.DataSource {
	return &rolesDataSource{}
}

type rolesDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

// Schema defines the schema for the data source.
func (d *rolesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = RolesDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *rolesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan RolesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	spaceId, err := uuid.Parse(plan.SpaceID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space id", fmt.Sprintf("space id should be in UUID format but was '%s'", plan.SpaceID.ValueString()))
		return
	}

	roles := []RolesModel{}
	params := RolesFromTfToAPIReadParams(plan)
	d.fetchPaginatedRoles(ctx, spaceId, &params, &roles, response)

	state = plan
	state.Items = roles

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (d *rolesDataSource) fetchPaginatedRoles(
	ctx context.Context,
	spaceID uuid.UUID,
	requestParams *numspot.ListRolesSpaceParams,
	permissionsHolder *[]RolesModel,
	response *datasource.ReadResponse,
) {
	var pageSize int32 = 50
	numspotClient, err := d.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.ListRolesSpaceResponse, error) {
		return numspotClient.ListRolesSpaceWithResponse(ctx, spaceID, requestParams)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty roles list")
	}

	for _, item := range res.JSON200.Items {
		tf := RegisteredRoleFromHTTPToTFDataSource(item)
		*permissionsHolder = append(*permissionsHolder, tf)
	}

	if res.JSON200.NextPageToken != nil {
		requestParams.Page = new(numspot.ListRolesPage)
		requestParams.Page.NextToken = res.JSON200.NextPageToken
		requestParams.Page.Size = &pageSize
		d.fetchPaginatedRoles(ctx, spaceID, requestParams, permissionsHolder, response)
	}
}
