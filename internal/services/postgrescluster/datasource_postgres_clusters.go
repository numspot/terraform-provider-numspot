package postgrescluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/postgrescluster/datasource_postgres_cluster"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &postgresclustersDataSource{}

// postgresclustersDataSource is the Terraform data source implementation for numspot_postgres_clusters.
type postgresclustersDataSource struct {
	services.BaseDataSource
}

// NewPostgresClustersDataSource returns a new postgresclustersDataSource instance.
func NewPostgresClustersDataSource() datasource.DataSource {
	return &postgresclustersDataSource{}
}

// Metadata sets the data source type name.
func (d *postgresclustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres_clusters"
}

// Schema defines the schema for the data source.
func (d *postgresclustersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_postgres_cluster.PostgresClusterDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *postgresclustersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_postgres_cluster.PostgresClusterModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	result, err := core.ReadPostgresClusters(ctx, d.Provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read postgres_clusters", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.PostgreSQLCluster) (datasource_postgres_cluster.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, valueDiags := datasource_postgres_cluster.NewItemsValue(
		datasource_postgres_cluster.ItemsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"created_on":             utils.TimeToTfString(item.CreatedOn),
			"extensions":             extensionsToDatasourceList(ctx, item.Extensions, &diags),
			"full_version":           types.StringValue(item.FullVersion),
			"host":                   types.StringValue(utils.ConvertStringPtrToString(item.Host)),
			"id":                     types.StringValue(item.Id.String()),
			"last_password_reset_at": utils.TimePtrToTfString(item.LastPasswordResetAt),
			"major_version":          types.StringValue(string(item.MajorVersion)),
			"name":                   types.StringValue(item.Name),
			"node_configuration":     mappingNodeConfiguration(ctx, &item.NodeConfiguration, &diags),
			"port":                   types.Int64Value(int64(item.Port)),
			"replica_count":          types.Int64Value(int64(item.ReplicaCount)),
			"status":                 mappingStatu(ctx, &item.Status, &diags),
			"user":                   types.StringValue(item.User),
			"visibility":             types.StringValue(string(item.Visibility)),
			"volume":                 volumeToDsObject(ctx, item.Volume, &diags),
		},
	)
	diags.Append(valueDiags...)
	return value, diags
}

func mappingNodeConfiguration(ctx context.Context, item *api.PostgreSQLNodeConfiguration, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_postgres_cluster.NodeConfigurationValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_postgres_cluster.NewNodeConfigurationValue(
		datasource_postgres_cluster.NodeConfigurationValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"memory_size_gi_b": types.Int64Value(int64(item.MemorySizeGiB)),
			"vcpu_count":       types.Int64Value(int64(item.VcpuCount)),
		},
	)
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingStatu(ctx context.Context, item *api.PostgreSQLClusterStatus, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_postgres_cluster.StatusValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_postgres_cluster.NewStatusValue(datasource_postgres_cluster.StatusValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"message": types.StringValue(item.Message),
		"state":   types.StringValue(string(item.State)),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}
