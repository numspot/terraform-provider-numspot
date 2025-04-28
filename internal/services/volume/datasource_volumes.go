package volume

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/volume/datasource_volume"
	"terraform-provider-numspot/internal/utils"
)

// Package volume provides the implementation of the Volumes data source
// for the NumSpot provider. It handles reading and listing volumes from NumSpot,
// including managing volume attributes such as size, type, state, and tags.

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &volumesDataSource{}
)

// volumesDataSource represents the Volumes data source implementation.
// It implements the Terraform datasource.DataSource interface and provides
// read operations for volumes in NumSpot.
type volumesDataSource struct {
	provider *client.NumSpotSDK
}

// NewVolumesDataSource creates a new instance of the Volumes data source.
// This is the factory function used by the provider to create new volume data source instances.
func NewVolumesDataSource() datasource.DataSource {
	return &volumesDataSource{}
}

// Configure implements the datasource.DataSource interface.
// It configures the data source with the provider's SDK client.
func (d *volumesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

// Metadata implements the datasource.DataSource interface.
// It sets the data source type name for the Volumes data source.
func (d *volumesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

// Schema implements the datasource.DataSource interface.
// It defines the schema for the Volumes data source, including all its attributes
// and their types.
func (d *volumesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_volume.VolumeDataSourceSchema(ctx)
}

// Read implements the datasource.DataSource interface.
// It reads the current state of volumes from NumSpot based on the provided filters.
// The function handles:
// - Converting Terraform configuration to NumSpot API parameters
// - Querying volumes from NumSpot
// - Converting the API response to Terraform state format
// - Updating the Terraform state with the retrieved data
func (d *volumesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_volume.VolumeModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeVolumeParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	volumes, err := core.ReadVolumeWithParams(ctx, d.provider, params)
	if err != nil {
		return
	}

	objectItems := utils.SerializeDatasourceItemsWithDiags(ctx, *volumes, &response.Diagnostics, mappingItemsValue)
	if response.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

// deserializeVolumeParams converts the Terraform data source model
// into the NumSpot API parameters format for querying volumes.
// It handles all filter parameters including:
// - Volume IDs and types
// - Creation dates
// - Volume sizes
// - Link volume parameters
// - Snapshot IDs
// - Availability zones
func deserializeVolumeParams(ctx context.Context, tf datasource_volume.VolumeModel, diags *diag.Diagnostics) api.ReadVolumesParams {
	var creationDatesPtr *[]time.Time
	var linkVolumeLinkDatesPtr *[]time.Time
	var volumeSizesPtr *[]int

	if !(tf.CreationDates.IsNull() || tf.CreationDates.IsUnknown()) {
		creationDates := utils.ConvertTfListToArrayOfTime(ctx, tf.CreationDates, "2020-06-30T00:00:00.000Z", diags)
		creationDatesPtr = creationDates
	}

	if !(tf.LinkVolumeLinkDates.IsNull() || tf.LinkVolumeLinkDates.IsUnknown()) {
		linkVolumeLinkDates := utils.ConvertTfListToArrayOfTime(ctx, tf.LinkVolumeLinkDates, "2020-06-30T00:00:00.000Z", diags)
		linkVolumeLinkDatesPtr = linkVolumeLinkDates
	}

	if !(tf.VolumeSizes.IsNull() || tf.VolumeSizes.IsUnknown()) {
		volumeSizes := utils.ConvertTfListToArrayOfInt(ctx, tf.VolumeSizes, diags)
		volumeSizesPtr = volumeSizes
	}

	return api.ReadVolumesParams{
		CreationDates:                creationDatesPtr,
		LinkVolumeDeleteOnVmDeletion: tf.LinkVolumeDeleteOnVmDeletion.ValueBoolPointer(),
		LinkVolumeDeviceNames:        utils.ConvertTfListToArrayOfString(ctx, tf.LinkVolumeDeviceNames, diags),
		LinkVolumeLinkDates:          linkVolumeLinkDatesPtr,
		LinkVolumeLinkStates:         utils.ConvertTfListToArrayOfString(ctx, tf.LinkVolumeLinkStates, diags),
		LinkVolumeVmIds:              utils.ConvertTfListToArrayOfString(ctx, tf.LinkVolumeVmIds, diags),
		SnapshotIds:                  utils.ConvertTfListToArrayOfString(ctx, tf.SnapshotIds, diags),
		VolumeSizes:                  volumeSizesPtr,
		VolumeStates:                 utils.ConvertTfListToArrayOfString(ctx, tf.VolumeStates, diags),
		VolumeTypes:                  utils.ConvertTfListToArrayOfString(ctx, tf.VolumeTypes, diags),
		AvailabilityZoneNames:        utils.ConvertTfListToArrayOfAzName(ctx, tf.AvailabilityZoneNames, diags),
		Ids:                          utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
	}
}

// mappingItemsValue converts a NumSpot API volume response into the Terraform
// data source model. It handles all volume attributes including:
// - Basic attributes (ID, size, type, state, etc.)
// - IOPS configuration
// - Linked volumes
// - Tags
func mappingItemsValue(ctx context.Context, volume api.Volume, diags *diag.Diagnostics) (datasource_volume.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_volume.ItemsValue{}.Type(ctx))
	linkedVolumesList := types.List{}

	if volume.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *volume.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_volume.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_volume.ItemsValue{}, serializeDiags
		}
	}

	if volume.LinkedVolumes != nil {
		var serializeDiags diag.Diagnostics

		linkedVolumesList, serializeDiags = mappingLinkedVolumes(ctx, volume, diags)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	return datasource_volume.NewItemsValue(datasource_volume.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(volume.AvailabilityZoneName)),
		"creation_date":          types.StringValue(volume.CreationDate.String()),
		"id":                     types.StringValue(utils.ConvertStringPtrToString(volume.Id)),
		"iops":                   types.Int64Value(utils.ConvertIntPtrToInt64(volume.Iops)),
		"linked_volumes":         linkedVolumesList,
		"size":                   types.Int64Value(utils.ConvertIntPtrToInt64(volume.Size)),
		"snapshot_id":            types.StringValue(utils.ConvertStringPtrToString(volume.SnapshotId)),
		"state":                  types.StringValue(utils.ConvertStringPtrToString(volume.State)),
		"tags":                   tagsList,
		"type":                   types.StringValue(utils.ConvertStringPtrToString(volume.Type)),
	})
}

// mappingTags converts a NumSpot API tag response into the Terraform
// data source tag model.
func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_volume.TagsValue, diag.Diagnostics) {
	return datasource_volume.NewTagsValue(datasource_volume.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

// mappingLinkedVolumes converts NumSpot API linked volume data into the Terraform
// data source linked volume model. It handles:
// - Volume attachment attributes
// - Device names
// - VM associations
// - Deletion policies
func mappingLinkedVolumes(ctx context.Context, volumes api.Volume, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	ll := len(*volumes.LinkedVolumes)
	elementValue := make([]datasource_volume.LinkedVolumesValue, ll)
	for y, lv := range *volumes.LinkedVolumes {
		elementValue[y], *diags = datasource_volume.NewLinkedVolumesValue(datasource_volume.LinkedVolumesValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(lv.DeleteOnVmDeletion),
			"device_name":           types.StringPointerValue(lv.DeviceName),
			"state":                 types.StringPointerValue(lv.State),
			"vm_id":                 types.StringPointerValue(lv.VmId),
			"id":                    types.StringPointerValue(lv.Id),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_volume.LinkedVolumesValue).Type(ctx), elementValue)
}
