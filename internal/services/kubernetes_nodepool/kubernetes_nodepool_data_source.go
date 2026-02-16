package kubernetes_nodepool

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/kubernetes_nodepool/datasource_kubernetes_nodepool"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &kubernetesNodepoolDataSource{}

func (d *kubernetesNodepoolDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

func NewKubernetesNodepoolDataSource() datasource.DataSource {
	return &kubernetesNodepoolDataSource{}
}

type kubernetesNodepoolDataSource struct {
	provider *client.NumSpotSDK
}

func (d *kubernetesNodepoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_nodepools"
}

func (d *kubernetesNodepoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_kubernetes_nodepool.KubernetesNodepoolDataSourceSchema(ctx)
}

func (d *kubernetesNodepoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state, plan datasource_kubernetes_nodepool.KubernetesNodepoolModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUuid, err := uuid.Parse(plan.ClusterId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to parse cluster id", err.Error())
		return
	}

	kubernetesNodePool, err := core.ReadKubernetesNodePools(ctx, d.provider, clusterUuid)
	if err != nil {
		resp.Diagnostics.AddError("unable to read kubernetes node pools", err.Error())
		return
	}

	kubernetesNodePoolsItems := utils.SerializeDatasourceItemsWithDiags(ctx, kubernetesNodePool.Items, &resp.Diagnostics, mappingItemsValue)
	if resp.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, kubernetesNodePoolsItems, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, nodePool api.KubernetesNodePool, diags *diag.Diagnostics) (datasource_kubernetes_nodepool.ItemsValue, diag.Diagnostics) {
	var serializeDiags diag.Diagnostics
	var rootDisk basetypes.ObjectValue
	var status basetypes.ObjectValue

	rootDiskValue, serializeDiags := mappingRootDisk(ctx, nodePool.RootDisk, serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	rootDisk, serializeDiags = rootDiskValue.ToObjectValue(ctx)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}

	statusValue, serializeDiags := mappingStatus(ctx, nodePool.Status, diags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	status, serializeDiags = statusValue.ToObjectValue(ctx)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}

	return datasource_kubernetes_nodepool.NewItemsValue(datasource_kubernetes_nodepool.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone": types.StringValue(utils.ConvertAzNamePtrToString(&nodePool.AvailabilityZone)),
		"id":                types.StringValue(nodePool.Id.String()),
		"name":              types.StringValue(utils.ConvertStringPtrToString(nodePool.Name)),
		"node_profile":      types.StringValue(string(nodePool.NodeProfile)),
		"replicas":          types.Int64Value(utils.ConvertIntPtrToInt64(&nodePool.Replicas)),
		"root_disk":         rootDisk,
		"status":            status,
	})
}

func mappingRootDisk(ctx context.Context, nodePoolDisk api.KubernetesNodePoolDisk, diags diag.Diagnostics) (datasource_kubernetes_nodepool.RootDiskValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_kubernetes_nodepool.NewRootDiskValue(datasource_kubernetes_nodepool.RootDiskValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"iops": types.Int64Value(int64(nodePoolDisk.Iops)),
		"size": types.Int64Value(int64(nodePoolDisk.Size)),
		"type": types.StringValue(string(nodePoolDisk.Type)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingStatus(ctx context.Context, status api.Status, diags *diag.Diagnostics) (datasource_kubernetes_nodepool.StatusValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_kubernetes_nodepool.NewStatusValue(datasource_kubernetes_nodepool.StatusValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"message": types.StringValue(status.Message),
		"state":   types.StringValue(string(status.State)),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}
