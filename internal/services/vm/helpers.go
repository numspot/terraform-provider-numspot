package vm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services/vm/datasource_vm"
	"github.com/numspot/terraform-provider-numspot/internal/services/vm/resource_vm"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// securityGroupIDsFromVm derives the flat security_group_ids set from the read
// response's SecurityGroups objects (api.Vm exposes no top-level id list).
func securityGroupIDsFromVm(ctx context.Context, groups *[]api.SecurityGroupLight, diags *diag.Diagnostics) types.Set {
	ids := make([]string, 0)
	if groups != nil {
		for _, g := range *groups {
			if g.SecurityGroupId != nil {
				ids = append(ids, *g.SecurityGroupId)
			}
		}
	}
	return utils.FromStringListPointerToTfStringSet(ctx, &ids, diags)
}

// securityGroupNamesFromVm derives the flat security_groups (names) set from the
// read response's SecurityGroups objects.
func securityGroupNamesFromVm(ctx context.Context, groups *[]api.SecurityGroupLight, diags *diag.Diagnostics) types.Set {
	names := make([]string, 0)
	if groups != nil {
		for _, g := range *groups {
			if g.SecurityGroupName != nil {
				names = append(names, *g.SecurityGroupName)
			}
		}
	}
	return utils.FromStringListPointerToTfStringSet(ctx, &names, diags)
}

// privateIpsFromVm mirrors the read response's single PrivateIp into the
// private_ips list (the API returns a scalar, the schema exposes a list).
func privateIpsFromVm(ctx context.Context, privateIp *string, diags *diag.Diagnostics) types.List {
	ips := make([]string, 0)
	if privateIp != nil {
		ips = append(ips, *privateIp)
	}
	return utils.StringListToTfListValue(ctx, ips, diags)
}

// privateIpsFromNicLight maps the NIC's PrivateIpLightForVm objects onto the
// generated nics[].private_ips list (ListNestedAttribute of PrivateIpsValue).
func privateIpsFromNicLight(ctx context.Context, privateIps *[]api.PrivateIpLightForVm, diags *diag.Diagnostics) types.List {
	items := make([]api.PrivateIpLightForVm, 0)
	if privateIps != nil {
		items = *privateIps
	}
	return utils.GenericListToTfListValue(ctx, serializeNicPrivateIp, items, diags)
}

func serializeNicPrivateIp(ctx context.Context, item *api.PrivateIpLightForVm, diags *diag.Diagnostics) resource_vm.PrivateIpsValue {
	value, d := resource_vm.NewPrivateIpsValue(resource_vm.PrivateIpsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"is_primary":                types.BoolPointerValue(item.IsPrimary),
		"private_dns_name":          types.StringPointerValue(item.PrivateDnsName),
		"private_ip":                types.StringPointerValue(item.PrivateIp),
		"private_ip_link_public_ip": serializeNicPrivateIpLinkPublicIp(ctx, item.LinkPublicIp, diags),
	})
	diags.Append(d...)
	return value
}

func serializeNicPrivateIpLinkPublicIp(ctx context.Context, item *api.LinkPublicIpLightForVm, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(resource_vm.PrivateIpLinkPublicIpValue{}.AttributeTypes(ctx))
	}
	value, d := resource_vm.NewPrivateIpLinkPublicIpValue(resource_vm.PrivateIpLinkPublicIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"public_dns_name": types.StringPointerValue(item.PublicDnsName),
		"public_ip":       types.StringPointerValue(item.PublicIp),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

// Datasource variants: the generated datasource_vm model has its own
// PrivateIps/PrivateIpLinkPublicIp value types, distinct from the resource ones.
func privateIpsFromNicLightDs(ctx context.Context, privateIps *[]api.PrivateIpLightForVm, diags *diag.Diagnostics) types.List {
	items := make([]api.PrivateIpLightForVm, 0)
	if privateIps != nil {
		items = *privateIps
	}
	return utils.GenericListToTfListValue(ctx, serializeNicPrivateIpDs, items, diags)
}

func serializeNicPrivateIpDs(ctx context.Context, item *api.PrivateIpLightForVm, diags *diag.Diagnostics) datasource_vm.PrivateIpsValue {
	value, d := datasource_vm.NewPrivateIpsValue(datasource_vm.PrivateIpsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"is_primary":                types.BoolPointerValue(item.IsPrimary),
		"private_dns_name":          types.StringPointerValue(item.PrivateDnsName),
		"private_ip":                types.StringPointerValue(item.PrivateIp),
		"private_ip_link_public_ip": serializeNicPrivateIpLinkPublicIpDs(ctx, item.LinkPublicIp, diags),
	})
	diags.Append(d...)
	return value
}

func serializeNicPrivateIpLinkPublicIpDs(ctx context.Context, item *api.LinkPublicIpLightForVm, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_vm.PrivateIpLinkPublicIpValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_vm.NewPrivateIpLinkPublicIpValue(datasource_vm.PrivateIpLinkPublicIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"public_dns_name": types.StringPointerValue(item.PublicDnsName),
		"public_ip":       types.StringPointerValue(item.PublicIp),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

// stringSetToPtrList converts a string set to *[]string, returning nil (not an
// empty slice) when the set is null/unknown so the request omits the field.
func stringSetToPtrList(ctx context.Context, set types.Set, diags *diag.Diagnostics) *[]string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	l := utils.FromTfStringSetToStringList(ctx, set, diags)
	return &l
}

// deserializeCreateBlockDeviceMappings builds the create body's block device
// mappings (api.BlockDeviceMappingVmCreation, distinct from the read type).
func deserializeCreateBlockDeviceMappings(ctx context.Context, list types.List, diags *diag.Diagnostics) *[]api.BlockDeviceMappingVmCreation {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	out := make([]api.BlockDeviceMappingVmCreation, 0, len(list.Elements()))
	for _, e := range list.Elements() {
		bdm, ok := e.(resource_vm.BlockDeviceMappingsValue)
		if !ok {
			diags.AddError("unable to cast block device mapping", "")
			return nil
		}
		bsuTF, d := resource_vm.NewBsuValue(bdm.Bsu.AttributeTypes(ctx), bdm.Bsu.Attributes())
		diags.Append(d...)
		out = append(out, api.BlockDeviceMappingVmCreation{
			Bsu:               bsuToCreate(&bsuTF),
			DeviceName:        utils.FromTfStringToStringPtr(bdm.DeviceName),
			NoDevice:          utils.FromTfStringToStringPtr(bdm.NoDevice),
			VirtualDeviceName: utils.FromTfStringToStringPtr(bdm.VirtualDeviceName),
		})
	}
	return &out
}

func bsuToCreate(bsu *resource_vm.BsuValue) *api.BsuToCreate {
	if bsu == nil || bsu.IsNull() || bsu.IsUnknown() {
		return nil
	}
	return &api.BsuToCreate{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(bsu.DeleteOnVmDeletion),
		Iops:               utils.FromTfInt64ToIntPtr(bsu.Iops),
		SnapshotId:         utils.FromTfStringToStringPtr(bsu.SnapshotId),
		VolumeSize:         utils.FromTfInt64ToIntPtr(bsu.VolumeSize),
		VolumeType:         utils.FromTfStringToStringPtr(bsu.VolumeType),
	}
}

// deserializeUpdateVm builds the shared update body. UpdateVMKeypair and
// UpdateVMAttributes each null out the fields they do not own.
func deserializeUpdateVm(ctx context.Context, tf *resource_vm.VmModel, diags *diag.Diagnostics) *api.UpdateVmJSONRequestBody {
	bdms := make([]api.BlockDeviceMappingVmUpdate, 0, len(tf.BlockDeviceMappings.Elements()))
	for _, e := range tf.BlockDeviceMappings.Elements() {
		bdm, ok := e.(resource_vm.BlockDeviceMappingsValue)
		if !ok {
			diags.AddError("unable to cast block device mapping", "")
			return &api.UpdateVmJSONRequestBody{}
		}
		bsuTF, d := resource_vm.NewBsuValue(bdm.Bsu.AttributeTypes(ctx), bdm.Bsu.Attributes())
		diags.Append(d...)
		bdms = append(bdms, api.BlockDeviceMappingVmUpdate{
			Bsu:               bsuToUpdate(&bsuTF),
			DeviceName:        bdm.DeviceName.ValueStringPointer(),
			NoDevice:          bdm.NoDevice.ValueStringPointer(),
			VirtualDeviceName: bdm.VirtualDeviceName.ValueStringPointer(),
		})
	}
	return &api.UpdateVmJSONRequestBody{
		DeletionProtection:          utils.FromTfBoolToBoolPtr(tf.DeletionProtection),
		KeypairName:                 utils.FromTfStringToStringPtr(tf.KeypairName),
		NestedVirtualization:        utils.FromTfBoolToBoolPtr(tf.NestedVirtualization),
		SecurityGroupIds:            stringSetToPtrList(ctx, tf.SecurityGroupIds, diags),
		UserData:                    utils.FromTfStringToStringPtr(tf.UserData),
		VmInitiatedShutdownBehavior: utils.FromTfStringToStringPtr(tf.InitiatedShutdownBehavior),
		Type:                        utils.PointerOf(api.UpdateVmType(tf.Type.ValueString())),
		BlockDeviceMappings:         &bdms,
		IsSourceDestChecked:         utils.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
	}
}

func bsuToUpdate(bsu *resource_vm.BsuValue) *api.BsuToUpdateVm {
	if bsu == nil || bsu.IsNull() || bsu.IsUnknown() {
		return nil
	}
	return &api.BsuToUpdateVm{
		DeleteOnVmDeletion: bsu.DeleteOnVmDeletion.ValueBoolPointer(),
		VolumeId:           bsu.VolumeId.ValueStringPointer(),
	}
}

// enrichBlockDeviceMappings fills the block device mapping creation fields
// (volume_size, volume_type, iops, snapshot_id) that the VM read payload
// (BsuCreated) does not carry, by reading each volume from the Volumes API.
// strict=true (right after Create): a volume read failure is an error.
// strict=false (refresh/import): a failure leaves the fields null.
func enrichBlockDeviceMappings(
	ctx context.Context,
	provider *client.NumspotSDK,
	state *resource_vm.VmModel,
	strict bool,
	diags *diag.Diagnostics,
) {
	if state.BlockDeviceMappings.IsNull() || state.BlockDeviceMappings.IsUnknown() {
		return
	}

	elements := state.BlockDeviceMappings.Elements()
	enriched := make([]attr.Value, 0, len(elements))
	changed := false

	for _, e := range elements {
		bdm, ok := e.(resource_vm.BlockDeviceMappingsValue)
		if !ok {
			diags.AddError("unable to cast block device mapping", "")
			return
		}

		bdmChanged := enrichBlockDeviceMapping(ctx, provider, &bdm, strict, diags)
		if diags.HasError() {
			return
		}
		changed = changed || bdmChanged
		enriched = append(enriched, bdm)
	}

	if !changed {
		return
	}
	list, listDiags := types.ListValueFrom(ctx, resource_vm.BlockDeviceMappingsValue{}.Type(ctx), enriched)
	diags.Append(listDiags...)
	if diags.HasError() {
		return
	}
	state.BlockDeviceMappings = list
}

// enrichBlockDeviceMapping enriches a single block device mapping by reading
// its volume from the Volumes API, updating bdm in place. It reports whether
// the mapping was enriched.
func enrichBlockDeviceMapping(
	ctx context.Context,
	provider *client.NumspotSDK,
	bdm *resource_vm.BlockDeviceMappingsValue,
	strict bool,
	diags *diag.Diagnostics,
) bool {
	if bdm.Bsu.IsNull() || bdm.Bsu.IsUnknown() {
		return false
	}

	bsu, bsuAsDiags := resource_vm.NewBsuValue(bdm.Bsu.AttributeTypes(ctx), bdm.Bsu.Attributes())
	diags.Append(bsuAsDiags...)
	if diags.HasError() {
		return false
	}

	volumeID := bsu.VolumeId.ValueString()
	if volumeID == "" {
		return false
	}

	volume, err := core.ReadVolume(ctx, provider, volumeID)
	if err != nil {
		if strict {
			diags.AddError("unable to read created vm volume", err.Error())
		}
		return false
	}

	bsuObj, ok := enrichedBsuObject(ctx, &bsu, volume, diags)
	if !ok {
		return false
	}

	newBdm, bdmDiags := resource_vm.NewBlockDeviceMappingsValue(
		resource_vm.BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":                 bsuObj,
			"device_name":         bdm.DeviceName,
			"no_device":           bdm.NoDevice,
			"virtual_device_name": bdm.VirtualDeviceName,
		},
	)
	diags.Append(bdmDiags...)
	if diags.HasError() {
		return false
	}
	*bdm = newBdm
	return true
}

// enrichedBsuObject rebuilds the Bsu object with the creation fields
// (volume_size, volume_type, iops, snapshot_id) filled from the volume read.
// It returns a null object and false when the diagnostics carry an error.
func enrichedBsuObject(
	ctx context.Context,
	bsu *resource_vm.BsuValue,
	volume *api.Volume,
	diags *diag.Diagnostics,
) (types.Object, bool) {
	enrichedBsu, bsuDiags := resource_vm.NewBsuValue(
		resource_vm.BsuValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": bsu.DeleteOnVmDeletion,
			"link_date":             bsu.LinkDate,
			"state":                 bsu.State,
			"volume_id":             bsu.VolumeId,
			"iops":                  utils.FromIntPtrToTfInt64(volume.Iops),
			"snapshot_id":           types.StringPointerValue(volume.SnapshotId),
			"volume_size":           utils.FromIntPtrToTfInt64(volume.Size),
			"volume_type":           types.StringPointerValue(volume.Type),
		},
	)
	diags.Append(bsuDiags...)
	if diags.HasError() {
		return types.ObjectNull(nil), false
	}
	bsuObj, objDiags := enrichedBsu.ToObjectValue(ctx)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(nil), false
	}
	return bsuObj, true
}
