package postgrescluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/postgrescluster/resource_postgres_cluster"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &postgresClusterResource{}
	_ resource.ResourceWithConfigure   = &postgresClusterResource{}
	_ resource.ResourceWithImportState = &postgresClusterResource{}
)

// postgresClusterResource is the Terraform resource implementation for numspot_postgres_cluster.
type postgresClusterResource struct {
	services.BaseResource
}

// NewPostgresClusterResource returns a new postgresClusterResource instance.
func NewPostgresClusterResource() resource.Resource {
	return &postgresClusterResource{}
}

// Metadata sets the resource type name.
func (r *postgresClusterResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_postgres_cluster"
}

// Schema defines the schema for the resource.
func (r *postgresClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_postgres_cluster.PostgresClusterResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *postgresClusterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_postgres_cluster.PostgresClusterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numSpotPostgresCluster, err := core.CreatePostgresCluster(
		ctx,
		r.Provider,
		deserializeCreatePostgresClusterRequest(ctx, &plan, &response.Diagnostics),
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create postgres_cluster", err.Error())
		return
	}

	state := serializePostgresCluster(ctx, numSpotPostgresCluster, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *postgresClusterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_postgres_cluster.PostgresClusterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotPostgresCluster, err := core.ReadPostgresCluster(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read postgres_cluster", err.Error())
		return
	}

	newState := serializePostgresCluster(ctx, numSpotPostgresCluster, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *postgresClusterResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *postgresClusterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_postgres_cluster.PostgresClusterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeletePostgresCluster(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete postgres_cluster", err.Error())
		return
	}
}

func serializePostgresCluster(
	ctx context.Context,
	http *api.PostgreSQLCluster,
	diags *diag.Diagnostics,
) resource_postgres_cluster.PostgresClusterModel {
	if http == nil {
		diags.AddError("nil PostgresCluster response", "the API returned an empty PostgresCluster")
		return resource_postgres_cluster.PostgresClusterModel{}
	}

	return resource_postgres_cluster.PostgresClusterModel{
		ClusterId:           types.StringValue(http.Id.String()),
		CreatedOn:           utils.TimeToTfString(http.CreatedOn),
		Extensions:          extensionsToTfList(ctx, http.Extensions, diags),
		FullVersion:         types.StringValue(http.FullVersion),
		Host:                types.StringPointerValue(http.Host),
		Id:                  types.StringValue(http.Id.String()),
		LastPasswordResetAt: utils.TimePtrToTfString(http.LastPasswordResetAt),
		MajorVersion:        types.StringValue(string(http.MajorVersion)),
		Name:                types.StringValue(http.Name),
		NodeConfiguration:   serializeNodeConfiguration(ctx, &http.NodeConfiguration, diags),
		Port:                types.Int64Value(int64(http.Port)),
		ReplicaCount:        types.Int64Value(int64(http.ReplicaCount)),
		Status:              serializeStatu(ctx, &http.Status, diags),
		User:                types.StringValue(http.User),
		Visibility:          types.StringValue(string(http.Visibility)),
		Volume:              serializeVolumeValue(ctx, http.Volume, diags),
	}
}

func deserializeCreatePostgresClusterRequest(
	ctx context.Context,
	tf *resource_postgres_cluster.PostgresClusterModel,
	diags *diag.Diagnostics,
) *api.PostgreSQLClusterCreationRequest {
	return &api.PostgreSQLClusterCreationRequest{
		Extensions:        extensionsToCreateBody(ctx, tf.Extensions, diags),
		MajorVersion:      utils.PointerOf(api.PostgreSQLMajorVersion(tf.MajorVersion.ValueString())),
		Name:              tf.Name.ValueString(),
		NodeConfiguration: deserializeNodeConfiguration(tf.NodeConfiguration),
		ReplicaCount:      utils.PointerOf(api.PostgreSQLReplicaCount(tf.ReplicaCount.ValueInt64())),
		User:              tf.User.ValueStringPointer(),
		Visibility:        api.PostgreSQLClusterVisibility(tf.Visibility.ValueString()),
		Volume:            deserializeVolumeUnion(tf.Volume),
	}
}

func serializeNodeConfiguration(
	ctx context.Context,
	item *api.PostgreSQLNodeConfiguration,
	diags *diag.Diagnostics,
) resource_postgres_cluster.NodeConfigurationValue {
	if item == nil {
		return resource_postgres_cluster.NewNodeConfigurationValueNull()
	}
	value, d := resource_postgres_cluster.NewNodeConfigurationValue(
		resource_postgres_cluster.NodeConfigurationValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"memory_size_gi_b": types.Int64Value(int64(item.MemorySizeGiB)),
			"vcpu_count":       types.Int64Value(int64(item.VcpuCount)),
		},
	)
	diags.Append(d...)
	return value
}

func deserializeNodeConfiguration(e resource_postgres_cluster.NodeConfigurationValue) api.PostgreSQLNodeConfiguration {
	return api.PostgreSQLNodeConfiguration{
		MemorySizeGiB: int32(e.MemorySizeGiB.ValueInt64()),
		VcpuCount:     int32(e.VcpuCount.ValueInt64()),
	}
}

func serializeStatu(ctx context.Context, item *api.PostgreSQLStatus, diags *diag.Diagnostics) resource_postgres_cluster.StatusValue {
	if item == nil {
		return resource_postgres_cluster.NewStatusValueNull()
	}
	value, d := resource_postgres_cluster.NewStatusValue(resource_postgres_cluster.StatusValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"message": types.StringValue(item.Message),
		"state":   types.StringValue(string(item.State)),
	})
	diags.Append(d...)
	return value
}
