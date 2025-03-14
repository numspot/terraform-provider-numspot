package publicip

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func NewPublicIpResource() resource.Resource {
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
	response.TypeName = request.ProviderTypeName + "_public_ip"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = PublicIpResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan PublicIpModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vmId := plan.VmId.ValueString()
	nicId := plan.NicId.ValueString()
	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)

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

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state PublicIpModel

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

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state PublicIpModel
	var numSpotPublicIp *numspot.PublicIp
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
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

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

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state PublicIpModel
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

func serializePublicIp(ctx context.Context, elt *numspot.PublicIp, diags *diag.Diagnostics) PublicIpModel {
	var tagsList types.List

	if elt.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *elt.Tags, diags)
		if diags.HasError() {
			return PublicIpModel{}
		}
	}

	return PublicIpModel{
		Id:             types.StringPointerValue(elt.Id),
		NicId:          types.StringPointerValue(elt.NicId),
		PrivateIp:      types.StringPointerValue(elt.PrivateIp),
		PublicIp:       types.StringPointerValue(elt.PublicIp),
		VmId:           types.StringPointerValue(elt.VmId),
		LinkPublicIpId: types.StringPointerValue(elt.LinkPublicIpId),
		Tags:           tagsList,
	}
}
