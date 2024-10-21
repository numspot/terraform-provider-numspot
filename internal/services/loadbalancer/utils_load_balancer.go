package loadbalancer

import (
	"context"
	"net/http"

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

func LoadBalancerFromHttpToTf(ctx context.Context, http *numspot.LoadBalancer, diags *diag.Diagnostics) *LoadBalancerModel {
	var tagsTf types.List

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
	if diags.HasError() {
		return nil
	}

	backendVmIds := utils.FromStringListPointerToTfStringSet(ctx, http.BackendVmIds, diags)
	if diags.HasError() {
		return nil
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

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
	if diags.HasError() {
		return nil
	}

	securityGroups := utils.FromStringListPointerToTfStringList(ctx, http.SecurityGroups, diags)
	if diags.HasError() {
		return nil
	}

	sourceSecurityGroup := SourceSecurityGroupValue{
		SecurityGroupName: types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupName),
	}

	subnets := utils.FromStringListPointerToTfStringList(ctx, http.Subnets, diags)
	if diags.HasError() {
		return nil
	}

	azNames := utils.FromStringListPointerToTfStringList(ctx, http.AvailabilityZoneNames, diags)
	if diags.HasError() {
		return nil
	}

	return &LoadBalancerModel{
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
		Tags:                            tagsTf,
	}
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

func LoadBalancerFromTfToCreateRequest(ctx context.Context, tf *LoadBalancerModel, diags *diag.Diagnostics) numspot.CreateLoadBalancerJSONRequestBody {
	var securityGroupsPtr *[]string
	if !(tf.SecurityGroups.IsNull() || tf.SecurityGroups.IsUnknown()) {
		securityGroups := utils.TfStringListToStringList(ctx, tf.SecurityGroups, diags)
		securityGroupsPtr = &securityGroups
	}
	subnets := utils.TfStringListToStringList(ctx, tf.Subnets, diags)
	listeners := utils.TfSetToGenericList(func(a ListenersValue) numspot.ListenerForCreation {
		return numspot.ListenerForCreation{
			BackendPort:          utils.FromTfInt64ToInt(a.BackendPort),
			BackendProtocol:      a.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils.FromTfInt64ToInt(a.LoadBalancerPort),
			LoadBalancerProtocol: a.LoadBalancerProtocol.ValueString(),
		}
	}, ctx, tf.Listeners, diags)

	return numspot.CreateLoadBalancerJSONRequestBody{
		Listeners:      listeners,
		Name:           tf.Name.ValueString(),
		PublicIp:       tf.PublicIp.ValueStringPointer(),
		SecurityGroups: securityGroupsPtr,
		Subnets:        subnets,
		Type:           tf.Type.ValueStringPointer(),
	}
}

func LoadBalancerFromTfToUpdateRequest(ctx context.Context, tf *LoadBalancerModel, diags *diag.Diagnostics) numspot.UpdateLoadBalancerJSONRequestBody {
	var (
		loadBalancerPort *int                 = nil
		policyNames      *[]string            = nil
		hc               *numspot.HealthCheck = nil
		publicIp         *string              = nil
		securedCookies   *bool                = nil
	)

	if !tf.HealthCheck.IsUnknown() {
		hc = &numspot.HealthCheck{
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
	securityGroups := utils.TfStringListToStringList(ctx, tf.SecurityGroups, diags)
	listeners := utils.TfSetToGenericSet(func(elt ListenersValue) numspot.Listener {
		policyNames := utils.TfStringListToStringList(ctx, elt.PolicyNames, diags)
		return numspot.Listener{
			BackendPort:          utils.FromTfInt64ToIntPtr(elt.BackendPort),
			BackendProtocol:      elt.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils.FromTfInt64ToIntPtr(elt.LoadBalancerPort),
			LoadBalancerProtocol: elt.BackendProtocol.ValueStringPointer(),
			PolicyNames:          &policyNames,
			ServerCertificateId:  elt.ServerCertificateId.ValueStringPointer(),
		}
	}, ctx, tf.Listeners, diags)

	if len(listeners) == 1 {
		loadBalancerPort = listeners[0].LoadBalancerPort
		policyNames = listeners[0].PolicyNames
	}

	return numspot.UpdateLoadBalancerJSONRequestBody{
		HealthCheck:      hc,
		LoadBalancerPort: loadBalancerPort,
		PolicyNames:      policyNames,
		PublicIp:         publicIp,
		SecuredCookies:   securedCookies,
		SecurityGroups:   &securityGroups,
	}
}

func CreateLoadBalancerTags(
	ctx context.Context,
	spaceId numspot.SpaceId,
	iaasClient *numspot.ClientWithResponses,
	loadBalancerName string,
	tagList types.List,
	diags *diag.Diagnostics,
) {
	tfTags := make([]tags.TagsValue, 0, len(tagList.Elements()))
	tagList.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]numspot.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, numspot.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	_ = utils.ExecuteRequest(func() (*numspot.CreateLoadBalancerTagsResponse, error) {
		return iaasClient.CreateLoadBalancerTagsWithResponse(ctx, spaceId, numspot.CreateLoadBalancerTagsRequest{
			Names: []string{loadBalancerName},
			Tags:  apiTags,
		})
	}, http.StatusNoContent, diags)
}

func DeleteLoadBalancerTags(
	ctx context.Context,
	spaceId numspot.SpaceId,
	iaasClient *numspot.ClientWithResponses,
	loadBalancerName string,
	tagList types.List,
	diags *diag.Diagnostics,
) {
	tfTags := make([]tags.TagsValue, 0, len(tagList.Elements()))
	tagList.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]numspot.ResourceLoadBalancerTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, numspot.ResourceLoadBalancerTag{
			Key: tfTag.Key.ValueStringPointer(),
		})
	}

	_ = utils.ExecuteRequest(func() (*numspot.DeleteLoadBalancerTagsResponse, error) {
		return iaasClient.DeleteLoadBalancerTagsWithResponse(ctx, spaceId, numspot.DeleteLoadBalancerTagsJSONRequestBody{
			Names: []string{loadBalancerName},
			Tags:  apiTags,
		})
	}, http.StatusNoContent, diags)
}
