package image

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services/image/resource_image"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// deserializeBsu maps the nested bsu object onto the create body. obj.As cannot
// target the generated attr.Value structs (the framework reflect layer rejects
// custom attr.Value targets), so the attribute values are mapped manually.
func deserializeBsu(obj types.Object) *api.BsuToCreate {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	attrs := obj.Attributes()
	getString := func(name string) *string {
		v, ok := attrs[name].(types.String)
		if !ok || v.IsNull() || v.IsUnknown() {
			return nil
		}
		return v.ValueStringPointer()
	}
	getInt := func(name string) *int {
		v, ok := attrs[name].(types.Int64)
		if !ok || v.IsNull() || v.IsUnknown() {
			return nil
		}
		return utils.FromTfInt64ToIntPtr(v)
	}
	getBool := func(name string) *bool {
		v, ok := attrs[name].(types.Bool)
		if !ok || v.IsNull() || v.IsUnknown() {
			return nil
		}
		return v.ValueBoolPointer()
	}
	return &api.BsuToCreate{
		DeleteOnVmDeletion: getBool("delete_on_vm_deletion"),
		Iops:               getInt("iops"),
		SnapshotId:         getString("snapshot_id"),
		VolumeSize:         getInt("volume_size"),
		VolumeType:         getString("volume_type"),
	}
}

// deserializeBlockDeviceMappings builds the create body's block device mappings
// from the plan. A null list yields an empty slice (sent as [] to the API), and
// an unknown bsu maps to a nil Bsu.
func deserializeBlockDeviceMappings(ctx context.Context, tfBlockDeviceMappings types.List) []api.BlockDeviceMappingImage {
	tfItems := make([]resource_image.BlockDeviceMappingsValue, 0, len(tfBlockDeviceMappings.Elements()))
	tfBlockDeviceMappings.ElementsAs(ctx, &tfItems, false)

	items := make([]api.BlockDeviceMappingImage, 0, len(tfItems))
	for i := range tfItems {
		e := &tfItems[i]
		items = append(items, api.BlockDeviceMappingImage{
			Bsu:               deserializeBsu(e.Bsu),
			DeviceName:        e.DeviceName.ValueStringPointer(),
			VirtualDeviceName: utils.FromTfStringToStringPtr(e.VirtualDeviceName),
		})
	}
	return items
}
