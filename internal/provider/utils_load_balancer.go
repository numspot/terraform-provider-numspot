package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_load_balancer"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_load_balancer"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func applicationStickyCookiePoliciesFromHTTP(ctx context.Context, elt iaas.ApplicationStickyCookiePolicy) (resource_load_balancer.ApplicationStickyCookiePoliciesValue, diag.Diagnostics) {
	return resource_load_balancer.NewApplicationStickyCookiePoliciesValue(
		resource_load_balancer.ApplicationStickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_name": types.StringPointerValue(elt.CookieName),
			"policy_name": types.StringPointerValue(elt.PolicyName),
		})
}

func listenersFromHTTP(ctx context.Context, elt iaas.Listener) (resource_load_balancer.ListenersValue, diag.Diagnostics) {
	tfPolicyNames, diags := utils.FromStringListPointerToTfStringList(ctx, elt.PolicyNames)
	if diags.HasError() {
		return resource_load_balancer.ListenersValue{}, diags
	}
	return resource_load_balancer.NewListenersValue(
		resource_load_balancer.ListenersValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"backend_port":           utils.FromIntPtrToTfInt64(elt.BackendPort),
			"backend_protocol":       types.StringPointerValue(elt.BackendProtocol),
			"load_balancer_port":     utils.FromIntPtrToTfInt64(elt.LoadBalancerPort),
			"load_balancer_protocol": types.StringPointerValue(elt.BackendProtocol),
			"policy_names":           tfPolicyNames,
			"server_certificate_id":  types.StringPointerValue(elt.BackendProtocol),
		})
}

func stickyCookiePoliciesFromHTTP(ctx context.Context, elt iaas.LoadBalancerStickyCookiePolicy) (resource_load_balancer.StickyCookiePoliciesValue, diag.Diagnostics) {
	return resource_load_balancer.NewStickyCookiePoliciesValue(
		resource_load_balancer.StickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_expiration_period": utils.FromIntPtrToTfInt64(elt.CookieExpirationPeriod),
			"policy_name":              types.StringPointerValue(elt.PolicyName),
		})
}

func LoadBalancerFromTfToHttp(tf *resource_load_balancer.LoadBalancerModel) *iaas.LoadBalancer {
	return &iaas.LoadBalancer{}
}

func LoadBalancerFromHttpToTf(ctx context.Context, http *iaas.LoadBalancer) resource_load_balancer.LoadBalancerModel {
	applicationStickyCookiePoliciestypes, diags := utils.GenericListToTfListValue(ctx, resource_load_balancer.ApplicationStickyCookiePoliciesValue{}, applicationStickyCookiePoliciesFromHTTP, *http.ApplicationStickyCookiePolicies)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}

	listeners, diags := utils.GenericListToTfListValue(ctx, resource_load_balancer.ListenersValue{}, listenersFromHTTP, *http.Listeners)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}

	stickyCookiePolicies, diags := utils.GenericListToTfListValue(ctx, resource_load_balancer.StickyCookiePoliciesValue{}, stickyCookiePoliciesFromHTTP, *http.StickyCookiePolicies)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}

	backendIps, _ := utils.FromStringListPointerToTfStringList(ctx, http.BackendIps)
	backendVmIds, _ := utils.FromStringListPointerToTfStringList(ctx, http.BackendVmIds)
	healthCheck, err := resource_load_balancer.NewHealthCheckValue(resource_load_balancer.HealthCheckValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"check_interval":      utils.FromIntToTfInt64(http.HealthCheck.CheckInterval),
			"healthy_threshold":   utils.FromIntToTfInt64(http.HealthCheck.HealthyThreshold),
			"path":                types.StringPointerValue(http.HealthCheck.Path),
			"port":                utils.FromIntToTfInt64(http.HealthCheck.Port),
			"protocol":            types.StringValue(http.HealthCheck.Protocol),
			"timeout":             utils.FromIntToTfInt64(http.HealthCheck.Timeout),
			"unhealthy_threshold": utils.FromIntToTfInt64(http.HealthCheck.UnhealthyThreshold),
		})
	if err != nil {
		return resource_load_balancer.LoadBalancerModel{}
	}
	// httpListeners := *http.Listeners
	securityGroups, _ := utils.FromStringListPointerToTfStringList(ctx, http.SecurityGroups)
	sourceSecurityGroup := resource_load_balancer.SourceSecurityGroupValue{
		SecurityGroupAccountId: types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupAccountId),
		SecurityGroupName:      types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupName),
	}
	subnets, _ := utils.FromStringListPointerToTfStringList(ctx, http.Subnets)
	azNames, _ := utils.FromStringListPointerToTfStringList(ctx, http.AvailabilityZoneNames)

	return resource_load_balancer.LoadBalancerModel{
		ApplicationStickyCookiePolicies: applicationStickyCookiePoliciestypes,
		BackendIps:                      backendIps,
		BackendVmIds:                    backendVmIds,
		DnsName:                         types.StringPointerValue(http.DnsName),
		HealthCheck:                     healthCheck,
		Id:                              types.StringPointerValue(http.Name),
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
		Type:                            types.StringPointerValue(http.Type),
	}
}

func LoadBalancerFromHttpToTfDatasource(ctx context.Context, http *iaas.LoadBalancer) datasource_load_balancer.LoadBalancerModel {
	applicationStickyCookiePoliciestypes, diags := utils.GenericListToTfListValue(ctx, resource_load_balancer.ApplicationStickyCookiePoliciesValue{}, applicationStickyCookiePoliciesFromHTTP, *http.ApplicationStickyCookiePolicies)
	if diags.HasError() {
		return datasource_load_balancer.LoadBalancerModel{}
	}

	listeners, diags := utils.GenericListToTfListValue(ctx, resource_load_balancer.ListenersValue{}, listenersFromHTTP, *http.Listeners)
	if diags.HasError() {
		return datasource_load_balancer.LoadBalancerModel{}
	}

	stickyCookiePolicies, diags := utils.GenericListToTfListValue(ctx, resource_load_balancer.StickyCookiePoliciesValue{}, stickyCookiePoliciesFromHTTP, *http.StickyCookiePolicies)
	if diags.HasError() {
		return datasource_load_balancer.LoadBalancerModel{}
	}

	backendIps, _ := utils.FromStringListPointerToTfStringList(ctx, http.BackendIps)
	backendVmIds, _ := utils.FromStringListPointerToTfStringList(ctx, http.BackendVmIds)
	healthCheck, err := datasource_load_balancer.NewHealthCheckValue(resource_load_balancer.HealthCheckValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"check_interval":      utils.FromIntToTfInt64(http.HealthCheck.CheckInterval),
			"healthy_threshold":   utils.FromIntToTfInt64(http.HealthCheck.HealthyThreshold),
			"path":                types.StringPointerValue(http.HealthCheck.Path),
			"port":                utils.FromIntToTfInt64(http.HealthCheck.Port),
			"protocol":            types.StringValue(http.HealthCheck.Protocol),
			"timeout":             utils.FromIntToTfInt64(http.HealthCheck.Timeout),
			"unhealthy_threshold": utils.FromIntToTfInt64(http.HealthCheck.UnhealthyThreshold),
		})
	if err != nil {
		return datasource_load_balancer.LoadBalancerModel{}
	}
	// httpListeners := *http.Listeners
	securityGroups, _ := utils.FromStringListPointerToTfStringList(ctx, http.SecurityGroups)
	sourceSecurityGroup := datasource_load_balancer.SourceSecurityGroupValue{
		SecurityGroupAccountId: types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupAccountId),
		SecurityGroupName:      types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupName),
	}
	subnets, _ := utils.FromStringListPointerToTfStringList(ctx, http.Subnets)
	azNames, _ := utils.FromStringListPointerToTfStringList(ctx, http.AvailabilityZoneNames)

	return datasource_load_balancer.LoadBalancerModel{
		ApplicationStickyCookiePolicies: applicationStickyCookiePoliciestypes,
		BackendIps:                      backendIps,
		BackendVmIds:                    backendVmIds,
		DnsName:                         types.StringPointerValue(http.DnsName),
		HealthCheck:                     healthCheck,
		Id:                              types.StringPointerValue(http.Name),
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
		Type:                            types.StringPointerValue(http.Type),
	}
}

func LoadBalancerFromTfToCreateRequest(ctx context.Context, tf *resource_load_balancer.LoadBalancerModel) iaas.CreateLoadBalancerJSONRequestBody {
	securityGroups := utils.TfStringListToStringList(ctx, tf.SecurityGroups)
	subnets := utils.TfStringListToStringList(ctx, tf.Subnets)
	listeners := utils.TfListToGenericList(func(a resource_load_balancer.ListenersValue) iaas.ListenerForCreation {
		return iaas.ListenerForCreation{
			BackendPort:          utils.FromTfInt64ToInt(a.BackendPort),
			BackendProtocol:      a.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils.FromTfInt64ToInt(a.LoadBalancerPort),
			LoadBalancerProtocol: a.LoadBalancerProtocol.ValueString(),
		}
	}, ctx, tf.Listeners)

	return iaas.CreateLoadBalancerJSONRequestBody{
		Listeners:      listeners,
		Name:           tf.Name.ValueString(),
		PublicIp:       tf.PublicIp.ValueStringPointer(),
		SecurityGroups: &securityGroups,
		Subnets:        &subnets,
		// Tags:           nil,
		Type: tf.Type.ValueStringPointer(),
	}
}

func LoadBalancerFromTfToUpdateRequest(ctx context.Context, tf *resource_load_balancer.LoadBalancerModel) iaas.UpdateLoadBalancerJSONRequestBody {
	var (
		loadBalancerPort *int              = nil
		policyNames      *[]string         = nil
		hc               *iaas.HealthCheck = nil
		publicIp         *string           = nil
		securedCookies   *bool             = nil
	)

	if !tf.HealthCheck.IsUnknown() {
		hc = &iaas.HealthCheck{
			CheckInterval:      utils.FromTfInt64ToInt(tf.HealthCheck.CheckInterval),
			HealthyThreshold:   utils.FromTfInt64ToInt(tf.HealthCheck.HealthyThreshold),
			Path:               tf.HealthCheck.Path.ValueStringPointer(),
			Port:               utils.FromTfInt64ToInt(tf.HealthCheck.Port),
			Protocol:           tf.HealthCheck.Protocol.ValueString(),
			Timeout:            utils.FromTfInt64ToInt(tf.HealthCheck.Timeout),
			UnhealthyThreshold: utils.FromTfInt64ToInt(tf.HealthCheck.UnhealthyThreshold),
		}
	}
	if !tf.PublicIp.IsUnknown() {
		publicIp = tf.PublicIp.ValueStringPointer()
	}
	if !tf.SecuredCookies.IsUnknown() {
		securedCookies = tf.SecuredCookies.ValueBoolPointer()
	}
	securityGroups := utils.TfStringListToStringList(ctx, tf.SecurityGroups)
	listeners := utils.TfListToGenericList(func(elt resource_load_balancer.ListenersValue) iaas.Listener {
		policyNames := utils.TfStringListToStringList(ctx, elt.PolicyNames)
		return iaas.Listener{
			BackendPort:          utils.FromTfInt64ToIntPtr(elt.BackendPort),
			BackendProtocol:      elt.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils.FromTfInt64ToIntPtr(elt.LoadBalancerPort),
			LoadBalancerProtocol: elt.BackendProtocol.ValueStringPointer(),
			PolicyNames:          &policyNames,
			ServerCertificateId:  elt.ServerCertificateId.ValueStringPointer(),
		}
	}, ctx, tf.Listeners)

	if len(listeners) == 1 {
		loadBalancerPort = listeners[0].LoadBalancerPort
		policyNames = listeners[0].PolicyNames
	}

	return iaas.UpdateLoadBalancerJSONRequestBody{
		HealthCheck:      hc,
		LoadBalancerPort: loadBalancerPort,
		PolicyNames:      policyNames,
		PublicIp:         publicIp,
		SecuredCookies:   securedCookies,
		SecurityGroups:   &securityGroups,
	}
}
