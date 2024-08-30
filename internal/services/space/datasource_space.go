package space

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &spaceDataSource{}
)

func (d *spaceDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewSpaceDataSource() datasource.DataSource {
	return &spaceDataSource{}
}

type spaceDataSource struct {
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *spaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

// Schema defines the schema for the data source.
func (d *spaceDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SpaceDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *spaceDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan SpaceModelDataSource
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	spaceId, err := uuid.Parse(plan.SpaceId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid space id", fmt.Sprintf("space id should be in UUID format but was '%s'", plan.SpaceId))
		return
	}

	organisationId, err := uuid.Parse(plan.OrganisationId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid organisation id", fmt.Sprintf("organisation id should be in UUID format but was '%s'", plan.OrganisationId))
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.GetSpaceByIdResponse, error) {
		return d.provider.GetNumspotClient().GetSpaceByIdWithResponse(ctx, organisationId, spaceId)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := SpaceFromHttpToTfDatasource(ctx, res.JSON200)
	if diags != nil {
		response.Diagnostics.AddError("Error while converting Space HTTP object to Terraform object", diags.Errors()[0].Detail())
	}
	state = *tf

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
