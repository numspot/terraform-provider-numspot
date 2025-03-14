package vm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/core"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewVmResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.provider = provider
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vm"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VmResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan VmModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)

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

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state VmModel
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

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err         error
		state, plan VmModel
		numSpotVM   *numspot.Vm
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)
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

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVM(ctx, r.provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete vm", err.Error())
		return
	}
}

func deserializeCreateNumSpotVM(ctx context.Context, tf VmModel, diags *diag.Diagnostics) numspot.CreateVmsJSONRequestBody {
	var blockDeviceMappingPtr *[]numspot.BlockDeviceMappingVmCreation
	var placement *numspot.Placement

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

func serializeNumSpotVM(ctx context.Context, http *numspot.Vm, diags *diag.Diagnostics) *VmModel {
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
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	if http.Nics != nil {
		nics = utils.GenericListToTfListValue(ctx, nicsFromApi, *http.Nics, diags)
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

func deserializeUpdateNumSpotVM(ctx context.Context, tf VmModel, diags *diag.Diagnostics) numspot.UpdateVmJSONRequestBody {
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
		VmInitiatedShutdownBehavior: utils.FromTfStringToStringPtr(tf.InitiatedShutdownBehavior),
		Type:                        utils.FromTfStringToStringPtr(tf.Type),
		BlockDeviceMappings:         &blockDeviceMapping,
		IsSourceDestChecked:         utils.FromTfBoolToBoolPtr(tf.IsSourceDestChecked),
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

func bsuFromTf(bsu BsuValue) *numspot.BsuToUpdateVm {
	if bsu.IsNull() || bsu.IsUnknown() {
		return nil
	}

	return &numspot.BsuToUpdateVm{
		DeleteOnVmDeletion: bsu.DeleteOnVmDeletion.ValueBoolPointer(),
		VolumeId:           bsu.VolumeId.ValueStringPointer(),
	}
}
