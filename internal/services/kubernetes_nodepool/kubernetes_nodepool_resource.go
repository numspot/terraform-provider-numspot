package kubernetes_nodepool

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/kubernetes_nodepool/resource_kubernetes_nodepool"
	"terraform-provider-numspot/internal/utils"
)

var _ resource.Resource = (*kubernetesNodepoolResource)(nil)

type kubernetesNodepoolResource struct {
	provider *client.NumSpotSDK
}

func NewKubernetesNodepoolResource() resource.Resource {
	return &kubernetesNodepoolResource{}
}

func (r *kubernetesNodepoolResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *kubernetesNodepoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_nodepool"
}

func (r *kubernetesNodepoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_kubernetes_nodepool.KubernetesNodepoolResourceSchema(ctx)
}

func (r *kubernetesNodepoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_kubernetes_nodepool.KubernetesNodepoolModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUuid, err := uuid.Parse(plan.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to parse cluster id", err.Error())
		return
	}

	createNodePoolRequest := deserializeCreateNodePool(plan)

	nodePool, err := core.CreateKubernetesNodePool(ctx, r.provider, createNodePoolRequest, clusterUuid)
	if err != nil {
		resp.Diagnostics.AddError("unable to create kubernetes node pool", err.Error())
		return
	}

	state := serializeNodePool(ctx, nodePool, &resp.Diagnostics, plan)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func deserializeCreateNodePool(tf resource_kubernetes_nodepool.KubernetesNodepoolModel) api.CreateKubernetesNodePoolJSONRequestBody {
	var rootDisk api.KubernetesNodePoolDisk
	var autoscalling *api.Autoscaling

	if !(tf.RootDisk.IsNull() || tf.RootDisk.IsUnknown()) {
		rootDisk = api.KubernetesNodePoolDisk{
			Iops: utils.FromTfInt64ToInt(tf.RootDisk.Iops),
			Size: utils.FromTfInt64ToInt(tf.RootDisk.Size),
			Type: api.KubernetesNodePoolDiskType(tf.RootDisk.RootDiskType.ValueString()),
		}
	}

	if !(tf.Autoscaling.IsNull() || tf.Autoscaling.IsUnknown()) {
		autoscalling = &api.Autoscaling{
			Max: utils.FromTfInt64ToInt(tf.RootDisk.Iops),
			Min: utils.FromTfInt64ToInt(tf.RootDisk.Size),
		}
	}

	return api.CreateKubernetesNodePoolJSONRequestBody{
		AvailabilityZone: api.AvailabilityZoneName(tf.AvailabilityZone.ValueString()),
		Name:             tf.Name.ValueStringPointer(),
		NodeProfile:      api.NodeProfiles(tf.NodeProfile.ValueString()),
		Replicas:         utils.FromTfInt64ToIntPtr(tf.Replicas),
		Autoscaling:      autoscalling,
		RootDisk:         rootDisk,
	}
}

func (r *kubernetesNodepoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_kubernetes_nodepool.KubernetesNodepoolModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUuid, err := uuid.Parse(state.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to parse cluster id", err.Error())
		return
	}
	nodePoolId := state.Id.ValueString()

	nodePool, err := core.ReadKubernetesNodePool(ctx, r.provider, clusterUuid, nodePoolId)
	if err != nil {
		resp.Diagnostics.AddError("unable to read kubernetes node pool", err.Error())
		return
	}

	newState := serializeNodePool(ctx, nodePool, &resp.Diagnostics, state)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *kubernetesNodepoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No implementation at this time
}

func (r *kubernetesNodepoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_kubernetes_nodepool.KubernetesNodepoolModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUuid, err := uuid.Parse(state.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to parse cluster id", err.Error())
		return
	}
	nodePoolId := state.Id.ValueString()

	if err := core.DeleteKubernetesNodePool(ctx, r.provider, clusterUuid, nodePoolId); err != nil {
		resp.Diagnostics.AddError("unable to delete kubernetes node pool", err.Error())
		return
	}
}

func serializeNodePool(ctx context.Context, nodePool *api.CreateKubernetesNodePool201Response, diags *diag.Diagnostics, plan resource_kubernetes_nodepool.KubernetesNodepoolModel) resource_kubernetes_nodepool.KubernetesNodepoolModel {
	rootDisk, diagnostics := resource_kubernetes_nodepool.NewRootDiskValue(resource_kubernetes_nodepool.RootDiskValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"iops": utils.FromIntToTfInt64(nodePool.RootDisk.Iops),
			"size": utils.FromIntToTfInt64(nodePool.RootDisk.Size),
			"type": types.StringValue(string(nodePool.RootDisk.Type)),
		})

	if diagnostics.HasError() {
		diags.Append(diagnostics...)
		return resource_kubernetes_nodepool.KubernetesNodepoolModel{}
	}

	status, diagnostics := resource_kubernetes_nodepool.NewStatusValue(resource_kubernetes_nodepool.StatusValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"message": types.StringValue(nodePool.Status.Message),
			"state":   types.StringValue(string(nodePool.Status.State)),
		},
	)
	if diagnostics.HasError() {
		diags.Append(diagnostics...)
	}

	return resource_kubernetes_nodepool.KubernetesNodepoolModel{
		AvailabilityZone: types.StringValue(string(nodePool.AvailabilityZone)),
		ClusterId:        types.StringValue(plan.ClusterId.ValueString()),
		Id:               types.StringValue(nodePool.Id.String()),
		Name:             types.StringPointerValue(nodePool.Name),
		NodePoolId:       types.StringValue(plan.NodePoolId.ValueString()),
		NodeProfile:      types.StringValue(string(nodePool.NodeProfile)),
		Replicas:         types.Int64Value(int64(nodePool.Replicas)),
		RootDisk:         rootDisk,
		Status:           status,
	}
}
