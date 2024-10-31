package nic

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &NicResource{}
	_ resource.ResourceWithConfigure   = &NicResource{}
	_ resource.ResourceWithImportState = &NicResource{}
)

type NicResource struct {
	provider *client.NumSpotSDK
}

func NewNicResource() resource.Resource {
	return &NicResource{}
}

func (r *NicResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *NicResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *NicResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_nic"
}

func (r *NicResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = NicResourceSchema(ctx)
}

func (r *NicResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan NicModel
	var linkNicBody *numspot.LinkNicJSONRequestBody
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
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

func (r *NicResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state NicModel
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

func (r *NicResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan NicModel
		numspotNic  *numspot.Nic
		err         error
	)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	nicId := state.Id.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

	// Update tags
	if !state.Tags.Equal(plan.Tags) {
		numspotNic, err = core.UpdateNicTags(ctx, r.provider, nicId, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update nic tags", err.Error())
			return
		}
	}

	// Update link
	if !utils.IsTfValueNull(plan.LinkNic) || !utils.IsTfValueNull(state.LinkNic) {
		numspotNic, err = core.UpdateNicLink(ctx, r.provider, nicId, deserializeUnlinkNic(state.LinkNic), deserializeLinkNic(plan.LinkNic))
		if err != nil {
			response.Diagnostics.AddError("unable to update nic link", err.Error())
			return
		}
	}

	// Update Nic
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

func (r *NicResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state NicModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteNic(ctx, r.provider, state.Id.ValueString(), deserializeUnlinkNic(state.LinkNic)); err != nil {
		response.Diagnostics.AddError("unable to delete Nic", err.Error())
		return
	}
}

func deserializeCreateNumSpotNic(ctx context.Context, tf NicModel, diags *diag.Diagnostics) numspot.CreateNicJSONRequestBody {
	var privateIpsPtr *[]numspot.PrivateIpLight
	var securityGroupIdsPtr *[]string

	if !(tf.PrivateIps.IsNull() || tf.PrivateIps.IsUnknown()) {
		privateIps := utils.TfSetToGenericSet(func(a PrivateIpsValue) numspot.PrivateIpLight {
			return numspot.PrivateIpLight{
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
	return numspot.CreateNicJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		PrivateIps:       privateIpsPtr,
		SecurityGroupIds: securityGroupIdsPtr,
		SubnetId:         tf.SubnetId.ValueString(),
	}
}

func deserializeUpdateNumSpotNic(ctx context.Context, tf NicModel, diags *diag.Diagnostics) numspot.UpdateNicJSONRequestBody {
	linkNic := numspot.LinkNicToUpdate{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.LinkNic.DeleteOnVmDeletion),
		LinkNicId:          utils.FromTfStringToStringPtr(tf.LinkNic.Id),
	}

	return numspot.UpdateNicJSONRequestBody{
		SecurityGroupIds: utils.TfStringListToStringPtrList(ctx, tf.SecurityGroupIds, diags),
		Description:      utils.FromTfStringToStringPtr(tf.Description),
		LinkNic:          &linkNic,
	}
}

func deserializeLinkNic(tf LinkNicValue) *numspot.LinkNicJSONRequestBody {
	if utils.IsTfValueNull(tf) {
		return nil
	}

	return &numspot.LinkNicJSONRequestBody{
		DeviceNumber: utils.FromTfInt64ToInt(tf.DeviceNumber),
		VmId:         tf.VmId.ValueString(),
	}
}

func deserializeUnlinkNic(tf LinkNicValue) *numspot.UnlinkNicJSONRequestBody {
	if utils.IsTfValueNull(tf) {
		return nil
	}

	return &numspot.UnlinkNicJSONRequestBody{
		LinkNicId: tf.Id.ValueString(),
	}
}

func serializeNumSpotNic(ctx context.Context, http *numspot.Nic, diags *diag.Diagnostics) *NicModel {
	var (
		linkPublicIpTf LinkPublicIpValue
		linkNicTf      LinkNicValue
		tagsTf         types.List
	)
	// Private IPs
	privateIps := utils.GenericSetToTfSetValue(ctx, serializeNumspotPrivateIps, utils.GetPtrValue(http.PrivateIps), diags)
	if diags.HasError() {
		return nil
	}

	if http.SecurityGroups == nil {
		return nil
	}
	// Retrieve security groups id
	securityGroupIds := make([]string, 0, len(*http.SecurityGroups))
	for _, e := range *http.SecurityGroups {
		if e.SecurityGroupId != nil {
			securityGroupIds = append(securityGroupIds, *e.SecurityGroupId)
		}
	}

	// Security Group Ids
	securityGroupsIdTf := utils.StringListToTfListValue(ctx, securityGroupIds, diags)
	if diags.HasError() {
		return nil
	}

	// Security Groups
	securityGroupsTf := utils.GenericListToTfListValue(ctx, serializeNumspotSecurityGroups, utils.GetPtrValue(http.SecurityGroups), diags)
	if diags.HasError() {
		return nil
	}

	// Link Public Ip
	if http.LinkPublicIp != nil {
		linkPublicIpTf = serializeLinkPublicIp(ctx, *http.LinkPublicIp, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		linkPublicIpTf = NewLinkPublicIpValueNull()
	}

	// Link NIC
	if http.LinkNic != nil {
		linkNicTf = serializeNumspotLinkNic(ctx, *http.LinkNic, diags)
		if diags.HasError() {
			return nil
		}
	} else {
		linkNicTf = NewLinkNicValueNull()
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

	return &NicModel{
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
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Tags:                 tagsTf,
	}
}

func serializeNumspotLinkNic(ctx context.Context, http numspot.LinkNic, diags *diag.Diagnostics) LinkNicValue {
	value, diagnostics := NewLinkNicValue(
		LinkNicValue{}.AttributeTypes(ctx),
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

func serializeNumspotPrivateIps(ctx context.Context, elt numspot.PrivateIp, diags *diag.Diagnostics) PrivateIpsValue {
	var (
		linkPublicIpTf  LinkPublicIpValue
		linkPublicIpObj basetypes.ObjectValue
	)

	if elt.LinkPublicIp != nil {
		linkPublicIpTf = serializeLinkPublicIp(ctx, *elt.LinkPublicIp, diags)
		if diags.HasError() {
			return PrivateIpsValue{}
		}

		var diagnostics diag.Diagnostics
		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		diags.Append(diagnostics...)
		if diags.HasError() {
			return PrivateIpsValue{}
		}
	} else {
		linkPublicIpTf = NewLinkPublicIpValueNull()
		var diagnostics diag.Diagnostics
		linkPublicIpObj, diagnostics = linkPublicIpTf.ToObjectValue(ctx)
		diags.Append(diagnostics...)
		if diagnostics.HasError() {
			return PrivateIpsValue{}
		}
	}

	value, diagnostics := NewPrivateIpsValue(
		PrivateIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_primary":       types.BoolPointerValue(elt.IsPrimary),
			"link_public_ip":   linkPublicIpObj,
			"private_dns_name": types.StringPointerValue(elt.PrivateDnsName),
			"private_ip":       types.StringPointerValue(elt.PrivateIp),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeNumspotSecurityGroups(ctx context.Context, elt numspot.SecurityGroupLight, diags *diag.Diagnostics) SecurityGroupsValue {
	value, diagnostics := NewSecurityGroupsValue(
		SecurityGroupsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"security_group_id":   types.StringPointerValue(elt.SecurityGroupId),
			"security_group_name": types.StringPointerValue(elt.SecurityGroupName),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeLinkPublicIp(ctx context.Context, elt numspot.LinkPublicIp, diags *diag.Diagnostics) LinkPublicIpValue {
	value, diagnostics := NewLinkPublicIpValue(
		LinkPublicIpValue{}.AttributeTypes(ctx),
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
