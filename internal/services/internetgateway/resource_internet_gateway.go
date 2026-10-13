package internetgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/internetgateway/resource_internet_gateway"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &internetGatewayResource{}
	_ resource.ResourceWithConfigure   = &internetGatewayResource{}
	_ resource.ResourceWithImportState = &internetGatewayResource{}
)

// internetGatewayResource is the Terraform resource implementation for numspot_internet_gateway.
type internetGatewayResource struct {
	services.BaseResource
}

// NewInternetGatewayResource returns a new internetGatewayResource instance.
func NewInternetGatewayResource() resource.Resource {
	return &internetGatewayResource{}
}

// Metadata sets the resource type name.
func (r *internetGatewayResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_internet_gateway"
}

// Schema defines the schema for the resource.
func (r *internetGatewayResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_internet_gateway.InternetGatewayResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *internetGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := internetgatewayTags(ctx, plan.Tags)
	numSpotInternetGateway, err := core.CreateInternetGateway(ctx, r.Provider, plan.VpcId.ValueString(), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create internet_gateway", err.Error())
		return
	}

	state := serializeInternetGateway(ctx, numSpotInternetGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *internetGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotInternetGateway, err := core.ReadInternetGateway(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read internet_gateway", err.Error())
		return
	}

	newState := serializeInternetGateway(ctx, numSpotInternetGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *internetGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                    error
		state, plan            resource_internet_gateway.InternetGatewayModel
		numSpotInternetGateway *api.InternetGateway
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	internetgatewayID := state.Id.ValueString()
	planTags := internetgatewayTags(ctx, plan.Tags)
	stateTags := internetgatewayTags(ctx, state.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotInternetGateway, err = core.UpdateInternetGatewayTags(ctx, r.Provider, internetgatewayID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update internet_gateway tags", err.Error())
			return
		}
	}

	newState := serializeInternetGateway(ctx, numSpotInternetGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *internetGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteInternetGateway(ctx, r.Provider, state.Id.ValueString(), state.VpcId.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete internet_gateway", err.Error())
		return
	}
}

func serializeInternetGateway(
	ctx context.Context,
	http *api.InternetGateway,
	diags *diag.Diagnostics,
) resource_internet_gateway.InternetGatewayModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil InternetGateway response", "the API returned an empty InternetGateway")
		return resource_internet_gateway.InternetGatewayModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_internet_gateway.InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		State: types.StringPointerValue(http.State),
		VpcId: types.StringPointerValue(http.VpcId),
		Tags:  tagsTf,
	}
}

func internetgatewayTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_internet_gateway.TagsValue, 0, len(tagsSet.Elements()))
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
