package volume

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type VolumesDataSourceModel struct {
	Items                        []DatasourceVolumeModel `tfsdk:"items"`
	AvailabilityZoneNames        types.List              `tfsdk:"availability_zone_names"`
	CreationDates                types.List              `tfsdk:"creation_dates"`
	Ids                          types.List              `tfsdk:"ids"`
	LinkVolumeDeleteOnVmDeletion types.Bool              `tfsdk:"link_volume_delete_on_vm_deletion"`
	LinkVolumeDeviceNames        types.List              `tfsdk:"link_volume_device_names"`
	LinkVolumeLinkDates          types.List              `tfsdk:"link_volume_link_dates"`
	LinkVolumeLinkStates         types.List              `tfsdk:"link_volume_link_states"`
	LinkVolumeVmIds              types.List              `tfsdk:"link_volume_vm_ids"`
	SnapshotIds                  types.List              `tfsdk:"snapshot_ids"`
	TagKeys                      types.List              `tfsdk:"tag_keys"`
	TagValues                    types.List              `tfsdk:"tag_values"`
	Tags                         types.List              `tfsdk:"tags"`
	VolumeSizes                  types.List              `tfsdk:"volume_sizes"`
	VolumeStates                 types.List              `tfsdk:"volume_states"`
	VolumeTypes                  types.List              `tfsdk:"volume_types"`
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

	res, err := d.provider.GetNumspotClient().ReadVolumesWithResponse(ctx, d.provider.GetSpaceID(), deserializeNumSpotVolumeDataSource(ctx, plan, &response.Diagnostics))
	if err != nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty volumes list")
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, serializeNumSpotDataSource, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeNumSpotVolumeDataSource(ctx context.Context, tf VolumesDataSourceModel, diags *diag.Diagnostics) *numspot.ReadVolumesParams {
	var creationDatesPtr *[]time.Time
	var linkVolumeLinkDatesPtr *[]time.Time
	var volumeSizesPtr *[]int

	if !(tf.CreationDates.IsNull() || tf.CreationDates.IsUnknown()) {
		creationDates := utils.TfStringListToTimeList(ctx, tf.CreationDates, "2020-06-30T00:00:00.000Z", diags)
		creationDatesPtr = &creationDates
	}

	if !(tf.LinkVolumeLinkDates.IsNull() || tf.LinkVolumeLinkDates.IsUnknown()) {
		linkVolumeLinkDates := utils.TfStringListToTimeList(ctx, tf.LinkVolumeLinkDates, "2020-06-30T00:00:00.000Z", diags)
		linkVolumeLinkDatesPtr = &linkVolumeLinkDates
	}

	if !(tf.VolumeSizes.IsNull() || tf.VolumeSizes.IsUnknown()) {
		volumeSizes := utils.TFInt64ListToIntList(ctx, tf.VolumeSizes, diags)
		volumeSizesPtr = &volumeSizes
	}
	return &numspot.ReadVolumesParams{
		CreationDates:                creationDatesPtr,
		LinkVolumeDeleteOnVmDeletion: tf.LinkVolumeDeleteOnVmDeletion.ValueBoolPointer(),
		LinkVolumeDeviceNames:        utils.TfStringListToStringPtrList(ctx, tf.LinkVolumeDeviceNames, diags),
		LinkVolumeLinkDates:          linkVolumeLinkDatesPtr,
		LinkVolumeLinkStates:         utils.TfStringListToStringPtrList(ctx, tf.LinkVolumeLinkStates, diags),
		LinkVolumeVmIds:              utils.TfStringListToStringPtrList(ctx, tf.LinkVolumeVmIds, diags),
		SnapshotIds:                  utils.TfStringListToStringPtrList(ctx, tf.SnapshotIds, diags),
		VolumeSizes:                  volumeSizesPtr,
		VolumeStates:                 utils.TfStringListToStringPtrList(ctx, tf.VolumeStates, diags),
		VolumeTypes:                  utils.TfStringListToStringPtrList(ctx, tf.VolumeTypes, diags),
		AvailabilityZoneNames:        utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames, diags),
		Ids:                          utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
	}
}

func serializeNumSpotDataSource(ctx context.Context, http *numspot.Volume, diags *diag.Diagnostics) *DatasourceVolumeModel {
	var (
		linkedVolumes = types.ListNull(LinkedVolumesValue{}.Type(ctx))
		tagsList      types.List
	)

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	if http.LinkedVolumes != nil {
		linkedVolumes = utils.GenericListToTfListValue(
			ctx,
			LinkedVolumesValue{},
			fromLinkedVolumeSchemaToTFVolumesList,
			*http.LinkedVolumes,
			diags,
		)
	}

	return &DatasourceVolumeModel{
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		CreationDate:         types.StringValue(http.CreationDate.String()),
		Id:                   types.StringPointerValue(http.Id),
		Iops:                 utils.FromIntPtrToTfInt64(http.Iops),
		LinkedVolumes:        linkedVolumes,
		Size:                 utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId:           types.StringPointerValue(http.SnapshotId),
		State:                types.StringPointerValue(http.State),
		Type:                 types.StringPointerValue(http.Type),
		Tags:                 tagsList,
	}
}

func fromLinkedVolumeSchemaToTFVolumesList(ctx context.Context, http numspot.LinkedVolume, diags *diag.Diagnostics) LinkedVolumesValue {
	value, diagnostics := NewLinkedVolumesValue(
		LinkedVolumesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_name":           types.StringPointerValue(http.DeviceName),
			"state":                 types.StringPointerValue(http.State),
			"vm_id":                 types.StringPointerValue(http.VmId),
			"id":                    types.StringPointerValue(http.Id),
		})
	diags.Append(diagnostics...)
	return value
}
