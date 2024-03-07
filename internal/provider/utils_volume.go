package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_volume"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_volume"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func VolumeFromTfToHttp(tf *resource_volume.VolumeModel) *api.Volume {
	return &api.Volume{}
}

func fromLinkedVolumeSchemaToTFVolumesList(ctx context.Context, http api.LinkedVolume) (resource_volume.LinkedVolumesValue, diag.Diagnostics) {
	return resource_volume.NewLinkedVolumesValue(
		resource_volume.LinkedVolumesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_name":           types.StringPointerValue(http.DeviceName),
			"state":                 types.StringPointerValue(http.State),
			"vm_id":                 types.StringPointerValue(http.VmId),
			"volume_id":             types.StringPointerValue(http.Id),
		})
}

func VolumeFromHttpToTf(ctx context.Context, http *api.Volume) (resource_volume.VolumeModel, diag.Diagnostics) {
	volumes, diags := utils.GenericListToTfListValue(
		ctx,
		resource_volume.LinkedVolumesValue{},
		fromLinkedVolumeSchemaToTFVolumesList,
		*http.LinkedVolumes,
	)
	if diags.HasError() {
		return resource_volume.VolumeModel{}, diags
	}

	return resource_volume.VolumeModel{
		CreationDate:         types.StringValue(http.CreationDate.String()),
		Id:                   types.StringPointerValue(http.Id),
		Iops:                 utils.FromIntPtrToTfInt64(http.Iops),
		Size:                 utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId:           types.StringPointerValue(http.SnapshotId),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Type:                 types.StringPointerValue(http.Type),
		LinkedVolumes:        volumes,
	}, diags
}

func VolumeFromTfToCreateRequest(tf *resource_volume.VolumeModel) api.CreateVolumeJSONRequestBody {
	var (
		httpIops   *int
		snapshotId *string
	)
	if !tf.Iops.IsUnknown() && !tf.Iops.IsNull() {
		httpIops = utils.FromTfInt64ToIntPtr(tf.Iops)
	}
	if !tf.SnapshotId.IsUnknown() {
		snapshotId = tf.SnapshotId.ValueStringPointer()
	}

	return api.CreateVolumeJSONRequestBody{
		Iops:                 httpIops,
		Size:                 utils.FromTfInt64ToIntPtr(tf.Size),
		SnapshotId:           snapshotId,
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueString(),
		Type:                 tf.Type.ValueStringPointer(),
	}
}

func ValueFromTfToUpdaterequest(tf *resource_volume.VolumeModel) api.UpdateVolumeJSONRequestBody {
	var httpIops *int
	if !tf.Iops.IsUnknown() && !tf.Iops.IsNull() {
		httpIops = utils.FromTfInt64ToIntPtr(tf.Iops)
	}

	return api.UpdateVolumeJSONRequestBody{
		Iops:       httpIops,
		Size:       utils.FromTfInt64ToIntPtr(tf.Size),
		VolumeType: tf.Type.ValueStringPointer(),
	}
}

func VolumeFromTfToAPIReadParams(ctx context.Context, tf VolumesDataSourceModel) api.ReadVolumesParams {
	creationDates := utils.TfStringListToTimeList(ctx, tf.CreationDates, "2020-06-30T00:00:00.000Z")
	linkVolumeLinkDates := utils.TfStringListToTimeList(ctx, tf.LinkVolumeLinkDates, "2020-06-30T00:00:00.000Z")
	volumeSizes := utils.TFInt64ListToIntList(ctx, tf.VolumeSizes)

	return api.ReadVolumesParams{
		CreationDates:                &creationDates,
		LinkVolumeDeleteOnVmDeletion: tf.LinkVolumeDeleteOnVmDeletion.ValueBoolPointer(),
		LinkVolumeDeviceNames:        utils.TfStringListToStringPtrList(ctx, tf.LinkVolumeDeviceNames),
		LinkVolumeLinkDates:          &linkVolumeLinkDates,
		LinkVolumeLinkStates:         utils.TfStringListToStringPtrList(ctx, tf.LinkVolumeLinkStates),
		LinkVolumeVmIds:              utils.TfStringListToStringPtrList(ctx, tf.LinkVolumeVmIds),
		SnapshotIds:                  utils.TfStringListToStringPtrList(ctx, tf.SnapshotIds),
		VolumeSizes:                  &volumeSizes,
		VolumeStates:                 utils.TfStringListToStringPtrList(ctx, tf.VolumeStates),
		VolumeTypes:                  utils.TfStringListToStringPtrList(ctx, tf.VolumeTypes),
		AvailabilityZoneNames:        utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames),
		Ids:                          utils.TfStringListToStringPtrList(ctx, tf.Ids),
	}
}

func VolumesFromHttpToTfDatasource(ctx context.Context, http *api.Volume) (*datasource_volume.VolumeModel, diag.Diagnostics) {
	var (
		linkedVolumes = types.ListNull(resource_volume.LinkedVolumesValue{}.Type(ctx))
		diags         diag.Diagnostics
	)
	if http.LinkedVolumes != nil {
		linkedVolumes, diags = utils.GenericListToTfListValue(
			ctx,
			resource_volume.LinkedVolumesValue{},
			fromLinkedVolumeSchemaToTFVolumesList,
			*http.LinkedVolumes,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &datasource_volume.VolumeModel{
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		CreationDate:         types.StringValue(http.CreationDate.String()),
		Id:                   types.StringPointerValue(http.Id),
		Iops:                 utils.FromIntPtrToTfInt64(http.Iops),
		LinkedVolumes:        linkedVolumes,
		Size:                 utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId:           types.StringPointerValue(http.SnapshotId),
		State:                types.StringPointerValue(http.State),
		Type:                 types.StringPointerValue(http.Type),
	}, nil
}
