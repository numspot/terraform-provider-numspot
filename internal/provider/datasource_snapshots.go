package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_snapshot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type SnapshotsDataSourceModel struct {
	Snapshots                                 []datasource_snapshot.SnapshotModel `tfsdk:"snapshots"`
	Descriptions                              types.List                          `tfsdk:"descriptions"`
	FromCreationDate                          types.String                        `tfsdk:"from_creation_date"`
	PermissionsToCreateVolumeGlobalPermission types.Bool                          `tfsdk:"permissions_to_create_volume_global_permission"`
	Progresses                                types.List                          `tfsdk:"progresses"`
	States                                    types.List                          `tfsdk:"states"`
	ToCreationDate                            types.String                        `tfsdk:"to_creation_date"`
	VolumeIds                                 types.List                          `tfsdk:"volume_ids"`
	VolumeSizes                               types.List                          `tfsdk:"volume_sizes"`
	TagKeys                                   types.List                          `tfsdk:"tag_keys"`
	TagValues                                 types.List                          `tfsdk:"tag_values"`
	Tags                                      types.List                          `tfsdk:"tags"`
	IDs                                       types.List                          `tfsdk:"ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &snapshotsDataSource{}
)

func (d *snapshotsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewSnapshotsDataSource() datasource.DataSource {
	return &snapshotsDataSource{}
}

type snapshotsDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *snapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshots"
}

// Schema defines the schema for the data source.
func (d *snapshotsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_snapshot.SnapshotDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *snapshotsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan SnapshotsDataSourceModel
	request.Config.Get(ctx, &plan)

	params := SnapshotsFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadSnapshotsResponse, error) {
		return d.provider.ApiClient.ReadSnapshotsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Snapshot list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := SnapshotsFromHttpToTfDatasource(ctx, &item)
		if diags != nil {
			response.Diagnostics.AddError("Error while converting Snapshot HTTP object to Terraform object", diags.Errors()[0].Detail())
		}
		state.Snapshots = append(state.Snapshots, *tf)
	}

	state.Descriptions = plan.Descriptions
	state.FromCreationDate = plan.FromCreationDate
	state.PermissionsToCreateVolumeGlobalPermission = plan.PermissionsToCreateVolumeGlobalPermission
	state.Progresses = plan.Progresses
	state.States = plan.States
	state.ToCreationDate = plan.ToCreationDate
	state.VolumeIds = plan.VolumeIds
	state.VolumeSizes = plan.VolumeSizes
	state.TagKeys = plan.TagKeys
	state.TagValues = plan.TagValues
	state.Tags = plan.Tags
	state.IDs = plan.IDs

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
