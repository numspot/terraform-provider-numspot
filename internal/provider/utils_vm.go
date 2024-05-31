package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func vmBsuFromApi(ctx context.Context, elt iaas.BsuCreated) (basetypes.ObjectValue, diag.Diagnostics) {
	obj, diags := resource_vm.NewBsuValue(
		resource_vm.BsuValue{}.AttributeTypes(ctx),
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

func vmBlockDeviceMappingFromApi(ctx context.Context, elt iaas.BlockDeviceMappingCreated) (resource_vm.BlockDeviceMappingsValue, diag.Diagnostics) {
	// Bsu
	bsuTf, diagnostics := vmBsuFromApi(ctx, *elt.Bsu)
	if diagnostics.HasError() {
		return resource_vm.BlockDeviceMappingsValue{}, diagnostics
	}

	return resource_vm.NewBlockDeviceMappingsValue(
		resource_vm.BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":         bsuTf,
			"device_name": types.StringPointerValue(elt.DeviceName),

			"no_device":           types.StringNull(),
			"virtual_device_name": types.StringNull(),
		},
	)
}

func linkNicsFromApi(ctx context.Context, linkNic iaas.LinkNicLight) (resource_vm.LinkNicValue, diag.Diagnostics) {
	deviceNumber := int64(*linkNic.DeviceNumber)
	return resource_vm.NewLinkNicValue(
		resource_vm.LinkNicValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(linkNic.DeleteOnVmDeletion),
			"device_number":         types.Int64PointerValue(&deviceNumber),
			"link_nic_id":           types.StringPointerValue(linkNic.LinkNicId),
			"state":                 types.StringPointerValue(linkNic.State),
		},
	)
}

func linkPublicIpVmFromApi(ctx context.Context, linkPublicIp iaas.LinkPublicIpLightForVm) (resource_vm.LinkPublicIpValue, diag.Diagnostics) {
	return resource_vm.NewLinkPublicIpValue(
		resource_vm.LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name": types.StringPointerValue(linkPublicIp.PublicDnsName),
			"public_ip":       types.StringPointerValue(linkPublicIp.PublicIp),
		},
	)
}

func privateIpsFromApi(ctx context.Context, privateIp iaas.PrivateIpLightForVm) (resource_vm.PrivateIpsValue, diag.Diagnostics) {
	linkPublicIp, diags := linkPublicIpVmFromApi(ctx, utils.GetPtrValue(privateIp.LinkPublicIp))
	if diags.HasError() {
		return resource_vm.PrivateIpsValue{}, diags
	}

	linkPublicIpObjectValue, diagnostics := linkPublicIp.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return resource_vm.PrivateIpsValue{}, diags
	}

	return resource_vm.NewPrivateIpsValue(
		resource_vm.PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(privateIp.IsPrimary),
			"link_public_ip":   linkPublicIpObjectValue,
			"private_dns_name": types.StringPointerValue(privateIp.PrivateDnsName),
			"private_ip":       types.StringPointerValue(privateIp.PrivateIp),
		},
	)
}

func securityGroupsFromApi(ctx context.Context, privateIp iaas.SecurityGroupLight) (resource_vm.SecurityGroupsValue, diag.Diagnostics) {
	return resource_vm.NewSecurityGroupsValue(
		resource_vm.SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(privateIp.SecurityGroupId),
			"security_group_name": types.StringPointerValue(privateIp.SecurityGroupName),
		},
	)
}

func nicsFromApi(ctx context.Context, nic iaas.NicLight) (resource_vm.NicsValue, diag.Diagnostics) {
	var (
		diagnosticsToReturn diag.Diagnostics
		diags               diag.Diagnostics
		linkNics            resource_vm.LinkNicValue
		linkPublicIp        resource_vm.LinkPublicIpValue
		privateIpsTf        basetypes.ListValue
		securityGroupsTf    basetypes.ListValue
	)

	deviceNumber := int64(*nic.LinkNic.DeviceNumber)

	if nic.LinkNic != nil {
		linkNics, diags = linkNicsFromApi(ctx, *nic.LinkNic)
		diagnosticsToReturn.Append(diags...)
	}
	linkNicsObjectValue, diagnostics := linkNics.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return resource_vm.NicsValue{}, diagnosticsToReturn
	}

	if nic.LinkPublicIp != nil {
		linkPublicIp, diags = linkPublicIpVmFromApi(ctx, *nic.LinkPublicIp)
		diagnosticsToReturn.Append(diags...)
	}
	linkPublicIpObjectValue, diagnostics := linkPublicIp.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return resource_vm.NicsValue{}, diagnosticsToReturn
	}

	var privateIps []iaas.PrivateIpLightForVm
	if nic.PrivateIps != nil {
		privateIps = *nic.PrivateIps
	}
	privateIpsTf, diags = utils.GenericListToTfListValue(
		ctx,
		resource_vm.PrivateIpsValue{},
		privateIpsFromApi,
		privateIps,
	)
	diagnosticsToReturn.Append(diags...)

	var securityGroups []iaas.SecurityGroupLight
	if nic.SecurityGroups != nil {
		securityGroups = *nic.SecurityGroups
	}
	securityGroupsTf, diags = utils.GenericListToTfListValue(
		ctx,
		resource_vm.SecurityGroupsValue{},
		securityGroupsFromApi,
		securityGroups,
	)
	diagnosticsToReturn.Append(diags...)

	if diagnosticsToReturn.HasError() {
		return resource_vm.NicsValue{}, diagnosticsToReturn
	}

	return resource_vm.NewNicsValue(
		resource_vm.NicsValue{}.AttributeTypes(ctx),
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

func placementFromHTTP(ctx context.Context, elt *iaas.Placement) (resource_vm.PlacementValue, diag.Diagnostics) {
	if elt == nil {
		return resource_vm.PlacementValue{}, diag.Diagnostics{}
	}
	return resource_vm.NewPlacementValue(
		resource_vm.PlacementValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"availability_zone_name": types.StringPointerValue(elt.AvailabilityZoneName),
			"tenancy":                types.StringPointerValue(elt.Tenancy),
		})
}

func VmFromHttpToTf(ctx context.Context, http *iaas.Vm) (*resource_vm.VmModel, diag.Diagnostics) {
	var (
		tagsTf types.List
		diags  diag.Diagnostics
		nics   = types.ListNull(resource_vm.NicsValue{}.Type(ctx))
	)

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

	if http.Nics != nil {
		nics, diags = utils.GenericListToTfListValue(ctx, resource_vm.NicsValue{}, nicsFromApi, *http.Nics)
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
		launchNumber = utils.FromIntPtrToTfInt64(http.LaunchNumber)
	}

	r := resource_vm.VmModel{
		//
		Architecture:        types.StringPointerValue(http.Architecture),
		BlockDeviceMappings: blockDeviceMappingTf,
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
		VpcId:                types.StringPointerValue(http.VpcId),
		Nics:                 nics,
		OsFamily:             types.StringPointerValue(http.OsFamily),
		Performance:          types.StringPointerValue(http.Performance),
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
		BootOnCreation:              types.BoolPointerValue(utils.EmptyTrueBoolPointer()), // TODO : need to have BootOnCreation in SDK
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

func VmFromTfToCreateRequest(ctx context.Context, tf *resource_vm.VmModel, diags *diag.Diagnostics) iaas.CreateVmsJSONRequestBody {
	nics := make([]iaas.NicForVmCreation, 0, len(tf.Nics.Elements()))
	diags.Append(tf.Nics.ElementsAs(ctx, &nics, true)...)

	blockDeviceMapping := make([]iaas.BlockDeviceMappingVmCreation, 0, len(tf.BlockDeviceMappings.Elements()))
	diags.Append(tf.BlockDeviceMappings.ElementsAs(ctx, &blockDeviceMapping, true)...)

	var placement *iaas.Placement
	if !(tf.Placement.IsNull() || tf.Placement.IsUnknown()) {
		placement = &iaas.Placement{
			AvailabilityZoneName: utils.FromTfStringToStringPtr(tf.Placement.AvailabilityZoneName),
			Tenancy:              utils.FromTfStringToStringPtr(tf.Placement.Tenancy),
		}
	}

	var performance *iaas.CreateVmsPerformance
	if !(tf.Performance.IsNull() || tf.Performance.IsUnknown()) {
		performance = (*iaas.CreateVmsPerformance)(utils.FromTfStringToStringPtr(tf.Performance))
	}

	return iaas.CreateVmsJSONRequestBody{
		BootOnCreation:              utils.FromTfBoolToBoolPtr(tf.BootOnCreation),
		BsuOptimized:                utils.FromTfBoolToBoolPtr(tf.BsuOptimized),
		ClientToken:                 utils.FromTfStringToStringPtr(tf.ClientToken),
		DeletionProtection:          utils.FromTfBoolToBoolPtr(tf.DeletionProtection),
		ImageId:                     tf.ImageId.ValueString(),
		KeypairName:                 utils.FromTfStringToStringPtr(tf.KeypairName),
		NestedVirtualization:        utils.FromTfBoolToBoolPtr(tf.NestedVirtualization),
		Nics:                        &nics,
		Performance:                 performance,
		Placement:                   placement,
		PrivateIps:                  utils.TfStringListToStringPtrList(ctx, tf.PrivateIps),
		SecurityGroupIds:            utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		SecurityGroups:              utils.TfStringListToStringPtrList(ctx, tf.SecurityGroups),
		SubnetId:                    utils.FromTfStringToStringPtr(tf.SubnetId),
		UserData:                    utils.FromTfStringToStringPtr(tf.UserData),
		VmInitiatedShutdownBehavior: utils.FromTfStringToStringPtr(tf.VmInitiatedShutdownBehavior),
		Type:                        utils.FromTfStringToStringPtr(tf.Type),
		BlockDeviceMappings:         &blockDeviceMapping,
		MaxVmsCount:                 utils.FromTfInt64ToIntPtr(tf.MaxVmsCount),
		MinVmsCount:                 utils.FromTfInt64ToIntPtr(tf.MinVmsCount),
	}
}

func VmFromTfToUpdaterequest(ctx context.Context, tf *resource_vm.VmModel, diagnostics *diag.Diagnostics) iaas.UpdateVmJSONRequestBody {
	blockDeviceMapping := utils.TfListToGenericList(func(a resource_vm.BlockDeviceMappingsValue) iaas.BlockDeviceMappingVmUpdate {
		var bsu iaas.BsuToUpdateVm
		a.Bsu.As(ctx, &bsu, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})

		return iaas.BlockDeviceMappingVmUpdate{
			Bsu:               &bsu,
			DeviceName:        utils.FromTfStringToStringPtr(a.DeviceName),
			NoDevice:          utils.FromTfStringToStringPtr(a.NoDevice),
			VirtualDeviceName: utils.FromTfStringToStringPtr(a.VirtualDeviceName),
		}
	}, ctx, tf.BlockDeviceMappings)

	return iaas.UpdateVmJSONRequestBody{
		BsuOptimized:                utils.FromTfBoolToBoolPtr(tf.BsuOptimized),
		DeletionProtection:          utils.FromTfBoolToBoolPtr(tf.DeletionProtection),
		KeypairName:                 utils.FromTfStringToStringPtr(tf.KeypairName),
		NestedVirtualization:        utils.FromTfBoolToBoolPtr(tf.NestedVirtualization),
		Performance:                 (*iaas.UpdateVmPerformance)(utils.FromTfStringToStringPtr(tf.Performance)),
		SecurityGroupIds:            utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		UserData:                    utils.FromTfStringToStringPtr(tf.UserData),
		VmInitiatedShutdownBehavior: utils.FromTfStringToStringPtr(tf.VmInitiatedShutdownBehavior),
		Type:                        utils.FromTfStringToStringPtr(tf.Type),
		BlockDeviceMappings:         &blockDeviceMapping,
		IsSourceDestChecked:         utils.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
	}
}

func VmsFromTfToAPIReadParams(ctx context.Context, tf VmsDataSourceModel) iaas.ReadVmsParams {
	blockDeviceMappingsLinkDates := make([]iaas.ReadVmsParams_BlockDeviceMappingLinkDates_Item, 0, len(tf.BlockDeviceMappingsLinkDates.Elements()))
	tf.BlockDeviceMappingsLinkDates.ElementsAs(ctx, &blockDeviceMappingsLinkDates, false)

	creationDates := make([]iaas.ReadVmsParams_CreationDates_Item, 0, len(tf.CreationDates.Elements()))
	tf.CreationDates.ElementsAs(ctx, &creationDates, false)

	return iaas.ReadVmsParams{
		TagKeys:                              utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:                            utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:                                 utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:                                  utils.TfStringListToStringPtrList(ctx, tf.IDs),
		Architectures:                        utils.TfStringListToStringPtrList(ctx, tf.Architectures),
		BlockDeviceMappingDeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.BlockDeviceMappingsDeleteOnVmDeletion),
		BlockDeviceMappingDeviceNames:        utils.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingsDeviceNames),
		BlockDeviceMappingLinkDates:          &blockDeviceMappingsLinkDates,
		BlockDeviceMappingStates:             utils.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingsStates),
		BlockDeviceMappingVolumeIds:          utils.TfStringListToStringPtrList(ctx, tf.BlockDeviceMappingsVolumeIds),
		ClientTokens:                         utils.TfStringListToStringPtrList(ctx, tf.ClientTokens),
		CreationDates:                        &creationDates,
		ImageIds:                             utils.TfStringListToStringPtrList(ctx, tf.ImageIds),
		IsSourceDestChecked:                  utils.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
		KeypairNames:                         utils.TfStringListToStringPtrList(ctx, tf.KeypairNames),
		LaunchNumbers:                        utils.TFInt64ListToIntListPointer(ctx, tf.LaunchNumbers),
		NicAccountIds:                        utils.TfStringListToStringPtrList(ctx, tf.NicAccountIds),
		NicDescriptions:                      utils.TfStringListToStringPtrList(ctx, tf.NicDescriptions),
		NicIsSourceDestChecked:               utils.FromTfBoolToBoolPtr(tf.NicIsSourceDestChecked),
		NicLinkNicDeleteOnVmDeletion:         utils.FromTfBoolToBoolPtr(tf.NicLinkNicDeleteOnVmDeletion),
		NicLinkNicDeviceNumbers:              utils.TFInt64ListToIntListPointer(ctx, tf.NicLinkNicDeviceNumbers),
		NicLinkNicLinkNicIds:                 utils.TfStringListToStringPtrList(ctx, tf.NicLinkNicLinkNicIds),
		NicLinkNicStates:                     utils.TfStringListToStringPtrList(ctx, tf.NicLinkNicStates),
		NicLinkPublicIpAccountIds:            utils.TfStringListToStringPtrList(ctx, tf.NicLinkPublicIpAccountIds),
		NicLinkPublicIpPublicIps:             utils.TfStringListToStringPtrList(ctx, tf.NicLinkPublicIpsPublicIps),
		NicMacAddresses:                      utils.TfStringListToStringPtrList(ctx, tf.NicMacAddresses),
		NicNicIds:                            utils.TfStringListToStringPtrList(ctx, tf.NicNicIds),
		NicPrivateIpsLinkPublicIpAccountIds:  utils.TfStringListToStringPtrList(ctx, tf.NicPrivateIpsLinkPublicIpAccountId),
		NicPrivateIpsLinkPublicIpIds:         utils.TfStringListToStringPtrList(ctx, tf.NicPrivateIpsLinkPublicIps),
		NicPrivateIpsPrimaryIp:               utils.FromTfBoolToBoolPtr(tf.NicPrivateIpsIsPrimary),
		NicPrivateIpsPrivateIps:              utils.TfStringListToStringPtrList(ctx, tf.NicPrivateIpsPrivateIps),
		NicSecurityGroupIds:                  utils.TfStringListToStringPtrList(ctx, tf.NicSecurityGroupIds),
		NicSecurityGroupNames:                utils.TfStringListToStringPtrList(ctx, tf.NicSecurityGroupNames),
		NicStates:                            utils.TfStringListToStringPtrList(ctx, tf.NicStates),
		NicSubnetIds:                         utils.TfStringListToStringPtrList(ctx, tf.NicSubnetIds),
		Platforms:                            utils.TfStringListToStringPtrList(ctx, tf.OsFamilies),
		PrivateIps:                           utils.TfStringListToStringPtrList(ctx, tf.PrivateIps),
		ProductCodes:                         utils.TfStringListToStringPtrList(ctx, tf.ProductCodes),
		PublicIps:                            utils.TfStringListToStringPtrList(ctx, tf.PublicIps),
		ReservationIds:                       utils.TfStringListToStringPtrList(ctx, tf.ReservationIds),
		RootDeviceNames:                      utils.TfStringListToStringPtrList(ctx, tf.RootDeviceNames),
		RootDeviceTypes:                      utils.TfStringListToStringPtrList(ctx, tf.RootDeviceTypes),
		SecurityGroupIds:                     utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds),
		SecurityGroupNames:                   utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupNames),
		StateReasonMessages:                  utils.TfStringListToStringPtrList(ctx, tf.StateReasonMessages),
		SubnetIds:                            utils.TfStringListToStringPtrList(ctx, tf.SubnetIds),
		Tenancies:                            utils.TfStringListToStringPtrList(ctx, tf.Tenancies),
		VmStateNames:                         utils.TfStringListToStringPtrList(ctx, tf.VmStateNames),
		Types:                                utils.TfStringListToStringPtrList(ctx, tf.VmTypes),
		VpcIds:                               utils.TfStringListToStringPtrList(ctx, tf.VpcIds),
		NicVpcIds:                            utils.TfStringListToStringPtrList(ctx, tf.NicVpcIds),
		AvailabilityZoneNames:                utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames),
	}
}

func fromBsuToTFBsu(ctx context.Context, http *iaas.BsuCreated) (datasource_vm.BsuValue, diag.Diagnostics) {
	if http == nil {
		return datasource_vm.BsuValue{}, diag.Diagnostics{}
	}

	var linkDateTf types.String

	if http.LinkDate != nil {
		linkDate := http.LinkDate.String()
		linkDateTf = types.StringPointerValue(&linkDate)
	}

	return datasource_vm.NewBsuValue(
		datasource_vm.BsuValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"link_date":             linkDateTf,
			"state":                 types.StringPointerValue(http.State),
			"volume_id":             types.StringPointerValue(http.VolumeId),
		},
	)
}

func fromBlockDeviceMappingsToBlockDeviceMappingsList(ctx context.Context, http iaas.BlockDeviceMappingCreated) (datasource_vm.BlockDeviceMappingsValue, diag.Diagnostics) {
	bsu, diags := fromBsuToTFBsu(ctx, http.Bsu)
	if diags.HasError() {
		return datasource_vm.BlockDeviceMappingsValue{}, diags
	}
	bsuObjectValue, diagnostics := bsu.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return datasource_vm.BlockDeviceMappingsValue{}, diags
	}

	return datasource_vm.NewBlockDeviceMappingsValue(
		datasource_vm.BlockDeviceMappingsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"bsu":         bsuObjectValue,
			"device_name": types.StringPointerValue(http.DeviceName),
		},
	)
}

func fromLinkNicToTFLinkNic(ctx context.Context, http *iaas.LinkNicLight) (datasource_vm.LinkNicValue, diag.Diagnostics) {
	if http == nil {
		return datasource_vm.LinkNicValue{}, diag.Diagnostics{}
	}
	var deviceNumberTf types.Int64

	if http.DeviceNumber != nil {
		deviceNumber := int64(*http.DeviceNumber)
		deviceNumberTf = types.Int64PointerValue(&deviceNumber)
	}
	return datasource_vm.NewLinkNicValue(
		datasource_vm.LinkNicValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_number":         deviceNumberTf,
			"link_nic_id":           types.StringPointerValue(http.LinkNicId),
			"state":                 types.StringPointerValue(http.State),
		},
	)
}

func linkPublicIpForVmFromHTTPDatasource(ctx context.Context, http *iaas.LinkPublicIpLightForVm) (datasource_vm.LinkPublicIpValue, diag.Diagnostics) {
	if http == nil {
		return datasource_vm.LinkPublicIpValue{}, diag.Diagnostics{}
	}

	return datasource_vm.NewLinkPublicIpValue(
		datasource_vm.LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name":      types.StringPointerValue(http.PublicDnsName),
			"public_ip":            types.StringPointerValue(http.PublicIp),
			"public_ip_account_id": types.StringPointerValue(utils.EmptyStrPointer()),
		})
}

func securityGroupsForVmFromHTTP(ctx context.Context, elt iaas.SecurityGroupLight) (resource_vm.SecurityGroupsValue, diag.Diagnostics) {
	return resource_vm.NewSecurityGroupsValue(
		resource_vm.SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		})
}

func fromNicsToNicsList(ctx context.Context, http iaas.NicLight) (datasource_vm.NicsValue, diag.Diagnostics) {
	linkNic, diags := fromLinkNicToTFLinkNic(ctx, http.LinkNic)
	if diags.HasError() {
		return datasource_vm.NicsValue{}, diags
	}

	linkPublicIp, diags := linkPublicIpForVmFromHTTPDatasource(ctx, http.LinkPublicIp)
	if diags.HasError() {
		return datasource_vm.NicsValue{}, diags
	}

	privateIps, diags := utils.GenericListToTfListValue(ctx, resource_vm.PrivateIpsValue{}, privateIpsFromApi, utils.GetPtrValue(http.PrivateIps))
	if diags.HasError() {
		return datasource_vm.NicsValue{}, diags
	}

	securityGroups, diags := utils.GenericListToTfListValue(ctx, resource_vm.SecurityGroupsValue{}, securityGroupsForVmFromHTTP, utils.GetPtrValue(http.SecurityGroups))
	if diags.HasError() {
		return datasource_vm.NicsValue{}, diags
	}

	return datasource_vm.NewNicsValue(
		datasource_vm.NicsValue{}.AttributeTypes(ctx),
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

func fromSecurityGroupToTFSecurityGroupList(ctx context.Context, http iaas.SecurityGroupLight) (datasource_vm.SecurityGroupsValue, diag.Diagnostics) {
	return datasource_vm.NewSecurityGroupsValue(
		datasource_vm.SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(http.SecurityGroupId),
			"security_group_name": types.StringPointerValue(http.SecurityGroupName),
		},
	)
}

func VmsFromHttpToTfDatasource(ctx context.Context, http *iaas.Vm) (*datasource_vm.VmModel, diag.Diagnostics) {
	var (
		blockDeviceMappings = types.ListNull(datasource_vm.BlockDeviceMappingsValue{}.Type(ctx))
		nics                = types.ListNull(datasource_vm.NicsValue{}.Type(ctx))
		securityGroups      = types.ListNull(datasource_vm.SecurityGroupsValue{}.Type(ctx))
		productCodes        types.List
		placement           datasource_vm.PlacementValue
		diags               diag.Diagnostics
		tagsList            types.List
		launchNumberTf      types.Int64
		creationDateTf      types.String
	)

	if http.BlockDeviceMappings != nil {
		blockDeviceMappings, diags = utils.GenericListToTfListValue(
			ctx,
			datasource_vm.BlockDeviceMappingsValue{},
			fromBlockDeviceMappingsToBlockDeviceMappingsList,
			*http.BlockDeviceMappings,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Nics != nil {
		nics, diags = utils.GenericListToTfListValue(
			ctx,
			datasource_vm.NicsValue{},
			fromNicsToNicsList,
			*http.Nics,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.SecurityGroups != nil {
		securityGroups, diags = utils.GenericListToTfListValue(
			ctx,
			datasource_vm.SecurityGroupsValue{},
			fromSecurityGroupToTFSecurityGroupList,
			*http.SecurityGroups,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Placement != nil {
		placement, diags = datasource_vm.NewPlacementValue(
			datasource_vm.PlacementValue{}.AttributeTypes(ctx),
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
		productCodes, diags = utils.StringListToTfListValue(ctx, *http.ProductCodes)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
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

	return &datasource_vm.VmModel{
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
