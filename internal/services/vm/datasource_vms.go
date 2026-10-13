package vm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/vm/datasource_vm"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &vmsDataSource{}

// vmsDataSource is the Terraform data source implementation for numspot_vms.
type vmsDataSource struct {
	services.BaseDataSource
}

// NewVMsDataSource returns a new vmsDataSource instance.
func NewVMsDataSource() datasource.DataSource {
	return &vmsDataSource{}
}

// Metadata sets the data source type name.
func (d *vmsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vms"
}

// Schema defines the schema for the data source.
func (d *vmsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_vm.VmDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *vmsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_vm.VmModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadVMs(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadVMsWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read vms", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.Vm) (datasource_vm.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	tagsList := types.ListNull(datasource_vm.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_vm.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_vm.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_vm.NewItemsValue(datasource_vm.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"architecture":                types.StringValue(utils.ConvertStringPtrToString(item.Architecture)),
		"block_device_mappings":       mappingBlockDeviceMappings(ctx, item, &diags),
		"bsu_optimized":               types.BoolValue(*item.BsuOptimized),
		"client_token":                types.StringValue(utils.ConvertStringPtrToString(item.ClientToken)),
		"creation_date":               utils.TimePtrToTfString(item.CreationDate),
		"deletion_protection":         types.BoolValue(*item.DeletionProtection),
		"hypervisor":                  types.StringValue(utils.ConvertStringPtrToString(item.Hypervisor)),
		"id":                          types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"image_id":                    types.StringValue(utils.ConvertStringPtrToString(item.ImageId)),
		"initiated_shutdown_behavior": types.StringValue(utils.ConvertStringPtrToString(item.InitiatedShutdownBehavior)),
		"is_source_dest_checked":      types.BoolValue(*item.IsSourceDestChecked),
		"keypair_name":                types.StringValue(utils.ConvertStringPtrToString(item.KeypairName)),
		"launch_number":               utils.FromIntPtrToTfInt64(item.LaunchNumber),
		"nested_virtualization":       types.BoolValue(*item.NestedVirtualization),
		"nics":                        mappingNics(ctx, item, &diags),
		"os_family":                   types.StringValue(utils.ConvertStringPtrToString(item.OsFamily)),
		"placement":                   mappingPlacement(ctx, item.Placement, &diags),
		"private_dns_name":            types.StringValue(utils.ConvertStringPtrToString(item.PrivateDnsName)),
		"private_ip":                  types.StringValue(utils.ConvertStringPtrToString(item.PrivateIp)),
		"product_codes":               utils.FromStringListPointerToTfStringList(ctx, item.ProductCodes, &diags),
		"public_dns_name":             types.StringValue(utils.ConvertStringPtrToString(item.PublicDnsName)),
		"public_ip":                   types.StringValue(utils.ConvertStringPtrToString(item.PublicIp)),
		"reservation_id":              types.StringValue(utils.ConvertStringPtrToString(item.ReservationId)),
		"root_device_name":            types.StringValue(utils.ConvertStringPtrToString(item.RootDeviceName)),
		"root_device_type":            types.StringValue(utils.ConvertStringPtrToString(item.RootDeviceType)),
		"security_groups":             mappingSecurityGroups(ctx, item, &diags),
		"state":                       types.StringValue(utils.ConvertStringPtrToString(item.State)),
		"state_reason":                types.StringValue(utils.ConvertStringPtrToString(item.StateReason)),
		"subnet_id":                   types.StringValue(utils.ConvertStringPtrToString(item.SubnetId)),
		"type":                        types.StringValue(utils.ConvertStringPtrToString(item.Type)),
		"user_data":                   types.StringValue(utils.ConvertStringPtrToString(item.UserData)),
		"vpc_id":                      types.StringValue(utils.ConvertStringPtrToString(item.VpcId)),
		"tags":                        tagsList,
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_vm.TagsValue, diag.Diagnostics) {
	return datasource_vm.NewTagsValue(datasource_vm.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingBsu(ctx context.Context, item *api.BsuCreated, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_vm.BsuValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_vm.NewBsuValue(datasource_vm.BsuValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(item.DeleteOnVmDeletion),
		"link_date":             utils.TimePtrToTfString(item.LinkDate),
		"state":                 types.StringPointerValue(item.State),
		"volume_id":             types.StringPointerValue(item.VolumeId),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingBlockDeviceMapping(ctx context.Context, item *api.BlockDeviceMappingCreated) (datasource_vm.BlockDeviceMappingsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_vm.NewBlockDeviceMappingsValue(datasource_vm.BlockDeviceMappingsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"bsu":         mappingBsu(ctx, item.Bsu, &diags),
		"device_name": types.StringPointerValue(item.DeviceName),
	})
	diags.Append(d...)
	return value, diags
}

func mappingBlockDeviceMappings(ctx context.Context, item *api.Vm, diags *diag.Diagnostics) types.List {
	if item.BlockDeviceMappings == nil {
		return types.ListNull(datasource_vm.BlockDeviceMappingsValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.BlockDeviceMappings, mappingBlockDeviceMapping)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_vm.BlockDeviceMappingsValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingLinkNic(ctx context.Context, item *api.LinkNicLight, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_vm.LinkNicValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_vm.NewLinkNicValue(datasource_vm.LinkNicValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(item.DeleteOnVmDeletion),
		"device_number":         utils.FromIntPtrToTfInt64(item.DeviceNumber),
		"link_nic_id":           types.StringPointerValue(item.LinkNicId),
		"state":                 types.StringPointerValue(item.State),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingNicLinkPublicIp(ctx context.Context, item *api.LinkPublicIpLightForVm, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_vm.NicLinkPublicIpValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_vm.NewNicLinkPublicIpValue(datasource_vm.NicLinkPublicIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"public_dns_name": types.StringPointerValue(item.PublicDnsName),
		"public_ip":       types.StringPointerValue(item.PublicIp),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingNicSecurityGroup(ctx context.Context, item *api.SecurityGroupLight) (datasource_vm.NicSecurityGroupsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_vm.NewNicSecurityGroupsValue(datasource_vm.NicSecurityGroupsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"security_group_id":   types.StringPointerValue(item.SecurityGroupId),
		"security_group_name": types.StringPointerValue(item.SecurityGroupName),
	})
	diags.Append(d...)
	return value, diags
}

func mappingNicSecurityGroups(ctx context.Context, item *api.NicLight, diags *diag.Diagnostics) types.List {
	if item.SecurityGroups == nil {
		return types.ListNull(datasource_vm.NicSecurityGroupsValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.SecurityGroups, mappingNicSecurityGroup)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_vm.NicSecurityGroupsValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingNic(ctx context.Context, item *api.NicLight) (datasource_vm.NicsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_vm.NewNicsValue(datasource_vm.NicsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"description":            types.StringPointerValue(item.Description),
		"is_source_dest_checked": types.BoolPointerValue(item.IsSourceDestChecked),
		"link_nic":               mappingLinkNic(ctx, item.LinkNic, &diags),
		"mac_address":            types.StringPointerValue(item.MacAddress),
		"nic_id":                 types.StringPointerValue(item.NicId),
		"nic_link_public_ip":     mappingNicLinkPublicIp(ctx, item.LinkPublicIp, &diags),
		"nic_security_groups":    mappingNicSecurityGroups(ctx, item, &diags),
		"private_dns_name":       types.StringPointerValue(item.PrivateDnsName),
		"private_ips":            privateIpsFromNicLightDs(ctx, item.PrivateIps, &diags),
		"state":                  types.StringPointerValue(item.State),
		"subnet_id":              types.StringPointerValue(item.SubnetId),
		"vpc_id":                 types.StringPointerValue(item.VpcId),
	})
	diags.Append(d...)
	return value, diags
}

func mappingNics(ctx context.Context, item *api.Vm, diags *diag.Diagnostics) types.List {
	if item.Nics == nil {
		return types.ListNull(datasource_vm.NicsValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Nics, mappingNic)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_vm.NicsValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingPlacement(ctx context.Context, item *api.Placement, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_vm.PlacementValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_vm.NewPlacementValue(datasource_vm.PlacementValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(item.AvailabilityZoneName)),
		"tenancy":                types.StringPointerValue(item.Tenancy),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingSecurityGroup(ctx context.Context, item *api.SecurityGroupLight) (datasource_vm.SecurityGroupsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_vm.NewSecurityGroupsValue(datasource_vm.SecurityGroupsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"security_group_id":   types.StringPointerValue(item.SecurityGroupId),
		"security_group_name": types.StringPointerValue(item.SecurityGroupName),
	})
	diags.Append(d...)
	return value, diags
}

func mappingSecurityGroups(ctx context.Context, item *api.Vm, diags *diag.Diagnostics) types.List {
	if item.SecurityGroups == nil {
		return types.ListNull(datasource_vm.SecurityGroupsValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.SecurityGroups, mappingSecurityGroup)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_vm.SecurityGroupsValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func deserializeReadVMs(ctx context.Context, tf *datasource_vm.VmModel, diags *diag.Diagnostics) *api.ReadVmsParams {
	return &api.ReadVmsParams{
		Architectures:                        utils.ConvertTfListToArrayOfString(ctx, tf.Architectures, diags),
		BlockDeviceMappingDeleteOnVmDeletion: tf.BlockDeviceMappingDeleteOnVmDeletion.ValueBoolPointer(),
		BlockDeviceMappingDeviceNames:        utils.ConvertTfListToArrayOfString(ctx, tf.BlockDeviceMappingDeviceNames, diags),
		BlockDeviceMappingStates:             utils.ConvertTfListToArrayOfString(ctx, tf.BlockDeviceMappingStates, diags),
		BlockDeviceMappingVolumeIds:          utils.ConvertTfListToArrayOfString(ctx, tf.BlockDeviceMappingVolumeIds, diags),
		ClientTokens:                         utils.ConvertTfListToArrayOfString(ctx, tf.ClientTokens, diags),
		ImageIds:                             utils.ConvertTfListToArrayOfString(ctx, tf.ImageIds, diags),
		IsSourceDestChecked:                  tf.IsSourceDestChecked.ValueBoolPointer(),
		KeypairNames:                         utils.ConvertTfListToArrayOfString(ctx, tf.KeypairNames, diags),
		LaunchNumbers:                        utils.ConvertTfListToArrayOfInt(ctx, tf.LaunchNumbers, diags),
		Lifecycles:                           utils.ConvertTfListToArrayOfString(ctx, tf.Lifecycles, diags),
		NicDescriptions:                      utils.ConvertTfListToArrayOfString(ctx, tf.NicDescriptions, diags),
		NicIsSourceDestChecked:               tf.NicIsSourceDestChecked.ValueBoolPointer(),
		NicLinkNicDeleteOnVmDeletion:         tf.NicLinkNicDeleteOnVmDeletion.ValueBoolPointer(),
		NicLinkNicDeviceNumbers:              utils.ConvertTfListToArrayOfInt(ctx, tf.NicLinkNicDeviceNumbers, diags),
		NicLinkNicLinkNicIds:                 utils.ConvertTfListToArrayOfString(ctx, tf.NicLinkNicLinkNicIds, diags),
		NicLinkNicStates:                     utils.ConvertTfListToArrayOfString(ctx, tf.NicLinkNicStates, diags),
		NicLinkNicVmIds:                      utils.ConvertTfListToArrayOfString(ctx, tf.NicLinkNicVmIds, diags),
		NicLinkPublicIpLinkPublicIpIds:       utils.ConvertTfListToArrayOfString(ctx, tf.NicLinkPublicIpLinkPublicIpIds, diags),
		NicLinkPublicIpPublicIpIds:           utils.ConvertTfListToArrayOfString(ctx, tf.NicLinkPublicIpPublicIpIds, diags),
		NicLinkPublicIpPublicIps:             utils.ConvertTfListToArrayOfString(ctx, tf.NicLinkPublicIpPublicIps, diags),
		NicMacAddresses:                      utils.ConvertTfListToArrayOfString(ctx, tf.NicMacAddresses, diags),
		NicNicIds:                            utils.ConvertTfListToArrayOfString(ctx, tf.NicNicIds, diags),
		NicPrivateIpsLinkPublicIpIds:         utils.ConvertTfListToArrayOfString(ctx, tf.NicPrivateIpsLinkPublicIpIds, diags),
		NicPrivateIpsPrimaryIp:               tf.NicPrivateIpsPrimaryIp.ValueBoolPointer(),
		NicPrivateIpsPrivateIps:              utils.ConvertTfListToArrayOfString(ctx, tf.NicPrivateIpsPrivateIps, diags),
		NicSecurityGroupIds:                  utils.ConvertTfListToArrayOfString(ctx, tf.NicSecurityGroupIds, diags),
		NicSecurityGroupNames:                utils.ConvertTfListToArrayOfString(ctx, tf.NicSecurityGroupNames, diags),
		NicStates:                            utils.ConvertTfListToArrayOfString(ctx, tf.NicStates, diags),
		NicSubnetIds:                         utils.ConvertTfListToArrayOfString(ctx, tf.NicSubnetIds, diags),
		Platforms:                            utils.ConvertTfListToArrayOfString(ctx, tf.Platforms, diags),
		PrivateIps:                           utils.ConvertTfListToArrayOfString(ctx, tf.PrivateIps, diags),
		ProductCodes:                         utils.ConvertTfListToArrayOfString(ctx, tf.ProductCodes, diags),
		PublicIps:                            utils.ConvertTfListToArrayOfString(ctx, tf.PublicIps, diags),
		ReservationIds:                       utils.ConvertTfListToArrayOfString(ctx, tf.ReservationIds, diags),
		RootDeviceNames:                      utils.ConvertTfListToArrayOfString(ctx, tf.RootDeviceNames, diags),
		RootDeviceTypes:                      utils.ConvertTfListToArrayOfString(ctx, tf.RootDeviceTypes, diags),
		SecurityGroupIds:                     utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupIds, diags),
		SecurityGroupNames:                   utils.ConvertTfListToArrayOfString(ctx, tf.SecurityGroupNames, diags),
		StateReasonCodes:                     utils.ConvertTfListToArrayOfInt(ctx, tf.StateReasonCodes, diags),
		StateReasonMessages:                  utils.ConvertTfListToArrayOfString(ctx, tf.StateReasonMessages, diags),
		StateReasons:                         utils.ConvertTfListToArrayOfString(ctx, tf.StateReasons, diags),
		SubnetIds:                            utils.ConvertTfListToArrayOfString(ctx, tf.SubnetIds, diags),
		TagKeys:                              utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:                            utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:                                 utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Tenancies:                            utils.ConvertTfListToArrayOfString(ctx, tf.Tenancies, diags),
		VmSecurityGroupIds:                   utils.ConvertTfListToArrayOfString(ctx, tf.VmSecurityGroupIds, diags),
		VmSecurityGroupNames:                 utils.ConvertTfListToArrayOfString(ctx, tf.VmSecurityGroupNames, diags),
		VmStateCodes:                         utils.ConvertTfListToArrayOfInt(ctx, tf.VmStateCodes, diags),
		VmStateNames:                         utils.ConvertTfListToArrayOfString(ctx, tf.VmStateNames, diags),
		Types:                                utils.ConvertTfListToArrayOfStringEnum[api.ReadVmsParamsTypes](ctx, tf.Types, diags),
		VpcIds:                               utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		NicVpcIds:                            utils.ConvertTfListToArrayOfString(ctx, tf.NicVpcIds, diags),
		NicAvailabilityZoneNames:             utils.ConvertTfListToArrayOfAzName(ctx, tf.NicAvailabilityZoneNames, diags),
		AvailabilityZoneNames:                utils.ConvertTfListToArrayOfAzName(ctx, tf.AvailabilityZoneNames, diags),
		Ids:                                  utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
