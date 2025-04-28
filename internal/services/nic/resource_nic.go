package nic

import (
	"context"
	"strings"

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
	"terraform-provider-numspot/internal/services/nic/resource_nic"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &nicResource{}
	_ resource.ResourceWithConfigure   = &nicResource{}
	_ resource.ResourceWithImportState = &nicResource{}
)

type nicResource struct {
	provider *client.NumSpotSDK
}

func NewNicResource() resource.Resource {
	return &nicResource{}
}

func (r *nicResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *nicResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *nicResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_nic"
}

func (r *nicResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_nic.NicResourceSchema(ctx)
}

func (r *nicResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_nic.NicModel
	var linkNicBody *api.LinkNicJSONRequestBody
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := nicTags(ctx, plan.Tags)
	body := deserializeCreateNumSpotNic(ctx, plan, &response.Diagnostics)

	if !utils.IsTfValueNull(plan.LinkNic) {
		linkNicBody = deserializeLinkNic(plan.LinkNic)
	}

	nic, err := core.CreateNic(ctx, r.provider, body, tagsValue, linkNicBody)
	if err != nil {
		response.Diagnostics.AddError("unable to create nic", err.Error())
		return
	}

	state := serializeNumSpotNic(ctx, nic, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *nicResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_nic.NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	nicID := state.Id.ValueString()
	numSpotNic, err := core.ReadNicWithID(ctx, r.provider, nicID)
	if err != nil {
		response.Diagnostics.AddError("unable to read Nic", err.Error())
		return
	}

	newState := serializeNumSpotNic(ctx, numSpotNic, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *nicResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan resource_nic.NicModel
		numspotNic  *api.Nic
		err         error
	)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	nicId := state.Id.ValueString()
	planTags := nicTags(ctx, plan.Tags)
	stateTags := nicTags(ctx, state.Tags)

	if !state.Tags.Equal(plan.Tags) {
		numspotNic, err = core.UpdateNicTags(ctx, r.provider, nicId, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update nic tags", err.Error())
			return
		}
	}

	if !utils.IsTfValueNull(plan.LinkNic) || !utils.IsTfValueNull(state.LinkNic) {
		numspotNic, err = core.UpdateNicLink(ctx, r.provider, nicId, deserializeUnlinkNic(state.LinkNic), deserializeLinkNic(plan.LinkNic))
		if err != nil {
			response.Diagnostics.AddError("unable to update nic link", err.Error())
			return
		}
	}

	if !utils.IsTfValueNull(plan.Description) && !plan.Description.Equal(state.Description) {
		body := deserializeUpdateNumSpotNic(ctx, plan, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
		numspotNic, err = core.UpdateNicAttributes(ctx, r.provider, body, nicId)
		if err != nil {
			response.Diagnostics.AddError("unable to update nic attributes", err.Error())
			return
		}
	}

	newState := serializeNumSpotNic(ctx, numspotNic, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *nicResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_nic.NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteNic(ctx, r.provider, state.Id.ValueString(), deserializeUnlinkNic(state.LinkNic)); err != nil {
		response.Diagnostics.AddError("unable to delete Nic", err.Error())
		return
	}
}

func deserializeCreateNumSpotNic(ctx context.Context, tf resource_nic.NicModel, diags *diag.Diagnostics) api.CreateNicJSONRequestBody {
	var privateIpsPtr *[]api.PrivateIpLight
	var securityGroupIdsPtr *[]string

	if !(tf.PrivateIps.IsNull() || tf.PrivateIps.IsUnknown()) {
		privateIps := utils.TfSetToGenericSet(func(a resource_nic.PrivateIpsValue) api.PrivateIpLight {
			return api.PrivateIpLight{
				IsPrimary: a.IsPrimary.ValueBoolPointer(),
				PrivateIp: a.PrivateIp.ValueStringPointer(),
			}
		}, ctx, tf.PrivateIps, diags)
		privateIpsPtr = &privateIps
	}

	if !(tf.SecurityGroupIds.IsNull() || tf.SecurityGroupIds.IsUnknown()) {
		securityGroupIds := utils.TfStringListToStringList(ctx, tf.SecurityGroupIds, diags)
		securityGroupIdsPtr = &securityGroupIds
	}
	return api.CreateNicJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		PrivateIps:       privateIpsPtr,
		SecurityGroupIds: securityGroupIdsPtr,
		SubnetId:         tf.SubnetId.ValueString(),
	}
}

func deserializeUpdateNumSpotNic(ctx context.Context, tf resource_nic.NicModel, diags *diag.Diagnostics) api.UpdateNicJSONRequestBody {
	linkNic := api.LinkNicToUpdate{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.LinkNic.DeleteOnVmDeletion),
		LinkNicId:          utils.FromTfStringToStringPtr(tf.LinkNic.Id),
	}

	return api.UpdateNicJSONRequestBody{
		SecurityGroupIds: utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		Description:      utils.FromTfStringToStringPtr(tf.Description),
		LinkNic:          &linkNic,
	}
}

func deserializeLinkNic(tf resource_nic.LinkNicValue) *api.LinkNicJSONRequestBody {
	if utils.IsTfValueNull(tf) {
		return nil
	}

	return &api.LinkNicJSONRequestBody{
		DeviceNumber: utils.FromTfInt64ToInt(tf.DeviceNumber),
		VmId:         tf.VmId.ValueString(),
	}
}

func deserializeUnlinkNic(tf resource_nic.LinkNicValue) *api.UnlinkNicJSONRequestBody {
	if utils.IsTfValueNull(tf) {
		return nil
	}

	return &api.UnlinkNicJSONRequestBody{
		LinkNicId: tf.Id.ValueString(),
	}
}

func serializeNumSpotNic(ctx context.Context, http *api.Nic, diags *diag.Diagnostics) *resource_nic.NicModel {
	var (
		linkPublicIpTf resource_nic.LinkPublicIpValue
		linkNicTf      resource_nic.LinkNicValue
		tagsTf         types.List
	)

	privateIps := utils.GenericSetToTfSetValue(ctx, serializeNumspotPrivateIps, utils.GetPtrValue(http.PrivateIps), diags)
	if diags.HasError() {
		return nil
	}

	if http.SecurityGroups == nil {
		return nil
	}

	securityGroupIds := make([]string, 0, len(*http.SecurityGroups))
	for _, e := range *http.SecurityGroups {
		if e.SecurityGroupId != nil {
			securityGroupIds = append(securityGroupIds, *e.SecurityGroupId)
		}
	}

	securityGroupsIdTf := utils.StringListToTfListValue(ctx, securityGroupIds, diags)
	if diags.HasError() {
		return nil
	}

	securityGroupsTf := utils.GenericListToTfListValue(ctx, serializeNumspotSecurityGroups, utils.GetPtrValue(http.SecurityGroups), diags)
	if diags.HasError() {
		return nil
	}

	if http.LinkPublicIp != nil {
		linkPublicIpTf = serializeLinkPublicIp(ctx, *http.LinkPublicIp, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		linkPublicIpTf = resource_nic.NewLinkPublicIpValueNull()
	}

	if http.LinkNic != nil {
		linkNicTf = serializeNumspotLinkNic(ctx, *http.LinkNic, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		linkNicTf = resource_nic.NewLinkNicValueNull()
	}

	var macAddress *string
	if http.MacAddress != nil {
		lowerMacAddr := strings.ToLower(*http.MacAddress)
		macAddress = &lowerMacAddr
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &resource_nic.NicModel{
		LinkNic:              linkNicTf,
		Description:          types.StringPointerValue(http.Description),
		Id:                   types.StringPointerValue(http.Id),
		IsSourceDestChecked:  types.BoolPointerValue(http.IsSourceDestChecked),
		LinkPublicIp:         linkPublicIpTf,
		MacAddress:           types.StringPointerValue(macAddress),
		VpcId:                types.StringPointerValue(http.VpcId),
		PrivateDnsName:       types.StringPointerValue(http.PrivateDnsName),
		PrivateIps:           privateIps,
		SecurityGroupIds:     securityGroupsIdTf,
		SecurityGroups:       securityGroupsTf,
		State:                types.StringPointerValue(http.State),
		SubnetId:             types.StringPointerValue(http.SubnetId),
		AvailabilityZoneName: types.StringValue(utils.ConvertAzNamePtrToString(http.AvailabilityZoneName)),
		Tags:                 tagsTf,
	}
}

func serializeNumspotLinkNic(ctx context.Context, http api.LinkNic, diags *diag.Diagnostics) resource_nic.LinkNicValue {
	value, diagnostics := resource_nic.NewLinkNicValue(
		resource_nic.LinkNicValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":                    types.StringPointerValue(http.Id),
			"delete_on_vm_deletion": types.BoolPointerValue(http.DeleteOnVmDeletion),
			"device_number":         utils.FromIntPtrToTfInt64(http.DeviceNumber),
			"state":                 types.StringPointerValue(http.State),
			"vm_id":                 types.StringPointerValue(http.VmId),
		})
	diags.Append(diagnostics...)
	return value
}

func serializeNumspotPrivateIps(ctx context.Context, elt api.PrivateIp, diags *diag.Diagnostics) resource_nic.PrivateIpsValue {
	var (
		linkPublicIpTf  resource_nic.LinkPublicIpPrivateIpValue
		linkPublicIpObj basetypes.ObjectValue
	)

	if elt.LinkPublicIp != nil {
		linkPublicIpTf = serializeLinkPublicIpPrivateIp(ctx, *elt.LinkPublicIp, diags)
		if diags.HasError() {
			return resource_nic.PrivateIpsValue{}
		}

		var diagnostics diag.Diagnostics
		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		diags.Append(diagnostics...)
		if diags.HasError() {
			return resource_nic.PrivateIpsValue{}
		}
	} else {
		linkPublicIpTf = resource_nic.NewLinkPublicIpPrivateIpValueNull()
		var diagnostics diag.Diagnostics
		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		diags.Append(diagnostics...)
		if diagnostics.HasError() {
			return resource_nic.PrivateIpsValue{}
		}
	}

	value, diagnostics := resource_nic.NewPrivateIpsValue(
		resource_nic.PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":                types.BoolPointerValue(elt.IsPrimary),
			"link_public_ip_private_ip": linkPublicIpObj,
			"private_dns_name":          types.StringPointerValue(elt.PrivateDnsName),
			"private_ip":                types.StringPointerValue(elt.PrivateIp),
		},
	)
	diags.Append(diagnostics...)
	if diagnostics.HasError() {
		return resource_nic.PrivateIpsValue{}
	}
	return value
}

func serializeNumspotSecurityGroups(ctx context.Context, elt api.SecurityGroupLight, diags *diag.Diagnostics) resource_nic.SecurityGroupsValue {
	value, diagnostics := resource_nic.NewSecurityGroupsValue(
		resource_nic.SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeLinkPublicIp(ctx context.Context, elt api.LinkPublicIp, diags *diag.Diagnostics) resource_nic.LinkPublicIpValue {
	value, diagnostics := resource_nic.NewLinkPublicIpValue(
		resource_nic.LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":              types.StringPointerValue(elt.Id),
			"public_dns_name": types.StringPointerValue(elt.PublicDnsName),
			"public_ip":       types.StringPointerValue(elt.PublicIp),
			"public_ip_id":    types.StringPointerValue(elt.PublicIpId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeLinkPublicIpPrivateIp(ctx context.Context, elt api.LinkPublicIp, diags *diag.Diagnostics) resource_nic.LinkPublicIpPrivateIpValue {
	value, diagnostics := resource_nic.NewLinkPublicIpPrivateIpValue(
		resource_nic.LinkPublicIpValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":              types.StringPointerValue(elt.Id),
			"public_dns_name": types.StringPointerValue(elt.PublicDnsName),
			"public_ip":       types.StringPointerValue(elt.PublicIp),
			"public_ip_id":    types.StringPointerValue(elt.PublicIpId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func nicTags(ctx context.Context, tags types.List) []api.ResourceTag {
	tfTags := make([]resource_nic.TagsValue, 0, len(tags.Elements()))
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
