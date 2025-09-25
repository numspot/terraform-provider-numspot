package publicip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/publicip/resource_public_ip"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &publicIpResource{}
	_ resource.ResourceWithConfigure   = &publicIpResource{}
	_ resource.ResourceWithImportState = &publicIpResource{}
)

type publicIpResource struct {
	provider *client.NumSpotSDK
}

func NewPublicIpResource() resource.Resource {
	return &publicIpResource{}
}

func (r *publicIpResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *publicIpResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *publicIpResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_public_ip"
}

func (r *publicIpResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_public_ip.PublicIpResourceSchema(ctx)
}

func (r *publicIpResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_public_ip.PublicIpModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vmId := plan.VmId.ValueString()
	nicId := plan.NicId.ValueString()
	tagsValue := publicIpTags(ctx, plan.Tags)

	publicIp, err := core.CreatePublicIp(ctx, r.provider, tagsValue, vmId, nicId)
	if err != nil {
		response.Diagnostics.AddError("unable to create public ip", err.Error())
		return
	}

	state := serializePublicIp(ctx, publicIp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *publicIpResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_public_ip.PublicIpModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	publicIpID := state.Id.ValueString()

	numSpotPublicIp, err := core.ReadPublicIp(ctx, r.provider, publicIpID)
	if err != nil {
		response.Diagnostics.AddError("unable to read public ip", err.Error())
		return
	}

	newState := serializePublicIp(ctx, numSpotPublicIp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func (r *publicIpResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resource_public_ip.PublicIpModel
	var numSpotPublicIp *api.PublicIp
	var err error

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	publicIpID := state.Id.ValueString()
	planTags := publicIpTags(ctx, plan.Tags)
	stateTags := publicIpTags(ctx, state.Tags)

	if !state.Tags.Equal(plan.Tags) {
		numSpotPublicIp, err = core.UpdatePublicIpTags(ctx, r.provider, stateTags, planTags, publicIpID)
		if err != nil {
			response.Diagnostics.AddError("unable to update public ip tags", err.Error())
			return
		}
	}

	state = serializePublicIp(ctx, numSpotPublicIp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *publicIpResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	publicIpID := state.Id.ValueString()
	linkPublicIpId := state.LinkPublicIpId.ValueString()
	err := core.DeletePublicIp(ctx, r.provider, publicIpID, linkPublicIpId)
	if err != nil {
		response.Diagnostics.AddError("unable to delete public ip", err.Error())
		return
	}
}

func serializePublicIp(ctx context.Context, elt *api.PublicIp, diags *diag.Diagnostics) resource_public_ip.PublicIpModel {
	var tagsList types.Set

	if elt.Tags != nil {
		tagsList = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *elt.Tags, diags)
		if diags.HasError() {
			return resource_public_ip.PublicIpModel{}
		}
	}

	return resource_public_ip.PublicIpModel{
		Id:             types.StringPointerValue(elt.Id),
		NicId:          types.StringPointerValue(elt.NicId),
		PrivateIp:      types.StringPointerValue(elt.PrivateIp),
		PublicIp:       types.StringPointerValue(elt.PublicIp),
		VmId:           types.StringPointerValue(elt.VmId),
		LinkPublicIpId: types.StringPointerValue(elt.LinkPublicIpId),
		Tags:           tagsList,
	}
}

func publicIpTags(ctx context.Context, tags types.Set) []api.ResourceTag {
	tfTags := make([]resource_public_ip.TagsValue, 0, len(tags.Elements()))
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
