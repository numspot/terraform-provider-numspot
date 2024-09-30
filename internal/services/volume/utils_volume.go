package volume

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func VolumeFromTfToHttp(tf *VolumeModel) *numspot.Volume {
	return &numspot.Volume{}
}

func VolumeFromTfToLinkRequest(tf *VolumeModel) numspot.LinkVolumeJSONRequestBody {
	var (
		deviceName string
		vmId       string
	)

	if !utils.IsTfValueNull(tf.LinkVM) {
		deviceName = tf.LinkVM.DeviceName.ValueString()
		vmId = tf.LinkVM.VmID.ValueString()
	}

	return numspot.LinkVolumeJSONRequestBody{
		DeviceName: deviceName,
		VmId:       vmId,
	}
}

//func VolumeFromTfToUnlinkRequest(tf *VolumeModel) numspot.UnlinkVolumeJSONRequestBody {
//	return numspot.UnlinkVolumeJSONRequestBody{}
//}

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
	//var (
	//	linkedVolumes = types.ListNull(LinkedVolumesValue{}.Type(ctx))
	//	diags         diag.Diagnostics
	//	tagsList      types.List
	//)
	//
	//if http.Tags != nil {
	//	tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
	//	if diags.HasError() {
	//		return nil, diags
	//	}
	//}
	//
	//if http.LinkedVolumes != nil {
	//	linkedVolumes, diags = utils.GenericListToTfListValue(
	//		ctx,
	//		LinkedVolumesValue{},
	//		fromLinkedVolumeSchemaToTFVolumesList,
	//		*http.LinkedVolumes,
	//	)
	//	if diags.HasError() {
	//		return nil, diags
	//	}
	//}

	return &VolumeModel{
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		CreationDate:         types.StringValue(http.CreationDate.String()),
		Id:                   types.StringPointerValue(http.Id),
		Iops:                 utils.FromIntPtrToTfInt64(http.Iops),
		//LinkedVolumes:        linkedVolumes,
		Size:       utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId: types.StringPointerValue(http.SnapshotId),
		State:      types.StringPointerValue(http.State),
		Type:       types.StringPointerValue(http.Type),
		//Tags:                 tagsList,
	}, nil
}
