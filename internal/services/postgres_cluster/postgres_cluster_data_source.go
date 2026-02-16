package postgres_cluster

import (
	"context"
	"time"

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
	var nodeConfiguration basetypes.ObjectValue
	var volume basetypes.ObjectValue
	var status basetypes.ObjectValue

	extensionList := types.ListNull(datasource_postgres_cluster.ExtensionsValue{}.Type(ctx))

	if cluster.Extensions != nil {
		extensionItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *cluster.Extensions, mappingDatasourceExtensions)
		if serializeDiags.HasError() {
			return datasource_postgres_cluster.ItemsValue{}, serializeDiags
		}
		extensionList = utils.CreateListValueItems(ctx, extensionItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_postgres_cluster.ItemsValue{}, serializeDiags
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

	var portValue types.Int64
	if cluster.Port != nil {
		portValue = types.Int64Value(int64(*cluster.Port))
	} else {
		portValue = types.Int64Value(0)
	}

	var majorVersion types.String
	if cluster.MajorVersion != nil {
		majorVersion = types.StringValue(string(*cluster.MajorVersion))
	} else {
		majorVersion = types.StringNull()
	}

	statusValue, serializeDiags := mappingStatus(ctx, cluster.Status, diags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}

	status, serializeDiags = statusValue.ToObjectValue(ctx)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}

	return datasource_postgres_cluster.NewItemsValue(datasource_postgres_cluster.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"created_on":         types.StringValue(cluster.CreatedOn.Format(time.RFC3339)),
		"extensions":         extensionList,
		"full_version":       types.StringPointerValue(cluster.FullVersion),
		"host":               types.StringPointerValue(cluster.Host),
		"id":                 types.StringValue(cluster.Id.String()),
		"major_version":      majorVersion,
		"name":               types.StringValue(cluster.Name),
		"node_configuration": nodeConfiguration,
		"port":               portValue,
		"replica_count":      types.Int64Value(int64(cluster.ReplicaCount)),
		"status":             status,
		"user":               types.StringValue(cluster.User),
		"visibility":         types.StringValue(string(cluster.Visibility)),
		"volume":             volume,
	})
}

func mappingDatasourceExtensions(ctx context.Context, ext api.PostgresExtension) (datasource_postgres_cluster.ExtensionsValue, diag.Diagnostics) {
	return datasource_postgres_cluster.NewExtensionsValue(datasource_postgres_cluster.ExtensionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"name": types.StringValue(string(ext.Name)),
	})
}

func mappingNodeConfiguration(ctx context.Context, nodeConfiguration api.PostgresNodeConfiguration, diags *diag.Diagnostics) (datasource_postgres_cluster.NodeConfigurationValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_postgres_cluster.NewNodeConfigurationValue(datasource_postgres_cluster.NodeConfigurationValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"memory_size_gi_b": types.Int64Value(int64(nodeConfiguration.MemorySizeGiB)),
		"vcpu_count":       types.Int64Value(int64(nodeConfiguration.VcpuCount)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingVolume(ctx context.Context, volume api.PostgresVolume, diags *diag.Diagnostics) (datasource_postgres_cluster.VolumeValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_postgres_cluster.NewVolumeValue(datasource_postgres_cluster.VolumeValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"size_gi_b": types.Int64Value(int64(volume.SizeGiB)),
		"type":      types.StringValue(string(volume.Type)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingStatus(ctx context.Context, status api.PostgresStatus, diags *diag.Diagnostics) (datasource_postgres_cluster.StatusValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_postgres_cluster.NewStatusValue(datasource_postgres_cluster.StatusValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"message": types.StringValue(status.Message),
		"state":   types.StringValue(string(status.State)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}
