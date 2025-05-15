package postgres_cluster

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
	"terraform-provider-numspot/internal/services/postgres_cluster/resource_postgres_cluster"
	"terraform-provider-numspot/internal/utils"
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
	body := deserializeCreatePostgresCluster(ctx, plan, &diags)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	res, err := core.CreatePostgresCluster(ctx, r.provider, body)
	if err != nil {
		response.Diagnostics.AddError("unable to create postgres cluster", err.Error())
		return
	}

	state := serializePostgresCluster(ctx, res, &response.Diagnostics)
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

	state := serializePostgresCluster(ctx, res, &resp.Diagnostics)

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
	res, err := core.DeletePostgresCluster(ctx, r.provider, clusterId, api.PostgreSQLDeleteClusterJSONRequestBody{})
	if err != nil {
		resp.Diagnostics.AddError("unable to delete postgres cluster", err.Error())
		return
	}

	state := serializePostgresCluster(ctx, res, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func deserializeCreatePostgresCluster(ctx context.Context, tf resource_postgres_cluster.PostgresClusterModel, diags *diag.Diagnostics) api.PostgresClusterCreationRequestWithVolume {
	var nodeConfiguration api.PostgresNodeConfiguration
	var volume api.PostgresAllVolumes
	var backupId *api.PostgresBackupId
	var tags *[]api.PostgresTag

	if !(tf.NodeConfiguration.IsNull() || tf.NodeConfiguration.IsUnknown()) {
		nodeConfiguration = api.PostgresNodeConfiguration{
			MemorySizeGiB:    utils.FromTfInt64ToInt(tf.NodeConfiguration.MemorySizeGiB),
			PerformanceLevel: api.PostgresNodeConfigurationPerformanceLevel(tf.NodeConfiguration.PerformanceLevel.ValueString()),
			VcpuCount:        utils.FromTfInt64ToInt(tf.NodeConfiguration.VcpuCount),
		}
	}

	if !(tf.SourceBackupId.IsNull() || tf.SourceBackupId.IsUnknown()) {
		id := uuid.MustParse(tf.SourceBackupId.ValueString())
		backupId = &id
	}

	if !(tf.Tags.IsNull() || tf.Tags.IsUnknown()) {
		tagList := make([]api.PostgresTag, 0, len(tf.Tags.Elements()))
		diags.Append(tf.Tags.ElementsAs(ctx, &tagList, true)...)
		tags = &tagList
	}

	if !(tf.Volume.IsNull() || tf.Volume.IsUnknown()) {
		volumeType := tf.Volume.VolumeType.ValueString()
		switch volumeType {
		case "GP2":
			volume = api.PostgresAllVolumes{
				Type:    volumeType,
				SizeGiB: utils.FromTfInt64ToInt(tf.Volume.SizeGiB),
			}

		case "IO1":
			volume = api.PostgresAllVolumes{
				Type:    volumeType,
				SizeGiB: utils.FromTfInt64ToInt(tf.Volume.SizeGiB),
				Iops:    utils.FromTfInt64ToIntPtr(tf.Volume.Iops),
			}

		default:
			diags.AddError("Unsupported volume type", "Type: "+volumeType+" is not recognized.")
		}
	}

	return api.PostgresClusterCreationRequestWithVolume{
		AllowedIpRanges:   utils.TfStringListToStringList(ctx, tf.AllowedIpRanges, diags),
		AutomaticBackup:   utils.FromTfBoolToBoolPtr(tf.AutomaticBackup),
		IsPublic:          utils.FromTfBoolToBoolPtr(tf.IsPublic),
		Name:              tf.Name.ValueString(),
		NetCidr:           tf.VpcCidr.ValueStringPointer(),
		NodeConfiguration: nodeConfiguration,
		SourceBackupId:    backupId,
		Tags:              tags,
		User:              tf.User.ValueString(),
		Volume:            volume,
	}
}

func serializePostgresCluster(ctx context.Context, cluster *api.PostgresCluster, diags *diag.Diagnostics) resource_postgres_cluster.PostgresClusterModel {
	var serializeDiags diag.Diagnostics

	allowedIpRangesList := types.ListNull(types.String{}.Type(ctx))
	availableOperationsList := types.ListNull(types.String{}.Type(ctx))
	tagsList := types.List{}
	maintenanceSchedule := resource_postgres_cluster.MaintenanceScheduleValue{}
	volume := resource_postgres_cluster.VolumeValue{}

	if cluster.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, cluster.Tags, mappingResourceTags)
		if serializeDiags.HasError() {
			return resource_postgres_cluster.PostgresClusterModel{}
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return resource_postgres_cluster.PostgresClusterModel{}
		}
	}

	if cluster.AllowedIpRanges != nil {
		allowedIpRangesList, serializeDiags = types.ListValueFrom(ctx, types.StringType, cluster.AllowedIpRanges)
		diags.Append(serializeDiags...)
	}

	if cluster.AvailableOperations != nil {
		availableOperationsList, serializeDiags = types.ListValueFrom(ctx, types.StringType, cluster.AvailableOperations)
		diags.Append(serializeDiags...)
	}

	if cluster.MaintenanceSchedule != nil {
		maintenanceSchedule = resource_postgres_cluster.MaintenanceScheduleValue{
			BeginAt:                 types.StringValue(cluster.MaintenanceSchedule.BeginAt),
			EndAt:                   types.StringValue(cluster.MaintenanceSchedule.EndAt),
			PotentialImpact:         types.StringValue(cluster.MaintenanceSchedule.PotentialImpact),
			MaintenanceScheduleType: types.StringValue(string(cluster.MaintenanceSchedule.Type)),
		}
	}

	nodeConfiguration, mappingDiags := resource_postgres_cluster.NewNodeConfigurationValue(resource_postgres_cluster.NodeConfigurationValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"memory_size_gi_b":  types.Int64Value(utils.ConvertIntPtrToInt64(&cluster.NodeConfiguration.MemorySizeGiB)),
		"performance_level": types.StringValue(string(cluster.NodeConfiguration.PerformanceLevel)),
		"vcpu_count":        types.Int64Value(utils.ConvertIntPtrToInt64(&cluster.NodeConfiguration.VcpuCount)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	valueD, err := cluster.Volume.ValueByDiscriminator()
	if err != nil {
		return resource_postgres_cluster.PostgresClusterModel{}
	}

	switch v := valueD.(type) {
	case api.PostgresVolumeGp2:
		volume, mappingDiags = resource_postgres_cluster.NewVolumeValue(resource_postgres_cluster.VolumeValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"size_gi_b": types.Int64Value(utils.ConvertIntPtrToInt64(&v.SizeGiB)),
			"type":      types.StringValue(string(v.Type)),
		})
		if mappingDiags.HasError() {
			diags.Append(mappingDiags...)
		}
	case api.PostgresVolumeIo1:
		volume, mappingDiags = resource_postgres_cluster.NewVolumeValue(resource_postgres_cluster.VolumeValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"size_gi_b": types.Int64Value(utils.ConvertIntPtrToInt64(&v.SizeGiB)),
			"iops":      types.Int64Value(utils.ConvertIntPtrToInt64(&v.Iops)),
			"type":      types.StringValue(string(v.Type)),
		})
		if mappingDiags.HasError() {
			diags.AddError("Unsupported volume type", "Type: "+string(v.Type)+" is not recognized.")
		}
	default:
		diags.Append(mappingDiags...)
	}

	return resource_postgres_cluster.PostgresClusterModel{
		AllowedIpRanges:     allowedIpRangesList,
		AutomaticBackup:     types.BoolValue(cluster.AutomaticBackup),
		AvailableOperations: availableOperationsList,
		CreatedOn:           types.StringValue(cluster.CreatedOn),
		ErrorReason:         types.StringPointerValue(cluster.ErrorReason),
		Host:                types.StringPointerValue(cluster.Host),
		Id:                  types.StringValue(cluster.Id.String()),
		LastOperationName:   types.StringValue(string(cluster.LastOperationName)),
		LastOperationResult: types.StringValue(string(cluster.LastOperationResult)),
		MaintenanceSchedule: maintenanceSchedule,
		Name:                types.StringValue(cluster.Name),
		NodeConfiguration:   nodeConfiguration,
		Port:                types.Int64Value(utils.ConvertIntPtrToInt64(cluster.Port)),
		PrivateHost:         types.StringPointerValue(cluster.PrivateHost),
		Status:              types.StringValue(string(cluster.Status)),
		Tags:                tagsList,
		User:                types.StringValue(cluster.User),
		Volume:              volume,
		VpcCidr:             types.StringPointerValue(cluster.NetCidr),
	}
}

func mappingResourceTags(ctx context.Context, tag api.PostgresTag) (resource_postgres_cluster.TagsValue, diag.Diagnostics) {
	return resource_postgres_cluster.NewTagsValue(resource_postgres_cluster.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}
