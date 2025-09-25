package vm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/services/vm/resource_vm"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &vmResource{}
	_ resource.ResourceWithConfigure   = &vmResource{}
	_ resource.ResourceWithImportState = &vmResource{}
)

type vmResource struct {
	provider *client.NumSpotSDK
}

func NewVmResource() resource.Resource {
	return &vmResource{}
}

func (r *vmResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *vmResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *vmResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vm"
}

func (r *vmResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vm.VmResourceSchema(ctx)
}

func (r *vmResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_vm.VmModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	tagsValue := vmTags(ctx, plan.Tags)

	var diags diag.Diagnostics
	numSpotCreateVM := deserializeCreateNumSpotVM(ctx, plan, &diags)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	numSpotVM, err := core.CreateVM(ctx, r.provider, numSpotCreateVM, tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create vm", err.Error())
		return
	}

	state := serializeNumSpotVM(ctx, numSpotVM, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *vmResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	vmID := state.Id.ValueString()

	numSpotVM, err := core.ReadVM(ctx, r.provider, vmID)
	if err != nil {
		response.Diagnostics.AddError("unable to read vm", err.Error())
	}

	newState := serializeNumSpotVM(ctx, numSpotVM, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *vmResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err         error
		state, plan resource_vm.VmModel
		numSpotVM   *api.Vm
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	planTags := vmTags(ctx, plan.Tags)
	stateTags := vmTags(ctx, state.Tags)
	vmID := state.Id.ValueString()

	numSpotUpdateVM := deserializeUpdateNumSpotVM(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.KeypairName.Equal(state.KeypairName) {
		numSpotVM, err = core.UpdateVMKeypair(ctx, r.provider, numSpotUpdateVM, vmID)
		if err != nil {
			response.Diagnostics.AddError("unable to update vm keypair", err.Error())
			return
		}
	}

	if !plan.DeletionProtection.Equal(state.DeletionProtection) ||
		!plan.NestedVirtualization.Equal(state.NestedVirtualization) ||
		!plan.SecurityGroupIds.Equal(state.SecurityGroupIds) ||
		!plan.Type.Equal(state.Type) ||
		!plan.UserData.Equal(state.UserData) ||
		!plan.VmInitiatedShutdownBehavior.Equal(state.VmInitiatedShutdownBehavior) {
		numSpotVM, err = core.UpdateVMAttributes(ctx, r.provider, numSpotUpdateVM, vmID)
		if err != nil {
			response.Diagnostics.AddError("unable to update vm attributes", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numSpotVM, err = core.UpdateVMTags(ctx, r.provider, stateTags, planTags, vmID)
		if err != nil {
			response.Diagnostics.AddError("unable to update vm tags", err.Error())
			return
		}
	}

	newState := serializeNumSpotVM(ctx, numSpotVM, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *vmResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVM(ctx, r.provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete vm", err.Error())
		return
	}
}

func deserializeCreateNumSpotVM(ctx context.Context, tf resource_vm.VmModel, diags *diag.Diagnostics) api.CreateVmsJSONRequestBody {
	var blockDeviceMappingPtr *[]api.BlockDeviceMappingVmCreation
	var placement *api.Placement

	if !(tf.BlockDeviceMappings.IsNull() || tf.BlockDeviceMappings.IsUnknown()) {
		blockDeviceMapping := make([]api.BlockDeviceMappingVmCreation, 0, len(tf.BlockDeviceMappings.Elements()))
		diags.Append(tf.BlockDeviceMappings.ElementsAs(ctx, &blockDeviceMapping, true)...)
		blockDeviceMappingPtr = &blockDeviceMapping
	}

	if !(tf.Placement.IsNull() || tf.Placement.IsUnknown()) {
		placement = &api.Placement{
			AvailabilityZoneName: utils.FromTfStringToAzNamePtr(tf.Placement.AvailabilityZoneName),
			Tenancy:              utils.FromTfStringToStringPtr(tf.Placement.Tenancy),
		}
	}

	bootOnCreation := true
	return api.CreateVmsJSONRequestBody{
		BootOnCreation:              &bootOnCreation,
		ClientToken:                 utils.FromTfStringToStringPtr(tf.ClientToken),
		DeletionProtection:          utils.FromTfBoolToBoolPtr(tf.DeletionProtection),
		ImageId:                     tf.ImageId.ValueString(),
		KeypairName:                 utils.FromTfStringToStringPtr(tf.KeypairName),
		NestedVirtualization:        utils.FromTfBoolToBoolPtr(tf.NestedVirtualization),
		Placement:                   placement,
		PrivateIps:                  utils.TfStringListToStringPtrList(ctx, tf.PrivateIps, diags),
		SecurityGroupIds:            utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		SecurityGroups:              utils.TfStringListToStringPtrList(ctx, tf.SecurityGroups, diags),
		SubnetId:                    tf.SubnetId.ValueString(),
		UserData:                    utils.FromTfStringToStringPtr(tf.UserData),
		VmInitiatedShutdownBehavior: utils.FromTfStringToStringPtr(tf.InitiatedShutdownBehavior),
		Type:                        tf.Type.ValueString(),
		BlockDeviceMappings:         blockDeviceMappingPtr,
	}
}

func nicsFromApi(ctx context.Context, nic api.NicLight, diags *diag.Diagnostics) resource_vm.NicsValue {
	var (
		linkNics         resource_vm.LinkNicValue
		linkPublicIp     resource_vm.NicLinkPublicIpValue
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

	var privateIps []api.PrivateIpLightForVm
	if nic.PrivateIps != nil {
		privateIps = *nic.PrivateIps
	}
	privateIpsTf = utils.GenericListToTfListValue(
		ctx,
		privateIpsFromApi,
		privateIps,
		diags,
	)

	var securityGroups []api.SecurityGroupLight
	if nic.SecurityGroups != nil {
		securityGroups = *nic.SecurityGroups
	}
	securityGroupsTf = utils.GenericListToTfListValue(
		ctx,
		securityGroupsFromApi,
		securityGroups,
		diags,
	)

	value, diagnostics := resource_vm.NewNicsValue(
		resource_vm.NicsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"description":            types.StringPointerValue(nic.Description),
			"is_source_dest_checked": types.BoolPointerValue(nic.IsSourceDestChecked),
			"link_nic":               linkNicsObjectValue,
			"nic_link_public_ip":     linkPublicIpObjectValue,
			"mac_address":            types.StringPointerValue(nic.MacAddress),
			"vpc_id":                 types.StringPointerValue(nic.VpcId),
			"nic_id":                 types.StringPointerValue(nic.NicId),
			"private_dns_name":       types.StringPointerValue(nic.PrivateDnsName),
			"private_ips":            privateIpsTf,
			"nic_security_groups":    securityGroupsTf,
			"state":                  types.StringPointerValue(nic.State),
			"subnet_id":              types.StringPointerValue(nic.SubnetId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func placementFromHTTP(ctx context.Context, elt *api.Placement, diags *diag.Diagnostics) resource_vm.PlacementValue {
	if elt == nil {
		return resource_vm.PlacementValue{}
	}
	value, diagnostics := resource_vm.NewPlacementValue(
		resource_vm.PlacementValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(elt.AvailabilityZoneName)),
			"tenancy":                types.StringPointerValue(elt.Tenancy),
		})
	diags.Append(diagnostics...)
	return value
}

func serializeNumSpotVM(ctx context.Context, http *api.Vm, diags *diag.Diagnostics) *resource_vm.VmModel {
	var (
		tagsTf types.Set
		nics   = types.ListNull(resource_vm.NicsValue{}.Type(ctx))
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
	var blockDeviceMappings []api.BlockDeviceMappingCreated
	blockDeviceMappingTf := types.ListNull(resource_vm.BlockDeviceMappingsValue{}.Type(ctx))
	if http.BlockDeviceMappings != nil {
		blockDeviceMappings = *http.BlockDeviceMappings
		blockDeviceMappingTf = utils.GenericListToTfListValue(
			ctx,
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
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	if http.Nics != nil {
		nics = utils.GenericListToTfListValue(ctx, nicsFromApi, *http.Nics, diags)
	}

	var launchNumber basetypes.Int64Value
	if http.LaunchNumber != nil {
		launchNumber = utils.FromIntPtrToTfInt64(http.LaunchNumber)
	}

	r := resource_vm.VmModel{
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

func securityGroupsFromApi(ctx context.Context, privateIp api.SecurityGroupLight, diags *diag.Diagnostics) resource_vm.NicSecurityGroupsValue {
	value, diagnostics := resource_vm.NewNicSecurityGroupsValue(
		resource_vm.NicSecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(privateIp.SecurityGroupId),
			"security_group_name": types.StringPointerValue(privateIp.SecurityGroupName),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func linkNicsFromApi(ctx context.Context, linkNic api.LinkNicLight, diags *diag.Diagnostics) resource_vm.LinkNicValue {
	if linkNic.DeviceNumber == nil {
		return resource_vm.LinkNicValue{}
	}
	deviceNumber := int64(*linkNic.DeviceNumber)
	value, diagnostics := resource_vm.NewLinkNicValue(
		resource_vm.LinkNicValue{}.AttributeTypes(ctx),
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

func privateIpsFromApi(ctx context.Context, privateIp api.PrivateIpLightForVm, diags *diag.Diagnostics) resource_vm.PrivateIpsValue {
	linkPublicIp := linkPublicIpPrivateVmFromApi(ctx, utils.GetPtrValue(privateIp.LinkPublicIp), diags)
	if diags.HasError() {
		return resource_vm.PrivateIpsValue{}
	}

	linkPublicIpObjectValue, diagnostics := linkPublicIp.ToObjectValue(ctx)
	if diagnostics.HasError() {
		return resource_vm.PrivateIpsValue{}
	}

	value, diagnostics := resource_vm.NewPrivateIpsValue(
		resource_vm.PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":                types.BoolPointerValue(privateIp.IsPrimary),
			"private_ip_link_public_ip": linkPublicIpObjectValue,
			"private_dns_name":          types.StringPointerValue(privateIp.PrivateDnsName),
			"private_ip":                types.StringPointerValue(privateIp.PrivateIp),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func linkPublicIpVmFromApi(ctx context.Context, linkPublicIp api.LinkPublicIpLightForVm, diags *diag.Diagnostics) resource_vm.NicLinkPublicIpValue {
	value, diagnostics := resource_vm.NewNicLinkPublicIpValue(
		resource_vm.NicLinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name": types.StringPointerValue(linkPublicIp.PublicDnsName),
			"public_ip":       types.StringPointerValue(linkPublicIp.PublicIp),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func linkPublicIpPrivateVmFromApi(ctx context.Context, linkPublicIp api.LinkPublicIpLightForVm, diags *diag.Diagnostics) resource_vm.PrivateIpLinkPublicIpValue {
	value, diagnostics := resource_vm.NewPrivateIpLinkPublicIpValue(
		resource_vm.NicLinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_dns_name": types.StringPointerValue(linkPublicIp.PublicDnsName),
			"public_ip":       types.StringPointerValue(linkPublicIp.PublicIp),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func vmBlockDeviceMappingFromApi(ctx context.Context, elt api.BlockDeviceMappingCreated, diags *diag.Diagnostics) resource_vm.BlockDeviceMappingsValue {
	if elt.Bsu == nil {
		return resource_vm.BlockDeviceMappingsValue{}
	}
	// Bsu
	bsuTf := vmBsuFromApi(ctx, *elt.Bsu, diags)
	if diags.HasError() {
		return resource_vm.BlockDeviceMappingsValue{}
	}

	value, diagnostics := resource_vm.NewBlockDeviceMappingsValue(
		resource_vm.BlockDeviceMappingsValue{}.AttributeTypes(ctx),
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

func vmBsuFromApi(ctx context.Context, elt api.BsuCreated, diags *diag.Diagnostics) basetypes.ObjectValue {
	obj, diagnostics := resource_vm.NewBsuValue(
		resource_vm.BsuValue{}.AttributeTypes(ctx),
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

func deserializeUpdateNumSpotVM(ctx context.Context, tf resource_vm.VmModel, diags *diag.Diagnostics) api.UpdateVmJSONRequestBody {
	blockDeviceMapping := make([]api.BlockDeviceMappingVmUpdate, 0, len(tf.BlockDeviceMappings.Elements()))

	for _, bdmTf := range tf.BlockDeviceMappings.Elements() {
		bdmTfRes, ok := bdmTf.(resource_vm.BlockDeviceMappingsValue)
		if !ok {
			diags.AddError("Failed to cast block device mapping resource", "")
			return api.UpdateVmJSONRequestBody{}
		}

		bdmApi := blockDeviceMappingFromTf(bdmTfRes)
		blockDeviceMapping = append(blockDeviceMapping, bdmApi)
	}

	return api.UpdateVmJSONRequestBody{
		DeletionProtection:          utils.FromTfBoolToBoolPtr(tf.DeletionProtection),
		KeypairName:                 utils.FromTfStringToStringPtr(tf.KeypairName),
		NestedVirtualization:        utils.FromTfBoolToBoolPtr(tf.NestedVirtualization),
		SecurityGroupIds:            utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		UserData:                    utils.FromTfStringToStringPtr(tf.UserData),
		VmInitiatedShutdownBehavior: utils.FromTfStringToStringPtr(tf.InitiatedShutdownBehavior),
		Type:                        utils.FromTfStringToStringPtr(tf.Type),
		BlockDeviceMappings:         &blockDeviceMapping,
		IsSourceDestChecked:         utils.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
	}
}

func blockDeviceMappingFromTf(bdm resource_vm.BlockDeviceMappingsValue) api.BlockDeviceMappingVmUpdate {
	attrtypes := bdm.Bsu.AttributeTypes(context.Background())
	attrVals := bdm.Bsu.Attributes()
	bsuTF, diags := resource_vm.NewBsuValue(attrtypes, attrVals)
	if diags.HasError() {
		return api.BlockDeviceMappingVmUpdate{}
	}
	bsu := bsuFromTf(bsuTF)
	return api.BlockDeviceMappingVmUpdate{
		Bsu:               bsu,
		DeviceName:        bdm.DeviceName.ValueStringPointer(),
		NoDevice:          bdm.NoDevice.ValueStringPointer(),
		VirtualDeviceName: bdm.VirtualDeviceName.ValueStringPointer(),
	}
}

func bsuFromTf(bsu resource_vm.BsuValue) *api.BsuToUpdateVm {
	if bsu.IsNull() || bsu.IsUnknown() {
		return nil
	}

	return &api.BsuToUpdateVm{
		DeleteOnVmDeletion: bsu.DeleteOnVmDeletion.ValueBoolPointer(),
		VolumeId:           bsu.VolumeId.ValueStringPointer(),
	}
}

func vmTags(ctx context.Context, tags types.Set) []api.ResourceTag {
	tfTags := make([]resource_vm.TagsValue, 0, len(tags.Elements()))
	tags.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}
