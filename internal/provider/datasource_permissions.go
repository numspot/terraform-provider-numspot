package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iam"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_permissions"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type PermissionsDataSourceModel struct {
	Items       []datasource_permissions.PermissionModel `tfsdk:"items"`
	SpaceID     types.String                             `tfsdk:"space_id"`
	Action      types.String                             `tfsdk:"action"`
	Resource    types.String                             `tfsdk:"resource"`
	Subresource types.String                             `tfsdk:"subresource"`
	Service     types.String                             `tfsdk:"service"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &serviceAccountsDataSource{}
)

func (d *permissionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewPermissionsDataSource() datasource.DataSource {
	return &permissionsDataSource{}
}

type permissionsDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *permissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions"
}

// Schema defines the schema for the data source.
func (d *permissionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_permissions.PermissionsDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *permissionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan PermissionsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	spaceId, err := uuid.Parse(plan.SpaceID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space id", fmt.Sprintf("space id should be in UUID format but was '%s'", plan.SpaceID.ValueString()))
		return
	}

	permissions := []datasource_permissions.PermissionModel{}
	params := PermissionsFromTfToAPIReadParams(plan)
	d.fetchPaginatedPermissions(ctx, spaceId, &params, &permissions, response)

	state = plan
	state.Items = permissions

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (d *permissionsDataSource) fetchPaginatedPermissions(
	ctx context.Context,
	spaceID uuid.UUID,
	requestParams *iam.ListPermissionsSpaceParams,
	permissionsHolder *[]datasource_permissions.PermissionModel,
	response *datasource.ReadResponse,
) {
	res := utils.ExecuteRequest(func() (*iam.ListPermissionsSpaceResponse, error) {
		return d.provider.IAMAccessManagerClient.ListPermissionsSpaceWithResponse(ctx, spaceID, requestParams)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty permissions list")
	}

	for _, item := range res.JSON200.Items {
		tf := RegisteredPermissionFromHTTPToTFDataSource(item)
		*permissionsHolder = append(*permissionsHolder, tf)
	}

	pagesize := new(int32)
	*pagesize = 15
	if res.JSON200.NextPageToken != nil {
		requestParams.Page = new(iam.ListPermissionsPage)
		requestParams.Page.NextToken = res.JSON200.NextPageToken
		requestParams.Page.Size = pagesize
		d.fetchPaginatedPermissions(ctx, spaceID, requestParams, permissionsHolder, response)
	}
}
