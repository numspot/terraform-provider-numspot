package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func vmBsuFromApi(ctx context.Context, elt iaas.BsuCreated) (basetypes.ObjectValue, diag.Diagnostics) {
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

func vmBlockDeviceMappingFromApi(ctx context.Context, elt iaas.BlockDeviceMappingCreated) (resource_vm.BlockDeviceMappingsValue, diag.Diagnostics) {
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

func VmFromHttpToTf(ctx context.Context, http *iaas.Vm) (*resource_vm.VmModel, diag.Diagnostics) {
	var (
		tagsTf types.List
		diags  diag.Diagnostics
	)

	vmsCount := utils.FromIntToTfInt64(1)

	// Private Ips
	var privateIps []string
	if http.PrivateIp != nil {
		privateIps = []string{*http.PrivateIp}
	}
	privateIpsTf, diags := utils.StringListToTfListValue(ctx, privateIps)
	if diags.HasError() {
		return nil, diags
	}

	// Product Code
	var productCodes []string
	if http.ProductCodes != nil {
		productCodes = *http.ProductCodes
	}
	productCodesTf, diags := utils.StringListToTfListValue(ctx, productCodes)
	if diags.HasError() {
		return nil, diags
	}

	// Security Group Ids & names
	var securityGroupIds []string
	var securityGroupNames []string

	if http.SecurityGroups != nil {
		securityGroupIds = make([]string, 0, len(*http.SecurityGroups))
		securityGroupNames = make([]string, 0, len(*http.SecurityGroups))
		for _, e := range *http.SecurityGroups {
			securityGroupIds = append(securityGroupIds, *e.SecurityGroupId)
			securityGroupNames = append(securityGroupNames, *e.SecurityGroupName)
		}
	}

	// Security Group Ids
	securityGroupIdsTf, diags := utils.StringListToTfListValue(ctx, securityGroupIds)
	if diags.HasError() {
		return nil, diags
	}

	// Security Groups names
	securityGroupsTf, diags := utils.StringListToTfListValue(ctx, securityGroupNames)
	if diags.HasError() {
		return nil, diags
	}

	// Block Device Mapping
	var blockDeviceMappings []iaas.BlockDeviceMappingCreated
	if http.BlockDeviceMappings != nil {
		blockDeviceMappings = *http.BlockDeviceMappings
	}
	blockDeviceMappingTf, diags := utils.GenericListToTfListValue(
		ctx,
		resource_vm.BlockDeviceMappingsValue{},
		vmBlockDeviceMappingFromApi,
		blockDeviceMappings,
	)
	if diags.HasError() {
		return nil, diags
	}

	// Creation Date
	var creationDate string
	if http.CreationDate != nil {
		creationDate = http.CreationDate.String()
	}

	// Tags
	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	r := resource_vm.VmModel{
		//
		Architecture:        types.StringPointerValue(http.Architecture),
		BlockDeviceMappings: blockDeviceMappingTf,
		BootOnCreation:      types.BoolValue(true), // FIXME Set value
		BsuOptimized:        types.BoolPointerValue(http.BsuOptimized),
		ClientToken:         types.StringPointerValue(http.ClientToken),
		CreationDate:        types.StringValue(creationDate),
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
		Tags:                        tagsTf,
	}

	if http.LaunchNumber != nil {
		launchNumber := utils.FromIntPtrToTfInt64(http.LaunchNumber)
		r.LaunchNumber = launchNumber
	}

	var securityGroups []string
	if http.SecurityGroups != nil {
		securityGroups = make([]string, 0, len(*http.SecurityGroups))
		for _, e := range *http.SecurityGroups {
			securityGroups = append(securityGroups, *e.SecurityGroupId)
		}
	}
	if http.SecurityGroups != nil {
		listValue, _ := types.ListValueFrom(ctx, types.StringType, securityGroups)
		r.SecurityGroupIds = listValue
	}

	return &r, diags
}

func VmFromTfToCreateRequest(ctx context.Context, tf *resource_vm.VmModel) iaas.CreateVmsJSONRequestBody {
	securityGroupIdsTf := make([]types.String, 0, len(tf.SecurityGroupIds.Elements()))
	tf.SecurityGroupIds.ElementsAs(ctx, &securityGroupIdsTf, false)
	securityGroupIds := []string{}
	for _, sgid := range securityGroupIdsTf {
		securityGroupIds = append(securityGroupIds, sgid.ValueString())
	}

	return iaas.CreateVmsJSONRequestBody{
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
