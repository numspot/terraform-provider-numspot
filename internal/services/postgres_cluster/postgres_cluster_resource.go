package postgres_cluster

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
	"terraform-provider-numspot/internal/services/postgres_cluster/resource_postgres_cluster"
)

var _ resource.Resource = &postgresClusterResource{}

func (r *postgresClusterResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func NewPostgresClusterResource() resource.Resource {
	return &postgresClusterResource{}
}

type postgresClusterResource struct {
	provider *client.NumSpotSDK
}

func (r *postgresClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres_cluster"
}

func (r *postgresClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_postgres_cluster.PostgresClusterResourceSchema(ctx)
}

func (r *postgresClusterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_postgres_cluster.PostgresClusterModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics
	body := deserializeCreatePostgresCluster(plan, &diags)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	res, err := core.CreatePostgresCluster(ctx, r.provider, *body)
	if err != nil {
		response.Diagnostics.AddError("unable to create postgres cluster", err.Error())
		return
	}

	state := serializePostgresCluster(ctx, res, &response.Diagnostics, plan)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *postgresClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan resource_postgres_cluster.PostgresClusterModel

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := uuid.MustParse(plan.Id.String())

	res, err := core.ReadPostgresCluster(ctx, r.provider, clusterId)
	if err != nil {
		resp.Diagnostics.AddError("unable to read postgres cluster", err.Error())
		return
	}

	state := serializePostgresCluster(ctx, res, &resp.Diagnostics, plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *postgresClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *postgresClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan resource_postgres_cluster.PostgresClusterModel

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := uuid.MustParse(plan.Id.String())

	if err := core.DeletePostgresCluster(ctx, r.provider, clusterId); err != nil {
		resp.Diagnostics.AddError("unable to delete postgres cluster", err.Error())
		return
	}
}

func deserializeCreatePostgresCluster(tf resource_postgres_cluster.PostgresClusterModel, diag *diag.Diagnostics) *api.PostgresClusterCreationRequest {
	var postgresExtensionsMappingApi []api.PostgresExtension
	var replicaCount *api.PostgresReplicaCount
	var nodeConfiguration api.PostgresNodeConfiguration
	var volume api.PostgresVolume
	var majorVersion *api.PostgresMajorVersion

	if !(tf.Extensions.IsNull() || tf.Extensions.IsUnknown()) {
		postgresExtensionsMappingApi = make([]api.PostgresExtension, 0, len(tf.Extensions.Elements()))
		for _, pemTf := range tf.Extensions.Elements() {
			pemTfRes, ok := pemTf.(resource_postgres_cluster.ExtensionsValue)
			if !ok {
				diag.AddError("unable to cast extension value to postgres resource", "")
				return nil
			}

			pemApi := deserializeExtensionsMapping(pemTfRes)
			postgresExtensionsMappingApi = append(postgresExtensionsMappingApi, pemApi)
		}
	}

	if !(tf.ReplicaCount.IsNull() || tf.ReplicaCount.IsUnknown()) {
		rc := api.PostgresReplicaCount(tf.ReplicaCount.ValueInt64())
		replicaCount = &rc
	}

	if !(tf.NodeConfiguration.IsNull() || tf.NodeConfiguration.IsUnknown()) {
		nodeConfiguration = api.PostgresNodeConfiguration{
			MemorySizeGiB: int32(tf.NodeConfiguration.MemorySizeGiB.ValueInt64()),
			VcpuCount:     int32(tf.NodeConfiguration.VcpuCount.ValueInt64()),
		}
	}

	if !(tf.MajorVersion.IsNull() || tf.MajorVersion.IsUnknown()) {
		v := api.PostgresMajorVersion(tf.MajorVersion.ValueString())
		majorVersion = &v
	} else {
		defaultVersion := types.StringValue("18")
		v := api.PostgresMajorVersion(defaultVersion.ValueString())
		majorVersion = &v
	}

	if !(tf.Volume.IsNull() || tf.Volume.IsUnknown()) {
		volumeType := tf.Volume.VolumeType.ValueString()
		switch volumeType {
		case "GP2", "IO1", "STANDARD":
			volume = api.PostgresVolume{
				Type:    api.PostgresVolumeType(volumeType),
				SizeGiB: int32(tf.Volume.SizeGiB.ValueInt64()),
			}

		default:
			diag.AddError(
				"Unsupported volume type",
				"Type: "+volumeType+" is not recognized.",
			)
		}
	}

	return &api.PostgresClusterCreationRequest{
		Extensions:        &postgresExtensionsMappingApi,
		ReplicaCount:      replicaCount,
		Visibility:        api.PostgresClusterVisibility(tf.Visibility.ValueString()),
		Name:              tf.Name.ValueString(),
		NodeConfiguration: nodeConfiguration,
		MajorVersion:      majorVersion,
		User:              tf.User.ValueString(),
		Volume:            volume,
	}
}

func deserializeExtensionsMapping(ext resource_postgres_cluster.ExtensionsValue) api.PostgresExtension {
	return api.PostgresExtension{
		Name: api.PostgresExtensionName(ext.Name.ValueString()),
	}
}

func serializePostgresCluster(ctx context.Context, cluster *api.PostgresCluster, diags *diag.Diagnostics, plan resource_postgres_cluster.PostgresClusterModel) resource_postgres_cluster.PostgresClusterModel {
	var serializeDiags diag.Diagnostics
	var volume resource_postgres_cluster.VolumeValue
	extensionList := types.List{}

	nodeConfiguration, mappingDiags := resource_postgres_cluster.NewNodeConfigurationValue(resource_postgres_cluster.NodeConfigurationValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"memory_size_gi_b": types.Int64Value(int64(cluster.NodeConfiguration.MemorySizeGiB)),
		"vcpu_count":       types.Int64Value(int64(cluster.NodeConfiguration.VcpuCount)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	volume, mappingDiags = resource_postgres_cluster.NewVolumeValue(resource_postgres_cluster.VolumeValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"size_gi_b": types.Int64Value(int64(cluster.Volume.SizeGiB)),
		"type":      types.StringValue(string(cluster.Volume.Type)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	var port types.Int64
	if cluster.Port != nil {
		port = types.Int64Value(int64(*cluster.Port))
	} else {
		port = types.Int64Null()
	}

	var majorVersion types.String

	if cluster.MajorVersion == nil {
		majorVersion = types.StringNull()
	} else {
		majorVersion = types.StringValue(string(*cluster.MajorVersion))
	}

	if cluster.Extensions != nil {
		extensionLen := len(*cluster.Extensions)
		elementValue := make([]resource_postgres_cluster.ExtensionsValue, extensionLen)

		for i, ext := range *cluster.Extensions {
			elementValue[i], serializeDiags = resource_postgres_cluster.NewExtensionsValue(resource_postgres_cluster.ExtensionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"name": types.StringValue(string(ext.Name)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		extensionList, serializeDiags = types.ListValueFrom(ctx, new(resource_postgres_cluster.ExtensionsValue).Type(ctx), elementValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
			return resource_postgres_cluster.PostgresClusterModel{}
		}
	}

	status, diagnostics := resource_postgres_cluster.NewStatusValue(resource_postgres_cluster.StatusValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"message": types.StringValue(cluster.Status.Message),
			"state":   types.StringValue(string(cluster.Status.State)),
		},
	)
	if diagnostics.HasError() {
		diags.Append(diagnostics...)
	}

	return resource_postgres_cluster.PostgresClusterModel{
		Visibility:        types.StringValue(plan.Visibility.ValueString()),
		CreatedOn:         types.StringValue(cluster.CreatedOn.Format(time.RFC3339)),
		Extensions:        extensionList,
		ReplicaCount:      types.Int64Value(int64(cluster.ReplicaCount)),
		Host:              types.StringPointerValue(cluster.Host),
		Id:                types.StringValue(cluster.Id.String()),
		Name:              types.StringValue(cluster.Name),
		NodeConfiguration: nodeConfiguration,
		MajorVersion:      majorVersion,
		Port:              port,
		Status:            status,
		User:              types.StringValue(cluster.User),
		Volume:            volume,
		FullVersion:       types.StringPointerValue(cluster.FullVersion),
	}
}
