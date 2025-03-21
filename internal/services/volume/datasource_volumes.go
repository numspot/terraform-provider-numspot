package volume

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/volume/datasource_volume"
	"terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &volumesDataSource{}
)

func (d *volumesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVolumesDataSource() datasource.DataSource {
	return &volumesDataSource{}
}

type volumesDataSource struct {
	provider *client.NumSpotSDK
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

	objectItems := serializeVolumes(ctx, volumes, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems.Items
	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

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
		AvailabilityZoneNames:        utils.ConvertTfListToArrayOfString(ctx, tf.AvailabilityZoneNames, diags),
		Ids:                          utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
	}
}

func serializeVolumes(ctx context.Context, volumes *[]api.Volume, diags *diag.Diagnostics) datasource_volume.VolumeModel {
	var volumesList types.List
	var serializeDiags diag.Diagnostics

	tagsList := types.List{}
	linkedVolumesList := types.List{}

	if len(*volumes) != 0 {
		ll := len(*volumes)
		itemsValue := make([]datasource_volume.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			if (*volumes)[i].Tags != nil {
				tagsList, serializeDiags = mappingVolumeTags(ctx, volumes, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*volumes)[i].LinkedVolumes != nil {
				linkedVolumesList, serializeDiags = mappingLinkedVolumes(ctx, volumes, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			itemsValue[i], serializeDiags = datasource_volume.NewItemsValue(datasource_volume.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"availability_zone_name": types.StringValue(utils.ConvertStringPtrToString((*volumes)[i].AvailabilityZoneName)),
				"creation_date":          types.StringValue((*volumes)[i].CreationDate.String()),
				"id":                     types.StringValue(utils.ConvertStringPtrToString((*volumes)[i].Id)),
				"iops":                   types.Int64Value(utils.ConvertIntPtrToInt64((*volumes)[i].Iops)),
				"linked_volumes":         linkedVolumesList,
				"size":                   types.Int64Value(utils.ConvertIntPtrToInt64((*volumes)[i].Size)),
				"snapshot_id":            types.StringValue(utils.ConvertStringPtrToString((*volumes)[i].SnapshotId)),
				"state":                  types.StringValue(utils.ConvertStringPtrToString((*volumes)[i].State)),
				"tags":                   tagsList,
				"type":                   types.StringValue(utils.ConvertStringPtrToString((*volumes)[i].Type)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		volumesList, serializeDiags = types.ListValueFrom(ctx, new(datasource_volume.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		volumesList = types.ListNull(new(datasource_volume.ItemsValue).Type(ctx))
	}

	return datasource_volume.VolumeModel{
		Items: volumesList,
	}
}

func mappingVolumeTags(ctx context.Context, volumes *[]api.Volume, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	lt := len(*(*volumes)[i].Tags)
	elementValue := make([]datasource_volume.TagsValue, lt)
	for y, tag := range *(*volumes)[i].Tags {
		elementValue[y], *diags = datasource_volume.NewTagsValue(datasource_volume.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"key":   types.StringValue(tag.Key),
			"value": types.StringValue(tag.Value),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_volume.TagsValue).Type(ctx), elementValue)
}

func mappingLinkedVolumes(ctx context.Context, volumes *[]api.Volume, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	ll := len(*(*volumes)[i].LinkedVolumes)
	elementValue := make([]datasource_volume.LinkedVolumesValue, ll)
	for y, lv := range *(*volumes)[i].LinkedVolumes {
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
