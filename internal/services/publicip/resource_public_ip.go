package publicip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/publicip/resource_public_ip"
	"github.com/numspot/terraform-provider-numspot/internal/services/tags"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
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
	numspotPublicIp, err := core.CreatePublicIp(ctx, r.Provider, plan.VmId.ValueString(), plan.NicId.ValueString(), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create public_ip", err.Error())
		return
	}

	state := serializePublicIp(ctx, numspotPublicIp, &response.Diagnostics)
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

	numspotPublicIp, err := core.ReadPublicIp(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		if utils.IsResourceGone(err) {
			tflog.Warn(ctx, "unable to read public_ip (not found or access forbidden); removing from state")
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("unable to read public_ip", err.Error())
		return
	}

	newState := serializePublicIp(ctx, numspotPublicIp, &response.Diagnostics)
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
		numspotPublicIp *api.PublicIp
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
	planTags := publicipTags(ctx, plan.Tags)
	stateTags := publicipTags(ctx, state.Tags)

	if !plan.VmId.Equal(state.VmId) || !plan.NicId.Equal(state.NicId) {
		numspotPublicIp, err = core.UpdatePublicIpTarget(ctx, r.Provider, publicipID, plan.VmId.ValueString(), plan.NicId.ValueString())
		if err != nil {
			response.Diagnostics.AddError("unable to update public_ip", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numspotPublicIp, err = core.UpdatePublicIpTags(ctx, r.Provider, publicipID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update public_ip tags", err.Error())
			return
		}
	}

	if numspotPublicIp == nil {
		numspotPublicIp, err = core.ReadPublicIp(ctx, r.Provider, state.Id.ValueString())
		if err != nil {
			response.Diagnostics.AddError("unable to read public_ip", err.Error())
			return
		}
	}

	newState := serializePublicIp(ctx, numspotPublicIp, &response.Diagnostics)
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
