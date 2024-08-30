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
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func applicationStickyCookiePoliciesFromHTTP(ctx context.Context, elt numspot.ApplicationStickyCookiePolicy) (ApplicationStickyCookiePoliciesValue, diag.Diagnostics) {
	return NewApplicationStickyCookiePoliciesValue(
		ApplicationStickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_name": types.StringPointerValue(elt.CookieName),
			"policy_name": types.StringPointerValue(elt.PolicyName),
		})
}

func listenersFromHTTP(ctx context.Context, elt numspot.Listener) (ListenersValue, diag.Diagnostics) {
	tfPolicyNames, diags := utils2.FromStringListPointerToTfStringList(ctx, elt.PolicyNames)
	if diags.HasError() {
		return ListenersValue{}, diags
	}
	return NewListenersValue(
		ListenersValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"backend_port":           utils2.FromIntPtrToTfInt64(elt.BackendPort),
			"backend_protocol":       types.StringPointerValue(elt.BackendProtocol),
			"load_balancer_port":     utils2.FromIntPtrToTfInt64(elt.LoadBalancerPort),
			"load_balancer_protocol": types.StringPointerValue(elt.BackendProtocol),
			"policy_names":           tfPolicyNames,
			"server_certificate_id":  types.StringPointerValue(elt.BackendProtocol),
		})
}

func stickyCookiePoliciesFromHTTP(ctx context.Context, elt numspot.LoadBalancerStickyCookiePolicy) (StickyCookiePoliciesValue, diag.Diagnostics) {
	return NewStickyCookiePoliciesValue(
		StickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_expiration_period": utils2.FromIntPtrToTfInt64(elt.CookieExpirationPeriod),
			"policy_name":              types.StringPointerValue(elt.PolicyName),
		})
}

func LoadBalancerFromTfToHttp(tf *LoadBalancerModel) *numspot.LoadBalancer {
	return &numspot.LoadBalancer{}
}

func LoadBalancerFromHttpToTf(ctx context.Context, http *numspot.LoadBalancer) (*LoadBalancerModel, diag.Diagnostics) {
	var (
		diags  diag.Diagnostics
		tagsTf types.List
	)

	applicationStickyCookiePoliciestypes, diags := utils2.GenericListToTfListValue(ctx, ApplicationStickyCookiePoliciesValue{}, applicationStickyCookiePoliciesFromHTTP, *http.ApplicationStickyCookiePolicies)
	if diags.HasError() {
		return nil, diags
	}

	listeners, diags := utils2.GenericSetToTfSetValue(ctx, ListenersValue{}, listenersFromHTTP, *http.Listeners)
	if diags.HasError() {
		return nil, diags
	}

	stickyCookiePolicies, diags := utils2.GenericListToTfListValue(ctx, StickyCookiePoliciesValue{}, stickyCookiePoliciesFromHTTP, *http.StickyCookiePolicies)
	if diags.HasError() {
		return nil, diags
	}

	backendIps, diags := utils2.FromStringListPointerToTfStringSet(ctx, http.BackendIps)
	if diags.HasError() {
		return nil, diags
	}

	backendVmIds, diags := utils2.FromStringListPointerToTfStringSet(ctx, http.BackendVmIds)
	if diags.HasError() {
		return nil, diags
	}

	if http.Tags != nil {
		tagsTf, diags = utils2.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	healthCheck, diags := NewHealthCheckValue(HealthCheckValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"check_interval":      utils2.FromIntToTfInt64(http.HealthCheck.CheckInterval),
			"healthy_threshold":   utils2.FromIntToTfInt64(http.HealthCheck.HealthyThreshold),
			"path":                types.StringPointerValue(http.HealthCheck.Path),
			"port":                utils2.FromIntToTfInt64(http.HealthCheck.Port),
			"protocol":            types.StringValue(http.HealthCheck.Protocol),
			"timeout":             utils2.FromIntToTfInt64(http.HealthCheck.Timeout),
			"unhealthy_threshold": utils2.FromIntToTfInt64(http.HealthCheck.UnhealthyThreshold),
		})

	if diags.HasError() {
		return nil, diags
	}

	securityGroups, diags := utils2.FromStringListPointerToTfStringList(ctx, http.SecurityGroups)
	if diags.HasError() {
		return nil, diags
	}

	sourceSecurityGroup := SourceSecurityGroupValue{
		SecurityGroupName: types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupName),
	}

	subnets, diags := utils2.FromStringListPointerToTfStringList(ctx, http.Subnets)
	if diags.HasError() {
		return nil, diags
	}

	azNames, diags := utils2.FromStringListPointerToTfStringList(ctx, http.AvailabilityZoneNames)
	if diags.HasError() {
		return nil, diags
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
	}, diags
}

func LoadBalancerFromHttpToTfDatasource(ctx context.Context, http *numspot.LoadBalancer) (*LoadBalancerModelDatasource, diag.Diagnostics) {
	var (
		tagsList types.List
		diags    diag.Diagnostics
	)

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	applicationStickyCookiePoliciestypes, diags := utils2.GenericListToTfListValue(ctx, ApplicationStickyCookiePoliciesValue{}, applicationStickyCookiePoliciesFromHTTP, *http.ApplicationStickyCookiePolicies)
	if diags.HasError() {
		return nil, diags
	}

	listeners, diags := utils2.GenericSetToTfSetValue(ctx, ListenersValue{}, listenersFromHTTP, *http.Listeners)
	if diags.HasError() {
		return nil, diags
	}

	stickyCookiePolicies, diags := utils2.GenericListToTfListValue(ctx, StickyCookiePoliciesValue{}, stickyCookiePoliciesFromHTTP, *http.StickyCookiePolicies)
	if diags.HasError() {
		return nil, diags
	}

	backendIps, _ := utils2.FromStringListPointerToTfStringSet(ctx, http.BackendIps)
	backendVmIds, _ := utils2.FromStringListPointerToTfStringSet(ctx, http.BackendVmIds)
	healthCheck, err := NewHealthCheckValue(HealthCheckValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"check_interval":      utils2.FromIntToTfInt64(http.HealthCheck.CheckInterval),
			"healthy_threshold":   utils2.FromIntToTfInt64(http.HealthCheck.HealthyThreshold),
			"path":                types.StringPointerValue(http.HealthCheck.Path),
			"port":                utils2.FromIntToTfInt64(http.HealthCheck.Port),
			"protocol":            types.StringValue(http.HealthCheck.Protocol),
			"timeout":             utils2.FromIntToTfInt64(http.HealthCheck.Timeout),
			"unhealthy_threshold": utils2.FromIntToTfInt64(http.HealthCheck.UnhealthyThreshold),
		})
	if err != nil {
		return nil, diags
	}
	securityGroups, _ := utils2.FromStringListPointerToTfStringList(ctx, http.SecurityGroups)
	sourceSecurityGroup := SourceSecurityGroupValue{
		SecurityGroupName: types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupName),
	}
	subnets, _ := utils2.FromStringListPointerToTfStringList(ctx, http.Subnets)
	azNames, _ := utils2.FromStringListPointerToTfStringList(ctx, http.AvailabilityZoneNames)

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
	}, nil
}

func LoadBalancerFromTfToCreateRequest(ctx context.Context, tf *LoadBalancerModel) numspot.CreateLoadBalancerJSONRequestBody {
	var securityGroupsPtr *[]string
	if !(tf.SecurityGroups.IsNull() || tf.SecurityGroups.IsUnknown()) {
		securityGroups := utils2.TfStringListToStringList(ctx, tf.SecurityGroups)
		securityGroupsPtr = &securityGroups
	}
	subnets := utils2.TfStringListToStringList(ctx, tf.Subnets)
	listeners := utils2.TfSetToGenericList(func(a ListenersValue) numspot.ListenerForCreation {
		return numspot.ListenerForCreation{
			BackendPort:          utils2.FromTfInt64ToInt(a.BackendPort),
			BackendProtocol:      a.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils2.FromTfInt64ToInt(a.LoadBalancerPort),
			LoadBalancerProtocol: a.LoadBalancerProtocol.ValueString(),
		}
	}, ctx, tf.Listeners)

	return numspot.CreateLoadBalancerJSONRequestBody{
		Listeners:      listeners,
		Name:           tf.Name.ValueString(),
		PublicIp:       tf.PublicIp.ValueStringPointer(),
		SecurityGroups: securityGroupsPtr,
		Subnets:        subnets,
		Type:           tf.Type.ValueStringPointer(),
	}
}

func LoadBalancerFromTfToUpdateRequest(ctx context.Context, tf *LoadBalancerModel) numspot.UpdateLoadBalancerJSONRequestBody {
	var (
		loadBalancerPort *int                 = nil
		policyNames      *[]string            = nil
		hc               *numspot.HealthCheck = nil
		publicIp         *string              = nil
		securedCookies   *bool                = nil
	)

	if !tf.HealthCheck.IsUnknown() {
		hc = &numspot.HealthCheck{
			CheckInterval:      utils2.FromTfInt64ToInt(tf.HealthCheck.CheckInterval),
			HealthyThreshold:   utils2.FromTfInt64ToInt(tf.HealthCheck.HealthyThreshold),
			Path:               tf.HealthCheck.Path.ValueStringPointer(),
			Port:               utils2.FromTfInt64ToInt(tf.HealthCheck.Port),
			Protocol:           tf.HealthCheck.Protocol.ValueString(),
			Timeout:            utils2.FromTfInt64ToInt(tf.HealthCheck.Timeout),
			UnhealthyThreshold: utils2.FromTfInt64ToInt(tf.HealthCheck.UnhealthyThreshold),
		}
	}
	if !tf.PublicIp.IsUnknown() {
		publicIp = tf.PublicIp.ValueStringPointer()
	}
	if !tf.SecuredCookies.IsUnknown() {
		securedCookies = tf.SecuredCookies.ValueBoolPointer()
	}
	securityGroups := utils2.TfStringListToStringList(ctx, tf.SecurityGroups)
	listeners := utils2.TfSetToGenericSet(func(elt ListenersValue) numspot.Listener {
		policyNames := utils2.TfStringListToStringList(ctx, elt.PolicyNames)
		return numspot.Listener{
			BackendPort:          utils2.FromTfInt64ToIntPtr(elt.BackendPort),
			BackendProtocol:      elt.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils2.FromTfInt64ToIntPtr(elt.LoadBalancerPort),
			LoadBalancerProtocol: elt.BackendProtocol.ValueStringPointer(),
			PolicyNames:          &policyNames,
			ServerCertificateId:  elt.ServerCertificateId.ValueStringPointer(),
		}
	}, ctx, tf.Listeners)

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

	_ = utils2.ExecuteRequest(func() (*numspot.CreateLoadBalancerTagsResponse, error) {
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

	_ = utils2.ExecuteRequest(func() (*numspot.DeleteLoadBalancerTagsResponse, error) {
		return iaasClient.DeleteLoadBalancerTagsWithResponse(ctx, spaceId, numspot.DeleteLoadBalancerTagsJSONRequestBody{
			Names: []string{loadBalancerName},
			Tags:  apiTags,
		})
	}, http.StatusNoContent, diags)
}
