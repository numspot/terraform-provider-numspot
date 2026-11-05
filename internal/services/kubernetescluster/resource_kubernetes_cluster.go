package kubernetescluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/kubernetescluster/resource_kubernetes_cluster"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &kubernetesClusterResource{}
	_ resource.ResourceWithConfigure   = &kubernetesClusterResource{}
	_ resource.ResourceWithImportState = &kubernetesClusterResource{}
)

// kubernetesClusterResource is the Terraform resource implementation for numspot_kubernetes_cluster.
type kubernetesClusterResource struct {
	services.BaseResource
}

// NewKubernetesClusterResource returns a new kubernetesClusterResource instance.
func NewKubernetesClusterResource() resource.Resource {
	return &kubernetesClusterResource{}
}

// Metadata sets the resource type name.
func (r *kubernetesClusterResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_kubernetes_cluster"
}

// Schema defines the schema for the resource.
func (r *kubernetesClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_kubernetes_cluster.KubernetesClusterResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *kubernetesClusterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_kubernetes_cluster.KubernetesClusterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numspotKubernetesCluster, err := core.CreateKubernetesCluster(
		ctx,
		r.Provider,
		deserializeCreateKubernetesClusterRequest(&plan),
		createTimeout(&plan),
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create kubernetes_cluster", err.Error())
		return
	}

	state := serializeKubernetesCluster(ctx, numspotKubernetesCluster, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	state.Timeouts = utils.PreserveFromState(plan.Timeouts)
	state.AirGap = utils.PreserveFromState(plan.AirGap)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *kubernetesClusterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_kubernetes_cluster.KubernetesClusterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotKubernetesCluster, err := core.ReadKubernetesCluster(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read kubernetes_cluster", err.Error())
		return
	}

	newState := serializeKubernetesCluster(ctx, numspotKubernetesCluster, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.Timeouts = state.Timeouts
	newState.AirGap = state.AirGap

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *kubernetesClusterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resource_kubernetes_cluster.KubernetesClusterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotKubernetesCluster, err := core.ReadKubernetesCluster(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read kubernetes_cluster", err.Error())
		return
	}

	newState := serializeKubernetesCluster(ctx, numspotKubernetesCluster, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.Timeouts = utils.PreserveFromState(plan.Timeouts)
	newState.AirGap = utils.PreserveFromState(plan.AirGap)

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *kubernetesClusterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_kubernetes_cluster.KubernetesClusterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteKubernetesCluster(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete kubernetes_cluster", err.Error())
		return
	}
}

func serializeKubernetesCluster(
	ctx context.Context,
	http *api.KubernetesCluster,
	diags *diag.Diagnostics,
) resource_kubernetes_cluster.KubernetesClusterModel {
	if http == nil {
		diags.AddError("nil KubernetesCluster response", "the API returned an empty KubernetesCluster")
		return resource_kubernetes_cluster.KubernetesClusterModel{}
	}

	return resource_kubernetes_cluster.KubernetesClusterModel{
		Cidr:          types.StringValue(http.Cidr),
		ClusterId:     types.StringValue(http.Id.String()),
		CreatedOn:     utils.TimePtrToTfString(http.CreatedOn),
		DedicatedEtcd: types.BoolValue(http.DedicatedEtcd),
		FullVersion:   types.StringValue(http.FullVersion),
		Id:            types.StringValue(http.Id.String()),
		Name:          types.StringValue(http.Name),
		Profile:       types.StringValue(http.Profile),
		Status:        serializeStatu(ctx, &http.Status, diags),
		Version:       types.StringValue(http.Version),
		Visibility:    types.StringValue(string(http.Visibility)),
	}
}

func deserializeCreateKubernetesClusterRequest(tf *resource_kubernetes_cluster.KubernetesClusterModel) *api.CreateKubernetesClusterJSONRequestBody {
	return &api.CreateKubernetesClusterJSONRequestBody{
		AirGap:        tf.AirGap.ValueBoolPointer(),
		Cidr:          tf.Cidr.ValueString(),
		DedicatedEtcd: tf.DedicatedEtcd.ValueBoolPointer(),
		Name:          tf.Name.ValueString(),
		Profile:       api.KubernetesClusterCreateProfile(tf.Profile.ValueString()),
		Version:       api.KubernetesClusterCreateVersion(tf.Version.ValueString()),
		Visibility:    api.KubernetesVisibility(tf.Visibility.ValueString()),
	}
}

func serializeStatu(ctx context.Context, item *api.KubernetesStatus, diags *diag.Diagnostics) resource_kubernetes_cluster.StatusValue {
	if item == nil {
		return resource_kubernetes_cluster.NewStatusValueNull()
	}
	value, d := resource_kubernetes_cluster.NewStatusValue(resource_kubernetes_cluster.StatusValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"message": types.StringValue(item.Message),
		"reason":  types.StringPointerValue((*string)(item.Reason)),
		"state":   types.StringValue(string(item.State)),
	})
	diags.Append(d...)
	return value
}
