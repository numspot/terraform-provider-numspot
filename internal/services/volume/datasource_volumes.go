package volume

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/volume/datasource_volume"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &volumesDataSource{}

// volumesDataSource is the Terraform data source implementation for numspot_volumes.
type volumesDataSource struct {
	services.BaseDataSource
}

// NewVolumesDataSource returns a new volumesDataSource instance.
func NewVolumesDataSource() datasource.DataSource {
	return &volumesDataSource{}
}

// Metadata sets the data source type name.
func (d *volumesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

// Schema defines the schema for the data source.
func (d *volumesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_volume.VolumeDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *volumesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_volume.VolumeModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadVolumes(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadVolumesWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read volumes", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.Volume) (datasource_volume.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	tagsList := types.ListNull(datasource_volume.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_volume.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_volume.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_volume.NewItemsValue(datasource_volume.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(item.AvailabilityZoneName)),
		"creation_date":          utils.TimePtrToTfString(item.CreationDate),
		"id":                     types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"iops":                   utils.FromIntPtrToTfInt64(item.Iops),
		"linked_volumes":         mappingLinkedVolumes(ctx, item, &diags),
		"size":                   utils.FromIntPtrToTfInt64(item.Size),
		"snapshot_id":            types.StringValue(utils.ConvertStringPtrToString(item.SnapshotId)),
		"state":                  types.StringValue(utils.ConvertStringPtrToString(item.State)),
		"type":                   types.StringValue(utils.ConvertStringPtrToString(item.Type)),
		"tags":                   tagsList,
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_volume.TagsValue, diag.Diagnostics) {
	return datasource_volume.NewTagsValue(datasource_volume.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingLinkedVolume(ctx context.Context, item *api.LinkedVolume) (datasource_volume.LinkedVolumesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_volume.NewLinkedVolumesValue(datasource_volume.LinkedVolumesValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(item.DeleteOnVmDeletion),
		"device_name":           types.StringPointerValue(item.DeviceName),
		"id":                    types.StringPointerValue(item.Id),
		"state":                 types.StringPointerValue(item.State),
		"vm_id":                 types.StringPointerValue(item.VmId),
	})
	diags.Append(d...)
	return value, diags
}

func mappingLinkedVolumes(ctx context.Context, item *api.Volume, diags *diag.Diagnostics) types.List {
	if item.LinkedVolumes == nil {
		return types.ListNull(datasource_volume.LinkedVolumesValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.LinkedVolumes, mappingLinkedVolume)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_volume.LinkedVolumesValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func deserializeReadVolumes(ctx context.Context, tf *datasource_volume.VolumeModel, diags *diag.Diagnostics) *api.ReadVolumesParams {
	return &api.ReadVolumesParams{
		CreationDates:                utils.ConvertTfListToArrayOfTime(ctx, tf.CreationDates, "2020-06-30T00:00:00.000Z", diags),
		LinkVolumeDeleteOnVmDeletion: tf.LinkVolumeDeleteOnVmDeletion.ValueBoolPointer(),
		LinkVolumeDeviceNames:        utils.ConvertTfListToArrayOfString(ctx, tf.LinkVolumeDeviceNames, diags),
		LinkVolumeLinkDates:          utils.ConvertTfListToArrayOfTime(ctx, tf.LinkVolumeLinkDates, "2020-06-30T00:00:00.000Z", diags),
		LinkVolumeLinkStates:         utils.ConvertTfListToArrayOfString(ctx, tf.LinkVolumeLinkStates, diags),
		LinkVolumeVmIds:              utils.ConvertTfListToArrayOfString(ctx, tf.LinkVolumeVmIds, diags),
		SnapshotIds:                  utils.ConvertTfListToArrayOfString(ctx, tf.SnapshotIds, diags),
		TagKeys:                      utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:                    utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:                         utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		VolumeSizes:                  utils.ConvertTfListToArrayOfInt(ctx, tf.VolumeSizes, diags),
		VolumeStates:                 utils.ConvertTfListToArrayOfString(ctx, tf.VolumeStates, diags),
		VolumeTypes:                  utils.ConvertTfListToArrayOfString(ctx, tf.VolumeTypes, diags),
		AvailabilityZoneNames:        utils.ConvertTfListToArrayOfAzName(ctx, tf.AvailabilityZoneNames, diags),
		Ids:                          utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
