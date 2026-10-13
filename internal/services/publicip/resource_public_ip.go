package publicip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

// publicIpResource is the Terraform resource implementation for numspot_public_ip.
type publicIpResource struct {
	services.BaseResource
}

// NewPublicIpResource returns a new publicIpResource instance.
func NewPublicIpResource() resource.Resource {
	return &publicIpResource{}
}

// Metadata sets the resource type name.
func (r *publicIpResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_public_ip"
}

// Schema defines the schema for the resource.
func (r *publicIpResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_public_ip.PublicIpResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *publicIpResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := publicipTags(ctx, plan.Tags)
	numSpotPublicIp, err := core.CreatePublicIp(ctx, r.Provider, plan.VmId.ValueString(), plan.NicId.ValueString(), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create public_ip", err.Error())
		return
	}

	state := serializePublicIp(ctx, numSpotPublicIp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *publicIpResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotPublicIp, err := core.ReadPublicIp(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read public_ip", err.Error())
		return
	}

	newState := serializePublicIp(ctx, numSpotPublicIp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *publicIpResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err             error
		state, plan     resource_public_ip.PublicIpModel
		numSpotPublicIp *api.PublicIp
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	publicipID := state.Id.ValueString()

	// Move the link to the planned VM/NIC when the target changed (e.g. the
	// linked VM was replaced during the same apply). The target is read from
	// the API first, so a stale state does not break the update. The IP itself
	// is never destroyed here: it keeps its address (SUPN2-643).
	targetChanged := !plan.VmId.Equal(state.VmId) || !plan.NicId.Equal(state.NicId)
	if targetChanged {
		numSpotPublicIp, err = core.UpdatePublicIpTarget(ctx, r.Provider, publicipID, plan.VmId.ValueString(), plan.NicId.ValueString())
		if err != nil {
			response.Diagnostics.AddError("unable to update public_ip target", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		planTags := publicipTags(ctx, plan.Tags)
		stateTags := publicipTags(ctx, state.Tags)

		numSpotPublicIp, err = core.UpdatePublicIpTags(ctx, r.Provider, publicipID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update public_ip tags", err.Error())
			return
		}
	}

	// Neither the target nor the tags changed (e.g. a planned value that was
	// unknown in the plan resolved to the state value at apply time): refresh
	// from the API so the serialized state is never built from a nil response.
	if numSpotPublicIp == nil {
		numSpotPublicIp, err = core.ReadPublicIp(ctx, r.Provider, publicipID)
		if err != nil {
			response.Diagnostics.AddError("unable to read public_ip", err.Error())
			return
		}
	}

	newState := serializePublicIp(ctx, numSpotPublicIp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *publicIpResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeletePublicIp(ctx, r.Provider, state.Id.ValueString(), state.LinkPublicIpId.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete public_ip", err.Error())
		return
	}
}

func serializePublicIp(ctx context.Context, http *api.PublicIp, diags *diag.Diagnostics) resource_public_ip.PublicIpModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil PublicIp response", "the API returned an empty PublicIp")
		return resource_public_ip.PublicIpModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_public_ip.PublicIpModel{
		Id:             types.StringPointerValue(http.Id),
		LinkPublicIpId: types.StringPointerValue(http.LinkPublicIpId),
		NicId:          types.StringPointerValue(http.NicId),
		PrivateIp:      types.StringPointerValue(http.PrivateIp),
		PublicIp:       types.StringPointerValue(http.PublicIp),
		VmId:           types.StringPointerValue(http.VmId),
		Tags:           tagsTf,
	}
}

func publicipTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_public_ip.TagsValue, 0, len(tagsSet.Elements()))
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
