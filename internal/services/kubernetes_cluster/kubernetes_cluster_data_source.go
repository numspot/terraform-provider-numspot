package kubernetes_cluster

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
	"terraform-provider-numspot/internal/services/kubernetes_cluster/datasource_kubernetes_cluster"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &kubernetesClusterDataSource{}

func (d *kubernetesClusterDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

func NewKubernetesClusterDataSource() datasource.DataSource {
	return &kubernetesClusterDataSource{}
}

type kubernetesClusterDataSource struct {
	provider *client.NumSpotSDK
}

func (d *kubernetesClusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_clusters"
}

func (d *kubernetesClusterDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_kubernetes_cluster.KubernetesClusterDataSourceSchema(ctx)
}

func (d *kubernetesClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state, plan datasource_kubernetes_cluster.KubernetesClusterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kubernetesCluster, err := core.ReadKubernetesClusters(ctx, d.provider)
	if err != nil {
		resp.Diagnostics.AddError("unable to read kubernetes clusters", err.Error())
		return
	}

	kubernetesClusterItems := utils.SerializeDatasourceItemsWithDiags(ctx, kubernetesCluster, &resp.Diagnostics, mappingItemsValue)
	if resp.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, kubernetesClusterItems, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, cluster api.KubernetesCluster, diags *diag.Diagnostics) (datasource_kubernetes_cluster.ItemsValue, diag.Diagnostics) {
	var status basetypes.ObjectValue

	statusValue, mappingDiag := mappingStatus(ctx, cluster.Status, diags)
	if mappingDiag.HasError() {
		diags.Append(mappingDiag...)
	}

	status, mappingDiag = statusValue.ToObjectValue(ctx)
	if mappingDiag.HasError() {
		diags.Append(mappingDiag...)
	}

	return datasource_kubernetes_cluster.NewItemsValue(datasource_kubernetes_cluster.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"created_on":   types.StringValue(cluster.CreatedOn.String()),
		"cidr":         types.StringValue(cluster.Cidr),
		"id":           types.StringValue(cluster.Id.String()),
		"name":         types.StringValue(cluster.Name),
		"profile":      types.StringValue(cluster.Profile),
		"full_version": types.StringValue(cluster.FullVersion),
		"status":       status,
		"version":      types.StringValue(cluster.Version),
		"visibility":   types.StringValue(string(cluster.Visibility)),
	})
}

func mappingStatus(ctx context.Context, status api.Status, diags *diag.Diagnostics) (datasource_kubernetes_cluster.StatusValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_kubernetes_cluster.NewStatusValue(datasource_kubernetes_cluster.StatusValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"message": types.StringValue(status.Message),
		"state":   types.StringValue(string(status.State)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}
