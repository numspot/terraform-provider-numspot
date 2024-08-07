package volume

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

type VolumesDataSourceModel struct {
	Items                        []VolumeModel `tfsdk:"items"`
	AvailabilityZoneNames        types.List    `tfsdk:"availability_zone_names"`
	CreationDates                types.List    `tfsdk:"creation_dates"`
	Ids                          types.List    `tfsdk:"ids"`
	LinkVolumeDeleteOnVmDeletion types.Bool    `tfsdk:"link_volume_delete_on_vm_deletion"`
	LinkVolumeDeviceNames        types.List    `tfsdk:"link_volume_device_names"`
	LinkVolumeLinkDates          types.List    `tfsdk:"link_volume_link_dates"`
	LinkVolumeLinkStates         types.List    `tfsdk:"link_volume_link_states"`
	LinkVolumeVmIds              types.List    `tfsdk:"link_volume_vm_ids"`
	SnapshotIds                  types.List    `tfsdk:"snapshot_ids"`
	TagKeys                      types.List    `tfsdk:"tag_keys"`
	TagValues                    types.List    `tfsdk:"tag_values"`
	Tags                         types.List    `tfsdk:"tags"`
	VolumeSizes                  types.List    `tfsdk:"volume_sizes"`
	VolumeStates                 types.List    `tfsdk:"volume_states"`
	VolumeTypes                  types.List    `tfsdk:"volume_types"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &volumesDataSource{}
)

func (d *volumesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVolumesDataSource() datasource.DataSource {
	return &volumesDataSource{}
}

type volumesDataSource struct {
	provider services.IProvider
}

// Metadata returns the data source type name.
func (d *volumesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

// Schema defines the schema for the data source.
func (d *volumesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = VolumeDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *volumesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VolumesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := VolumeFromTfToAPIReadParams(ctx, plan)
	res := utils2.ExecuteRequest(func() (*numspot.ReadVolumesResponse, error) {
		return d.provider.GetNumspotClient().ReadVolumesWithResponse(ctx, d.provider.GetSpaceID(), &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty volumes list")
	}

	objectItems, diags := utils2.FromHttpGenericListToTfList(ctx, res.JSON200.Items, VolumesFromHttpToTfDatasource)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
