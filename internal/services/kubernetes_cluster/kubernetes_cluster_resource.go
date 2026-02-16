package kubernetes_cluster

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/kubernetes_cluster/resource_kubernetes_cluster"
)

var _ resource.Resource = &kubernetesClusterResource{}

type kubernetesClusterResource struct {
	provider *client.NumSpotSDK
}

func NewKubernetesClusterResource() resource.Resource {
	return &kubernetesClusterResource{}
}

func (r *kubernetesClusterResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *kubernetesClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}

func (r *kubernetesClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_kubernetes_cluster.KubernetesClusterResourceSchema(ctx)
}

func (r *kubernetesClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_kubernetes_cluster.KubernetesClusterModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createClusterRequest := deserializeCreateCluster(plan)

	cluster, err := core.CreateKubernetesCluster(ctx, r.provider, createClusterRequest)
	if err != nil {
		resp.Diagnostics.AddError("unable to create kubernetes cluster", err.Error())
		return
	}

	state := serializeCluster(ctx, cluster, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func deserializeCreateCluster(tf resource_kubernetes_cluster.KubernetesClusterModel) api.CreateKubernetesClusterJSONRequestBody {
	return api.CreateKubernetesClusterJSONRequestBody{
		Cidr:       tf.Cidr.ValueString(),
		Name:       tf.Name.ValueString(),
		Profile:    api.KubernetesClusterCreateProfile(tf.Profile.ValueString()),
		Version:    api.KubernetesClusterCreateVersion(tf.Version.ValueString()),
		Visibility: api.Visibility(tf.Visibility.ValueString()),
	}
}

func (r *kubernetesClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_kubernetes_cluster.KubernetesClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUuid, err := uuid.Parse(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to parse cluster id", err.Error())
		return
	}

	numSpotCluster, err := core.ReadKubernetesCluster(ctx, r.provider, clusterUuid)
	if err != nil {
		resp.Diagnostics.AddError("unable to read kubernetes cluster", err.Error())
		return
	}

	newState := serializeCluster(ctx, numSpotCluster, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *kubernetesClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No implementation at this time
}

func (r *kubernetesClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_kubernetes_cluster.KubernetesClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUuid, err := uuid.Parse(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to parse cluster id", err.Error())
		return
	}

	if err := core.DeleteKubernetesCluster(ctx, r.provider, clusterUuid); err != nil {
		resp.Diagnostics.AddError("unable to delete kubernetes cluster", err.Error())
		return
	}
}

func serializeCluster(ctx context.Context, cluster *api.GetKubernetesCluster200Response, diags *diag.Diagnostics) resource_kubernetes_cluster.KubernetesClusterModel {
	var createdOn types.String

	status, diagnostics := resource_kubernetes_cluster.NewStatusValue(resource_kubernetes_cluster.StatusValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"message": types.StringValue(cluster.Status.Message),
			"state":   types.StringValue(string(cluster.Status.State)),
		},
	)
	if diagnostics.HasError() {
		diags.Append(diagnostics...)
	}

	if cluster.CreatedOn != nil {
		createdOn = types.StringValue(cluster.CreatedOn.Format(time.RFC3339))
	} else {
		createdOn = types.StringNull()
	}

	return resource_kubernetes_cluster.KubernetesClusterModel{
		Cidr:        types.StringValue(cluster.Cidr),
		ClusterId:   types.StringValue(cluster.Id.String()),
		CreatedOn:   createdOn,
		Id:          types.StringValue(cluster.Id.String()),
		FullVersion: types.StringValue(cluster.FullVersion),
		Name:        types.StringValue(cluster.Name),
		Profile:     types.StringValue(cluster.Profile),
		Status:      status,
		Version:     types.StringValue(cluster.Version),
		Visibility:  types.StringValue(string(cluster.Visibility)),
	}
}
