package postgres_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/postgres_cluster/datasource_postgres_cluster"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &postgresClusterDataSource{}

func (d *postgresClusterDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

func NewPostgresClusterDataSource() datasource.DataSource {
	return &postgresClusterDataSource{}
}

type postgresClusterDataSource struct {
	provider *client.NumSpotSDK
}

func (d *postgresClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres_clusters"
}

func (d *postgresClusterDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_postgres_cluster.PostgresClusterDataSourceSchema(ctx)
}

func (d *postgresClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state, plan datasource_postgres_cluster.PostgresClusterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusters, err := core.ReadPostgresClusters(ctx, d.provider)
	if err != nil {
		resp.Diagnostics.AddError("unable to read postgres clusters", err.Error())
		return
	}

	clusterItems := utils.SerializeDatasourceItemsWithDiags(ctx, clusters.Items, &resp.Diagnostics, mappingItemsValue)
	if resp.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, clusterItems, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func mappingItemsValue(ctx context.Context, cluster api.PostgresCluster, diags *diag.Diagnostics) (datasource_postgres_cluster.ItemsValue, diag.Diagnostics) {
	var serializeDiags diag.Diagnostics
	var maintenanceSchedule basetypes.ObjectValue
	var nodeConfiguration basetypes.ObjectValue
	var volume basetypes.ObjectValue

	tagsList := types.ListNull(datasource_postgres_cluster.ItemsValue{}.Type(ctx))
	allowedIpRangesList := types.ListNull(types.String{}.Type(ctx))
	availableOperationsList := types.ListNull(types.String{}.Type(ctx))

	if cluster.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, cluster.Tags, mappingDatasourceTags)
		if serializeDiags.HasError() {
			return datasource_postgres_cluster.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_postgres_cluster.ItemsValue{}, serializeDiags
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
		maintenanceScheduleValue, serializePlacementDiags := mappingMaintenanceSchedule(ctx, cluster.MaintenanceSchedule, diags)
		if serializePlacementDiags.HasError() {
			diags.Append(serializePlacementDiags...)
		}

		maintenanceSchedule, serializeDiags = maintenanceScheduleValue.ToObjectValue(ctx)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		maintenanceSchedule, serializeDiags = datasource_postgres_cluster.NewMaintenanceScheduleValueNull().ToObjectValue(ctx)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	nodeConfigurationValue, serializePlacementDiags := mappingNodeConfiguration(ctx, cluster.NodeConfiguration, diags)
	if serializePlacementDiags.HasError() {
		diags.Append(serializePlacementDiags...)
	}

	nodeConfiguration, serializeDiags = nodeConfigurationValue.ToObjectValue(ctx)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}

	volumeValue, serializePlacementDiags := mappingVolume(ctx, cluster.Volume, diags)
	if serializePlacementDiags.HasError() {
		diags.Append(serializePlacementDiags...)
	}

	volume, serializeDiags = volumeValue.ToObjectValue(ctx)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}

	return datasource_postgres_cluster.NewItemsValue(datasource_postgres_cluster.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"allowed_ip_ranges":     allowedIpRangesList,
		"automatic_backup":      types.BoolValue(cluster.AutomaticBackup),
		"available_operations":  availableOperationsList,
		"created_on":            types.StringValue(cluster.CreatedOn),
		"error_reason":          types.StringPointerValue(cluster.ErrorReason),
		"host":                  types.StringPointerValue(cluster.Host),
		"id":                    types.StringValue(cluster.Id.String()),
		"last_operation_name":   types.StringValue(string(cluster.LastOperationName)),
		"last_operation_result": types.StringValue(string(cluster.LastOperationResult)),
		"maintenance_schedule":  maintenanceSchedule,
		"name":                  types.StringValue(cluster.Name),
		"node_configuration":    nodeConfiguration,
		"port":                  types.Int64Value(utils.ConvertIntPtrToInt64(cluster.Port)),
		"private_host":          types.StringPointerValue(cluster.PrivateHost),
		"status":                types.StringValue(string(cluster.Status)),
		"tags":                  tagsList,
		"user":                  types.StringValue(cluster.User),
		"vpc_cidr":              types.StringPointerValue(cluster.NetCidr),
		"volume":                volume,
	})
}

func mappingDatasourceTags(ctx context.Context, tag api.PostgresTag) (datasource_postgres_cluster.TagsValue, diag.Diagnostics) {
	return datasource_postgres_cluster.NewTagsValue(datasource_postgres_cluster.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingMaintenanceSchedule(ctx context.Context, maintenanceSchedule *api.PostgresClusterMaintenanceSchedule, diags *diag.Diagnostics) (datasource_postgres_cluster.MaintenanceScheduleValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_postgres_cluster.NewMaintenanceScheduleValue(datasource_postgres_cluster.MaintenanceScheduleValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"begin_at":         types.StringValue(maintenanceSchedule.BeginAt),
		"end_at":           types.StringValue(maintenanceSchedule.EndAt),
		"potential_impact": types.StringValue(maintenanceSchedule.PotentialImpact),
		"type":             types.StringValue(string(maintenanceSchedule.Type)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingNodeConfiguration(ctx context.Context, nodeConfiguration api.PostgresNodeConfiguration, diags *diag.Diagnostics) (datasource_postgres_cluster.NodeConfigurationValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_postgres_cluster.NewNodeConfigurationValue(datasource_postgres_cluster.NodeConfigurationValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"memory_size_gi_b":  types.Int64Value(utils.ConvertIntPtrToInt64(&nodeConfiguration.MemorySizeGiB)),
		"performance_level": types.StringValue(string(nodeConfiguration.PerformanceLevel)),
		"vcpu_count":        types.Int64Value(utils.ConvertIntPtrToInt64(&nodeConfiguration.VcpuCount)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingVolume(ctx context.Context, volume api.PostgresVolume, diags *diag.Diagnostics) (datasource_postgres_cluster.VolumeValue, diag.Diagnostics) {
	var elementValue datasource_postgres_cluster.VolumeValue
	var mappingDiags diag.Diagnostics
	valueD, err := volume.ValueByDiscriminator()
	if err != nil {
		return datasource_postgres_cluster.VolumeValue{}, nil
	}

	switch v := valueD.(type) {
	case api.PostgresVolumeGp2:
		elementValue, mappingDiags = datasource_postgres_cluster.NewVolumeValue(datasource_postgres_cluster.VolumeValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"size_gi_b": types.Int64Value(utils.ConvertIntPtrToInt64(&v.SizeGiB)),
			"type":      types.StringValue(string(v.Type)),
		})
		if mappingDiags.HasError() {
			diags.Append(mappingDiags...)
		}
	case api.PostgresVolumeIo1:
		elementValue, mappingDiags = datasource_postgres_cluster.NewVolumeValue(datasource_postgres_cluster.VolumeValue{}.AttributeTypes(ctx), map[string]attr.Value{
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
	return elementValue, mappingDiags
}
