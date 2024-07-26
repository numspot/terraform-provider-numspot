package snapshot

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

type SnapshotsDataSourceModel struct {
	Items            []SnapshotModelDatasource `tfsdk:"items"`
	Descriptions     types.List                `tfsdk:"descriptions"`
	FromCreationDate types.String              `tfsdk:"from_creation_date"`
	Ids              types.List                `tfsdk:"ids"`
	IsPublic         types.Bool                `tfsdk:"is_public"`
	Progresses       types.List                `tfsdk:"progresses"`
	States           types.List                `tfsdk:"states"`
	TagKeys          types.List                `tfsdk:"tag_keys"`
	TagValues        types.List                `tfsdk:"tag_values"`
	Tags             types.List                `tfsdk:"tags"`
	ToCreationDate   types.String              `tfsdk:"to_creation_date"`
	VolumeIds        types.List                `tfsdk:"volume_ids"`
	VolumeSizes      types.List                `tfsdk:"volume_sizes"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &snapshotsDataSource{}
)

func (d *snapshotsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewSnapshotsDataSource() datasource.DataSource {
	return &snapshotsDataSource{}
}

type snapshotsDataSource struct {
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *snapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshots"
}

// Schema defines the schema for the data source.
func (d *snapshotsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SnapshotDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *snapshotsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan SnapshotsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := SnapshotsFromTfToAPIReadParams(ctx, plan)
	res := utils2.ExecuteRequest(func() (*numspot.ReadSnapshotsResponse, error) {
		return d.provider.GetNumspotClient().ReadSnapshotsWithResponse(ctx, d.provider.GetSpaceID(), &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Snapshot list")
	}

	objectItems, diags := utils2.FromHttpGenericListToTfList(ctx, res.JSON200.Items, SnapshotsFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
