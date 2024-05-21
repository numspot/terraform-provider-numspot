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
		Type:                        tf.VmType.ValueStringPointer(),
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

func fromLinkPublicIpToTFLinkPublicIp(ctx context.Context, http *iaas.LinkPublicIpLightForVm) (datasource_vm.LinkPublicIpValue, diag.Diagnostics) {
	if http == nil {
		return datasource_vm.LinkPublicIpValue{}, diag.Diagnostics{}
	}

	return datasource_vm.NewLinkPublicIpValue(
		datasource_vm.LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name": types.StringPointerValue(http.PublicDnsName),
			"public_ip":       types.StringPointerValue(http.PublicIp),
		},
	)
}

func linkPublicIpForVmFromHTTP(ctx context.Context, http iaas.LinkPublicIpLightForVm) (datasource_vm.LinkPublicIpValue, diag.Diagnostics) {
	return datasource_vm.NewLinkPublicIpValue(
		datasource_vm.LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name": types.StringPointerValue(http.PublicDnsName),
			"public_ip":       types.StringPointerValue(http.PublicIp),
		})
}

func privateIpsForVmFromHTTP(ctx context.Context, elt iaas.PrivateIpLightForVm) (resource_vm.PrivateIpsValue, diag.Diagnostics) {
	linkPublicIp, diags := linkPublicIpForVmFromHTTP(ctx, utils.GetPtrValue(elt.LinkPublicIp))
	if diags.HasError() {
		return resource_vm.PrivateIpsValue{}, diags
	}

	return resource_vm.NewPrivateIpsValue(
		resource_vm.PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(elt.IsPrimary),
			"link_public_ip":   linkPublicIp,
			"private_dns_name": types.StringPointerValue(elt.PrivateDnsName),
			"private_ip":       types.StringPointerValue(elt.PrivateIp),
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

	linkPublicIp, diags := fromLinkPublicIpToTFLinkPublicIp(ctx, http.LinkPublicIp)
	if diags.HasError() {
		return datasource_vm.NicsValue{}, diags
	}

	privateIps, diags := utils.GenericListToTfListValue(ctx, resource_vm.PrivateIpsValue{}, privateIpsForVmFromHTTP, utils.GetPtrValue(http.PrivateIps))
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
