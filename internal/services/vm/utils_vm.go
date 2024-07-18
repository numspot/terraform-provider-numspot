package vm

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func StopVm(ctx context.Context, provider services.IProvider, id string) diag.Diagnostics {
	var diags diag.Diagnostics

	forceStop := true
	body := numspot.StopVm{
		ForceStop: &forceStop,
	}

	// Stop the VM
	_ = utils2.ExecuteRequest(func() (*numspot.StopVmResponse, error) {
		return provider.GetNumspotClient().StopVmWithResponse(ctx, provider.GetSpaceID(), id, body)
	}, http.StatusOK, &diags)

	if diags.HasError() {
		return diags
	}

	_, err := utils2.RetryReadUntilStateValid(
		ctx,
		id,
		provider.GetSpaceID(),
		[]string{"stopping"},
		[]string{"stopped"},
		provider.GetNumspotClient().ReadVmsByIdWithResponse,
	)
	if err != nil {
		diags.AddError("error while waiting for VM to stop", err.Error())
		return diags
	}

	return diags
}

func StartVm(ctx context.Context, provider services.IProvider, id string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Start the VM
	_ = utils2.ExecuteRequest(func() (*numspot.StartVmResponse, error) {
		return provider.GetNumspotClient().StartVmWithResponse(ctx, provider.GetSpaceID(), id)
	}, http.StatusOK, &diags)

	if diags.HasError() {
		return diags
	}

	_, err := utils2.RetryReadUntilStateValid(
		ctx,
		id,
		provider.GetSpaceID(),
		[]string{"pending"},
		[]string{"running"},
		provider.GetNumspotClient().ReadVmsByIdWithResponse,
	)
	if err != nil {
		diags.AddError("error while waiting for VM to start", err.Error())
		return diags
	}

	return diags
}

func vmBsuFromApi(ctx context.Context, elt numspot.BsuCreated) (basetypes.ObjectValue, diag.Diagnostics) {
	obj, diags := NewBsuValue(
		BsuValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(elt.DeleteOnVmDeletion),
			"link_date":             types.StringValue(elt.LinkDate.String()),
			"state":                 types.StringPointerValue(elt.State),
			"volume_id":             types.StringPointerValue(elt.VolumeId),

			"iops":        types.Int64Null(),
			"snapshot_id": types.StringNull(),
			"volume_size": types.Int64Null(),
			"volume_type": types.StringNull(),
		},
	)
	if diags != nil {
		return basetypes.ObjectValue{}, diags
	}
	return obj.ToObjectValue(ctx)
}

func vmBlockDeviceMappingFromApi(ctx context.Context, elt numspot.BlockDeviceMappingCreated) (BlockDeviceMappingsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	if elt.Bsu == nil {
		return BlockDeviceMappingsValue{}, diags
	}
	// Bsu
	bsuTf, diagnostics := vmBsuFromApi(ctx, *elt.Bsu)
	if diagnostics.HasError() {
		return BlockDeviceMappingsValue{}, diagnostics
	}

	return NewBlockDeviceMappingsValue(
		BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":         bsuTf,
			"device_name": types.StringPointerValue(elt.DeviceName),

			"no_device":           types.StringNull(),
			"virtual_device_name": types.StringNull(),
		},
	)
}

func linkNicsFromApi(ctx context.Context, linkNic numspot.LinkNicLight) (LinkNicValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	if linkNic.DeviceNumber == nil {
		return LinkNicValue{}, diags
	}
	deviceNumber := int64(*linkNic.DeviceNumber)
	return NewLinkNicValue(
		LinkNicValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(linkNic.DeleteOnVmDeletion),
			"device_number":         types.Int64PointerValue(&deviceNumber),
			"link_nic_id":           types.StringPointerValue(linkNic.LinkNicId),
			"state":                 types.StringPointerValue(linkNic.State),
		},
	)
}

func linkPublicIpVmFromApi(ctx context.Context, linkPublicIp numspot.LinkPublicIpLightForVm) (LinkPublicIpValue, diag.Diagnostics) {
	return NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name": types.StringPointerValue(linkPublicIp.PublicDnsName),
			"public_ip":       types.StringPointerValue(linkPublicIp.PublicIp),
		},
	)
}

func privateIpsFromApi(ctx context.Context, privateIp numspot.PrivateIpLightForVm) (PrivateIpsValue, diag.Diagnostics) {
	linkPublicIp, diags := linkPublicIpVmFromApi(ctx, utils2.GetPtrValue(privateIp.LinkPublicIp))
	if diags.HasError() {
		return PrivateIpsValue{}, diags
	}

	linkPublicIpObjectValue, diagnostics := linkPublicIp.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return PrivateIpsValue{}, diags
	}

	return NewPrivateIpsValue(
		PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(privateIp.IsPrimary),
			"link_public_ip":   linkPublicIpObjectValue,
			"private_dns_name": types.StringPointerValue(privateIp.PrivateDnsName),
			"private_ip":       types.StringPointerValue(privateIp.PrivateIp),
		},
	)
}

func securityGroupsFromApi(ctx context.Context, privateIp numspot.SecurityGroupLight) (SecurityGroupsValue, diag.Diagnostics) {
	return NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(privateIp.SecurityGroupId),
			"security_group_name": types.StringPointerValue(privateIp.SecurityGroupName),
		},
	)
}

func nicsFromApi(ctx context.Context, nic numspot.NicLight) (NicsValue, diag.Diagnostics) {
	var (
		diagnosticsToReturn diag.Diagnostics
		diags               diag.Diagnostics
		linkNics            LinkNicValue
		linkPublicIp        LinkPublicIpValue
		privateIpsTf        basetypes.ListValue
		securityGroupsTf    basetypes.ListValue
		deviceNumber        int64
	)

	if nic.LinkNic != nil {
		if nic.LinkNic.DeviceNumber == nil {
			return NicsValue{}, diags
		}
		deviceNumber = int64(*nic.LinkNic.DeviceNumber)

		linkNics, diags = linkNicsFromApi(ctx, *nic.LinkNic)
		diagnosticsToReturn.Append(diags...)
	}
	linkNicsObjectValue, diagnostics := linkNics.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return NicsValue{}, diagnosticsToReturn
	}

	if nic.LinkPublicIp != nil {
		linkPublicIp, diags = linkPublicIpVmFromApi(ctx, *nic.LinkPublicIp)
		diagnosticsToReturn.Append(diags...)
	}
	linkPublicIpObjectValue, diagnostics := linkPublicIp.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return NicsValue{}, diagnosticsToReturn
	}

	var privateIps []numspot.PrivateIpLightForVm
	if nic.PrivateIps != nil {
		privateIps = *nic.PrivateIps
	}
	privateIpsTf, diags = utils2.GenericListToTfListValue(
		ctx,
		PrivateIpsValue{},
		privateIpsFromApi,
		privateIps,
	)
	diagnosticsToReturn.Append(diags...)

	var securityGroups []numspot.SecurityGroupLight
	if nic.SecurityGroups != nil {
		securityGroups = *nic.SecurityGroups
	}
	securityGroupsTf, diags = utils2.GenericListToTfListValue(
		ctx,
		SecurityGroupsValue{},
		securityGroupsFromApi,
		securityGroups,
	)
	diagnosticsToReturn.Append(diags...)

	if diagnosticsToReturn.HasError() {
		return NicsValue{}, diagnosticsToReturn
	}

	return NewNicsValue(
		NicsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion":  types.BoolPointerValue(nic.LinkNic.DeleteOnVmDeletion),
			"description":            types.StringPointerValue(nic.Description),
			"device_number":          types.Int64PointerValue(&deviceNumber),
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

			"security_group_ids":         types.ListNull(types.StringType),
			"secondary_private_ip_count": types.Int64Null(),
		},
	)
}

func placementFromHTTP(ctx context.Context, elt *numspot.Placement) (PlacementValue, diag.Diagnostics) {
	if elt == nil {
		return PlacementValue{}, diag.Diagnostics{}
	}
	return NewPlacementValue(
		PlacementValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"availability_zone_name": types.StringPointerValue(elt.AvailabilityZoneName),
			"tenancy":                types.StringPointerValue(elt.Tenancy),
		})
}

func VmFromHttpToTf(ctx context.Context, http *numspot.Vm) (*VmModel, diag.Diagnostics) {
	var (
		tagsTf types.List
		diags  diag.Diagnostics
		nics   = types.ListNull(NicsValue{}.Type(ctx))
	)

	// Private Ips
	var privateIps []string
	if http.PrivateIp != nil {
		privateIps = []string{*http.PrivateIp}
	}
	privateIpsTf, diags := utils2.StringListToTfListValue(ctx, privateIps)
	if diags.HasError() {
		return nil, diags
	}

	// Product Code
	var productCodes []string
	if http.ProductCodes != nil {
		productCodes = *http.ProductCodes
	}
	productCodesTf, diags := utils2.StringListToTfListValue(ctx, productCodes)
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
			if e.SecurityGroupId != nil && e.SecurityGroupName != nil {
				securityGroupIds = append(securityGroupIds, *e.SecurityGroupId)
				securityGroupNames = append(securityGroupNames, *e.SecurityGroupName)
			}
		}
	}

	// Security Group Ids
	securityGroupIdsTf, diags := utils2.StringListToTfListValue(ctx, securityGroupIds)
	if diags.HasError() {
		return nil, diags
	}

	// Security Groups names
	securityGroupsTf, diags := utils2.StringListToTfListValue(ctx, securityGroupNames)
	if diags.HasError() {
		return nil, diags
	}

	// Block Device Mapping
	var blockDeviceMappings []numspot.BlockDeviceMappingCreated
	if http.BlockDeviceMappings != nil {
		blockDeviceMappings = *http.BlockDeviceMappings
	}
	blockDeviceMappingTf, diags := utils2.GenericListToTfListValue(
		ctx,
		BlockDeviceMappingsValue{},
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
		tagsTf, diags = utils2.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Nics != nil {
		nics, diags = utils2.GenericListToTfListValue(ctx, NicsValue{}, nicsFromApi, *http.Nics)
		if diags.HasError() {
			return nil, diags
		}
	}

	placement, diags := placementFromHTTP(ctx, http.Placement)
	if diags.HasError() {
		return nil, diags
	}

	var launchNumber basetypes.Int64Value
	if http.LaunchNumber != nil {
		launchNumber = utils2.FromIntPtrToTfInt64(http.LaunchNumber)
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
		Placement:            placement,
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

	return &r, diags
}

func VmFromTfToCreateRequest(ctx context.Context, tf *VmModel, diags *diag.Diagnostics) numspot.CreateVmsJSONRequestBody {
	nics := make([]numspot.NicForVmCreation, 0, len(tf.Nics.Elements()))
	diags.Append(tf.Nics.ElementsAs(ctx, &nics, true)...)

	blockDeviceMapping := make([]numspot.BlockDeviceMappingVmCreation, 0, len(tf.BlockDeviceMappings.Elements()))
	diags.Append(tf.BlockDeviceMappings.ElementsAs(ctx, &blockDeviceMapping, true)...)

	var placement *numspot.Placement
	if !(tf.Placement.IsNull() || tf.Placement.IsUnknown()) {
		placement = &numspot.Placement{
			AvailabilityZoneName: utils2.FromTfStringToStringPtr(tf.Placement.AvailabilityZoneName),
			Tenancy:              utils2.FromTfStringToStringPtr(tf.Placement.Tenancy),
		}
	}

	bootOnCreation := true
	return numspot.CreateVmsJSONRequestBody{
		BootOnCreation:              &bootOnCreation,
		ClientToken:                 utils2.FromTfStringToStringPtr(tf.ClientToken),
		DeletionProtection:          utils2.FromTfBoolToBoolPtr(tf.DeletionProtection),
		ImageId:                     tf.ImageId.ValueString(),
		KeypairName:                 utils2.FromTfStringToStringPtr(tf.KeypairName),
		NestedVirtualization:        utils2.FromTfBoolToBoolPtr(tf.NestedVirtualization),
		Nics:                        &nics,
		Placement:                   placement,
		PrivateIps:                  utils2.TfStringListToStringPtrList(ctx, tf.PrivateIps),
		SecurityGroupIds:            utils2.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		SecurityGroups:              utils2.TfStringListToStringPtrList(ctx, tf.SecurityGroups),
		SubnetId:                    tf.SubnetId.ValueString(),
		UserData:                    utils2.FromTfStringToStringPtr(tf.UserData),
		VmInitiatedShutdownBehavior: utils2.FromTfStringToStringPtr(tf.VmInitiatedShutdownBehavior),
		Type:                        utils2.FromTfStringToStringPtr(tf.Type),
		BlockDeviceMappings:         &blockDeviceMapping,
		MaxVmsCount:                 utils2.FromTfInt64ToIntPtr(tf.MaxVmsCount),
		MinVmsCount:                 utils2.FromTfInt64ToIntPtr(tf.MinVmsCount),
	}
}

func VmFromTfToUpdaterequest(ctx context.Context, tf *VmModel, diagnostics *diag.Diagnostics) numspot.UpdateVmJSONRequestBody {
	blockDeviceMapping := utils2.TfListToGenericList(func(a BlockDeviceMappingsValue) numspot.BlockDeviceMappingVmUpdate {
		var bsu numspot.BsuToUpdateVm
		a.Bsu.As(ctx, &bsu, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})

		return numspot.BlockDeviceMappingVmUpdate{
			Bsu:               &bsu,
			DeviceName:        utils2.FromTfStringToStringPtr(a.DeviceName),
			NoDevice:          utils2.FromTfStringToStringPtr(a.NoDevice),
			VirtualDeviceName: utils2.FromTfStringToStringPtr(a.VirtualDeviceName),
		}
	}, ctx, tf.BlockDeviceMappings)

	return numspot.UpdateVmJSONRequestBody{
		DeletionProtection:          utils2.FromTfBoolToBoolPtr(tf.DeletionProtection),
		KeypairName:                 utils2.FromTfStringToStringPtr(tf.KeypairName),
		NestedVirtualization:        utils2.FromTfBoolToBoolPtr(tf.NestedVirtualization),
		SecurityGroupIds:            utils2.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		UserData:                    utils2.FromTfStringToStringPtr(tf.UserData),
		VmInitiatedShutdownBehavior: utils2.FromTfStringToStringPtr(tf.VmInitiatedShutdownBehavior),
		Type:                        utils2.FromTfStringToStringPtr(tf.Type),
		BlockDeviceMappings:         &blockDeviceMapping,
		IsSourceDestChecked:         utils2.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
	}
}

func VmsFromTfToAPIReadParams(ctx context.Context, tf VmsDataSourceModel) numspot.ReadVmsParams {
	blockDeviceMappingsLinkDates := make([]numspot.ReadVmsParams_BlockDeviceMappingLinkDates_Item, 0, len(tf.BlockDeviceMappingsLinkDates.Elements()))
	tf.BlockDeviceMappingsLinkDates.ElementsAs(ctx, &blockDeviceMappingsLinkDates, false)

	creationDates := make([]numspot.ReadVmsParams_CreationDates_Item, 0, len(tf.CreationDates.Elements()))
	tf.CreationDates.ElementsAs(ctx, &creationDates, false)

	return numspot.ReadVmsParams{
		TagKeys:                              utils2.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:                            utils2.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:                                 utils2.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:                                  utils2.TfStringListToStringPtrList(ctx, tf.IDs),
		Architectures:                        utils2.TfStringListToStringPtrList(ctx, tf.Architectures),
		BlockDeviceMappingDeleteOnVmDeletion: utils2.FromTfBoolToBoolPtr(tf.BlockDeviceMappingsDeleteOnVmDeletion),
		BlockDeviceMappingDeviceNames:        utils2.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingsDeviceNames),
		BlockDeviceMappingLinkDates:          &blockDeviceMappingsLinkDates,
		BlockDeviceMappingStates:             utils2.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingsStates),
		BlockDeviceMappingVolumeIds:          utils2.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingsVolumeIds),
		ClientTokens:                         utils2.TfStringListToStringPtrList(ctx, tf.ClientTokens),
		CreationDates:                        &creationDates,
		ImageIds:                             utils2.TfStringListToStringPtrList(ctx, tf.ImageIds),
		IsSourceDestChecked:                  utils2.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
		KeypairNames:                         utils2.TfStringListToStringPtrList(ctx, tf.KeypairNames),
		LaunchNumbers:                        utils2.TFInt64ListToIntListPointer(ctx, tf.LaunchNumbers),
		NicDescriptions:                      utils2.TfStringListToStringPtrList(ctx, tf.NicDescriptions),
		NicIsSourceDestChecked:               utils2.FromTfBoolToBoolPtr(tf.NicIsSourceDestChecked),
		NicLinkNicDeleteOnVmDeletion:         utils2.FromTfBoolToBoolPtr(tf.NicLinkNicDeleteOnVmDeletion),
		NicLinkNicDeviceNumbers:              utils2.TFInt64ListToIntListPointer(ctx, tf.NicLinkNicDeviceNumbers),
		NicLinkNicLinkNicIds:                 utils2.TfStringListToStringPtrList(ctx, tf.NicLinkNicLinkNicIds),
		NicLinkNicStates:                     utils2.TfStringListToStringPtrList(ctx, tf.NicLinkNicStates),
		NicLinkPublicIpPublicIps:             utils2.TfStringListToStringPtrList(ctx, tf.NicLinkPublicIpsPublicIps),
		NicMacAddresses:                      utils2.TfStringListToStringPtrList(ctx, tf.NicMacAddresses),
		NicNicIds:                            utils2.TfStringListToStringPtrList(ctx, tf.NicNicIds),
		NicPrivateIpsLinkPublicIpIds:         utils2.TfStringListToStringPtrList(ctx, tf.NicPrivateIpsLinkPublicIps),
		NicPrivateIpsPrimaryIp:               utils2.FromTfBoolToBoolPtr(tf.NicPrivateIpsIsPrimary),
		NicPrivateIpsPrivateIps:              utils2.TfStringListToStringPtrList(ctx, tf.NicPrivateIpsPrivateIps),
		NicSecurityGroupIds:                  utils2.TfStringListToStringPtrList(ctx, tf.NicSecurityGroupIds),
		NicSecurityGroupNames:                utils2.TfStringListToStringPtrList(ctx, tf.NicSecurityGroupNames),
		NicStates:                            utils2.TfStringListToStringPtrList(ctx, tf.NicStates),
		NicSubnetIds:                         utils2.TfStringListToStringPtrList(ctx, tf.NicSubnetIds),
		Platforms:                            utils2.TfStringListToStringPtrList(ctx, tf.OsFamilies),
		PrivateIps:                           utils2.TfStringListToStringPtrList(ctx, tf.PrivateIps),
		ProductCodes:                         utils2.TfStringListToStringPtrList(ctx, tf.ProductCodes),
		PublicIps:                            utils2.TfStringListToStringPtrList(ctx, tf.PublicIps),
		ReservationIds:                       utils2.TfStringListToStringPtrList(ctx, tf.ReservationIds),
		RootDeviceNames:                      utils2.TfStringListToStringPtrList(ctx, tf.RootDeviceNames),
		RootDeviceTypes:                      utils2.TfStringListToStringPtrList(ctx, tf.RootDeviceTypes),
		SecurityGroupIds:                     utils2.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		SecurityGroupNames:                   utils2.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames),
		StateReasonMessages:                  utils2.TfStringListToStringPtrList(ctx, tf.StateReasonMessages),
		SubnetIds:                            utils2.TfStringListToStringPtrList(ctx, tf.SubnetIds),
		Tenancies:                            utils2.TfStringListToStringPtrList(ctx, tf.Tenancies),
		VmStateNames:                         utils2.TfStringListToStringPtrList(ctx, tf.VmStateNames),
		Types:                                utils2.TfStringListToStringPtrList(ctx, tf.VmTypes),
		VpcIds:                               utils2.TfStringListToStringPtrList(ctx, tf.VpcIds),
		NicVpcIds:                            utils2.TfStringListToStringPtrList(ctx, tf.NicVpcIds),
		AvailabilityZoneNames:                utils2.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames),
	}
}

func fromBsuToTFBsu(ctx context.Context, http *numspot.BsuCreated) (BsuValue, diag.Diagnostics) {
	if http == nil {
		return BsuValue{}, diag.Diagnostics{}
	}

	var linkDateTf types.String

	if http.LinkDate != nil {
		linkDate := http.LinkDate.String()
		linkDateTf = types.StringPointerValue(&linkDate)
	}

	return NewBsuValue(
		BsuValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"link_date":             linkDateTf,
			"state":                 types.StringPointerValue(http.State),
			"volume_id":             types.StringPointerValue(http.VolumeId),
		},
	)
}

func fromBlockDeviceMappingsToBlockDeviceMappingsList(ctx context.Context, http numspot.BlockDeviceMappingCreated) (BlockDeviceMappingsValue, diag.Diagnostics) {
	bsu, diags := fromBsuToTFBsu(ctx, http.Bsu)
	if diags.HasError() {
		return BlockDeviceMappingsValue{}, diags
	}
	bsuObjectValue, diagnostics := bsu.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return BlockDeviceMappingsValue{}, diags
	}

	return NewBlockDeviceMappingsValue(
		BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":         bsuObjectValue,
			"device_name": types.StringPointerValue(http.DeviceName),
		},
	)
}

func fromLinkNicToTFLinkNic(ctx context.Context, http *numspot.LinkNicLight) (LinkNicValue, diag.Diagnostics) {
	if http == nil {
		return LinkNicValue{}, diag.Diagnostics{}
	}
	var deviceNumberTf types.Int64

	if http.DeviceNumber != nil {
		deviceNumber := int64(*http.DeviceNumber)
		deviceNumberTf = types.Int64PointerValue(&deviceNumber)
	}
	return NewLinkNicValue(
		LinkNicValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_number":         deviceNumberTf,
			"link_nic_id":           types.StringPointerValue(http.LinkNicId),
			"state":                 types.StringPointerValue(http.State),
		},
	)
}

func linkPublicIpForVmFromHTTPDatasource(ctx context.Context, http *numspot.LinkPublicIpLightForVm) (LinkPublicIpValue, diag.Diagnostics) {
	if http == nil {
		return LinkPublicIpValue{}, diag.Diagnostics{}
	}

	return NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name":      types.StringPointerValue(http.PublicDnsName),
			"public_ip":            types.StringPointerValue(http.PublicIp),
			"public_ip_account_id": types.StringPointerValue(utils2.EmptyStrPointer()),
		})
}

func securityGroupsForVmFromHTTP(ctx context.Context, elt numspot.SecurityGroupLight) (SecurityGroupsValue, diag.Diagnostics) {
	return NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		})
}

func fromNicsToNicsList(ctx context.Context, http numspot.NicLight) (NicsValue, diag.Diagnostics) {
	linkNic, diags := fromLinkNicToTFLinkNic(ctx, http.LinkNic)
	if diags.HasError() {
		return NicsValue{}, diags
	}

	linkPublicIp, diags := linkPublicIpForVmFromHTTPDatasource(ctx, http.LinkPublicIp)
	if diags.HasError() {
		return NicsValue{}, diags
	}

	privateIps, diags := utils2.GenericListToTfListValue(ctx, PrivateIpsValue{}, privateIpsFromApi, utils2.GetPtrValue(http.PrivateIps))
	if diags.HasError() {
		return NicsValue{}, diags
	}

	securityGroups, diags := utils2.GenericListToTfListValue(ctx, SecurityGroupsValue{}, securityGroupsForVmFromHTTP, utils2.GetPtrValue(http.SecurityGroups))
	if diags.HasError() {
		return NicsValue{}, diags
	}

	return NewNicsValue(
		NicsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"description":            types.StringPointerValue(http.Description),
			"is_source_dest_checked": types.BoolPointerValue(http.IsSourceDestChecked),
			"link_nic":               linkNic,
			"link_public_ip":         linkPublicIp,
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
}

func fromSecurityGroupToTFSecurityGroupList(ctx context.Context, http numspot.SecurityGroupLight) (SecurityGroupsValue, diag.Diagnostics) {
	return NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(http.SecurityGroupId),
			"security_group_name": types.StringPointerValue(http.SecurityGroupName),
		},
	)
}

func VmsFromHttpToTfDatasource(ctx context.Context, http *numspot.Vm) (*VmModel, diag.Diagnostics) {
	var (
		blockDeviceMappings = types.ListNull(BlockDeviceMappingsValue{}.Type(ctx))
		nics                = types.ListNull(NicsValue{}.Type(ctx))
		securityGroups      = types.ListNull(SecurityGroupsValue{}.Type(ctx))
		productCodes        types.List
		placement           PlacementValue
		diags               diag.Diagnostics
		tagsList            types.List
		launchNumberTf      types.Int64
		creationDateTf      types.String
	)

	if http.BlockDeviceMappings != nil {
		blockDeviceMappings, diags = utils2.GenericListToTfListValue(
			ctx,
			BlockDeviceMappingsValue{},
			fromBlockDeviceMappingsToBlockDeviceMappingsList,
			*http.BlockDeviceMappings,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Nics != nil {
		nics, diags = utils2.GenericListToTfListValue(
			ctx,
			NicsValue{},
			fromNicsToNicsList,
			*http.Nics,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.SecurityGroups != nil {
		securityGroups, diags = utils2.GenericListToTfListValue(
			ctx,
			SecurityGroupsValue{},
			fromSecurityGroupToTFSecurityGroupList,
			*http.SecurityGroups,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Placement != nil {
		placement, diags = NewPlacementValue(
			PlacementValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"availability_zone_name": types.StringPointerValue(http.Placement.AvailabilityZoneName),
				"tenancy":                types.StringPointerValue(http.Placement.Tenancy),
			},
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.ProductCodes != nil {
		productCodes, diags = utils2.StringListToTfListValue(ctx, *http.ProductCodes)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Tags != nil {
		tagsList, diags = utils2.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
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

	return &VmModel{
		Id:                        types.StringPointerValue(http.Id),
		State:                     types.StringPointerValue(http.State),
		Tags:                      tagsList,
		Architecture:              types.StringPointerValue(http.Architecture),
		BlockDeviceMappings:       blockDeviceMappings,
		BsuOptimized:              types.BoolPointerValue(http.BsuOptimized),
		ClientToken:               types.StringPointerValue(http.ClientToken),
		CreationDate:              creationDateTf,
		DeletionProtection:        types.BoolPointerValue(http.DeletionProtection),
		Hypervisor:                types.StringPointerValue(http.Hypervisor),
		ImageId:                   types.StringPointerValue(http.ImageId),
		InitiatedShutdownBehavior: types.StringPointerValue(http.InitiatedShutdownBehavior),
		IsSourceDestChecked:       types.BoolPointerValue(http.IsSourceDestChecked),
		KeypairName:               types.StringPointerValue(http.KeypairName),
		LaunchNumber:              launchNumberTf,
		NestedVirtualization:      types.BoolPointerValue(http.NestedVirtualization),
		Nics:                      nics,
		OsFamily:                  types.StringPointerValue(http.OsFamily),
		Performance:               types.StringPointerValue(http.Performance),
		Placement:                 placement,
		PrivateDnsName:            types.StringPointerValue(http.PrivateDnsName),
		PrivateIp:                 types.StringPointerValue(http.PrivateIp),
		ProductCodes:              productCodes,
		PublicDnsName:             types.StringPointerValue(http.PublicDnsName),
		PublicIp:                  types.StringPointerValue(http.PublicIp),
		ReservationId:             types.StringPointerValue(http.ReservationId),
		RootDeviceName:            types.StringPointerValue(http.RootDeviceName),
		RootDeviceType:            types.StringPointerValue(http.RootDeviceType),
		SecurityGroups:            securityGroups,
		StateReason:               types.StringPointerValue(http.StateReason),
		SubnetId:                  types.StringPointerValue(http.SubnetId),
		Type:                      types.StringPointerValue(http.Type),
		UserData:                  types.StringPointerValue(http.UserData),
		VpcId:                     types.StringPointerValue(http.VpcId),
	}, nil
}
