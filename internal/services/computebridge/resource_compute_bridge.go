package computebridge

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/computebridge/resource_compute_bridge"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &computeBridgeResource{}
	_ resource.ResourceWithConfigure   = &computeBridgeResource{}
	_ resource.ResourceWithImportState = &computeBridgeResource{}
)

// computeBridgeResource is the Terraform resource implementation for numspot_compute_bridge.
type computeBridgeResource struct {
	services.BaseResource
}

// NewComputeBridgeResource returns a new computeBridgeResource instance.
func NewComputeBridgeResource() resource.Resource {
	return &computeBridgeResource{}
}

// Metadata sets the resource type name.
func (r *computeBridgeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_compute_bridge"
}

// Schema defines the schema for the resource.
func (r *computeBridgeResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_compute_bridge.ComputeBridgeResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *computeBridgeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_compute_bridge.ComputeBridgeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numSpotComputeBridge, err := core.CreateComputeBridge(ctx, r.Provider, deserializeCreateComputeBridgeRequest(&plan))
	if err != nil {
		response.Diagnostics.AddError("unable to create compute_bridge", err.Error())
		return
	}

	state := serializeComputeBridge(numSpotComputeBridge)
	state.DestinationVpcId = utils.PreserveFromState(plan.DestinationVpcId)
	state.SourceVpcId = utils.PreserveFromState(plan.SourceVpcId)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *computeBridgeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_compute_bridge.ComputeBridgeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotComputeBridge, err := core.ReadComputeBridge(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read compute_bridge", err.Error())
		return
	}

	newState := serializeComputeBridge(numSpotComputeBridge)
	newState.DestinationVpcId = state.DestinationVpcId
	newState.SourceVpcId = state.SourceVpcId

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *computeBridgeResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *computeBridgeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_compute_bridge.ComputeBridgeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteComputeBridge(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete compute_bridge", err.Error())
		return
	}
}

func serializeComputeBridge(http *api.ComputeBridge) resource_compute_bridge.ComputeBridgeModel {
	return resource_compute_bridge.ComputeBridgeModel{
		Destination:        types.StringValue(http.Destination),
		DestinationId:      types.StringValue(http.DestinationId),
		DestinationIpRange: types.StringValue(http.DestinationIpRange),
		GatewayId:          types.StringValue(http.GatewayId),
		Id:                 types.StringValue(http.Id.String()),
		Source:             types.StringValue(http.Source),
		SourceId:           types.StringValue(http.SourceId),
		SourceIpRange:      types.StringValue(http.SourceIpRange),
	}
}

func deserializeCreateComputeBridgeRequest(tf *resource_compute_bridge.ComputeBridgeModel) *api.CreateComputeBridgeJSONRequestBody {
	return &api.CreateComputeBridgeJSONRequestBody{
		DestinationVpcId: tf.DestinationVpcId.ValueString(),
		SourceVpcId:      tf.SourceVpcId.ValueString(),
	}
}
