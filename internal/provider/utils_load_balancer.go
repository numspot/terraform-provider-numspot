package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_load_balancer"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func LoadBalancerFromTfToHttp(tf *resource_load_balancer.LoadBalancerModel) *api.LoadBalancerSchema {
	return &api.LoadBalancerSchema{}
}

func LoadBalancerFromHttpToTf(ctx context.Context, http *api.LoadBalancerSchema) resource_load_balancer.LoadBalancerModel {
	//TODO: handle returned diags instead of surpassing it => x, diags := types.ListValueFrom()
	applicationStickyCookiePoliciestypes, _ := types.ListValueFrom(ctx, resource_load_balancer.ApplicationStickyCookiePoliciesType{}, http.ApplicationStickyCookiePolicies)
	backendIps, _ := utils.FromStringListToTfStringList(ctx, *http.BackendIps)
	backendVmIds, _ := utils.FromStringListToTfStringList(ctx, *http.BackendVmIds)
	healthCheck := resource_load_balancer.HealthCheckValue{
		CheckInterval:      utils.FromIntToTfInt64(http.HealthCheck.CheckInterval),
		HealthyThreshold:   utils.FromIntToTfInt64(http.HealthCheck.HealthyThreshold),
		Path:               types.StringValue(*http.HealthCheck.Path),
		Port:               utils.FromIntToTfInt64(http.HealthCheck.Port),
		Protocol:           types.StringValue(http.HealthCheck.Protocol),
		Timeout:            utils.FromIntToTfInt64(http.HealthCheck.Timeout),
		UnhealthyThreshold: utils.FromIntToTfInt64(http.HealthCheck.UnhealthyThreshold),
	}
	listeners, _ := types.ListValueFrom(ctx, resource_load_balancer.ListenersType{}, http.Listeners)
	securityGroups, _ := types.ListValueFrom(ctx, resource_load_balancer.SourceSecurityGroupType{}, http.SecurityGroups)
	sourceSecurityGroup := resource_load_balancer.SourceSecurityGroupValue{
		SecurityGroupAccountId: types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupAccountId),
		SecurityGroupName:      types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupName),
	}
	stickyCookiePolicies, _ := types.ListValueFrom(ctx, resource_load_balancer.StickyCookiePoliciesType{}, http.StickyCookiePolicies)
	subnets, _ := utils.FromStringListToTfStringList(ctx, *http.Subnets)
	subregionNames, _ := utils.FromStringListToTfStringList(ctx, *http.SubregionNames)

	return resource_load_balancer.LoadBalancerModel{
		ApplicationStickyCookiePolicies: applicationStickyCookiePoliciestypes,
		BackendIps:                      backendIps,
		BackendVmIds:                    backendVmIds,
		DnsName:                         types.StringPointerValue(http.DnsName),
		HealthCheck:                     healthCheck,
		Id:                              types.StringPointerValue(http.Name),
		Listeners:                       listeners,
		Name:                            types.StringPointerValue(http.Name),
		NetId:                           types.StringPointerValue(http.NetId),
		PublicIp:                        types.StringPointerValue(http.PublicIp),
		SecuredCookies:                  types.BoolPointerValue(http.SecuredCookies),
		SecurityGroups:                  securityGroups,
		SourceSecurityGroup:             sourceSecurityGroup,
		StickyCookiePolicies:            stickyCookiePolicies,
		Subnets:                         subnets,
		SubregionNames:                  subregionNames,
		Type:                            types.StringPointerValue(http.Type),
	}
}

func LoadBalancerFromTfToCreateRequest(ctx context.Context, tf *resource_load_balancer.LoadBalancerModel) api.CreateLoadBalancerJSONRequestBody {

	securityGroups := utils.TfStringListToStringList(ctx, tf.SecurityGroups)
	subnets := utils.TfStringListToStringList(ctx, tf.Subnets)
	listeners := utils.TfListToGenericList(func(a resource_load_balancer.ListenersValue) api.ListenerForCreationSchema {
		return api.ListenerForCreationSchema{
			BackendPort:          utils.FromTfInt64ToInt(a.BackendPort),
			BackendProtocol:      a.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils.FromTfInt64ToInt(a.LoadBalancerPort),
			LoadBalancerProtocol: a.LoadBalancerProtocol.ValueString(),
		}
	}, ctx, tf.Listeners)

	return api.CreateLoadBalancerJSONRequestBody{
		Listeners:      listeners,
		Name:           tf.Name.ValueString(),
		PublicIp:       tf.PublicIp.ValueStringPointer(),
		SecurityGroups: &securityGroups,
		Subnets:        &subnets,
		//Tags:           nil,
		Type: tf.Type.ValueStringPointer(),
	}
}
