package clientgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/clientgateway/resource_client_gateway"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &clientGatewayResource{}
	_ resource.ResourceWithConfigure   = &clientGatewayResource{}
	_ resource.ResourceWithImportState = &clientGatewayResource{}
)

// clientGatewayResource is the Terraform resource implementation for numspot_client_gateway.
type clientGatewayResource struct {
	services.BaseResource
}

// NewClientGatewayResource returns a new clientGatewayResource instance.
func NewClientGatewayResource() resource.Resource {
	return &clientGatewayResource{}
}

// Metadata sets the resource type name.
func (r *clientGatewayResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_client_gateway"
}

// Schema defines the schema for the resource.
func (r *clientGatewayResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_client_gateway.ClientGatewayResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *clientGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numSpotClientGateway, err := core.CreateClientGateway(ctx, r.Provider, deserializeCreateClientGatewayRequest(&plan))
	if err != nil {
		response.Diagnostics.AddError("unable to create client_gateway", err.Error())
		return
	}

	state := serializeClientGateway(numSpotClientGateway)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *clientGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotClientGateway, err := core.ReadClientGateway(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read client_gateway", err.Error())
		return
	}

	newState := serializeClientGateway(numSpotClientGateway)

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *clientGatewayResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *clientGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_client_gateway.ClientGatewayModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteClientGateway(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete client_gateway", err.Error())
		return
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

func deserializeCreateClientGatewayRequest(tf *resource_client_gateway.ClientGatewayModel) *api.CreateClientGatewayJSONRequestBody {
	return &api.CreateClientGatewayJSONRequestBody{
		BgpAsn:         utils.FromTfInt64ToInt(tf.BgpAsn),
		ConnectionType: tf.ConnectionType.ValueString(),
		PublicIp:       tf.PublicIp.ValueString(),
	}
}
