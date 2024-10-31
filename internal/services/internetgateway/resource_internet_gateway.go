package internetgateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &InternetGatewayResource{}
	_ resource.ResourceWithConfigure   = &InternetGatewayResource{}
	_ resource.ResourceWithImportState = &InternetGatewayResource{}
)

type InternetGatewayResource struct {
	provider *client.NumSpotSDK
}

func NewInternetGatewayResource() resource.Resource {
	return &InternetGatewayResource{}
}

func (r *InternetGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *InternetGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *InternetGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_internet_gateway"
}

func (r *InternetGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = InternetGatewayResourceSchema(ctx)
}

func (r *InternetGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan InternetGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
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

func (r *InternetGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state InternetGatewayModel
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

func (r *InternetGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan            InternetGatewayModel
		numSpotInternetGateway *numspot.InternetGateway
		err                    error
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	internetGatewayID := state.Id.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

	if !plan.Tags.Equal(state.Tags) {
		numSpotInternetGateway, err = core.UpdateInternetGatewayTags(ctx, r.provider, internetGatewayID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update internet gateway tags", err.Error())
			return
		}
	}

	newState := serializeNumSpotInternetGateway(ctx, numSpotInternetGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *InternetGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state InternetGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteInternetGateway(ctx, r.provider, state.Id.ValueString(), state.VpcId.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete internet gateway", err.Error())
		return
	}
}

func serializeNumSpotInternetGateway(ctx context.Context, http *numspot.InternetGateway, diags *diag.Diagnostics) *InternetGatewayModel {
	var tagsTf types.List

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &InternetGatewayModel{
		Id:    types.StringPointerValue(http.Id),
		VpcId: types.StringPointerValue(http.VpcId),
		State: types.StringPointerValue(http.State),
		Tags:  tagsTf,
	}
}
