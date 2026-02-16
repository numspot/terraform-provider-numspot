package openshift_cluster

//
//import (
//	"context"
//	"fmt"
//
//	"github.com/google/uuid"
//	"github.com/hashicorp/terraform-plugin-framework/attr"
//	"github.com/hashicorp/terraform-plugin-framework/diag"
//	"github.com/hashicorp/terraform-plugin-framework/path"
//	"github.com/hashicorp/terraform-plugin-framework/resource"
//	"github.com/hashicorp/terraform-plugin-framework/types"
//	"terraform-provider-numspot/internal/client"
//	"terraform-provider-numspot/internal/core"
//	"terraform-provider-numspot/internal/sdk/api"
//	"terraform-provider-numspot/internal/services/openshift_cluster/resource_cluster"
//	"terraform-provider-numspot/internal/utils"
//)
//
//const (
//	resourceTypeName = "_openshift_cluster"
//)
//
//var (
//	_ resource.Resource                = &Resource{}
//	_ resource.ResourceWithConfigure   = &Resource{}
//	_ resource.ResourceWithImportState = &Resource{}
//)
//
//type Resource struct {
//	provider *client.NumSpotSDK
//}
//
//func NewClusterResource() resource.Resource {
//	return &Resource{}
//}
//
//func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
//	if request.ProviderData == nil {
//		return
//	}
//
//	provider, ok := request.ProviderData.(*client.NumSpotSDK)
//	if !ok {
//		response.Diagnostics.AddError(
//			"unexpected resource configure type",
//			fmt.Sprintf("expected *http.Client, got: %T please report this issue to the provider developers", request.ProviderData),
//		)
//
//		return
//	}
//
//	r.provider = provider
//}
//
//func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
//	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
//}
//
//func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
//	response.TypeName = request.ProviderTypeName + resourceTypeName
//}
//
//func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
//	response.Schema = resource_cluster.ClusterResourceSchema(ctx)
//}
//
//func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
//	var plan resource_cluster.ClusterModel
//	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
//	if response.Diagnostics.HasError() {
//		return
//	}
//
//	var diags diag.Diagnostics
//	createClusterRequest := deserializeCreateCluster(ctx, plan, &diags)
//	if diags.HasError() {
//		response.Diagnostics.Append(diags...)
//		return
//	}
//
//	cluster, err := core.CreateOpenshiftCluster(ctx, r.provider, createClusterRequest)
//	if err != nil {
//		response.Diagnostics.AddError("unable to create openshift cluster", err.Error())
//		return
//	}
//
//	state := serializeCluster(ctx, cluster, diags)
//	if response.Diagnostics.HasError() {
//		return
//	}
//
//	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
//}
//
//func deserializeCreateCluster(ctx context.Context, tf resource_cluster.ClusterModel, diags *diag.Diagnostics) api.CreateClusterJSONRequestBody {
//	var ocpAdminUserId *uuid.UUID
//	if !tf.OcpAdminUserId.IsNull() && !tf.OcpAdminUserId.IsUnknown() {
//		id := uuid.MustParse(tf.OcpAdminUserId.ValueString())
//		ocpAdminUserId = utils.PointerOf(id)
//	}
//
//	return api.CreateClusterJSONRequestBody{
//		AvailabilityZoneName: utils.FromTfStringToAzNamePtr(tf.AvailabilityZoneName),
//		Cidr:                 tf.Cidr.ValueString(),
//		Description:          tf.Description.ValueStringPointer(),
//		Name:                 tf.Name.ValueString(),
//		NodePools:            deserializeNodePools(ctx, tf.NodePools, diags),
//		OcpAdminUserId:       ocpAdminUserId,
//		Version:              tf.Version.ValueString(),
//	}
//}
//
//func deserializeNodePools(ctx context.Context, tfNodePools types.List, diags *diag.Diagnostics) []api.OpenShiftNodepool {
//	if tfNodePools.IsNull() || tfNodePools.IsUnknown() {
//		return nil
//	}
//
//	var nodePools []resource_cluster.NodePoolsValue
//	diags.Append(tfNodePools.ElementsAs(ctx, &nodePools, false)...)
//
//	var result []api.OpenShiftNodepool
//	for _, np := range nodePools {
//		nodePool := api.OpenShiftNodepool{
//			Name:        np.Name.ValueString(),
//			NodeCount:   int(np.NodeCount.ValueInt64()),
//			NodeProfile: api.NodeProfile(np.NodeProfile.ValueString()),
//		}
//
//		if !np.AvailabilityZoneName.IsNull() && !np.AvailabilityZoneName.IsUnknown() {
//			az := api.AvailabilityZoneName(np.AvailabilityZoneName.ValueString())
//			nodePool.AvailabilityZoneName = &az
//		}
//
//		if !np.Gpu.IsNull() && !np.Gpu.IsUnknown() {
//			gpu := api.Gpu(np.Gpu.ValueString())
//			nodePool.Gpu = &gpu
//		}
//
//		if !np.Tina.IsNull() && !np.Tina.IsUnknown() {
//			tina := np.Tina.ValueString()
//			nodePool.Tina = &tina
//		}
//
//		result = append(result, nodePool)
//	}
//
//	return result
//}
//
//func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
//	var state resource_cluster.ClusterModel
//	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
//	if response.Diagnostics.HasError() {
//		return
//	}
//
//	clusterId := state.Id.ValueString()
//
//	numSpotCluster, err := core.ReadOpenshiftCluster(ctx, r.provider, clusterId)
//	if err != nil {
//		response.Diagnostics.AddError("unable to read openshift cluster", err.Error())
//		return
//	}
//
//	newState := serializeCluster(ctx, numSpotCluster, response.Diagnostics)
//	if response.Diagnostics.HasError() {
//		return
//	}
//
//	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
//}
//
//func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
//}
//
//func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
//	var state resource_cluster.ClusterModel
//	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
//	if response.Diagnostics.HasError() {
//		return
//	}
//
//	clusterId := state.Id.ValueString()
//
//	_, err := core.DeleteOpenshiftCluster(ctx, r.provider, clusterId)
//	if err != nil {
//		response.Diagnostics.AddError("unable to delete openshift cluster", err.Error())
//		return
//	}
//}
//
//func serializeCluster(ctx context.Context, cluster *api.OpenShiftCluster, diag diag.Diagnostics) resource_cluster.ClusterModel {
//	var availabilityZoneName string
//	if cluster.AvailabilityZoneName != nil {
//		az := string(*cluster.AvailabilityZoneName)
//		availabilityZoneName = az
//	}
//
//	var clusterState string
//	if cluster.State != nil {
//		st := *cluster.State
//		clusterState = st
//	}
//
//	urlsType := map[string]attr.Type{
//		"api":     types.StringType,
//		"console": types.StringType,
//	}
//	urlsValue := map[string]attr.Value{
//		"api":     types.StringPointerValue(cluster.Urls.Api),
//		"console": types.StringPointerValue(cluster.Urls.Console),
//	}
//
//	nodePoolsList := types.List{}
//
//	if cluster.NodePools != nil {
//		ll := len(*cluster.NodePools)
//		elementValue := make([]resource_cluster.NodePoolsValue, ll)
//
//		for i := 0; ll > i; i++ {
//			var az string
//			if (*cluster.NodePools)[i].AvailabilityZoneName != nil {
//				az = string(*(*cluster.NodePools)[i].AvailabilityZoneName)
//			}
//
//			var gpu string
//			if (*cluster.NodePools)[i].Gpu != nil {
//				gpu = string(*(*cluster.NodePools)[i].Gpu)
//			}
//
//			elementValue[i], diag = resource_cluster.NewNodePoolsValue(resource_cluster.NodePoolsValue{}.AttributeTypes(ctx), map[string]attr.Value{
//				"availability_zone_name": types.StringValue(az),
//				"gpu":                    types.StringValue(gpu),
//				"name":                   types.StringValue((*cluster.NodePools)[i].Name),
//				"node_count":             types.Int64Value(int64((*cluster.NodePools)[i].NodeCount)),
//				"node_profile":           types.StringValue(string((*cluster.NodePools)[i].NodeProfile)),
//				"tina":                   types.StringPointerValue((*cluster.NodePools)[i].Tina),
//			})
//			if diag.HasError() {
//				diag.Append(diag...)
//				continue
//			}
//
//		}
//
//		nodePoolsList, diag = types.ListValueFrom(ctx, new(resource_cluster.NodePoolsValue).Type(ctx), elementValue)
//		if diag.HasError() {
//			diag.Append(diag...)
//			return resource_cluster.ClusterModel{}
//		}
//	}
//
//	return resource_cluster.ClusterModel{
//		AvailabilityZoneName: types.StringValue(availabilityZoneName),
//		Cidr:                 types.StringPointerValue(cluster.Cidr),
//		Description:          types.StringPointerValue(cluster.Description),
//		Id:                   types.StringValue(cluster.Id.String()),
//		Name:                 types.StringPointerValue(cluster.Name),
//		NodePools:            nodePoolsList,
//		State:                types.StringValue(clusterState),
//		Urls:                 resource_cluster.NewUrlsValueMust(urlsType, urlsValue),
//		Version:              types.StringPointerValue(cluster.Version),
//	}
//}
