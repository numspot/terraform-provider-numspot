package volume

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func VolumeFromTfToHttp(tf *VolumeModel) *numspot.Volume {
	return &numspot.Volume{}
}

func fromLinkedVolumeSchemaToTFVolumesList(ctx context.Context, http numspot.LinkedVolume) (LinkedVolumesValue, diag.Diagnostics) {
	return NewLinkedVolumesValue(
		LinkedVolumesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_name":           types.StringPointerValue(http.DeviceName),
			"state":                 types.StringPointerValue(http.State),
			"vm_id":                 types.StringPointerValue(http.VmId),
			"volume_id":             types.StringPointerValue(http.Id),
		})
}

func VolumeFromHttpToTf(ctx context.Context, http *numspot.Volume) (*VolumeModel, diag.Diagnostics) {
	var (
		volumes = types.ListNull(LinkedVolumesValue{}.Type(ctx))
		tagsTf  types.List
		diags   diag.Diagnostics
	)

	if http.LinkedVolumes != nil {
		volumes, diags = utils.GenericListToTfListValue(
			ctx,
			LinkedVolumesValue{},
			fromLinkedVolumeSchemaToTFVolumesList,
			*http.LinkedVolumes,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &VolumeModel{
		CreationDate:         types.StringValue(http.CreationDate.String()),
		Id:                   types.StringPointerValue(http.Id),
		Iops:                 utils.FromIntPtrToTfInt64(http.Iops),
		Size:                 utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId:           types.StringPointerValue(http.SnapshotId),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Type:                 types.StringPointerValue(http.Type),
		LinkedVolumes:        volumes,
		Tags:                 tagsTf,
	}, diags
}

func VolumeFromTfToCreateRequest(tf *VolumeModel) numspot.CreateVolumeJSONRequestBody {
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

	return numspot.CreateVolumeJSONRequestBody{
		Iops:                 httpIops,
		Size:                 utils.FromTfInt64ToIntPtr(tf.Size),
		SnapshotId:           snapshotId,
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueString(),
		Type:                 tf.Type.ValueStringPointer(),
	}
}

func ValueFromTfToUpdaterequest(tf *VolumeModel) numspot.UpdateVolumeJSONRequestBody {
	var httpIops *int
	if !tf.Iops.IsUnknown() && !tf.Iops.IsNull() {
		httpIops = utils.FromTfInt64ToIntPtr(tf.Iops)
	}

	return numspot.UpdateVolumeJSONRequestBody{
		Iops:       httpIops,
		Size:       utils.FromTfInt64ToIntPtr(tf.Size),
		VolumeType: tf.Type.ValueStringPointer(),
	}
}

func VolumeFromTfToAPIReadParams(ctx context.Context, tf VolumesDataSourceModel) numspot.ReadVolumesParams {
	var creationDatesPtr *[]time.Time
	var linkVolumeLinkDatesPtr *[]time.Time
	var volumeSizesPtr *[]int

	if !(tf.CreationDates.IsNull() || tf.CreationDates.IsUnknown()) {
		creationDates := utils.TfStringListToTimeList(ctx, tf.CreationDates, "2020-06-30T00:00:00.000Z")
		creationDatesPtr = &creationDates
	}

	if !(tf.LinkVolumeLinkDates.IsNull() || tf.LinkVolumeLinkDates.IsUnknown()) {
		linkVolumeLinkDates := utils.TfStringListToTimeList(ctx, tf.LinkVolumeLinkDates, "2020-06-30T00:00:00.000Z")
		linkVolumeLinkDatesPtr = &linkVolumeLinkDates
	}

	if !(tf.VolumeSizes.IsNull() || tf.VolumeSizes.IsUnknown()) {
		volumeSizes := utils.TFInt64ListToIntList(ctx, tf.VolumeSizes)
		volumeSizesPtr = &volumeSizes
	}
	return numspot.ReadVolumesParams{
		CreationDates:                creationDatesPtr,
		LinkVolumeDeleteOnVmDeletion: tf.LinkVolumeDeleteOnVmDeletion.ValueBoolPointer(),
		LinkVolumeDeviceNames:        utils.TfStringListToStringPtrList(ctx, tf.LinkVolumeDeviceNames),
		LinkVolumeLinkDates:          linkVolumeLinkDatesPtr,
		LinkVolumeLinkStates:         utils.TfStringListToStringPtrList(ctx, tf.LinkVolumeLinkStates),
		LinkVolumeVmIds:              utils.TfStringListToStringPtrList(ctx, tf.LinkVolumeVmIds),
		SnapshotIds:                  utils.TfStringListToStringPtrList(ctx, tf.SnapshotIds),
		VolumeSizes:                  volumeSizesPtr,
		VolumeStates:                 utils.TfStringListToStringPtrList(ctx, tf.VolumeStates),
		VolumeTypes:                  utils.TfStringListToStringPtrList(ctx, tf.VolumeTypes),
		AvailabilityZoneNames:        utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames),
		Ids:                          utils.TfStringListToStringPtrList(ctx, tf.Ids),
	}
}

func VolumesFromHttpToTfDatasource(ctx context.Context, http *numspot.Volume) (*VolumeModel, diag.Diagnostics) {
	var (
		linkedVolumes = types.ListNull(LinkedVolumesValue{}.Type(ctx))
		diags         diag.Diagnostics
		tagsList      types.List
	)

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.LinkedVolumes != nil {
		linkedVolumes, diags = utils.GenericListToTfListValue(
			ctx,
			LinkedVolumesValue{},
			fromLinkedVolumeSchemaToTFVolumesList,
			*http.LinkedVolumes,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &VolumeModel{
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
	}, nil
}
