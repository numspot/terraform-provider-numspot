package internetgateway

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
	"terraform-provider-numspot/internal/services/internetgateway/resource_internet_gateway"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &internetGatewayResource{}
	_ resource.ResourceWithConfigure   = &internetGatewayResource{}
	_ resource.ResourceWithImportState = &internetGatewayResource{}
)

type internetGatewayResource struct {
	provider *client.NumSpotSDK
}

func NewInternetGatewayResource() resource.Resource {
	return &internetGatewayResource{}
}

func (r *internetGatewayResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *internetGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *internetGatewayResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_internet_gateway"
}

func (r *internetGatewayResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_internet_gateway.InternetGatewayResourceSchema(ctx)
}

func (r *internetGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := internetGatewayTags(ctx, plan.Tags)
	vpcId := plan.VpcId.ValueString()

	internetGateway, err := core.CreateInternetGateway(ctx, r.provider, tagsValue, vpcId)
	if err != nil {
		response.Diagnostics.AddError("unable to create internet gateway", err.Error())
		return
	}

	state := serializeNumSpotInternetGateway(ctx, internetGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *internetGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	internetGatewayID := state.Id.ValueString()

	numSpotVolume, err := core.ReadInternetGatewaysWithID(ctx, r.provider, internetGatewayID)
	if err != nil {
		response.Diagnostics.AddError("unable to read internet gateway", err.Error())
		return
	}

	newState := serializeNumSpotInternetGateway(ctx, numSpotVolume, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *internetGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan            resource_internet_gateway.InternetGatewayModel
		numSpotInternetGateway *api.InternetGateway
		err                    error
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	internetGatewayID := state.Id.ValueString()
	planTags := internetGatewayTags(ctx, plan.Tags)
	stateTags := internetGatewayTags(ctx, state.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotInternetGateway, err = core.UpdateInternetGatewayTags(ctx, r.provider, internetGatewayID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update internet gateway tags", err.Error())
			return
		}

		newState := serializeNumSpotInternetGateway(ctx, numSpotInternetGateway, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
	}
}

func (r *internetGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteInternetGateway(ctx, r.provider, state.Id.ValueString(), state.VpcId.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete internet gateway", err.Error())
		return
	}
}

func serializeNumSpotInternetGateway(ctx context.Context, http *api.InternetGateway, diags *diag.Diagnostics) *resource_internet_gateway.InternetGatewayModel {
	var tagsTf types.List

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &resource_internet_gateway.InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		VpcId: types.StringPointerValue(http.VpcId),
		State: types.StringPointerValue(http.State),
		Tags:  tagsTf,
	}
}

func internetGatewayTags(ctx context.Context, tags types.List) []api.ResourceTag {
	tfTags := make([]resource_internet_gateway.TagsValue, 0, len(tags.Elements()))
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
