package hybridbridge

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/hybridbridge/resource_hybrid_bridge"
)

var _ resource.Resource = &hybridBridgeResource{}

type hybridBridgeResource struct {
	provider *client.NumSpotSDK
}

func NewHybridBridgeResource() resource.Resource {
	return &hybridBridgeResource{}
}

func (r *hybridBridgeResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *hybridBridgeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *hybridBridgeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hybrid_bridge"
}

func (r *hybridBridgeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_hybrid_bridge.HybridBridgeResourceSchema(ctx)
}

func (r *hybridBridgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_hybrid_bridge.HybridBridgeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpcId := plan.VpcId.ValueString()
	serviceManagedId := plan.ManagedServiceId.ValueString()
	body, err := deserializeHybridBridge(plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to deserialize hybrid bridge payload", err.Error())
		return
	}

	numSpot, err := core.CreateHybridBridge(ctx, r.provider, body)
	if err != nil {
		resp.Diagnostics.AddError("unable to create hybrid bridge", err.Error())
		return
	}

	data := serializeHybridBridge(numSpot, vpcId, serviceManagedId)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hybridBridgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan resource_hybrid_bridge.HybridBridgeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpcId := plan.VpcId.ValueString()
	serviceManagedId := plan.ManagedServiceId.ValueString()
	id, err := uuid.Parse(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to read managed service bridge", err.Error())
		return
	}

	hybridBridge, err := core.ReadHybridBridge(ctx, r.provider, id)
	if err != nil {
		resp.Diagnostics.AddError("unable to read hybrid bridge", err.Error())
		return
	}

	newPlan := serializeHybridBridge(hybridBridge, vpcId, serviceManagedId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newPlan)...)
}

func (r *hybridBridgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *hybridBridgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan resource_hybrid_bridge.HybridBridgeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to read managed service bridge", err.Error())
		return
	}

	if err := core.DeleteHybridBridge(ctx, r.provider, id); err != nil {
		resp.Diagnostics.AddError("unable to delete hybrid bridge", err.Error())
		return
	}
}

func deserializeHybridBridge(tf resource_hybrid_bridge.HybridBridgeModel) (api.CreateHybridBridgeRequest, error) {
	managedServiceId, err := uuid.Parse(tf.ManagedServiceId.ValueString())
	if err != nil {
		return api.CreateHybridBridgeRequest{}, err
	}

	return api.CreateHybridBridgeRequest{
		VpcId:            tf.VpcId.ValueString(),
		ManagedServiceId: managedServiceId,
	}, nil
}

func serializeHybridBridge(http *api.HybridBridge, vpcId, managedServiceId string) resource_hybrid_bridge.HybridBridgeModel {
	return resource_hybrid_bridge.HybridBridgeModel{
		Id:               types.StringValue(http.Id.String()),
		VpcId:            types.StringValue(vpcId),
		ManagedServiceId: types.StringValue(managedServiceId),
		Route: resource_hybrid_bridge.RouteValue{
			GatewayId:          types.StringValue(http.Route.GatewayId),
			DestinationIpRange: types.StringValue(http.Route.DestinationIpRange),
		},
	}
}
