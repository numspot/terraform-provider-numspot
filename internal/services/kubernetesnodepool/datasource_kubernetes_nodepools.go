package kubernetesnodepool

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/kubernetesnodepool/datasource_kubernetes_nodepool"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &kubernetesnodepoolsDataSource{}

// kubernetesnodepoolsDataSource is the Terraform data source implementation for numspot_kubernetes_nodepools.
type kubernetesnodepoolsDataSource struct {
	services.BaseDataSource
}

// NewKubernetesNodepoolsDataSource returns a new kubernetesnodepoolsDataSource instance.
func NewKubernetesNodepoolsDataSource() datasource.DataSource {
	return &kubernetesnodepoolsDataSource{}
}

// Metadata sets the data source type name.
func (d *kubernetesnodepoolsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_nodepools"
}

// Schema defines the schema for the data source.
func (d *kubernetesnodepoolsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_kubernetes_nodepool.KubernetesNodepoolDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *kubernetesnodepoolsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_kubernetes_nodepool.KubernetesNodepoolModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	result, err := core.ReadKubernetesNodePools(ctx, d.Provider, plan.ClusterId.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read kubernetes_nodepools", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.KubernetesNodePool) (datasource_kubernetes_nodepool.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, valueDiags := datasource_kubernetes_nodepool.NewItemsValue(
		datasource_kubernetes_nodepool.ItemsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"autoscaling_max":    utils.FromIntPtrToTfInt64(item.AutoscalingMax),
			"autoscaling_min":    utils.FromIntPtrToTfInt64(item.AutoscalingMin),
			"availability_zone":  types.StringValue(string(item.AvailabilityZone)),
			"id":                 types.StringValue(item.Id.String()),
			"name":               types.StringValue(utils.ConvertStringPtrToString(item.Name)),
			"node_profile":       types.StringValue(string(item.NodeProfile)),
			"replicas":           utils.FromIntToTfInt64(item.Replicas),
			"replicas_available": utils.FromIntPtrToTfInt64(item.ReplicasAvailable),
			"root_disk":          mappingRootDisk(ctx, &item.RootDisk, &diags),
			"status":             mappingStatu(ctx, &item.Status, &diags),
		},
	)
	diags.Append(valueDiags...)
	return value, diags
}

func mappingRootDisk(ctx context.Context, item *api.KubernetesNodePoolDisk, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_kubernetes_nodepool.RootDiskValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_kubernetes_nodepool.NewRootDiskValue(
		datasource_kubernetes_nodepool.RootDiskValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"iops": utils.FromIntToTfInt64(item.Iops),
			"size": utils.FromIntToTfInt64(item.Size),
			"type": types.StringValue(string(item.Type)),
		},
	)
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingStatu(ctx context.Context, item *api.KubernetesStatus, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_kubernetes_nodepool.StatusValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_kubernetes_nodepool.NewStatusValue(datasource_kubernetes_nodepool.StatusValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"message": types.StringValue(item.Message),
		"reason":  types.StringPointerValue((*string)(item.Reason)),
		"state":   types.StringValue(string(item.State)),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}
