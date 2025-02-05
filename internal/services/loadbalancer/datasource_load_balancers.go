package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

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

	loadBalancerParams := deserializeReadLoadBalancers(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	loadBalancers, err := core.ReadLoadBalancers(ctx, d.provider, loadBalancerParams)
	if err != nil {
		response.Diagnostics.AddError("unable to read load balancers", err.Error())
		return
	}

	loadBalancerItems := serializeLoadBalancers(ctx, loadBalancers, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = loadBalancerItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func deserializeReadLoadBalancers(ctx context.Context, plan loadBalancersDataSourceModel, diags *diag.Diagnostics) numspot.ReadLoadBalancersParams {
	params := numspot.ReadLoadBalancersParams{}
	if !plan.LoadBalancerNames.IsNull() {
		lbNames := utils.TfStringListToStringList(ctx, plan.LoadBalancerNames, diags)
		params.LoadBalancerNames = &lbNames
	}
	return params
}

func serializeLoadBalancers(ctx context.Context, loadBalancers *[]numspot.LoadBalancer, diags *diag.Diagnostics) []LoadBalancerModelDatasource {
	return utils.FromHttpGenericListToTfList(ctx, loadBalancers, func(ctx context.Context, loadBalancer *numspot.LoadBalancer, diags *diag.Diagnostics) *LoadBalancerModelDatasource {
		var tagsList types.List

		if loadBalancer.Tags != nil {
			tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *loadBalancer.Tags, diags)
			if diags.HasError() {
				return nil
			}
		}

		applicationStickyCookiePoliciesTypes := utils.GenericListToTfListValue(ctx, applicationStickyCookiePoliciesFromHTTP, *loadBalancer.ApplicationStickyCookiePolicies, diags)
		if diags.HasError() {
			return nil
		}

		listeners := utils.GenericSetToTfSetValue(ctx, listenersFromHTTP, *loadBalancer.Listeners, diags)
		if diags.HasError() {
			return nil
		}

		stickyCookiePolicies := utils.GenericListToTfListValue(ctx, stickyCookiePoliciesFromHTTP, *loadBalancer.StickyCookiePolicies, diags)
		if diags.HasError() {
			return nil
		}

		backendIps := utils.FromStringListPointerToTfStringSet(ctx, loadBalancer.BackendIps, diags)
		backendVmIds := utils.FromStringListPointerToTfStringSet(ctx, loadBalancer.BackendVmIds, diags)
		healthCheck, diagnostics := NewHealthCheckValue(HealthCheckValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"check_interval":      utils.FromIntToTfInt64(loadBalancer.HealthCheck.CheckInterval),
				"healthy_threshold":   utils.FromIntToTfInt64(loadBalancer.HealthCheck.HealthyThreshold),
				"path":                types.StringPointerValue(loadBalancer.HealthCheck.Path),
				"port":                utils.FromIntToTfInt64(loadBalancer.HealthCheck.Port),
				"protocol":            types.StringValue(loadBalancer.HealthCheck.Protocol),
				"timeout":             utils.FromIntToTfInt64(loadBalancer.HealthCheck.Timeout),
				"unhealthy_threshold": utils.FromIntToTfInt64(loadBalancer.HealthCheck.UnhealthyThreshold),
			})

		diags.Append(diagnostics...)
		securityGroups := utils.FromStringListPointerToTfStringList(ctx, loadBalancer.SecurityGroups, diags)
		sourceSecurityGroup := SourceSecurityGroupValue{
			SecurityGroupName: types.StringPointerValue(loadBalancer.SourceSecurityGroup.SecurityGroupName),
		}
		subnets := utils.FromStringListPointerToTfStringList(ctx, loadBalancer.Subnets, diags)
		azNames := utils.FromStringListPointerToTfStringList(ctx, loadBalancer.AvailabilityZoneNames, diags)

		return &LoadBalancerModelDatasource{
			ApplicationStickyCookiePolicies: applicationStickyCookiePoliciesTypes,
			BackendIps:                      backendIps,
			BackendVmIds:                    backendVmIds,
			DnsName:                         types.StringPointerValue(loadBalancer.DnsName),
			HealthCheck:                     healthCheck,
			Listeners:                       listeners,
			Name:                            types.StringPointerValue(loadBalancer.Name),
			VpcId:                           types.StringPointerValue(loadBalancer.VpcId),
			PublicIp:                        types.StringPointerValue(loadBalancer.PublicIp),
			SecuredCookies:                  types.BoolPointerValue(loadBalancer.SecuredCookies),
			SecurityGroups:                  securityGroups,
			SourceSecurityGroup:             sourceSecurityGroup,
			StickyCookiePolicies:            stickyCookiePolicies,
			Subnets:                         subnets,
			AvailabilityZoneNames:           azNames,
			ItemsType:                       types.StringPointerValue(loadBalancer.Type),
			Tags:                            tagsList,
		}
	}, diags)
}
