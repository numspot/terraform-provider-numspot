package kubernetescluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/kubernetescluster/datasource_kubernetes_cluster"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &kubernetesclustersDataSource{}

// kubernetesclustersDataSource is the Terraform data source implementation for numspot_kubernetes_clusters.
type kubernetesclustersDataSource struct {
	services.BaseDataSource
}

// NewKubernetesClustersDataSource returns a new kubernetesclustersDataSource instance.
func NewKubernetesClustersDataSource() datasource.DataSource {
	return &kubernetesclustersDataSource{}
}

// Metadata sets the data source type name.
func (d *kubernetesclustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_clusters"
}

// Schema defines the schema for the data source.
func (d *kubernetesclustersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_kubernetes_cluster.KubernetesClusterDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *kubernetesclustersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_kubernetes_cluster.KubernetesClusterModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	result, err := core.ReadKubernetesClusters(ctx, d.Provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read kubernetes_clusters", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.KubernetesCluster) (datasource_kubernetes_cluster.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, valueDiags := datasource_kubernetes_cluster.NewItemsValue(
		datasource_kubernetes_cluster.ItemsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cidr":           types.StringValue(item.Cidr),
			"created_on":     utils.TimePtrToTfString(item.CreatedOn),
			"dedicated_etcd": types.BoolValue(item.DedicatedEtcd),
			"full_version":   types.StringValue(item.FullVersion),
			"id":             types.StringValue(item.Id.String()),
			"name":           types.StringValue(item.Name),
			"profile":        types.StringValue(item.Profile),
			"status":         mappingStatu(ctx, &item.Status, &diags),
			"version":        types.StringValue(item.Version),
			"visibility":     types.StringValue(string(item.Visibility)),
		},
	)
	diags.Append(valueDiags...)
	return value, diags
}

func mappingStatu(ctx context.Context, item *api.KubernetesStatus, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_kubernetes_cluster.StatusValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_kubernetes_cluster.NewStatusValue(datasource_kubernetes_cluster.StatusValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"message": types.StringValue(item.Message),
		"reason":  types.StringPointerValue((*string)(item.Reason)),
		"state":   types.StringValue(string(item.State)),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}
