package vm

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func StopVmNoDiag(ctx context.Context, provider *client.NumSpotSDK, vm string) (err error) {
	// Already stopped
	if vm == "" {
		return nil
	}

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	var vmStatus *numspot.ReadVmsByIdResponse
	if vmStatus, err = numspotClient.ReadVmsByIdWithResponse(ctx, provider.SpaceID, vm); err != nil {
		return err
	}

	// VM does not exist
	if vmStatus == nil || vmStatus.JSON200 == nil {
		return nil
	}
	if *vmStatus.JSON200.State == "stopped" || *vmStatus.JSON200.State == "terminated" {
		return nil
	}

	//////////////////
	forceStop := true
	// Stop the VM
	if _, err = numspotClient.StopVmWithResponse(ctx, provider.SpaceID, vm, numspot.StopVm{ForceStop: &forceStop}); err != nil {
		return err
	}

	if _, err = utils.RetryReadUntilStateValid(ctx, vm, provider.SpaceID, []string{"stopping"}, []string{"stopped", "terminated"},
		numspotClient.ReadVmsByIdWithResponse); err != nil {
		return err
	}

	return nil
}

func StartVmNoDiag(ctx context.Context, provider *client.NumSpotSDK, vm string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	// Already running
	if vm == "" {
		return nil
	}

	var vmStatus *numspot.ReadVmsByIdResponse
	if vmStatus, err = numspotClient.ReadVmsByIdWithResponse(
		ctx,
		provider.SpaceID,
		vm,
	); err != nil {
		return err
	}

	// VM does not exist
	if vmStatus == nil || vmStatus.JSON200 == nil {
		return nil
	}

	if *vmStatus.JSON200.State == "running" || *vmStatus.JSON200.State == "terminated" {
		return nil
	}

	//////////////////
	// Start the VM
	if _, err = numspotClient.StartVmWithResponse(ctx, provider.SpaceID, vm); err != nil {
		return err
	}

	_, err = utils.RetryReadUntilStateValid(
		ctx,
		vm,
		provider.SpaceID,
		[]string{"pending"},
		[]string{"running"},
		numspotClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		return err
	}

	return nil
}

func StopVm(ctx context.Context, provider *client.NumSpotSDK, id string, diags *diag.Diagnostics) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	forceStop := true
	body := numspot.StopVm{
		ForceStop: &forceStop,
	}

	// Stop the VM
	_ = utils.ExecuteRequest(func() (*numspot.StopVmResponse, error) {
		return numspotClient.StopVmWithResponse(ctx, provider.SpaceID, id, body)
	}, http.StatusOK, diags)

	if diags.HasError() {
		return
	}

	_, err = utils.RetryReadUntilStateValid(
		ctx,
		id,
		provider.SpaceID,
		[]string{"stopping"},
		[]string{"stopped", "terminated"},
		numspotClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		diags.AddError("error while waiting for VM to stop", err.Error())
		return
	}
}

func StartVm(ctx context.Context, provider *client.NumSpotSDK, id string, diags *diag.Diagnostics) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	// Start the VM
	_ = utils.ExecuteRequest(func() (*numspot.StartVmResponse, error) {
		return numspotClient.StartVmWithResponse(ctx, provider.SpaceID, id)
	}, http.StatusOK, diags)

	if diags.HasError() {
		return
	}

	_, err = utils.RetryReadUntilStateValid(
		ctx,
		id,
		provider.SpaceID,
		[]string{"pending"},
		[]string{"running"},
		numspotClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		diags.AddError("error while waiting for VM to start", err.Error())
		return
	}
}

func VmIsDeleted(id string) bool {
	// TODO : Implement this function once Inventory allows us to check for deleted resources
	return true
}

func vmBsuFromApi(ctx context.Context, elt numspot.BsuCreated, diags *diag.Diagnostics) basetypes.ObjectValue {
	obj, diagnostics := NewBsuValue(
		BsuValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(elt.DeleteOnVmDeletion),
			"link_date":             types.StringValue(elt.LinkDate.String()),
			"state":                 types.StringPointerValue(elt.State),
			"volume_id":             types.StringPointerValue(elt.VolumeId),
			"iops":                  types.Int64Null(),
			"snapshot_id":           types.StringNull(),
			"volume_size":           types.Int64Null(),
			"volume_type":           types.StringNull(),
		},
	)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return basetypes.ObjectValue{}
	}
	objectValue, diagnostics := obj.ToObjectValue(ctx)
	diags.Append(diagnostics...)
	return objectValue
}

func vmBlockDeviceMappingFromApi(ctx context.Context, elt numspot.BlockDeviceMappingCreated, diags *diag.Diagnostics) BlockDeviceMappingsValue {
	if elt.Bsu == nil {
		return BlockDeviceMappingsValue{}
	}
	// Bsu
	bsuTf := vmBsuFromApi(ctx, *elt.Bsu, diags)
	if diags.HasError() {
		return BlockDeviceMappingsValue{}
	}

	value, diagnostics := NewBlockDeviceMappingsValue(
		BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":                 bsuTf,
			"device_name":         types.StringPointerValue(elt.DeviceName),
			"no_device":           types.StringNull(),
			"virtual_device_name": types.StringNull(),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func linkNicsFromApi(ctx context.Context, linkNic numspot.LinkNicLight, diags *diag.Diagnostics) LinkNicValue {
	if linkNic.DeviceNumber == nil {
		return LinkNicValue{}
	}
	deviceNumber := int64(*linkNic.DeviceNumber)
	value, diagnostics := NewLinkNicValue(
		LinkNicValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(linkNic.DeleteOnVmDeletion),
			"device_number":         types.Int64PointerValue(&deviceNumber),
			"link_nic_id":           types.StringPointerValue(linkNic.LinkNicId),
			"state":                 types.StringPointerValue(linkNic.State),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func linkPublicIpVmFromApi(ctx context.Context, linkPublicIp numspot.LinkPublicIpLightForVm, diags *diag.Diagnostics) LinkPublicIpValue {
	value, diagnostics := NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name": types.StringPointerValue(linkPublicIp.PublicDnsName),
			"public_ip":       types.StringPointerValue(linkPublicIp.PublicIp),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func privateIpsFromApi(ctx context.Context, privateIp numspot.PrivateIpLightForVm, diags *diag.Diagnostics) PrivateIpsValue {
	linkPublicIp := linkPublicIpVmFromApi(ctx, utils.GetPtrValue(privateIp.LinkPublicIp), diags)
	if diags.HasError() {
		return PrivateIpsValue{}
	}

	linkPublicIpObjectValue, diagnostics := linkPublicIp.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return PrivateIpsValue{}
	}

	value, diagnostics := NewPrivateIpsValue(
		PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(privateIp.IsPrimary),
			"link_public_ip":   linkPublicIpObjectValue,
			"private_dns_name": types.StringPointerValue(privateIp.PrivateDnsName),
			"private_ip":       types.StringPointerValue(privateIp.PrivateIp),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func securityGroupsFromApi(ctx context.Context, privateIp numspot.SecurityGroupLight, diags *diag.Diagnostics) SecurityGroupsValue {
	value, diagnostics := NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(privateIp.SecurityGroupId),
			"security_group_name": types.StringPointerValue(privateIp.SecurityGroupName),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func nicsFromApi(ctx context.Context, nic numspot.NicLight, diags *diag.Diagnostics) NicsValue {
	var (
		linkNics         LinkNicValue
		linkPublicIp     LinkPublicIpValue
		privateIpsTf     basetypes.ListValue
		securityGroupsTf basetypes.ListValue
	)

	if nic.LinkNic != nil {
		linkNics = linkNicsFromApi(ctx, *nic.LinkNic, diags)
	}
	linkNicsObjectValue, diagnostics := linkNics.ToObjectValue(ctx)
	diags.Append(diagnostics...)

	if nic.LinkPublicIp != nil {
		linkPublicIp = linkPublicIpVmFromApi(ctx, *nic.LinkPublicIp, diags)
	}
	linkPublicIpObjectValue, diagnostics := linkPublicIp.ToObjectValue(ctx)
	diags.Append(diagnostics...)

	var privateIps []numspot.PrivateIpLightForVm
	if nic.PrivateIps != nil {
		privateIps = *nic.PrivateIps
	}
	privateIpsTf = utils.GenericListToTfListValue(
		ctx,
		PrivateIpsValue{},
		privateIpsFromApi,
		privateIps,
		diags,
	)

	var securityGroups []numspot.SecurityGroupLight
	if nic.SecurityGroups != nil {
		securityGroups = *nic.SecurityGroups
	}
	securityGroupsTf = utils.GenericListToTfListValue(
		ctx,
		SecurityGroupsValue{},
		securityGroupsFromApi,
		securityGroups,
		diags,
	)

	value, diagnostics := NewNicsValue(
		NicsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"description":            types.StringPointerValue(nic.Description),
			"is_source_dest_checked": types.BoolPointerValue(nic.IsSourceDestChecked),
			"link_nic":               linkNicsObjectValue,
			"link_public_ip":         linkPublicIpObjectValue,
			"mac_address":            types.StringPointerValue(nic.MacAddress),
			"vpc_id":                 types.StringPointerValue(nic.VpcId),
			"nic_id":                 types.StringPointerValue(nic.NicId),
			"private_dns_name":       types.StringPointerValue(nic.PrivateDnsName),
			"private_ips":            privateIpsTf,
			"security_groups":        securityGroupsTf,
			"state":                  types.StringPointerValue(nic.State),
			"subnet_id":              types.StringPointerValue(nic.SubnetId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func placementFromHTTP(ctx context.Context, elt *numspot.Placement, diags *diag.Diagnostics) PlacementValue {
	if elt == nil {
		return PlacementValue{}
	}
	value, diagnostics := NewPlacementValue(
		PlacementValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"availability_zone_name": types.StringPointerValue(elt.AvailabilityZoneName),
			"tenancy":                types.StringPointerValue(elt.Tenancy),
		})
	diags.Append(diagnostics...)
	return value
}

func VmFromHttpToTf(ctx context.Context, http *numspot.Vm, diags *diag.Diagnostics) *VmModel {
	var (
		tagsTf types.List
		nics   = types.ListNull(NicsValue{}.Type(ctx))
	)

	// Private Ips
	var privateIps []string
	if http.PrivateIp != nil {
		privateIps = []string{*http.PrivateIp}
	}

	// Product Code
	var productCodes []string
	if http.ProductCodes != nil {
		productCodes = *http.ProductCodes
	}

	// Security Group Ids & names
	var securityGroupIds []string
	var securityGroupNames []string

	if http.SecurityGroups != nil {
		securityGroupIds = make([]string, 0, len(*http.SecurityGroups))
		securityGroupNames = make([]string, 0, len(*http.SecurityGroups))
		for _, e := range *http.SecurityGroups {
			if e.SecurityGroupId != nil && e.SecurityGroupName != nil {
				securityGroupIds = append(securityGroupIds, *e.SecurityGroupId)
				securityGroupNames = append(securityGroupNames, *e.SecurityGroupName)
			}
		}
	}

	// Block Device Mapping
	var blockDeviceMappings []numspot.BlockDeviceMappingCreated
	blockDeviceMappingTf := types.ListNull(BlockDeviceMappingsValue{}.Type(ctx))
	if http.BlockDeviceMappings != nil {
		blockDeviceMappings = *http.BlockDeviceMappings
		blockDeviceMappingTf = utils.GenericListToTfListValue(
			ctx,
			BlockDeviceMappingsValue{},
			vmBlockDeviceMappingFromApi,
			blockDeviceMappings,
			diags,
		)
	}

	// Creation Date
	var creationDate string
	if http.CreationDate != nil {
		creationDate = http.CreationDate.String()
	}

	// Tags
	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	if http.Nics != nil {
		nics = utils.GenericListToTfListValue(ctx, NicsValue{}, nicsFromApi, *http.Nics, diags)
	}

	var launchNumber basetypes.Int64Value
	if http.LaunchNumber != nil {
		launchNumber = utils.FromIntPtrToTfInt64(http.LaunchNumber)
	}

	r := VmModel{
		//
		Architecture:        types.StringPointerValue(http.Architecture),
		BlockDeviceMappings: blockDeviceMappingTf,
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
		VpcId:                types.StringPointerValue(http.VpcId),
		Nics:                 nics,
		OsFamily:             types.StringPointerValue(http.OsFamily),
		Placement:            placementFromHTTP(ctx, http.Placement, diags),
		PrivateDnsName:       types.StringPointerValue(http.PrivateDnsName),
		PrivateIp:            types.StringPointerValue(http.PrivateIp),
		//
		PrivateIps:                  utils.StringListToTfListValue(ctx, privateIps, diags),
		ProductCodes:                utils.StringListToTfListValue(ctx, productCodes, diags),
		PublicDnsName:               types.StringPointerValue(http.PublicDnsName),
		PublicIp:                    types.StringPointerValue(http.PublicIp),
		ReservationId:               types.StringPointerValue(http.ReservationId),
		RootDeviceName:              types.StringPointerValue(http.RootDeviceName),
		RootDeviceType:              types.StringPointerValue(http.RootDeviceType),
		SecurityGroupIds:            utils.StringListToTfListValue(ctx, securityGroupIds, diags),
		SecurityGroups:              utils.StringListToTfListValue(ctx, securityGroupNames, diags),
		State:                       types.StringPointerValue(http.State),
		StateReason:                 types.StringPointerValue(http.StateReason),
		SubnetId:                    types.StringPointerValue(http.SubnetId),
		Type:                        types.StringPointerValue(http.Type),
		UserData:                    types.StringPointerValue(http.UserData),
		VmInitiatedShutdownBehavior: types.StringPointerValue(http.InitiatedShutdownBehavior),
		Tags:                        tagsTf,
		LaunchNumber:                launchNumber,
	}

	var securityGroups []string
	if http.SecurityGroups != nil {
		securityGroups = make([]string, 0, len(*http.SecurityGroups))
		for _, e := range *http.SecurityGroups {
			if e.SecurityGroupId != nil {
				securityGroups = append(securityGroups, *e.SecurityGroupId)
			}
		}
	}
	if http.SecurityGroups != nil {
		listValue, _ := types.ListValueFrom(ctx, types.StringType, securityGroups)
		r.SecurityGroupIds = listValue
	}

	return &r
}

func VmFromTfToCreateRequest(ctx context.Context, tf *VmModel, diags *diag.Diagnostics) numspot.CreateVmsJSONRequestBody {
	var nicsPtr *[]numspot.NicForVmCreation
	var blockDeviceMappingPtr *[]numspot.BlockDeviceMappingVmCreation
	var placement *numspot.Placement

	if !(tf.Nics.IsNull() || tf.Nics.IsUnknown()) {
		nics := make([]numspot.NicForVmCreation, 0, len(tf.Nics.Elements()))
		diags.Append(tf.Nics.ElementsAs(ctx, &nics, true)...)
		nicsPtr = &nics
	}
	if !(tf.BlockDeviceMappings.IsNull() || tf.BlockDeviceMappings.IsUnknown()) {
		blockDeviceMapping := make([]numspot.BlockDeviceMappingVmCreation, 0, len(tf.BlockDeviceMappings.Elements()))
		diags.Append(tf.BlockDeviceMappings.ElementsAs(ctx, &blockDeviceMapping, true)...)
		blockDeviceMappingPtr = &blockDeviceMapping
	}

	if !(tf.Placement.IsNull() || tf.Placement.IsUnknown()) {
		placement = &numspot.Placement{
			AvailabilityZoneName: utils.FromTfStringToStringPtr(tf.Placement.AvailabilityZoneName),
			Tenancy:              utils.FromTfStringToStringPtr(tf.Placement.Tenancy),
		}
	}

	bootOnCreation := true
	return numspot.CreateVmsJSONRequestBody{
		BootOnCreation:              &bootOnCreation,
		ClientToken:                 utils.FromTfStringToStringPtr(tf.ClientToken),
		DeletionProtection:          utils.FromTfBoolToBoolPtr(tf.DeletionProtection),
		ImageId:                     tf.ImageId.ValueString(),
		KeypairName:                 utils.FromTfStringToStringPtr(tf.KeypairName),
		NestedVirtualization:        utils.FromTfBoolToBoolPtr(tf.NestedVirtualization),
		Nics:                        nicsPtr,
		Placement:                   placement,
		PrivateIps:                  utils.TfStringListToStringPtrList(ctx, tf.PrivateIps, diags),
		SecurityGroupIds:            utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		SecurityGroups:              utils.TfStringListToStringPtrList(ctx, tf.SecurityGroups, diags),
		SubnetId:                    tf.SubnetId.ValueString(),
		UserData:                    utils.FromTfStringToStringPtr(tf.UserData),
		VmInitiatedShutdownBehavior: utils.FromTfStringToStringPtr(tf.VmInitiatedShutdownBehavior),
		Type:                        utils.FromTfStringToStringPtr(tf.Type),
		BlockDeviceMappings:         blockDeviceMappingPtr,
		MaxVmsCount:                 utils.FromTfInt64ToIntPtr(tf.MaxVmsCount),
		MinVmsCount:                 utils.FromTfInt64ToIntPtr(tf.MinVmsCount),
	}
}

func bsuFromTf(bsu BsuValue) *numspot.BsuToUpdateVm {
	if bsu.IsNull() || bsu.IsUnknown() {
		return nil
	}

	return &numspot.BsuToUpdateVm{
		DeleteOnVmDeletion: bsu.DeleteOnVmDeletion.ValueBoolPointer(),
		VolumeId:           bsu.VolumeId.ValueStringPointer(),
	}
}

func blockDeviceMappingFromTf(bdm BlockDeviceMappingsValue) numspot.BlockDeviceMappingVmUpdate {
	attrtypes := bdm.Bsu.AttributeTypes(context.Background())
	attrVals := bdm.Bsu.Attributes()
	bsuTF, diags := NewBsuValue(attrtypes, attrVals)
	if diags.HasError() {
		return numspot.BlockDeviceMappingVmUpdate{}
	}
	bsu := bsuFromTf(bsuTF)
	return numspot.BlockDeviceMappingVmUpdate{
		Bsu:               bsu,
		DeviceName:        bdm.DeviceName.ValueStringPointer(),
		NoDevice:          bdm.NoDevice.ValueStringPointer(),
		VirtualDeviceName: bdm.VirtualDeviceName.ValueStringPointer(),
	}
}

func VmFromTfToUpdaterequest(ctx context.Context, tf *VmModel, diags *diag.Diagnostics) numspot.UpdateVmJSONRequestBody {
	blockDeviceMapping := make([]numspot.BlockDeviceMappingVmUpdate, 0, len(tf.BlockDeviceMappings.Elements()))

	for _, bdmTf := range tf.BlockDeviceMappings.Elements() {
		bdmTfRes, ok := bdmTf.(BlockDeviceMappingsValue)
		if !ok {
			diags.AddError("Failed to cast block device mapping resource", "")
			return numspot.UpdateVmJSONRequestBody{}
		}

		bdmApi := blockDeviceMappingFromTf(bdmTfRes)
		blockDeviceMapping = append(blockDeviceMapping, bdmApi)
	}

	return numspot.UpdateVmJSONRequestBody{
		DeletionProtection:          utils.FromTfBoolToBoolPtr(tf.DeletionProtection),
		KeypairName:                 utils.FromTfStringToStringPtr(tf.KeypairName),
		NestedVirtualization:        utils.FromTfBoolToBoolPtr(tf.NestedVirtualization),
		SecurityGroupIds:            utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		UserData:                    utils.FromTfStringToStringPtr(tf.UserData),
		VmInitiatedShutdownBehavior: utils.FromTfStringToStringPtr(tf.VmInitiatedShutdownBehavior),
		Type:                        utils.FromTfStringToStringPtr(tf.Type),
		BlockDeviceMappings:         &blockDeviceMapping,
		IsSourceDestChecked:         utils.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
	}
}

func VmsFromTfToAPIReadParams(ctx context.Context, tf VmsDataSourceModel, diags *diag.Diagnostics) numspot.ReadVmsParams {
	return numspot.ReadVmsParams{
		TagKeys:                              utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:                            utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:                                 utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:                                  utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		Architectures:                        utils.TfStringListToStringPtrList(ctx, tf.Architectures, diags),
		BlockDeviceMappingDeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.BlockDeviceMappingDeleteOnVmDeletion),
		BlockDeviceMappingDeviceNames:        utils.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingDeviceNames, diags),
		BlockDeviceMappingStates:             utils.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingsDataSourcetates, diags),
		BlockDeviceMappingVolumeIds:          utils.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingVolumeIds, diags),
		ClientTokens:                         utils.TfStringListToStringPtrList(ctx, tf.ClientTokens, diags),
		ImageIds:                             utils.TfStringListToStringPtrList(ctx, tf.ImageIds, diags),
		IsSourceDestChecked:                  utils.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
		KeypairNames:                         utils.TfStringListToStringPtrList(ctx, tf.KeypairNames, diags),
		LaunchNumbers:                        utils.TFInt64ListToIntListPointer(ctx, tf.LaunchNumbers, diags),
		NicDescriptions:                      utils.TfStringListToStringPtrList(ctx, tf.NicDescriptions, diags),
		NicIsSourceDestChecked:               utils.FromTfBoolToBoolPtr(tf.NicIsSourceDestChecked),
		NicLinkNicDeleteOnVmDeletion:         utils.FromTfBoolToBoolPtr(tf.NicLinkNicDeleteOnVmDeletion),
		NicLinkNicDeviceNumbers:              utils.TFInt64ListToIntListPointer(ctx, tf.NicLinkNicDeviceNumbers, diags),
		NicLinkNicLinkNicIds:                 utils.TfStringListToStringPtrList(ctx, tf.NicLinkNicLinkNicIds, diags),
		NicLinkNicStates:                     utils.TfStringListToStringPtrList(ctx, tf.NicLinkNicStates, diags),
		NicLinkPublicIpPublicIps:             utils.TfStringListToStringPtrList(ctx, tf.NicLinkPublicIpPublicIpIds, diags),
		NicMacAddresses:                      utils.TfStringListToStringPtrList(ctx, tf.NicMacAddresses, diags),
		NicNicIds:                            utils.TfStringListToStringPtrList(ctx, tf.NicNicIds, diags),
		NicPrivateIpsLinkPublicIpIds:         utils.TfStringListToStringPtrList(ctx, tf.NicPrivateIpsLinkPublicIpIds, diags),
		NicPrivateIpsPrimaryIp:               utils.FromTfBoolToBoolPtr(tf.NicPrivateIpsPrimaryIp),
		NicPrivateIpsPrivateIps:              utils.TfStringListToStringPtrList(ctx, tf.NicPrivateIpsPrivateIps, diags),
		NicSecurityGroupIds:                  utils.TfStringListToStringPtrList(ctx, tf.NicSecurityGroupIds, diags),
		NicSecurityGroupNames:                utils.TfStringListToStringPtrList(ctx, tf.NicSecurityGroupNames, diags),
		NicStates:                            utils.TfStringListToStringPtrList(ctx, tf.NicStates, diags),
		NicSubnetIds:                         utils.TfStringListToStringPtrList(ctx, tf.NicSubnetIds, diags),
		Platforms:                            utils.TfStringListToStringPtrList(ctx, tf.Platforms, diags),
		PrivateIps:                           utils.TfStringListToStringPtrList(ctx, tf.PrivateIps, diags),
		ProductCodes:                         utils.TfStringListToStringPtrList(ctx, tf.ProductCodes, diags),
		PublicIps:                            utils.TfStringListToStringPtrList(ctx, tf.PublicIps, diags),
		ReservationIds:                       utils.TfStringListToStringPtrList(ctx, tf.ReservationIds, diags),
		RootDeviceNames:                      utils.TfStringListToStringPtrList(ctx, tf.RootDeviceNames, diags),
		RootDeviceTypes:                      utils.TfStringListToStringPtrList(ctx, tf.RootDeviceTypes, diags),
		SecurityGroupIds:                     utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:                   utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames, diags),
		StateReasonMessages:                  utils.TfStringListToStringPtrList(ctx, tf.StateReasonMessages, diags),
		SubnetIds:                            utils.TfStringListToStringPtrList(ctx, tf.SubnetIds, diags),
		Tenancies:                            utils.TfStringListToStringPtrList(ctx, tf.Tenancies, diags),
		VmStateNames:                         utils.TfStringListToStringPtrList(ctx, tf.VmStateNames, diags),
		Types:                                utils.TfStringListToStringPtrList(ctx, tf.Types, diags),
		VpcIds:                               utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		NicVpcIds:                            utils.TfStringListToStringPtrList(ctx, tf.NicVpcIds, diags),
		AvailabilityZoneNames:                utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames, diags),
	}
}

func fromBsuToTFBsu(ctx context.Context, http *numspot.BsuCreated, diags *diag.Diagnostics) BsuDataSourceValue {
	if http == nil {
		return BsuDataSourceValue{}
	}

	var linkDateTf types.String

	if http.LinkDate != nil {
		linkDate := http.LinkDate.String()
		linkDateTf = types.StringPointerValue(&linkDate)
	}

	value, diagnostics := NewBsuDataSourceValue(
		BsuDataSourceValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"link_date":             linkDateTf,
			"state":                 types.StringPointerValue(http.State),
			"volume_id":             types.StringPointerValue(http.VolumeId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func fromBlockDeviceMappingsToBlockDeviceMappingsList(ctx context.Context, http numspot.BlockDeviceMappingCreated, diags *diag.Diagnostics) BlockDeviceMappingsDataSourceValue {
	bsu := fromBsuToTFBsu(ctx, http.Bsu, diags)
	if diags.HasError() {
		return BlockDeviceMappingsDataSourceValue{}
	}
	bsuObjectValue, diagnostics := bsu.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return BlockDeviceMappingsDataSourceValue{}
	}

	value, diagnostics := NewBlockDeviceMappingsDataSourceValue(
		BlockDeviceMappingsDataSourceValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":         bsuObjectValue,
			"device_name": types.StringPointerValue(http.DeviceName),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func fromLinkNicToTFLinkNic(ctx context.Context, http *numspot.LinkNicLight, diags *diag.Diagnostics) LinkNicValue {
	if http == nil {
		return LinkNicValue{}
	}
	var deviceNumberTf types.Int64

	if http.DeviceNumber != nil {
		deviceNumber := int64(*http.DeviceNumber)
		deviceNumberTf = types.Int64PointerValue(&deviceNumber)
	} else {
		deviceNumberTf = types.Int64Value(0)
	}

	value, diagnostics := NewLinkNicValue(
		LinkNicValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_number":         deviceNumberTf,
			"link_nic_id":           types.StringPointerValue(http.LinkNicId),
			"state":                 types.StringPointerValue(http.State),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func linkPublicIpForVmFromHTTPDatasource(ctx context.Context, http *numspot.LinkPublicIpLightForVm, diags *diag.Diagnostics) LinkPublicIpValue {
	if http == nil {
		return LinkPublicIpValue{}
	}

	value, diagnostics := NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name": types.StringPointerValue(http.PublicDnsName),
			"public_ip":       types.StringPointerValue(http.PublicIp),
		})
	diags.Append(diagnostics...)
	return value
}

func securityGroupsForVmFromHTTP(ctx context.Context, elt numspot.SecurityGroupLight, diags *diag.Diagnostics) SecurityGroupsValue {
	value, diagnostics := NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		})
	diags.Append(diagnostics...)
	return value
}

func fromNicsToNicsList(ctx context.Context, http numspot.NicLight, diags *diag.Diagnostics) NicsValue {
	linkNic := fromLinkNicToTFLinkNic(ctx, http.LinkNic, diags)
	linkNICObject, diagnostics := linkNic.ToObjectValue(ctx)
	diags.Append(diagnostics...)

	linkPublicIP := linkPublicIpForVmFromHTTPDatasource(ctx, http.LinkPublicIp, diags)
	linkPublicIPObject, diagnostics := linkPublicIP.ToObjectValue(ctx)
	diags.Append(diagnostics...)

	privateIps := utils.GenericListToTfListValue(ctx, PrivateIpsValue{}, privateIpsFromApi, utils.GetPtrValue(http.PrivateIps), diags)
	securityGroups := utils.GenericListToTfListValue(ctx, SecurityGroupsValue{}, securityGroupsForVmFromHTTP, utils.GetPtrValue(http.SecurityGroups), diags)

	value, diagnostics := NewNicsValue(
		NicsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"description":            types.StringPointerValue(http.Description),
			"is_source_dest_checked": types.BoolPointerValue(http.IsSourceDestChecked),
			"link_nic":               linkNICObject,
			"link_public_ip":         linkPublicIPObject,
			"mac_address":            types.StringPointerValue(http.MacAddress),
			"nic_id":                 types.StringPointerValue(http.NicId),
			"private_dns_name":       types.StringPointerValue(http.PrivateDnsName),
			"private_ips":            privateIps,
			"security_groups":        securityGroups,
			"state":                  types.StringPointerValue(http.State),
			"subnet_id":              types.StringPointerValue(http.SubnetId),
			"vpc_id":                 types.StringPointerValue(http.VpcId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func fromSecurityGroupToTFSecurityGroupList(ctx context.Context, http numspot.SecurityGroupLight, diags *diag.Diagnostics) SecurityGroupsValue {
	value, diagnostics := NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(http.SecurityGroupId),
			"security_group_name": types.StringPointerValue(http.SecurityGroupName),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func VmsFromHttpToTfDatasource(ctx context.Context, http *numspot.Vm, diags *diag.Diagnostics) *VmModelItemDataSource {
	var (
		blockDeviceMappings = types.ListNull(BlockDeviceMappingsValue{}.Type(ctx))
		nics                = types.ListNull(NicsValue{}.Type(ctx))
		securityGroups      = types.ListNull(SecurityGroupsValue{}.Type(ctx))
		productCodes        types.List
		placement           PlacementValue
		tagsList            types.List
		launchNumberTf      types.Int64
		creationDateTf      types.String
	)

	if http.BlockDeviceMappings != nil {
		blockDeviceMappings = utils.GenericListToTfListValue(
			ctx,
			BlockDeviceMappingsDataSourceValue{},
			fromBlockDeviceMappingsToBlockDeviceMappingsList,
			*http.BlockDeviceMappings,
			diags,
		)
	}

	if http.Nics != nil {
		nics = utils.GenericListToTfListValue(
			ctx,
			NicsValue{},
			fromNicsToNicsList,
			*http.Nics,
			diags,
		)
	}

	if http.SecurityGroups != nil {
		securityGroups = utils.GenericListToTfListValue(
			ctx,
			SecurityGroupsValue{},
			fromSecurityGroupToTFSecurityGroupList,
			*http.SecurityGroups,
			diags,
		)
	}

	if http.Placement != nil {
		var diagnostics diag.Diagnostics
		placement, diagnostics = NewPlacementValue(
			PlacementValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"availability_zone_name": types.StringPointerValue(http.Placement.AvailabilityZoneName),
				"tenancy":                types.StringPointerValue(http.Placement.Tenancy),
			},
		)
		diags.Append(diagnostics...)
	}

	if http.ProductCodes != nil {
		productCodes = utils.StringListToTfListValue(ctx, *http.ProductCodes, diags)
	}

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.LaunchNumber != nil {
		launchNumber := int64(*http.LaunchNumber)
		launchNumberTf = types.Int64PointerValue(&launchNumber)
	}

	if http.CreationDate != nil {
		creationDate := http.CreationDate.String()
		creationDateTf = types.StringPointerValue(&creationDate)
	}

	return &VmModelItemDataSource{
		Id:                            types.StringPointerValue(http.Id),
		State:                         types.StringPointerValue(http.State),
		BsuOptimized:                  types.BoolPointerValue(http.BsuOptimized),
		Performance:                   types.StringPointerValue(http.Performance),
		Tags:                          tagsList,
		Architecture:                  types.StringPointerValue(http.Architecture),
		BlockDeviceMappingsDataSource: blockDeviceMappings,
		ClientToken:                   types.StringPointerValue(http.ClientToken),
		CreationDate:                  creationDateTf,
		DeletionProtection:            types.BoolPointerValue(http.DeletionProtection),
		Hypervisor:                    types.StringPointerValue(http.Hypervisor),
		ImageId:                       types.StringPointerValue(http.ImageId),
		InitiatedShutdownBehavior:     types.StringPointerValue(http.InitiatedShutdownBehavior),
		IsSourceDestChecked:           types.BoolPointerValue(http.IsSourceDestChecked),
		KeypairName:                   types.StringPointerValue(http.KeypairName),
		LaunchNumber:                  launchNumberTf,
		NestedVirtualization:          types.BoolPointerValue(http.NestedVirtualization),
		Nics:                          nics,
		OsFamily:                      types.StringPointerValue(http.OsFamily),
		Placement:                     placement,
		PrivateDnsName:                types.StringPointerValue(http.PrivateDnsName),
		PrivateIp:                     types.StringPointerValue(http.PrivateIp),
		ProductCodes:                  productCodes,
		PublicDnsName:                 types.StringPointerValue(http.PublicDnsName),
		PublicIp:                      types.StringPointerValue(http.PublicIp),
		ReservationId:                 types.StringPointerValue(http.ReservationId),
		RootDeviceName:                types.StringPointerValue(http.RootDeviceName),
		RootDeviceType:                types.StringPointerValue(http.RootDeviceType),
		SecurityGroups:                securityGroups,
		StateReason:                   types.StringPointerValue(http.StateReason),
		SubnetId:                      types.StringPointerValue(http.SubnetId),
		Type:                          types.StringPointerValue(http.Type),
		UserData:                      types.StringPointerValue(http.UserData),
		VpcId:                         types.StringPointerValue(http.VpcId),
	}
}
