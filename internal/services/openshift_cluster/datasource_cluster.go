package openshift_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/openshift_cluster/datasource_cluster"
	"terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &clustersDataSource{}
)

func (d *clustersDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
}

func NewClustersDataSource() datasource.DataSource {
	return &clustersDataSource{}
}

type clustersDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *clustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_openshift_clusters"
}

// Schema defines the schema for the data source.
func (d *clustersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_cluster.ClusterDataSourceSchema(ctx)
}

func (d *clustersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_cluster.ClusterModel

	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	clusterParams := deserializeReadClusters(plan)
	if response.Diagnostics.HasError() {
		return
	}

	clusters, err := core.ReadClusters(ctx, d.provider, clusterParams)
	if err != nil {
		response.Diagnostics.AddError("unable to read openshift clusters", err.Error())
		return
	}

	clusterItems := utils.SerializeDatasourceItemsWithDiags(ctx, *clusters.Items, &response.Diagnostics, mappingItemsValue)
	if response.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, clusterItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	nextPageToken := mappingNextPageToken(*clusters)
	totalSize := mappingTotalSize(*clusters)
	page := mappingPage(clusterParams)

	state = plan
	state.Items = listValueItems
	state.NextPageToken = nextPageToken
	state.TotalSize = totalSize
	state.Page = page

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func deserializeReadClusters(plan datasource_cluster.ClusterModel) api.ListClustersParams {
	params := api.ListClustersParams{}

	if !plan.Page.IsNull() {
		params.Page = &api.OpenShiftPage{}

		if !plan.Page.NextToken.IsNull() {
			params.Page.NextToken = plan.Page.NextToken.ValueStringPointer()
		}

		if !plan.Page.Size.IsNull() {
			params.Page.Size = utils.PointerOf(int32(plan.Page.Size.ValueInt64()))
		}
	}

	return params
}

func mappingPage(clusterParams api.ListClustersParams) datasource_cluster.PageValue {
	var page datasource_cluster.PageValue
	pageType := map[string]attr.Type{
		"next_token": types.StringType,
		"size":       types.Int64Type,
	}
	if clusterParams.Page != nil {
		pageValue := map[string]attr.Value{
			"next_token": types.StringPointerValue(clusterParams.Page.NextToken),
			"size":       types.Int64Value(int64(*clusterParams.Page.Size)),
		}
		page = datasource_cluster.NewPageValueMust(pageType, pageValue)
	} else {
		pageValue := map[string]attr.Value{
			"next_token": types.StringValue(""),
			"size":       types.Int64Value(0),
		}
		page = datasource_cluster.NewPageValueMust(pageType, pageValue)
	}
	return page
}

func mappingTotalSize(clusters api.OpenShiftClusters) types.Int64 {
	var totalSizeTf types.Int64
	if clusters.TotalSize != nil {
		ts := int64(*clusters.TotalSize)
		totalSizeTf = types.Int64PointerValue(&ts)
	} else {
		totalSizeTf = types.Int64Value(int64(len(*clusters.Items)))
	}
	return totalSizeTf
}

func mappingItemsValue(ctx context.Context, cluster api.OpenShiftCluster, diags *diag.Diagnostics) (datasource_cluster.ItemsValue, diag.Diagnostics) {
	var serializeDiags diag.Diagnostics

	nodePoolsList := types.ListNull(datasource_cluster.NodepoolsValue{}.Type(ctx))

	var state string
	if cluster.State != nil {
		state = *cluster.State
	}

	if cluster.NodePools != nil {
		ln := len(*cluster.NodePools)

		elementValue := make([]datasource_cluster.NodepoolsValue, ln)

		for y := 0; ln > y; y++ {
			nodePools := *cluster.NodePools
			if nodePools != nil {
				var aznp string
				if nodePools[y].AvailabilityZoneName != nil {
					aznp = string(*nodePools[y].AvailabilityZoneName)
				}

				var gpunp string
				if nodePools[y].Gpu != nil {
					gpunp = string(*nodePools[y].Gpu)
				}

				elementValue[y], serializeDiags = datasource_cluster.NewNodepoolsValue(datasource_cluster.NodepoolsValue{}.AttributeTypes(ctx), map[string]attr.Value{
					"availability_zone_name": types.StringValue(aznp),
					"gpu":                    types.StringValue(gpunp),
					"name":                   types.StringValue(nodePools[y].Name),
					"node_count":             types.Int64Value(int64(nodePools[y].NodeCount)),
					"node_profile":           types.StringValue(string(nodePools[y].NodeProfile)),
					"tina":                   types.StringPointerValue(nodePools[y].Tina),
				})
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
					continue
				}
			}
		}

		nodePoolsList = utils.CreateListValueItems(ctx, elementValue, diags)
	}

	urlsValueObject := mappingUrlsValue(ctx, cluster, diags)

	return datasource_cluster.NewItemsValue(datasource_cluster.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":                     types.StringValue(cluster.Id.String()),
		"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(cluster.AvailabilityZoneName)),
		"description":            types.StringPointerValue(cluster.Description),
		"name":                   types.StringPointerValue(cluster.Name),
		"cidr":                   types.StringPointerValue(cluster.Cidr),
		"state":                  types.StringValue(state),
		"version":                types.StringPointerValue(cluster.Version),
		"nodepools":              nodePoolsList,
		"urls":                   urlsValueObject,
	})
}

func mappingUrlsValue(ctx context.Context, cluster api.OpenShiftCluster, diags *diag.Diagnostics) basetypes.ObjectValue {
	urlsType := map[string]attr.Type{
		"api":     types.StringType,
		"console": types.StringType,
	}
	urlsValue := map[string]attr.Value{
		"api":     types.StringPointerValue(cluster.Urls.Api),
		"console": types.StringPointerValue(cluster.Urls.Console),
	}

	urlsValueObject, diagUrl := datasource_cluster.NewUrlsValueMust(urlsType, urlsValue).ToObjectValue(ctx)
	if diagUrl.HasError() {
		diags.Append(diagUrl...)
	}
	return urlsValueObject
}

func mappingNextPageToken(clusters api.OpenShiftClusters) types.String {
	var nextPageToken types.String
	if clusters.NextPageToken != nil {
		nextPageToken = types.StringPointerValue(clusters.NextPageToken)
	} else {
		nextPageToken = types.StringValue("")
	}
	return nextPageToken
}
