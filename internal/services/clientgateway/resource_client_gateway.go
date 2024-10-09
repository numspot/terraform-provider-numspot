package clientgateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &ClientGatewayResource{}
	_ resource.ResourceWithConfigure   = &ClientGatewayResource{}
	_ resource.ResourceWithImportState = &ClientGatewayResource{}
)

type ClientGatewayResource struct {
	provider services.IProvider
}

func NewClientGatewayResource() resource.Resource {
	return &ClientGatewayResource{}
}

func (r *ClientGatewayResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(services.IProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *ClientGatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *ClientGatewayResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_client_gateway"
}

func (r *ClientGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = ClientGatewayResourceSchema(ctx)
}

func (r *ClientGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan ClientGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)

	clientGateway, err := core.CreateClientGateway(ctx, r.provider, deserializeCreateClientGateway(plan), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}

	state, diags := serializeClientGateway(ctx, clientGateway)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *ClientGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ClientGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	clientGatewayID := state.Id.ValueString()

	numSpotClientGateway, err := core.ReadClientGateway(ctx, r.provider, clientGatewayID)
	if err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}

	state, diags := serializeClientGateway(ctx, numSpotClientGateway)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *ClientGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                  error
		numSpotClientGateway *numspot.ClientGateway
		state, plan          ClientGatewayModel
		diags                diag.Diagnostics
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	clientGatewayID := state.Id.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

	if !state.Tags.Equal(plan.Tags) {
		numSpotClientGateway, err = core.UpdateClientGatewayTags(ctx, r.provider, stateTags, planTags, clientGatewayID)
		if err != nil {
			response.Diagnostics.AddError("", err.Error())
			return
		}
	}

	state, diags = serializeClientGateway(ctx, numSpotClientGateway)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *ClientGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state ClientGatewayModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	clientGatewayID := state.Id.ValueString()

	err := core.DeleteClientGateway(ctx, r.provider, clientGatewayID)
	if err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}
}

func deserializeCreateClientGateway(tf ClientGatewayModel) numspot.CreateClientGatewayJSONRequestBody {
	return numspot.CreateClientGatewayJSONRequestBody{
		BgpAsn:         utils.FromTfInt64ToInt(tf.BgpAsn),
		ConnectionType: tf.ConnectionType.ValueString(),
		PublicIp:       tf.PublicIp.ValueString(),
	}
}

func serializeClientGateway(ctx context.Context, http *numspot.ClientGateway) (ClientGatewayModel, diag.Diagnostics) {
	var (
		tagsTf types.List
		diags  diag.Diagnostics
	)

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, &diags)
		if diags.HasError() {
			return ClientGatewayModel{}, diags
		}
	}

	return ClientGatewayModel{
		BgpAsn:         utils.FromIntPtrToTfInt64(http.BgpAsn),
		ConnectionType: types.StringPointerValue(http.ConnectionType),
		Id:             types.StringPointerValue(http.Id),
		PublicIp:       types.StringPointerValue(http.PublicIp),
		State:          types.StringPointerValue(http.State),
		Tags:           tagsTf,
	}, diags
}
