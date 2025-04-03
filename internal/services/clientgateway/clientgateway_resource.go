package clientgateway

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/clientgateway/resource_client_gateway"
	"terraform-provider-numspot/internal/utils"
)

const resourceTypeName = "_client_gateway"

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewClientGatewayResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"unexpected resource configure type",
			fmt.Sprintf("expected *http.Client, got: %T please report this issue to the provider developers", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + resourceTypeName
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_client_gateway.ClientGatewayResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	clientGateway, err := core.CreateClientGateway(ctx, r.provider, deserializeCreateClientGateway(plan))
	if err != nil {
		response.Diagnostics.AddError("unable to create client gateway", err.Error())
		return
	}

	state := serializeClientGateway(clientGateway)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	clientGatewayID, err := uuid.Parse(state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to parse id from state", err.Error())
		return
	}

	numSpotClientGateway, err := core.ReadClientGateway(ctx, r.provider, clientGatewayID)
	if err != nil {
		response.Diagnostics.AddError("unable to read client gateway", err.Error())
		return
	}

	newState := serializeClientGateway(numSpotClientGateway)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_client_gateway.ClientGatewayModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	clientGatewayID, err := uuid.Parse(state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to parse id from state", err.Error())
		return
	}

	err = core.DeleteClientGateway(ctx, r.provider, clientGatewayID)
	if err != nil {
		response.Diagnostics.AddError("unable to delete client gateway", err.Error())
		return
	}
}

func deserializeCreateClientGateway(tf resource_client_gateway.ClientGatewayModel) api.CreateClientGatewayJSONRequestBody {
	return api.CreateClientGatewayJSONRequestBody{
		BgpAsn:         utils.FromTfInt64ToInt(tf.BgpAsn),
		ConnectionType: tf.ConnectionType.ValueString(),
		PublicIp:       tf.PublicIp.ValueString(),
	}
}

func serializeClientGateway(http *api.ClientGateway) resource_client_gateway.ClientGatewayModel {
	return resource_client_gateway.ClientGatewayModel{
		BgpAsn:         utils.FromIntToTfInt64(http.BgpAsn),
		ConnectionType: types.StringValue(http.ConnectionType),
		Id:             types.StringValue(http.Id.String()),
		PublicIp:       types.StringValue(http.PublicIp),
		State:          types.StringValue(http.State),
	}
}
