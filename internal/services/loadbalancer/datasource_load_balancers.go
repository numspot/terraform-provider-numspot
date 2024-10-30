package loadbalancer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type loadBalancersDataSourceModel struct {
	Items             []LoadBalancerModelDatasource `tfsdk:"items"`
	LoadBalancerNames types.List                    `tfsdk:"load_balancer_names"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &loadBalancersDataSource{}
)

func (d *loadBalancersDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewLoadBalancersDataSource() datasource.DataSource {
	return &loadBalancersDataSource{}
}

type loadBalancersDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *loadBalancersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancers"
}

// Schema defines the schema for the data source.
func (d *loadBalancersDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = LoadBalancerDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *loadBalancersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan loadBalancersDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := d.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	params := numspot.ReadLoadBalancersParams{}
	if !plan.LoadBalancerNames.IsNull() {
		lbNames := utils.TfStringListToStringList(ctx, plan.LoadBalancerNames, &response.Diagnostics)
		params.LoadBalancerNames = &lbNames
	}
	res := utils.ExecuteRequest(func() (*numspot.ReadLoadBalancersResponse, error) {
		return numspotClient.ReadLoadBalancersWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty load balancers list")
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, LoadBalancerFromHttpToTfDatasource, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func LoadBalancerFromHttpToTfDatasource(ctx context.Context, http *numspot.LoadBalancer, diags *diag.Diagnostics) *LoadBalancerModelDatasource {
	var tagsList types.List

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	applicationStickyCookiePoliciestypes := utils.GenericListToTfListValue(ctx, applicationStickyCookiePoliciesFromHTTP, *http.ApplicationStickyCookiePolicies, diags)
	if diags.HasError() {
		return nil
	}

	listeners := utils.GenericSetToTfSetValue(ctx, listenersFromHTTP, *http.Listeners, diags)
	if diags.HasError() {
		return nil
	}

	stickyCookiePolicies := utils.GenericListToTfListValue(ctx, stickyCookiePoliciesFromHTTP, *http.StickyCookiePolicies, diags)
	if diags.HasError() {
		return nil
	}

	backendIps := utils.FromStringListPointerToTfStringSet(ctx, http.BackendIps, diags)
	backendVmIds := utils.FromStringListPointerToTfStringSet(ctx, http.BackendVmIds, diags)
	healthCheck, diagnostics := NewHealthCheckValue(HealthCheckValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"check_interval":      utils.FromIntToTfInt64(http.HealthCheck.CheckInterval),
			"healthy_threshold":   utils.FromIntToTfInt64(http.HealthCheck.HealthyThreshold),
			"path":                types.StringPointerValue(http.HealthCheck.Path),
			"port":                utils.FromIntToTfInt64(http.HealthCheck.Port),
			"protocol":            types.StringValue(http.HealthCheck.Protocol),
			"timeout":             utils.FromIntToTfInt64(http.HealthCheck.Timeout),
			"unhealthy_threshold": utils.FromIntToTfInt64(http.HealthCheck.UnhealthyThreshold),
		})

	diags.Append(diagnostics...)
	securityGroups := utils.FromStringListPointerToTfStringList(ctx, http.SecurityGroups, diags)
	sourceSecurityGroup := SourceSecurityGroupValue{
		SecurityGroupName: types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupName),
	}
	subnets := utils.FromStringListPointerToTfStringList(ctx, http.Subnets, diags)
	azNames := utils.FromStringListPointerToTfStringList(ctx, http.AvailabilityZoneNames, diags)

	return &LoadBalancerModelDatasource{
		ApplicationStickyCookiePolicies: applicationStickyCookiePoliciestypes,
		BackendIps:                      backendIps,
		BackendVmIds:                    backendVmIds,
		DnsName:                         types.StringPointerValue(http.DnsName),
		HealthCheck:                     healthCheck,
		Listeners:                       listeners,
		Name:                            types.StringPointerValue(http.Name),
		VpcId:                           types.StringPointerValue(http.VpcId),
		PublicIp:                        types.StringPointerValue(http.PublicIp),
		SecuredCookies:                  types.BoolPointerValue(http.SecuredCookies),
		SecurityGroups:                  securityGroups,
		SourceSecurityGroup:             sourceSecurityGroup,
		StickyCookiePolicies:            stickyCookiePolicies,
		Subnets:                         subnets,
		AvailabilityZoneNames:           azNames,
		ItemsType:                       types.StringPointerValue(http.Type),
		Tags:                            tagsList,
	}
}
