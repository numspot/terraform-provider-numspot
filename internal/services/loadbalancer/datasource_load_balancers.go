package loadbalancer

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
	"terraform-provider-numspot/internal/services/loadbalancer/datasource_load_balancer"
	"terraform-provider-numspot/internal/utils"
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
	resp.Schema = datasource_load_balancer.LoadBalancerDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *loadBalancersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_load_balancer.LoadBalancerModel

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
	state.Items = loadBalancerItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func deserializeReadLoadBalancers(ctx context.Context, plan datasource_load_balancer.LoadBalancerModel, diags *diag.Diagnostics) api.ReadLoadBalancersParams {
	return api.ReadLoadBalancersParams{
		LoadBalancerNames: utils.ConvertTfListToArrayOfString(ctx, plan.LoadBalancerNames, diags),
	}
}

func serializeLoadBalancers(ctx context.Context, loadBalancers *[]api.LoadBalancer, diags *diag.Diagnostics) datasource_load_balancer.LoadBalancerModel {
	var serializeDiags diag.Diagnostics
	var loadBalancersList types.List

	tagsList := types.List{}
	applicationStickyCookiePoliciesList := types.List{}
	listenersList := types.List{}
	stickyCookiePoliciesList := types.List{}
	availabilityZoneNamesList := types.List{}
	backendIpsList := types.ListNull(types.String{}.Type(ctx))
	backendVmIdsList := types.List{}
	securityGroupsList := types.List{}
	subnetsList := types.List{}
	healthCheck := basetypes.ObjectValue{}

	if len(*loadBalancers) != 0 {
		ll := len(*loadBalancers)
		itemsValue := make([]datasource_load_balancer.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			if (*loadBalancers)[i].Tags != nil {
				tagsList, serializeDiags = mappingLoadBalancerTags(ctx, loadBalancers, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*loadBalancers)[i].ApplicationStickyCookiePolicies != nil {
				applicationStickyCookiePoliciesList, serializeDiags = mappingApplicationStickyCookiePolicies(ctx, loadBalancers, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*loadBalancers)[i].HealthCheck != nil {
				healthCheckValue, serializeHealthDiags := mappingHealthCheck(ctx, loadBalancers, diags, i)
				if serializeHealthDiags.HasError() {
					diags.Append(serializeHealthDiags...)
				}
				healthCheck, serializeDiags = healthCheckValue.ToObjectValue(ctx)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*loadBalancers)[i].Listeners != nil {
				listenersList, serializeDiags = mappingListeners(ctx, loadBalancers, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*loadBalancers)[i].StickyCookiePolicies != nil {
				stickyCookiePoliciesList, serializeDiags = mappingStickyCookiePolicies(ctx, loadBalancers, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*loadBalancers)[i].AvailabilityZoneNames != nil {
				availabilityZoneNamesList, serializeDiags = types.ListValueFrom(ctx, types.StringType, (*loadBalancers)[i].AvailabilityZoneNames)
				diags.Append(serializeDiags...)
			}

			if (*loadBalancers)[i].BackendIps != nil {
				backendIpsList, serializeDiags = types.ListValueFrom(ctx, types.StringType, (*loadBalancers)[i].BackendIps)
				diags.Append(serializeDiags...)
			}

			if (*loadBalancers)[i].BackendVmIds != nil {
				backendVmIdsList, serializeDiags = types.ListValueFrom(ctx, types.StringType, (*loadBalancers)[i].BackendVmIds)
				diags.Append(serializeDiags...)
			}

			if (*loadBalancers)[i].SecurityGroups != nil {
				securityGroupsList, serializeDiags = types.ListValueFrom(ctx, types.StringType, (*loadBalancers)[i].SecurityGroups)
				diags.Append(serializeDiags...)
			}

			if (*loadBalancers)[i].Subnets != nil {
				subnetsList, serializeDiags = types.ListValueFrom(ctx, types.StringType, (*loadBalancers)[i].Subnets)
				diags.Append(serializeDiags...)
			}

			itemsValue[i], serializeDiags = datasource_load_balancer.NewItemsValue(datasource_load_balancer.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"application_sticky_cookie_policies": applicationStickyCookiePoliciesList,
				"availability_zone_names":            availabilityZoneNamesList,
				"backend_ips":                        backendIpsList,
				"backend_vm_ids":                     backendVmIdsList,
				"dns_name":                           types.StringValue(utils.ConvertStringPtrToString((*loadBalancers)[i].DnsName)),
				"health_check":                       healthCheck,
				"listeners":                          listenersList,
				"name":                               types.StringValue(utils.ConvertStringPtrToString((*loadBalancers)[i].Name)),
				"public_ip":                          types.StringValue(utils.ConvertStringPtrToString((*loadBalancers)[i].PublicIp)),
				"secured_cookies":                    types.BoolPointerValue((*loadBalancers)[i].SecuredCookies),
				"security_groups":                    securityGroupsList,
				"sticky_cookie_policies":             stickyCookiePoliciesList,
				"subnets":                            subnetsList,
				"tags":                               tagsList,
				"type":                               types.StringValue(utils.ConvertStringPtrToString((*loadBalancers)[i].Type)),
				"vpc_id":                             types.StringValue(utils.ConvertStringPtrToString((*loadBalancers)[i].VpcId)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		loadBalancersList, serializeDiags = types.ListValueFrom(ctx, new(datasource_load_balancer.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		loadBalancersList = types.ListNull(new(datasource_load_balancer.ItemsValue).Type(ctx))
	}

	return datasource_load_balancer.LoadBalancerModel{
		Items: loadBalancersList,
	}
}

func mappingApplicationStickyCookiePolicies(ctx context.Context, loadBalancers *[]api.LoadBalancer, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	la := len(*(*loadBalancers)[i].ApplicationStickyCookiePolicies)
	elementValue := make([]datasource_load_balancer.ApplicationStickyCookiePoliciesValue, la)
	for y, cookiePolicy := range *(*loadBalancers)[i].ApplicationStickyCookiePolicies {
		elementValue[y], *diags = datasource_load_balancer.NewApplicationStickyCookiePoliciesValue(datasource_load_balancer.ApplicationStickyCookiePoliciesValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"cookie_name": types.StringPointerValue(cookiePolicy.CookieName),
			"policy_name": types.StringPointerValue(cookiePolicy.PolicyName),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_load_balancer.ApplicationStickyCookiePoliciesValue).Type(ctx), elementValue)
}

func mappingHealthCheck(ctx context.Context, loadBalancers *[]api.LoadBalancer, diags *diag.Diagnostics, i int) (datasource_load_balancer.HealthCheckValue, diag.Diagnostics) {
	elementValue, mappingDiags := datasource_load_balancer.NewHealthCheckValue(datasource_load_balancer.HealthCheckValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"check_interval":      utils.FromIntToTfInt64((*loadBalancers)[i].HealthCheck.CheckInterval),
		"healthy_threshold":   utils.FromIntToTfInt64((*loadBalancers)[i].HealthCheck.HealthyThreshold),
		"path":                types.StringPointerValue((*loadBalancers)[i].HealthCheck.Path),
		"port":                utils.FromIntToTfInt64((*loadBalancers)[i].HealthCheck.Port),
		"protocol":            types.StringValue((*loadBalancers)[i].HealthCheck.Protocol),
		"timeout":             utils.FromIntToTfInt64((*loadBalancers)[i].HealthCheck.Timeout),
		"unhealthy_threshold": utils.FromIntToTfInt64((*loadBalancers)[i].HealthCheck.UnhealthyThreshold),
	})
	if mappingDiags.HasError() {
		diags.Append(mappingDiags...)
	}

	return elementValue, mappingDiags
}

func mappingLoadBalancerTags(ctx context.Context, loadBalancers *[]api.LoadBalancer, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	lt := len(*(*loadBalancers)[i].Tags)
	elementValue := make([]datasource_load_balancer.TagsValue, lt)
	for y, tag := range *(*loadBalancers)[i].Tags {
		elementValue[y], *diags = datasource_load_balancer.NewTagsValue(datasource_load_balancer.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"key":   types.StringValue(tag.Key),
			"value": types.StringValue(tag.Value),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_load_balancer.TagsValue).Type(ctx), elementValue)
}

func mappingListeners(ctx context.Context, loadBalancers *[]api.LoadBalancer, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	var mappingDiags diag.Diagnostics

	ll := len(*(*loadBalancers)[i].Listeners)
	elementValue := make([]datasource_load_balancer.ListenersValue, ll)
	policyNamesList := types.ListNull(types.String{}.Type(ctx))

	for y, listener := range *(*loadBalancers)[i].Listeners {

		if listener.PolicyNames != nil {
			policyNamesList, mappingDiags = types.ListValueFrom(ctx, types.StringType, listener.PolicyNames)
			diags.Append(mappingDiags...)
		}

		elementValue[y], *diags = datasource_load_balancer.NewListenersValue(datasource_load_balancer.ListenersValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"backend_port":           types.Int64Value(utils.ConvertIntPtrToInt64(listener.BackendPort)),
			"backend_protocol":       types.StringPointerValue(listener.BackendProtocol),
			"load_balancer_port":     types.Int64Value(utils.ConvertIntPtrToInt64(listener.LoadBalancerPort)),
			"load_balancer_protocol": types.StringPointerValue(listener.LoadBalancerProtocol),
			"policy_names":           policyNamesList,
			"server_certificate_id":  types.StringPointerValue(listener.ServerCertificateId),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_load_balancer.ListenersValue).Type(ctx), elementValue)
}

func mappingStickyCookiePolicies(ctx context.Context, loadBalancers *[]api.LoadBalancer, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	ls := len(*(*loadBalancers)[i].StickyCookiePolicies)
	elementValue := make([]datasource_load_balancer.StickyCookiePoliciesValue, ls)
	for y, stickyCookiePolicy := range *(*loadBalancers)[i].StickyCookiePolicies {
		elementValue[y], *diags = datasource_load_balancer.NewStickyCookiePoliciesValue(datasource_load_balancer.StickyCookiePoliciesValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"cookie_expiration_period": types.Int64Value(utils.ConvertIntPtrToInt64(stickyCookiePolicy.CookieExpirationPeriod)),
			"policy_name":              types.StringPointerValue(stickyCookiePolicy.PolicyName),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_load_balancer.StickyCookiePoliciesValue).Type(ctx), elementValue)
}
