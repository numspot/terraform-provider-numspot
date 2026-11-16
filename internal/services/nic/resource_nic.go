package nic

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/nic/resource_nic"
	"github.com/numspot/terraform-provider-numspot/internal/services/tags"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &nicResource{}
	_ resource.ResourceWithConfigure   = &nicResource{}
	_ resource.ResourceWithImportState = &nicResource{}
)

// nicResource is the Terraform resource implementation for numspot_nic.
type nicResource struct {
	services.BaseResource
}

// NewNicResource returns a new nicResource instance.
func NewNicResource() resource.Resource {
	return &nicResource{}
}

// Metadata sets the resource type name.
func (r *nicResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_nic"
}

// Schema defines the schema for the resource.
func (r *nicResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_nic.NicResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *nicResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_nic.NicModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := nicTags(ctx, plan.Tags)
	numspotNic, err := core.CreateNic(
		ctx,
		r.Provider,
		deserializeCreateNicRequest(ctx, &plan, &response.Diagnostics),
		linkNicBody(&plan.LinkNic),
		tagsValue,
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create nic", err.Error())
		return
	}

	state := serializeNic(ctx, numspotNic, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *nicResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_nic.NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotNic, err := core.ReadNic(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		if utils.IsResourceGone(err) {
			tflog.Warn(ctx, "unable to read nic (not found or access forbidden); removing from state")
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("unable to read nic", err.Error())
		return
	}

	newState := serializeNic(ctx, numspotNic, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *nicResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err         error
		state, plan resource_nic.NicModel
		numspotNic  *api.Nic
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	nicID := state.Id.ValueString()
	planTags := nicTags(ctx, plan.Tags)
	stateTags := nicTags(ctx, state.Tags)

	if !plan.LinkNic.Equal(state.LinkNic) {
		numspotNic, err = core.UpdateNicLink(ctx, r.Provider, state.Id.ValueString(), unlinkNicBody(&state.LinkNic), linkNicBody(&plan.LinkNic))
		if err != nil {
			response.Diagnostics.AddError("unable to update nic", err.Error())
			return
		}
	}

	if !plan.Description.Equal(state.Description) {
		numspotNic, err = core.UpdateNicAttributes(
			ctx,
			r.Provider,
			deserializeUpdateNicBody(ctx, &plan, &response.Diagnostics),
			state.Id.ValueString(),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update nic", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numspotNic, err = core.UpdateNicTags(ctx, r.Provider, nicID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update nic tags", err.Error())
			return
		}
	}

	newState := serializeNic(ctx, numspotNic, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *nicResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_nic.NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteNic(ctx, r.Provider, state.Id.ValueString(), unlinkNicBody(&state.LinkNic)); err != nil {
		response.Diagnostics.AddError("unable to delete nic", err.Error())
		return
	}
}

func serializeNic(ctx context.Context, http *api.Nic, diags *diag.Diagnostics) resource_nic.NicModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil Nic response", "the API returned an empty Nic")
		return resource_nic.NicModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_nic.NicModel{
		AvailabilityZoneName: types.StringValue(utils.ConvertAzNamePtrToString(http.AvailabilityZoneName)),
		Description:          types.StringPointerValue(http.Description),
		Id:                   types.StringPointerValue(http.Id),
		IsSourceDestChecked:  types.BoolPointerValue(http.IsSourceDestChecked),
		LinkNic:              serializeLinkNic(ctx, http.LinkNic, diags),
		LinkPublicIp:         serializeLinkPublicIp(ctx, http.LinkPublicIp, diags),
		MacAddress:           lowerMacAddress(http.MacAddress),
		PrivateDnsName:       types.StringPointerValue(http.PrivateDnsName),
		PrivateIps:           utils.GenericSetToTfSetValue(ctx, serializePrivateIp, *http.PrivateIps, diags),
		SecurityGroupIds:     securityGroupIDsFromGroups(ctx, http.SecurityGroups, diags),
		SecurityGroups:       utils.GenericListToTfListValue(ctx, serializeSecurityGroup, *http.SecurityGroups, diags),
		State:                types.StringPointerValue(http.State),
		SubnetId:             types.StringPointerValue(http.SubnetId),
		VpcId:                types.StringPointerValue(http.VpcId),
		Tags:                 tagsTf,
	}
}

func deserializeCreateNicRequest(ctx context.Context, tf *resource_nic.NicModel, diags *diag.Diagnostics) *api.CreateNicJSONRequestBody {
	return &api.CreateNicJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		PrivateIps:       createPrivateIpsBody(ctx, tf.PrivateIps, diags),
		SecurityGroupIds: utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		SubnetId:         tf.SubnetId.ValueString(),
	}
}

func nicTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_nic.TagsValue, 0, len(tagsSet.Elements()))
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

func serializeLinkNic(ctx context.Context, item *api.LinkNic, diags *diag.Diagnostics) resource_nic.LinkNicValue {
	if item == nil {
		return resource_nic.NewLinkNicValueNull()
	}
	value, d := resource_nic.NewLinkNicValue(resource_nic.LinkNicValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"delete_on_vm_deletion": types.BoolPointerValue(item.DeleteOnVmDeletion),
		"device_number":         utils.FromIntPtrToTfInt64(item.DeviceNumber),
		"id":                    types.StringPointerValue(item.Id),
		"state":                 types.StringPointerValue(item.State),
		"vm_id":                 types.StringPointerValue(item.VmId),
	})
	diags.Append(d...)
	return value
}

func serializeLinkPublicIp(ctx context.Context, item *api.LinkPublicIp, diags *diag.Diagnostics) resource_nic.LinkPublicIpValue {
	if item == nil {
		return resource_nic.NewLinkPublicIpValueNull()
	}
	value, d := resource_nic.NewLinkPublicIpValue(resource_nic.LinkPublicIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":              types.StringPointerValue(item.Id),
		"public_dns_name": types.StringPointerValue(item.PublicDnsName),
		"public_ip":       types.StringPointerValue(item.PublicIp),
		"public_ip_id":    types.StringPointerValue(item.PublicIpId),
	})
	diags.Append(d...)
	return value
}

func serializeLinkPublicIpPrivateIp(ctx context.Context, item *api.LinkPublicIp, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(resource_nic.LinkPublicIpPrivateIpValue{}.AttributeTypes(ctx))
	}
	value, d := resource_nic.NewLinkPublicIpPrivateIpValue(resource_nic.LinkPublicIpPrivateIpValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":              types.StringPointerValue(item.Id),
		"public_dns_name": types.StringPointerValue(item.PublicDnsName),
		"public_ip":       types.StringPointerValue(item.PublicIp),
		"public_ip_id":    types.StringPointerValue(item.PublicIpId),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func serializePrivateIp(ctx context.Context, item *api.PrivateIp, diags *diag.Diagnostics) resource_nic.PrivateIpsValue {
	value, d := resource_nic.NewPrivateIpsValue(resource_nic.PrivateIpsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"is_primary":                types.BoolPointerValue(item.IsPrimary),
		"link_public_ip_private_ip": serializeLinkPublicIpPrivateIp(ctx, item.LinkPublicIp, diags),
		"private_dns_name":          types.StringPointerValue(item.PrivateDnsName),
		"private_ip":                types.StringPointerValue(item.PrivateIp),
	})
	diags.Append(d...)
	return value
}

func serializeSecurityGroup(ctx context.Context, item *api.SecurityGroupLight, diags *diag.Diagnostics) resource_nic.SecurityGroupsValue {
	value, d := resource_nic.NewSecurityGroupsValue(resource_nic.SecurityGroupsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"security_group_id":   types.StringPointerValue(item.SecurityGroupId),
		"security_group_name": types.StringPointerValue(item.SecurityGroupName),
	})
	diags.Append(d...)
	return value
}
