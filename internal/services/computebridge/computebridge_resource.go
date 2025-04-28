package computebridge

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
	"terraform-provider-numspot/internal/services/computebridge/resource_compute_bridge"
)

var _ resource.Resource = &computeBridgeResource{}

type computeBridgeResource struct {
	provider *client.NumSpotSDK
}

func NewComputeBridgeResource() resource.Resource {
	return &computeBridgeResource{}
}

func (r *computeBridgeResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *computeBridgeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *computeBridgeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_bridge"
}

func (r *computeBridgeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_compute_bridge.ComputeBridgeResourceSchema(ctx)
}

func (r *computeBridgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_compute_bridge.ComputeBridgeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	VpcA := plan.SourceVpcId.ValueString()
	VpcB := plan.DestinationVpcId.ValueString()

	body := deserializeComputeBridge(plan)

	numSpot, err := core.CreateComputeBridge(ctx, r.provider, body)
	if err != nil {
		resp.Diagnostics.AddError("unable to create compute bridge", err.Error())
		return
	}

	data := serializeComputeBridge(numSpot, VpcA, VpcB)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *computeBridgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan resource_compute_bridge.ComputeBridgeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	VpcA := plan.SourceVpcId.ValueString()
	VpcB := plan.DestinationVpcId.ValueString()
	id := uuid.MustParse(plan.Id.ValueString())
	computeBridge, err := core.ReadComputeBridge(ctx, r.provider, id)
	if err != nil {
		resp.Diagnostics.AddError("unable to read compute bridge", err.Error())
		return
	}

	newPlan := serializeComputeBridge(computeBridge, VpcA, VpcB)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newPlan)...)
}

func (r *computeBridgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *computeBridgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan resource_compute_bridge.ComputeBridgeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := uuid.MustParse(plan.Id.ValueString())
	if err := core.DeleteComputeBridge(ctx, r.provider, id); err != nil {
		resp.Diagnostics.AddError("unable to delete Compute bridge", err.Error())
		return
	}
}

func deserializeComputeBridge(tf resource_compute_bridge.ComputeBridgeModel) api.CreateComputeBridgeRequest {
	return api.CreateComputeBridgeRequest{
		SourceVpcId:      tf.SourceVpcId.ValueString(),
		DestinationVpcId: tf.DestinationVpcId.ValueString(),
	}
}

func serializeComputeBridge(http *api.ComputeBridge, sourceVpcId, destinationVpcId string) resource_compute_bridge.ComputeBridgeModel {
	return resource_compute_bridge.ComputeBridgeModel{
		SourceVpcId:        types.StringValue(sourceVpcId),
		DestinationVpcId:   types.StringValue(destinationVpcId),
		DestinationIpRange: types.StringValue(http.DestinationIpRange),
		SourceIpRange:      types.StringValue(http.SourceIpRange),
		GatewayId:          types.StringValue(http.GatewayId),
		Id:                 types.StringValue(http.Id.String()),
	}
}
