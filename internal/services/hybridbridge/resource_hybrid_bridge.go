package hybridbridge

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/hybridbridge/resource_hybrid_bridge"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &hybridBridgeResource{}
	_ resource.ResourceWithConfigure   = &hybridBridgeResource{}
	_ resource.ResourceWithImportState = &hybridBridgeResource{}
)

// hybridBridgeResource is the Terraform resource implementation for numspot_hybrid_bridge.
type hybridBridgeResource struct {
	services.BaseResource
}

// NewHybridBridgeResource returns a new hybridBridgeResource instance.
func NewHybridBridgeResource() resource.Resource {
	return &hybridBridgeResource{}
}

// Metadata sets the resource type name.
func (r *hybridBridgeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_hybrid_bridge"
}

// Schema defines the schema for the resource.
func (r *hybridBridgeResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_hybrid_bridge.HybridBridgeResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *hybridBridgeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_hybrid_bridge.HybridBridgeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numspotHybridBridge, err := core.CreateHybridBridge(ctx, r.Provider, deserializeCreateHybridBridgeRequest(&plan))
	if err != nil {
		response.Diagnostics.AddError("unable to create hybrid_bridge", err.Error())
		return
	}

	state := serializeHybridBridge(ctx, numspotHybridBridge, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	state.ManagedServiceId = utils.PreserveFromState(plan.ManagedServiceId)
	state.VpcId = utils.PreserveFromState(plan.VpcId)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *hybridBridgeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_hybrid_bridge.HybridBridgeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotHybridBridge, err := core.ReadHybridBridge(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read hybrid_bridge", err.Error())
		return
	}

	newState := serializeHybridBridge(ctx, numspotHybridBridge, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.ManagedServiceId = state.ManagedServiceId
	newState.VpcId = state.VpcId

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *hybridBridgeResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *hybridBridgeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_hybrid_bridge.HybridBridgeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteHybridBridge(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete hybrid_bridge", err.Error())
		return
	}
}

func serializeHybridBridge(ctx context.Context, http *api.HybridBridge, diags *diag.Diagnostics) resource_hybrid_bridge.HybridBridgeModel {
	if http == nil {
		diags.AddError("nil HybridBridge response", "the API returned an empty HybridBridge")
		return resource_hybrid_bridge.HybridBridgeModel{}
	}

	return resource_hybrid_bridge.HybridBridgeModel{
		Destination:   types.StringValue(http.Destination),
		DestinationId: types.StringValue(http.DestinationId),
		Id:            types.StringValue(http.Id.String()),
		Route:         serializeRoute(ctx, &http.Route, diags),
		Source:        types.StringValue(http.Source),
		SourceId:      types.StringValue(http.SourceId.String()),
	}
}

func deserializeCreateHybridBridgeRequest(tf *resource_hybrid_bridge.HybridBridgeModel) *api.CreateHybridBridgeJSONRequestBody {
	return &api.CreateHybridBridgeJSONRequestBody{
		ManagedServiceId: utils.FromTfStringToUuid(tf.ManagedServiceId),
		VpcId:            tf.VpcId.ValueString(),
	}
}

func serializeRoute(ctx context.Context, item *api.RouteConnectivity, diags *diag.Diagnostics) resource_hybrid_bridge.RouteValue {
	if item == nil {
		return resource_hybrid_bridge.NewRouteValueNull()
	}
	value, d := resource_hybrid_bridge.NewRouteValue(resource_hybrid_bridge.RouteValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"destination_ip_range": types.StringValue(item.DestinationIpRange),
		"gateway_id":           types.StringValue(item.GatewayId),
	})
	diags.Append(d...)
	return value
}
