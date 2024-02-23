package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_volume"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func VolumeFromTfToHttp(tf *resource_volume.VolumeModel) *api.VolumeSchema {
	return &api.VolumeSchema{}
}

func fromLinkedVolumeSchemaToTFVolumesList(ctx context.Context, http api.LinkedVolumeSchema) (resource_volume.VolumesValue, diag.Diagnostics) {
	return resource_volume.NewVolumesValue(
		resource_volume.VolumesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_name":           types.StringPointerValue(http.DeviceName),
			"state":                 types.StringPointerValue(http.State),
			"vm_id":                 types.StringPointerValue(http.VmId),
			"volume_id":             types.StringPointerValue(http.VolumeId),
		})
}

func VolumeFromHttpToTf(ctx context.Context, http *api.VolumeSchema) (resource_volume.VolumeModel, diag.Diagnostics) {
	volumes, diags := utils.GenericListToTfListValue(ctx, resource_volume.VolumesValue{}, fromLinkedVolumeSchemaToTFVolumesList, *http.LinkedVolumes)
	if diags.HasError() {
		return resource_volume.VolumeModel{}, diags
	}
	return resource_volume.VolumeModel{
		CreationDate:  types.StringValue(http.CreationDate.String()),
		Id:            types.StringPointerValue(http.Id),
		Iops:          utils.FromIntPtrToTfInt64(http.Iops),
		Size:          utils.FromIntPtrToTfInt64(http.Size),
		SnapshotId:    types.StringPointerValue(http.SnapshotId),
		State:         types.StringPointerValue(http.State),
		SubregionName: types.StringPointerValue(http.SubregionName),
		Type:          types.StringPointerValue(http.Type),
		Volumes:       volumes,
	}, diags
}

func VolumeFromTfToCreateRequest(tf *resource_volume.VolumeModel) api.CreateVolumeJSONRequestBody {
	return api.CreateVolumeJSONRequestBody{
		Iops:          utils.FromTfInt64ToIntPtr(tf.Iops),
		Size:          utils.FromTfInt64ToIntPtr(tf.Size),
		SnapshotId:    tf.SnapshotId.ValueStringPointer(),
		SubregionName: tf.SubregionName.ValueString(),
		Type:          tf.Type.ValueStringPointer(),
	}
}
