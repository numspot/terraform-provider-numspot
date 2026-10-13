package vm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

// vmResource is the Terraform resource implementation for numspot_vm.
type vmResource struct {
	services.BaseResource
}

// NewVMResource returns a new vmResource instance.
func NewVMResource() resource.Resource {
	return &vmResource{}
}

// Metadata sets the resource type name.
func (r *vmResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vm"
}

// Schema defines the schema for the resource.
func (r *vmResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vm.VmResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vmResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_vm.VmModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := vmTags(ctx, plan.Tags)
	numSpotVM, err := core.CreateVM(ctx, r.Provider, deserializeCreateVMRequest(ctx, &plan, &response.Diagnostics), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create vm", err.Error())
		return
	}

	state := serializeVM(ctx, numSpotVM, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	enrichBlockDeviceMappings(ctx, r.Provider, &state, true, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vmResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotVM, err := core.ReadVM(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read vm", err.Error())
		return
	}

	newState := serializeVM(ctx, numSpotVM, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	enrichBlockDeviceMappings(ctx, r.Provider, &newState, false, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
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

	vmID := state.Id.ValueString()
	planTags := vmTags(ctx, plan.Tags)
	stateTags := vmTags(ctx, state.Tags)

	if plannedChange(plan.KeypairName, state.KeypairName) {
		numSpotVM, err = core.UpdateVMKeypair(ctx, r.Provider, deserializeUpdateVm(ctx, &plan, &response.Diagnostics), state.Id.ValueString())
		if err != nil {
			response.Diagnostics.AddError("unable to update vm", err.Error())
			return
		}
	}

	attributesNeedStop := plannedChange(plan.DeletionProtection, state.DeletionProtection) ||
		plannedChange(plan.NestedVirtualization, state.NestedVirtualization) ||
		plannedChange(plan.Type, state.Type) ||
		plannedChange(plan.UserData, state.UserData) ||
		plannedChange(plan.VmInitiatedShutdownBehavior, state.VmInitiatedShutdownBehavior)
	securityGroupsChanged := plannedChange(plan.SecurityGroupIds, state.SecurityGroupIds)

	if securityGroupsChanged || attributesNeedStop {
		if attributesNeedStop {
			// Type, user data or nested virtualization changes require the VM
			// to be stopped: the update body carries all attributes.
			numSpotVM, err = core.UpdateVMAttributes(ctx, r.Provider, deserializeUpdateVm(ctx, &plan, &response.Diagnostics), state.Id.ValueString())
		} else {
			// Security-group-only change: the API applies it to a running VM,
			// so the VM is not stopped/restarted (CLOS-3261).
			numSpotVM, err = core.UpdateVMSecurityGroups(
				ctx,
				r.Provider,
				deserializeUpdateVm(ctx, &plan, &response.Diagnostics),
				state.Id.ValueString(),
			)
		}
		if err != nil {
			response.Diagnostics.AddError("unable to update vm", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numSpotVM, err = core.UpdateVMTags(ctx, r.Provider, vmID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update vm tags", err.Error())
			return
		}
	}

	newState := serializeVM(ctx, numSpotVM, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	// The VM read payload does not carry the block device mapping creation
	// fields (volume_size, volume_type, iops, snapshot_id): without the
	// enrichment the update would overwrite them with null and Terraform Core
	// would abort the apply with an inconsistent result (SUPN2-718).
	enrichBlockDeviceMappings(ctx, r.Provider, &newState, true, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vmResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVM(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete vm", err.Error())
		return
	}
}

func serializeVM(ctx context.Context, http *api.Vm, diags *diag.Diagnostics) resource_vm.VmModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil Vm response", "the API returned an empty Vm")
		return resource_vm.VmModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_vm.VmModel{
		Architecture:                types.StringPointerValue(http.Architecture),
		BlockDeviceMappings:         utils.GenericListToTfListValue(ctx, serializeBlockDeviceMapping, *http.BlockDeviceMappings, diags),
		ClientToken:                 types.StringPointerValue(http.ClientToken),
		CreationDate:                utils.TimePtrToTfString(http.CreationDate),
		DeletionProtection:          types.BoolPointerValue(http.DeletionProtection),
		Hypervisor:                  types.StringPointerValue(http.Hypervisor),
		Id:                          types.StringPointerValue(http.Id),
		ImageId:                     types.StringPointerValue(http.ImageId),
		InitiatedShutdownBehavior:   types.StringPointerValue(http.InitiatedShutdownBehavior),
		IsSourceDestChecked:         types.BoolPointerValue(http.IsSourceDestChecked),
		KeypairName:                 types.StringPointerValue(http.KeypairName),
		LaunchNumber:                utils.FromIntPtrToTfInt64(http.LaunchNumber),
		NestedVirtualization:        types.BoolPointerValue(http.NestedVirtualization),
		Nics:                        utils.GenericListToTfListValue(ctx, serializeNic, *http.Nics, diags),
		OsFamily:                    types.StringPointerValue(http.OsFamily),
		Placement:                   serializePlacement(ctx, http.Placement, diags),
		PrivateDnsName:              types.StringPointerValue(http.PrivateDnsName),
		PrivateIp:                   types.StringPointerValue(http.PrivateIp),
		PrivateIps:                  privateIpsFromVm(ctx, http.PrivateIp, diags),
		ProductCodes:                utils.FromStringListPointerToTfStringList(ctx, http.ProductCodes, diags),
		PublicDnsName:               types.StringPointerValue(http.PublicDnsName),
		PublicIp:                    types.StringPointerValue(http.PublicIp),
		ReservationId:               types.StringPointerValue(http.ReservationId),
		RootDeviceName:              types.StringPointerValue(http.RootDeviceName),
		RootDeviceType:              types.StringPointerValue(http.RootDeviceType),
		SecurityGroupIds:            securityGroupIDsFromVm(ctx, http.SecurityGroups, diags),
		SecurityGroups:              securityGroupNamesFromVm(ctx, http.SecurityGroups, diags),
		State:                       types.StringPointerValue(http.State),
		StateReason:                 types.StringPointerValue(http.StateReason),
		SubnetId:                    types.StringPointerValue(http.SubnetId),
		Type:                        types.StringPointerValue(http.Type),
		UserData:                    types.StringPointerValue(http.UserData),
		VmInitiatedShutdownBehavior: types.StringPointerValue(http.InitiatedShutdownBehavior),
		VpcId:                       types.StringPointerValue(http.VpcId),
		Tags:                        tagsTf,
	}
}

func deserializeCreateVMRequest(ctx context.Context, tf *resource_vm.VmModel, diags *diag.Diagnostics) *api.CreateVmsJSONRequestBody {
	return &api.CreateVmsJSONRequestBody{
		BlockDeviceMappings:         deserializeCreateBlockDeviceMappings(ctx, tf.BlockDeviceMappings, diags),
		DeletionProtection:          tf.DeletionProtection.ValueBoolPointer(),
		ImageId:                     tf.ImageId.ValueString(),
		VmInitiatedShutdownBehavior: utils.FromTfStringToStringPtr(tf.InitiatedShutdownBehavior),
		KeypairName:                 tf.KeypairName.ValueStringPointer(),
		NestedVirtualization:        tf.NestedVirtualization.ValueBoolPointer(),
		Placement:                   deserializePlacement(tf.Placement),
		PrivateIps:                  utils.TfStringListToStringPtrList(ctx, tf.PrivateIps, diags),
		SecurityGroupIds:            stringSetToPtrList(ctx, tf.SecurityGroupIds, diags),
		SecurityGroups:              stringSetToPtrList(ctx, tf.SecurityGroups, diags),
		SubnetId:                    tf.SubnetId.ValueString(),
		Type:                        api.CreateVmsType(tf.Type.ValueString()),
		UserData:                    tf.UserData.ValueStringPointer(),
	}
}

func vmTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_vm.TagsValue, 0, len(tagsSet.Elements()))
	tagsSet.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}

func serializeBsu(ctx context.Context, item *api.BsuCreated, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(resource_vm.BsuValue{}.AttributeTypes(ctx))
	}
	value, d := resource_vm.NewBsuValue(resource_vm.BsuValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(item.DeleteOnVmDeletion),
		"iops":                  types.Int64Null(),
		"link_date":             utils.TimePtrToTfString(item.LinkDate),
		"snapshot_id":           types.StringNull(),
		"state":                 types.StringPointerValue(item.State),
		"volume_id":             types.StringPointerValue(item.VolumeId),
		"volume_size":           types.Int64Null(),
		"volume_type":           types.StringNull(),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func serializeBlockDeviceMapping(
	ctx context.Context,
	item *api.BlockDeviceMappingCreated,
	diags *diag.Diagnostics,
) resource_vm.BlockDeviceMappingsValue {
	value, d := resource_vm.NewBlockDeviceMappingsValue(resource_vm.BlockDeviceMappingsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"bsu":                 serializeBsu(ctx, item.Bsu, diags),
		"device_name":         types.StringPointerValue(item.DeviceName),
		"no_device":           types.StringNull(),
		"virtual_device_name": types.StringNull(),
	})
	diags.Append(d...)
	return value
}

func serializeLinkNic(ctx context.Context, item *api.LinkNicLight, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(resource_vm.LinkNicValue{}.AttributeTypes(ctx))
	}
	value, d := resource_vm.NewLinkNicValue(resource_vm.LinkNicValue{}.AttributeTypes(ctx), map[string]attr.Value{
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

func serializeNicLinkPublicIp(ctx context.Context, item *api.LinkPublicIpLightForVm, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(resource_vm.NicLinkPublicIpValue{}.AttributeTypes(ctx))
	}
	value, d := resource_vm.NewNicLinkPublicIpValue(resource_vm.NicLinkPublicIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"public_dns_name": types.StringPointerValue(item.PublicDnsName),
		"public_ip":       types.StringPointerValue(item.PublicIp),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func serializeNicSecurityGroup(ctx context.Context, item *api.SecurityGroupLight, diags *diag.Diagnostics) resource_vm.NicSecurityGroupsValue {
	value, d := resource_vm.NewNicSecurityGroupsValue(resource_vm.NicSecurityGroupsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"security_group_id":   types.StringPointerValue(item.SecurityGroupId),
		"security_group_name": types.StringPointerValue(item.SecurityGroupName),
	})
	diags.Append(d...)
	return value
}

func serializeNic(ctx context.Context, item *api.NicLight, diags *diag.Diagnostics) resource_vm.NicsValue {
	value, d := resource_vm.NewNicsValue(resource_vm.NicsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"description":            types.StringPointerValue(item.Description),
		"is_source_dest_checked": types.BoolPointerValue(item.IsSourceDestChecked),
		"link_nic":               serializeLinkNic(ctx, item.LinkNic, diags),
		"mac_address":            types.StringPointerValue(item.MacAddress),
		"nic_id":                 types.StringPointerValue(item.NicId),
		"nic_link_public_ip":     serializeNicLinkPublicIp(ctx, item.LinkPublicIp, diags),
		"nic_security_groups":    utils.GenericListToTfListValue(ctx, serializeNicSecurityGroup, *item.SecurityGroups, diags),
		"private_dns_name":       types.StringPointerValue(item.PrivateDnsName),
		"private_ips":            privateIpsFromNicLight(ctx, item.PrivateIps, diags),
		"state":                  types.StringPointerValue(item.State),
		"subnet_id":              types.StringPointerValue(item.SubnetId),
		"vpc_id":                 types.StringPointerValue(item.VpcId),
	})
	diags.Append(d...)
	return value
}

func serializePlacement(ctx context.Context, item *api.Placement, diags *diag.Diagnostics) resource_vm.PlacementValue {
	if item == nil {
		return resource_vm.NewPlacementValueNull()
	}
	value, d := resource_vm.NewPlacementValue(resource_vm.PlacementValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(item.AvailabilityZoneName)),
		"tenancy":                types.StringPointerValue(item.Tenancy),
	})
	diags.Append(d...)
	return value
}

func deserializePlacement(e resource_vm.PlacementValue) *api.Placement {
	if e.IsNull() {
		return nil
	}
	return &api.Placement{
		AvailabilityZoneName: utils.FromTfStringToAzNamePtr(e.AvailabilityZoneName),
		Tenancy:              e.Tenancy.ValueStringPointer(),
	}
}

// plannedChange reports whether a planned attribute value is an explicit,
// known change relative to the prior state. Optional+Computed attributes that
// are not configured resolve to unknown at plan time; those framework
// placeholders must not be treated as user changes, otherwise every update
// would call the keypair and attribute update paths (forcing a stop/start,
// CLOS-3261).
func plannedChange(plan, state attr.Value) bool {
	return !plan.IsUnknown() && !plan.Equal(state)
}
