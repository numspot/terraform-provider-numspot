package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_volume"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VolumesDataSourceModel struct {
	Volumes                      []datasource_volume.VolumeModel `tfsdk:"volumes"`
	CreationDates                types.List                      `tfsdk:"creation_dates"`
	LinkVolumeDeleteOnVmDeletion types.Bool                      `tfsdk:"link_volume_delete_on_vm_deletion"`
	LinkVolumeDeviceNames        types.List                      `tfsdk:"link_volume_device_names"`
	LinkVolumeLinkDates          types.List                      `tfsdk:"link_volume_link_dates"`
	LinkVolumeLinkStates         types.List                      `tfsdk:"link_volume_link_states"`
	LinkVolumeVmIds              types.List                      `tfsdk:"link_volume_vm_ids"`
	SnapshotIds                  types.List                      `tfsdk:"snapshot_ids"`
	VolumeSizes                  types.List                      `tfsdk:"volume_sizes"`
	VolumeStates                 types.List                      `tfsdk:"volume_states"`
	VolumeTypes                  types.List                      `tfsdk:"volume_types"`
	AvailabilityZoneNames        types.List                      `tfsdk:"availability_zone_names"`
	Ids                          types.List                      `tfsdk:"ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &volumesDataSource{}
)

func (d *volumesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
func NewVolumesDataSource() datasource.DataSource {
	return &volumesDataSource{}
}

// coffeesDataSource is the data source implementation.
type volumesDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *volumesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

// Schema defines the schema for the data source.
func (d *volumesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_volume.VolumeDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *volumesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VolumesDataSourceModel
	request.Config.Get(ctx, &plan)

	params := VolumeFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadVolumesResponse, error) {
		return d.provider.ApiClient.ReadVolumesWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty volumes list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := VolumesFromHttpToTfDatasource(ctx, &item)
		if diags != nil {
			response.Diagnostics.AddError("Error while converting Volume HTTP object to Terraform object", diags.Errors()[0].Detail())
		}
		state.Volumes = append(state.Volumes, *tf)
	}
	state.AvailabilityZoneNames = plan.AvailabilityZoneNames
	state.CreationDates = plan.CreationDates
	state.Ids = plan.Ids
	state.LinkVolumeDeleteOnVmDeletion = plan.LinkVolumeDeleteOnVmDeletion
	state.LinkVolumeDeviceNames = plan.LinkVolumeDeviceNames
	state.LinkVolumeLinkDates = plan.LinkVolumeLinkDates
	state.LinkVolumeLinkStates = plan.LinkVolumeLinkStates
	state.LinkVolumeVmIds = plan.LinkVolumeVmIds
	state.SnapshotIds = plan.SnapshotIds
	state.VolumeSizes = plan.VolumeSizes
	state.VolumeStates = plan.VolumeStates
	state.VolumeTypes = plan.VolumeTypes
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
