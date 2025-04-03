package managedservicebridge

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
	"terraform-provider-numspot/internal/services/managedservicebridge/resource_managed_service_bridges"
)

var _ resource.Resource = &Resource{}

type Resource struct {
	provider *client.NumSpotSDK
}

func NewManagedServiceBridgeResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_service_bridge"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_managed_service_bridges.ManagedServiceBridgesResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_managed_service_bridges.ManagedServiceBridgesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceId := plan.SourceManagedServiceId.ValueString()
	destId := plan.DestinationManagedServiceId.ValueString()
	body := deserializeServiceManagesBridge(plan)

	numSpot, err := core.CreateManagedServiceBridge(ctx, r.provider, body)
	if err != nil {
		resp.Diagnostics.AddError("unable to create managed service bridge", err.Error())
		return
	}

	data := serializeServiceManagedBridge(numSpot, sourceId, destId)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan resource_managed_service_bridges.ManagedServiceBridgesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceId := plan.SourceManagedServiceId.ValueString()
	destId := plan.DestinationManagedServiceId.ValueString()
	id, err := uuid.Parse(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to read managed service bridge", err.Error())
		return
	}

	serviceManagedBridge, err := core.ReadManagedServiceBridge(ctx, r.provider, id)
	if err != nil {
		resp.Diagnostics.AddError("unable to read hybrid bridge", err.Error())
		return
	}

	newPlan := serializeServiceManagedBridge(serviceManagedBridge, sourceId, destId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newPlan)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan resource_managed_service_bridges.ManagedServiceBridgesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to parse uuid managed service bridge", err.Error())
		return
	}

	if err := core.DeleteManagedServiceBridge(ctx, r.provider, id); err != nil {
		resp.Diagnostics.AddError("unable to delete managed service bridge", err.Error())
		return
	}
}

func deserializeServiceManagesBridge(tf resource_managed_service_bridges.ManagedServiceBridgesModel) api.CreateManagedServicesBridgeRequest {
	return api.CreateManagedServicesBridgeRequest{
		DestinationManagedServiceId: uuid.MustParse(tf.DestinationManagedServiceId.ValueString()),
		SourceManagedServiceId:      uuid.MustParse(tf.SourceManagedServiceId.ValueString()),
	}
}

func serializeServiceManagedBridge(http *api.ManagedServicesBridge, sourceId, destId string) resource_managed_service_bridges.ManagedServiceBridgesModel {
	return resource_managed_service_bridges.ManagedServiceBridgesModel{
		DestinationManagedServiceId: types.StringValue(destId),
		SourceManagedServiceId:      types.StringValue(sourceId),
		Id:                          types.StringValue(http.Id.String()),
	}
}
