package kubernetesnodepool

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/kubernetesnodepool/resource_kubernetes_nodepool"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &kubernetesNodepoolResource{}
	_ resource.ResourceWithConfigure   = &kubernetesNodepoolResource{}
	_ resource.ResourceWithImportState = &kubernetesNodepoolResource{}
)

// kubernetesNodepoolResource is the Terraform resource implementation for numspot_kubernetes_nodepool.
type kubernetesNodepoolResource struct {
	services.BaseResource
}

// NewKubernetesNodepoolResource returns a new kubernetesNodepoolResource instance.
func NewKubernetesNodepoolResource() resource.Resource {
	return &kubernetesNodepoolResource{}
}

// Metadata sets the resource type name.
func (r *kubernetesNodepoolResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_kubernetes_nodepool"
}

// Schema defines the schema for the resource.
func (r *kubernetesNodepoolResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_kubernetes_nodepool.KubernetesNodepoolResourceSchema(ctx)
}

// ImportState imports the resource state.
func (r *kubernetesNodepoolResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, "/")
	if len(parts) != 2 {
		response.Diagnostics.AddError("invalid import ID format", "expected format: cluster_id/id")
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &resource_kubernetes_nodepool.KubernetesNodepoolModel{
		ClusterId: types.StringValue(parts[0]),
		Id:        types.StringValue(parts[1]),
	})...)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *kubernetesNodepoolResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_kubernetes_nodepool.KubernetesNodepoolModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numspotKubernetesNodepool, err := core.CreateKubernetesNodePool(
		ctx,
		r.Provider,
		deserializeCreateKubernetesNodePoolRequest(&plan),
		plan.ClusterId.ValueString(),
		createTimeout(&plan),
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create kubernetes_nodepool", err.Error())
		return
	}

	state := serializeKubernetesNodePool(ctx, numspotKubernetesNodepool, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	state.Timeouts = utils.PreserveFromState(plan.Timeouts)
	state.ClusterId = utils.PreserveFromState(plan.ClusterId)
	state.Autoscaling = utils.PreserveFromState(plan.Autoscaling)
	state.Replicas = utils.PreserveFromState(plan.Replicas)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *kubernetesNodepoolResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_kubernetes_nodepool.KubernetesNodepoolModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotKubernetesNodepool, err := core.ReadKubernetesNodePool(ctx, r.Provider, state.ClusterId.ValueString(), state.Id.ValueString())
	if err != nil {
		if utils.IsResourceGone(err) {
			tflog.Warn(ctx, "unable to read kubernetes_nodepool (not found or access forbidden); removing from state")
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("unable to read kubernetes_nodepool", err.Error())
		return
	}

	newState := serializeKubernetesNodePool(ctx, numspotKubernetesNodepool, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.Timeouts = state.Timeouts
	newState.ClusterId = state.ClusterId
	newState.Autoscaling = state.Autoscaling
	newState.Replicas = state.Replicas

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *kubernetesNodepoolResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resource_kubernetes_nodepool.KubernetesNodepoolModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotKubernetesNodepool, err := core.ReadKubernetesNodePool(ctx, r.Provider, state.ClusterId.ValueString(), state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read kubernetes_nodepool", err.Error())
		return
	}

	newState := serializeKubernetesNodePool(ctx, numspotKubernetesNodepool, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.Timeouts = utils.PreserveFromState(plan.Timeouts)
	newState.ClusterId = utils.PreserveFromState(plan.ClusterId)
	newState.Autoscaling = utils.PreserveFromState(plan.Autoscaling)
	newState.Replicas = utils.PreserveFromState(plan.Replicas)

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *kubernetesNodepoolResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_kubernetes_nodepool.KubernetesNodepoolModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteKubernetesNodePool(ctx, r.Provider, state.ClusterId.ValueString(), state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete kubernetes_nodepool", err.Error())
		return
	}
}

func serializeKubernetesNodePool(
	ctx context.Context,
	http *api.KubernetesNodePool,
	diags *diag.Diagnostics,
) resource_kubernetes_nodepool.KubernetesNodepoolModel {
	if http == nil {
		diags.AddError("nil KubernetesNodepool response", "the API returned an empty KubernetesNodepool")
		return resource_kubernetes_nodepool.KubernetesNodepoolModel{}
	}

	return resource_kubernetes_nodepool.KubernetesNodepoolModel{
		AutoscalingMax:    utils.FromIntPtrToTfInt64(http.AutoscalingMax),
		AutoscalingMin:    utils.FromIntPtrToTfInt64(http.AutoscalingMin),
		AvailabilityZone:  types.StringValue(string(http.AvailabilityZone)),
		Id:                types.StringValue(http.Id.String()),
		Name:              types.StringValue(http.Name),
		NodePoolId:        types.StringValue(http.Id.String()),
		NodeProfile:       types.StringValue(string(http.NodeProfile)),
		ReplicasAvailable: utils.FromIntToTfInt64(http.ReplicasAvailable),
		RootDisk:          serializeRootDisk(ctx, &http.RootDisk, diags),
		Status:            serializeStatu(ctx, &http.Status, diags),
	}
}

func deserializeCreateKubernetesNodePoolRequest(
	tf *resource_kubernetes_nodepool.KubernetesNodepoolModel,
) *api.CreateKubernetesNodePoolJSONRequestBody {
	return &api.CreateKubernetesNodePoolJSONRequestBody{
		Autoscaling:      deserializeAutoscaling(tf.Autoscaling),
		AvailabilityZone: api.AvailabilityZoneName(tf.AvailabilityZone.ValueString()),
		Name:             tf.Name.ValueString(),
		NodeProfile:      api.NodeProfiles(tf.NodeProfile.ValueString()),
		Replicas:         utils.FromTfInt64ToIntPtr(tf.Replicas),
		RootDisk:         deserializeRootDisk(tf.RootDisk),
	}
}

func serializeRootDisk(ctx context.Context, item *api.KubernetesNodePoolDisk, diags *diag.Diagnostics) resource_kubernetes_nodepool.RootDiskValue {
	if item == nil {
		return resource_kubernetes_nodepool.NewRootDiskValueNull()
	}
	value, d := resource_kubernetes_nodepool.NewRootDiskValue(resource_kubernetes_nodepool.RootDiskValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"iops": utils.FromIntToTfInt64(item.Iops),
		"size": utils.FromIntToTfInt64(item.Size),
		"type": types.StringValue(string(item.Type)),
	})
	diags.Append(d...)
	return value
}

func deserializeRootDisk(e resource_kubernetes_nodepool.RootDiskValue) api.KubernetesNodePoolDisk {
	return api.KubernetesNodePoolDisk{
		Iops: utils.FromTfInt64ToInt(e.Iops),
		Size: utils.FromTfInt64ToInt(e.Size),
		Type: api.KubernetesNodePoolDiskType(e.RootDiskType.ValueString()),
	}
}

func serializeStatu(ctx context.Context, item *api.KubernetesStatus, diags *diag.Diagnostics) resource_kubernetes_nodepool.StatusValue {
	if item == nil {
		return resource_kubernetes_nodepool.NewStatusValueNull()
	}
	value, d := resource_kubernetes_nodepool.NewStatusValue(resource_kubernetes_nodepool.StatusValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"message": types.StringValue(item.Message),
		"reason":  types.StringPointerValue((*string)(item.Reason)),
		"state":   types.StringValue(string(item.State)),
	})
	diags.Append(d...)
	return value
}

func deserializeAutoscaling(e resource_kubernetes_nodepool.AutoscalingValue) *api.KubernetesAutoscaling {
	if e.IsNull() || e.IsUnknown() {
		return nil
	}
	return &api.KubernetesAutoscaling{
		Max: utils.FromTfInt64ToInt(e.Max),
		Min: utils.FromTfInt64ToInt(e.Min),
	}
}
