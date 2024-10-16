package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func applicationStickyCookiePoliciesFromHTTP(ctx context.Context, elt numspot.ApplicationStickyCookiePolicy, diags *diag.Diagnostics) ApplicationStickyCookiePoliciesValue {
	value, diagnostics := NewApplicationStickyCookiePoliciesValue(
		ApplicationStickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_name": types.StringPointerValue(elt.CookieName),
			"policy_name": types.StringPointerValue(elt.PolicyName),
		})
	diags.Append(diagnostics...)
	return value
}

func listenersFromHTTP(ctx context.Context, elt numspot.Listener, diags *diag.Diagnostics) ListenersValue {
	tfPolicyNames := utils.FromStringListPointerToTfStringList(ctx, elt.PolicyNames, diags)
	if diags.HasError() {
		return ListenersValue{}
	}
	value, diagnostics := NewListenersValue(
		ListenersValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"backend_port":           utils.FromIntPtrToTfInt64(elt.BackendPort),
			"backend_protocol":       types.StringPointerValue(elt.BackendProtocol),
			"load_balancer_port":     utils.FromIntPtrToTfInt64(elt.LoadBalancerPort),
			"load_balancer_protocol": types.StringPointerValue(elt.BackendProtocol),
			"policy_names":           tfPolicyNames,
			"server_certificate_id":  types.StringPointerValue(elt.BackendProtocol),
		})
	diags.Append(diagnostics...)
	return value
}

func stickyCookiePoliciesFromHTTP(ctx context.Context, elt numspot.LoadBalancerStickyCookiePolicy, diags *diag.Diagnostics) StickyCookiePoliciesValue {
	value, diagnostics := NewStickyCookiePoliciesValue(
		StickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_expiration_period": utils.FromIntPtrToTfInt64(elt.CookieExpirationPeriod),
			"policy_name":              types.StringPointerValue(elt.PolicyName),
		})
	diags.Append(diagnostics...)
	return value
}

func LoadBalancerFromTfToHttp(tf *LoadBalancerModel) *numspot.LoadBalancer {
	return &numspot.LoadBalancer{}
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
