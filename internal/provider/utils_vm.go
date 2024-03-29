package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func vmBsuFromApi(ctx context.Context, elt api.BsuCreated) (basetypes.ObjectValue, diag.Diagnostics) {
	obj, diags := resource_vm.NewBsuValue(
		resource_vm.BsuValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(elt.DeleteOnVmDeletion),
			"iops":                  types.Int64Null(), // FIXME Not set
			"link_date":             types.StringValue(elt.LinkDate.String()),
			"snapshot_id":           types.StringNull(), // FIXME Not set
			"state":                 types.StringNull(), // FIXME Not set
			"volume_id":             types.StringPointerValue(elt.VolumeId),
			"volume_size":           types.Int64Null(),  // FIXME Not set
			"volume_type":           types.StringNull(), // FIXME Not set
		},
	)
	if diags != nil {
		return basetypes.ObjectValue{}, diags
	}
	return obj.ToObjectValue(ctx)

}

func vmBlockDeviceMappingFromApi(ctx context.Context, elt api.BlockDeviceMappingCreated) (resource_vm.BlockDeviceMappingsValue, diag.Diagnostics) {
	// Bsu
	bsuTf, diagnostics := vmBsuFromApi(ctx, *elt.Bsu)
	if diagnostics.HasError() {
		return resource_vm.BlockDeviceMappingsValue{}, diagnostics
	}

	return resource_vm.NewBlockDeviceMappingsValue(
		resource_vm.BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":                 bsuTf,
			"device_name":         types.StringPointerValue(elt.DeviceName),
			"no_device":           types.StringNull(),
			"virtual_device_name": types.StringNull(),
		},
	)
}

func VmFromHttpToTf(ctx context.Context, http *api.Vm) (*resource_vm.VmModel, diag.Diagnostics) {
	vmsCount := utils.FromIntToTfInt64(1)

	// Private Ips
	privateIpsTf, diagnostics := utils.StringListToTfListValue(ctx, []string{*http.PrivateIp})
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Product Code
	productCodesTf, diagnostics := utils.StringListToTfListValue(ctx, *http.ProductCodes)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Security Group Ids
	securityGroupIds := make([]string, 0, len(*http.SecurityGroups))
	for _, e := range *http.SecurityGroups {
		securityGroupIds = append(securityGroupIds, *e.SecurityGroupId)
	}

	securityGroupIdsTf, diagnostics := utils.StringListToTfListValue(ctx, securityGroupIds)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Security Groups names
	securityGroupNames := make([]string, 0, len(*http.SecurityGroups))
	for _, e := range *http.SecurityGroups {
		securityGroupNames = append(securityGroupNames, *e.SecurityGroupName)
	}

	securityGroupsTf, diagnostics := utils.StringListToTfListValue(ctx, securityGroupNames)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Block Device Mapping
	blockDeviceMappingTf, diagnostics := utils.GenericListToTfListValue(
		ctx,
		resource_vm.BlockDeviceMappingsValue{},
		vmBlockDeviceMappingFromApi,
		*http.BlockDeviceMappings,
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	r := resource_vm.VmModel{
		//
		Architecture:        types.StringPointerValue(http.Architecture),
		BlockDeviceMappings: blockDeviceMappingTf,
		BootOnCreation:      types.BoolValue(true), // FIXME Set value
		BsuOptimized:        types.BoolPointerValue(http.BsuOptimized),
		ClientToken:         types.StringPointerValue(http.ClientToken),
		CreationDate:        types.StringValue(http.CreationDate.String()),
		//
		DeletionProtection:        types.BoolPointerValue(http.DeletionProtection),
		Hypervisor:                types.StringPointerValue(http.Hypervisor),
		Id:                        types.StringPointerValue(http.Id),
		ImageId:                   types.StringPointerValue(http.ImageId),
		InitiatedShutdownBehavior: types.StringPointerValue(http.InitiatedShutdownBehavior),
		IsSourceDestChecked:       types.BoolPointerValue(http.IsSourceDestChecked),
		KeypairName:               types.StringPointerValue(http.KeypairName),
		//
		NestedVirtualization: types.BoolPointerValue(http.NestedVirtualization),
		NetId:                types.StringPointerValue(http.VpcId),
		Nics:                 types.ListNull(resource_vm.NicsValue{}.Type(ctx)), // FIXME Set value
		OsFamily:             types.StringPointerValue(http.OsFamily),
		Performance:          types.StringPointerValue(http.Performance),
		Placement:            resource_vm.PlacementValue{}, // FIXME Set value
		PrivateDnsName:       types.StringPointerValue(http.PrivateDnsName),
		PrivateIp:            types.StringPointerValue(http.PrivateIp),
		//
		PrivateIps:                  privateIpsTf,
		ProductCodes:                productCodesTf,
		PublicDnsName:               types.StringPointerValue(http.PublicDnsName),
		PublicIp:                    types.StringPointerValue(http.PublicIp),
		ReservationId:               types.StringPointerValue(http.ReservationId),
		RootDeviceName:              types.StringPointerValue(http.RootDeviceName),
		RootDeviceType:              types.StringPointerValue(http.RootDeviceType),
		SecurityGroupIds:            securityGroupIdsTf,
		SecurityGroups:              securityGroupsTf,
		State:                       types.StringPointerValue(http.State),
		StateReason:                 types.StringPointerValue(http.StateReason),
		SubnetId:                    types.StringPointerValue(http.SubnetId),
		Type:                        types.StringPointerValue(http.Type),
		UserData:                    types.StringPointerValue(http.UserData),
		VmInitiatedShutdownBehavior: types.StringPointerValue(http.InitiatedShutdownBehavior),
		VmType:                      types.StringPointerValue(http.Type),
		VmsCount:                    vmsCount,
	}

	if http.LaunchNumber != nil {
		launchNumber := utils.FromIntPtrToTfInt64(http.LaunchNumber)
		r.LaunchNumber = launchNumber
	}

	if http.SecurityGroups != nil {
		sg := make([]string, 0, len(*http.SecurityGroups))
		for _, e := range *http.SecurityGroups {
			sg = append(sg, *e.SecurityGroupId)
		}
		listValue, _ := types.ListValueFrom(ctx, types.StringType, sg)
		r.SecurityGroupIds = listValue
	}

	return &r, nil
}

func VmFromTfToCreateRequest(ctx context.Context, tf *resource_vm.VmModel) api.CreateVmsJSONRequestBody {
	securityGroupIdsTf := make([]types.String, 0, len(tf.SecurityGroupIds.Elements()))
	tf.SecurityGroupIds.ElementsAs(ctx, &securityGroupIdsTf, false)
	securityGroupIds := []string{}
	for _, sgid := range securityGroupIdsTf {
		securityGroupIds = append(securityGroupIds, sgid.ValueString())
	}

	return api.CreateVmsJSONRequestBody{
		BootOnCreation:              nil,
		BsuOptimized:                nil,
		ClientToken:                 nil,
		DeletionProtection:          nil,
		ImageId:                     tf.ImageId.ValueString(),
		KeypairName:                 tf.KeypairName.ValueStringPointer(),
		NestedVirtualization:        nil,
		Nics:                        nil,
		Performance:                 nil,
		Placement:                   nil,
		PrivateIps:                  nil,
		SecurityGroupIds:            &securityGroupIds,
		SecurityGroups:              nil,
		SubnetId:                    tf.SubnetId.ValueStringPointer(),
		UserData:                    nil,
		VmInitiatedShutdownBehavior: nil,
		VmType:                      tf.VmType.ValueStringPointer(),
	}
}
