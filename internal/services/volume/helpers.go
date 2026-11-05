package volume

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services/volume/resource_volume"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

const (
	linkVmDeviceNameAttr = "device_name"
	linkVmVmIDAttr       = "vm_id"
)

// linkVmFromVolume derives the link_vm block from the first linked volume. The
// read payload has no top-level link_vm field; it is surfaced from
// LinkedVolumes[0] (or null when the volume is unattached).
func linkVmFromVolume(ctx context.Context, linkedVolumes *[]api.LinkedVolume, diags *diag.Diagnostics) resource_volume.LinkVmValue {
	if linkedVolumes == nil || len(*linkedVolumes) == 0 {
		return resource_volume.NewLinkVmValueNull()
	}

	first := (*linkedVolumes)[0]
	value, d := resource_volume.NewLinkVmValue(resource_volume.LinkVmValue{}.AttributeTypes(ctx), map[string]attr.Value{
		linkVmDeviceNameAttr: types.StringPointerValue(first.DeviceName),
		linkVmVmIDAttr:       types.StringPointerValue(first.VmId),
	})
	diags.Append(d...)
	return value
}

// deserializeUpdateVolume builds the UpdateVolume body for the attribute update
// call (size/type/iops).
func deserializeUpdateVolume(tf *resource_volume.VolumeModel) *api.UpdateVolumeJSONRequestBody {
	var httpIOPS *int
	if !tf.Iops.IsUnknown() && !tf.Iops.IsNull() {
		httpIOPS = utils.FromTfInt64ToIntPtr(tf.Iops)
	}

	return &api.UpdateVolumeJSONRequestBody{
		Iops:       httpIOPS,
		Size:       utils.FromTfInt64ToIntPtr(tf.Size),
		VolumeType: tf.Type.ValueStringPointer(),
	}
}
