package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/loadbalancer/resource_load_balancer"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &loadBalancerResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerResource{}
	_ resource.ResourceWithImportState = &loadBalancerResource{}
)

type loadBalancerResource struct {
	provider *client.NumSpotSDK
}

func NewLoadBalancerResource() resource.Resource {
	return &loadBalancerResource{}
}

func (r *loadBalancerResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *loadBalancerResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), request, response)
}

func (r *loadBalancerResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_load_balancer"
}

func (r *loadBalancerResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_load_balancer.LoadBalancerResourceSchema(ctx)
}

func (r *loadBalancerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_load_balancer.LoadBalancerModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := loadBalancerTags(ctx, plan.Tags)
	backendIP := utils.FromTfStringSetToStringList(ctx, plan.BackendIps, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	backendVM := utils.FromTfStringSetToStringList(ctx, plan.BackendVmIds, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	updateNumSpotLoadBalancer := deserializeUpdateNumSpotLoadBalancer(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	createNumSpotLoadBalancer := deserializeCreateNumSpotLoadBalancer(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotLoadBalancer, err := core.CreateLoadBalancer(ctx, r.provider, createNumSpotLoadBalancer, updateNumSpotLoadBalancer, tagsValue, backendVM, backendIP)
	if err != nil {
		response.Diagnostics.AddError("unable to create load balancer", err.Error())
		return
	}

	state := serializeNumSpotLoadBalancer(ctx, numSpotLoadBalancer, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *loadBalancerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_load_balancer.LoadBalancerModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	loadBalancerName := state.Name.ValueString()

	numSpotLoadBalancer, err := core.ReadLoadBalancer(ctx, r.provider, loadBalancerName)
	if err != nil {
		response.Diagnostics.AddError("unable to read load balancer", err.Error())
		return
	}

	newState := serializeNumSpotLoadBalancer(ctx, numSpotLoadBalancer, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *loadBalancerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan         resource_load_balancer.LoadBalancerModel
		numSpotLoadBalancer *api.LoadBalancer
		err                 error
	)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	planBackendIP := utils.FromTfStringSetToStringList(ctx, plan.BackendIps, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	planBackendVM := utils.FromTfStringSetToStringList(ctx, plan.BackendVmIds, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	stateBackendIP := utils.FromTfStringSetToStringList(ctx, state.BackendIps, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	stateBackendVM := utils.FromTfStringSetToStringList(ctx, state.BackendVmIds, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	loadBalancerName := state.Name.ValueString()
	updateNumSpotLoadBalancer := deserializeUpdateNumSpotLoadBalancer(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	planTags := loadBalancerTags(ctx, plan.Tags)
	stateTags := deserializeTagsToDelete(ctx, state.Tags)
	statePublicIP := state.PublicIp.ValueString()
	planPublicIP := plan.PublicIp.ValueString()

	if !plan.HealthCheck.Equal(state.HealthCheck) || planPublicIP != statePublicIP {
		numSpotLoadBalancer, err = core.UpdateLoadBalancerAttributes(ctx, r.provider, loadBalancerName, updateNumSpotLoadBalancer)
		if err != nil {
			response.Diagnostics.AddError("unable to update load balancer attributes", err.Error())
			return
		}
	}

	if !plan.SecurityGroups.Equal(state.SecurityGroups) {
		numSpotLoadBalancer, err = core.UpdateLoadBalancerSecurityGroup(ctx, r.provider, loadBalancerName, updateNumSpotLoadBalancer)
		if err != nil {
			response.Diagnostics.AddError("unable to update load balancer security groups", err.Error())
			return
		}
	}

	if !plan.BackendVmIds.Equal(state.BackendVmIds) || !plan.BackendIps.Equal(state.BackendIps) {
		numSpotLoadBalancer, err = core.UpdateLoadBalancerBackend(ctx, r.provider, loadBalancerName, stateBackendVM, planBackendVM, stateBackendIP, planBackendIP)
		if err != nil {
			response.Diagnostics.AddError("unable to update load balancer backend", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numSpotLoadBalancer, err = core.UpdateLoadBalancerTags(ctx, r.provider, loadBalancerName, planTags, stateTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update load balancer tags", err.Error())
			return
		}
	}

	newState := serializeNumSpotLoadBalancer(ctx, numSpotLoadBalancer, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *loadBalancerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_load_balancer.LoadBalancerModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	loadBalancerName := state.Name.ValueString()

	if err := core.DeleteLoadBalancer(ctx, r.provider, loadBalancerName); err != nil {
		response.Diagnostics.AddError("unable to delete load balancer", err.Error())
		return
	}
}

func deserializeCreateNumSpotLoadBalancer(ctx context.Context, tf resource_load_balancer.LoadBalancerModel, diags *diag.Diagnostics) api.CreateLoadBalancerJSONRequestBody {
	var securityGroupsPtr *[]string
	if !(tf.SecurityGroups.IsNull() || tf.SecurityGroups.IsUnknown()) {
		securityGroups := utils.TfStringListToStringList(ctx, tf.SecurityGroups, diags)
		securityGroupsPtr = &securityGroups
	}
	subnets := utils.TfStringListToStringList(ctx, tf.Subnets, diags)
	listeners := utils.TfSetToGenericList(func(a resource_load_balancer.ListenersValue) api.ListenerForCreation {
		var backendProtocol *string
		if !a.BackendProtocol.IsNull() && a.BackendProtocol.ValueString() != "" {
			bp := a.BackendProtocol.ValueString()
			backendProtocol = &bp
		}

		var serverCertificateId *string
		if !a.ServerCertificateId.IsNull() && a.ServerCertificateId.ValueString() != "" {
			sc := a.ServerCertificateId.ValueString()
			serverCertificateId = &sc
		}
		return api.ListenerForCreation{
			BackendPort:          utils.FromTfInt64ToInt(a.BackendPort),
			BackendProtocol:      backendProtocol,
			LoadBalancerPort:     utils.FromTfInt64ToInt(a.LoadBalancerPort),
			LoadBalancerProtocol: a.LoadBalancerProtocol.ValueString(),
			ServerCertificateId:  serverCertificateId,
		}
	}, ctx, tf.Listeners, diags)

	return api.CreateLoadBalancerJSONRequestBody{
		Listeners:      listeners,
		Name:           tf.Name.ValueString(),
		PublicIp:       tf.PublicIp.ValueStringPointer(),
		SecurityGroups: securityGroupsPtr,
		Subnets:        subnets,
		Type:           tf.Type.ValueStringPointer(),
	}
}

func deserializeUpdateNumSpotLoadBalancer(ctx context.Context, tf resource_load_balancer.LoadBalancerModel, diags *diag.Diagnostics) api.UpdateLoadBalancerJSONRequestBody {
	var (
		loadBalancerPort *int = nil
		policyNames           = make([]string, 0)
		hc                    = &api.HealthCheck{
			CheckInterval:      30,
			HealthyThreshold:   10,
			Path:               nil,
			Port:               80,
			Protocol:           "TCP",
			Timeout:            5,
			UnhealthyThreshold: 2,
		} // Default health check
		publicIp       *string   = nil
		securedCookies *bool     = nil
		securityGroups *[]string = nil
	)

	if !tf.HealthCheck.IsUnknown() {
		hc = &api.HealthCheck{
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

	if !tf.SecurityGroups.IsUnknown() {
		sg := utils.TfStringListToStringList(ctx, tf.SecurityGroups, diags)
		securityGroups = &sg
	}

	listeners := utils.TfSetToGenericSet(func(elt resource_load_balancer.ListenersValue) api.Listener {
		policyNamesListener := utils.TfStringListToStringList(ctx, elt.PolicyNames, diags)
		if policyNamesListener == nil {
			policyNamesListener = make([]string, 0)
		}
		return api.Listener{
			BackendPort:          utils.FromTfInt64ToIntPtr(elt.BackendPort),
			BackendProtocol:      elt.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils.FromTfInt64ToIntPtr(elt.LoadBalancerPort),
			LoadBalancerProtocol: elt.BackendProtocol.ValueStringPointer(),
			PolicyNames:          &policyNamesListener,
			ServerCertificateId:  elt.ServerCertificateId.ValueStringPointer(),
		}
	}, ctx, tf.Listeners, diags)

	if len(listeners) == 1 {
		loadBalancerPort = listeners[0].LoadBalancerPort
		policyNames = *listeners[0].PolicyNames
	}

	return api.UpdateLoadBalancerJSONRequestBody{
		HealthCheck:      hc,
		LoadBalancerPort: loadBalancerPort,
		PolicyNames:      &policyNames,
		PublicIp:         publicIp,
		SecuredCookies:   securedCookies,
		SecurityGroups:   securityGroups,
	}
}

func serializeNumSpotLoadBalancer(ctx context.Context, http *api.LoadBalancer, diags *diag.Diagnostics) resource_load_balancer.LoadBalancerModel {
	var tagsTf types.Set

	applicationStickyCookiePoliciesTypes := utils.GenericListToTfListValue(ctx, applicationStickyCookiePoliciesFromHTTP, *http.ApplicationStickyCookiePolicies, diags)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}

	listeners := utils.GenericSetToTfSetValue(ctx, listenersFromHTTP, *http.Listeners, diags)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}

	stickyCookiePolicies := utils.GenericListToTfListValue(ctx, stickyCookiePoliciesFromHTTP, *http.StickyCookiePolicies, diags)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}

	backendIps := utils.FromStringListPointerToTfStringSet(ctx, http.BackendIps, diags)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}

	backendVmIds := utils.FromStringListPointerToTfStringSet(ctx, http.BackendVmIds, diags)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return resource_load_balancer.LoadBalancerModel{}
		}
	}

	healthCheck, diagnostics := resource_load_balancer.NewHealthCheckValue(resource_load_balancer.HealthCheckValue{}.AttributeTypes(ctx),
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
		return resource_load_balancer.LoadBalancerModel{}
	}

	securityGroups := utils.FromStringListPointerToTfStringList(ctx, http.SecurityGroups, diags)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}
	subnets := utils.FromStringListPointerToTfStringList(ctx, http.Subnets, diags)
	if diags.HasError() {
		return resource_load_balancer.LoadBalancerModel{}
	}

	return resource_load_balancer.LoadBalancerModel{
		Id:                              types.StringPointerValue(http.Name),
		ApplicationStickyCookiePolicies: applicationStickyCookiePoliciesTypes,
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
		StickyCookiePolicies:            stickyCookiePolicies,
		Subnets:                         subnets,
		Type:                            types.StringPointerValue(http.Type),
		Tags:                            tagsTf,
	}
}

func deserializeTagsToDelete(ctx context.Context, t types.Set) []api.ResourceLoadBalancerTag {
	lbTags := make([]api.ResourceLoadBalancerTag, len(t.Elements()))
	stateTags := loadBalancerTags(ctx, t)
	for idx, tag := range stateTags {
		key := &tag.Key
		lbTags[idx] = api.ResourceLoadBalancerTag{
			Key: key,
		}
	}
	return lbTags
}

func applicationStickyCookiePoliciesFromHTTP(ctx context.Context, elt api.ApplicationStickyCookiePolicy, diags *diag.Diagnostics) resource_load_balancer.ApplicationStickyCookiePoliciesValue {
	value, diagnostics := resource_load_balancer.NewApplicationStickyCookiePoliciesValue(
		resource_load_balancer.ApplicationStickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_name": types.StringPointerValue(elt.CookieName),
			"policy_name": types.StringPointerValue(elt.PolicyName),
		})
	diags.Append(diagnostics...)
	return value
}

func listenersFromHTTP(ctx context.Context, elt api.Listener, diags *diag.Diagnostics) resource_load_balancer.ListenersValue {
	tfPolicyNames := utils.FromStringListPointerToTfStringList(ctx, elt.PolicyNames, diags)
	if diags.HasError() {
		return resource_load_balancer.ListenersValue{}
	}
	value, diagnostics := resource_load_balancer.NewListenersValue(
		resource_load_balancer.ListenersValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"backend_port":           utils.FromIntPtrToTfInt64(elt.BackendPort),
			"backend_protocol":       types.StringPointerValue(elt.BackendProtocol),
			"load_balancer_port":     utils.FromIntPtrToTfInt64(elt.LoadBalancerPort),
			"load_balancer_protocol": types.StringPointerValue(elt.BackendProtocol),
			"policy_names":           tfPolicyNames,
			"server_certificate_id":  types.StringPointerValue(elt.ServerCertificateId),
		})
	diags.Append(diagnostics...)
	return value
}

func stickyCookiePoliciesFromHTTP(ctx context.Context, elt api.LoadBalancerStickyCookiePolicy, diags *diag.Diagnostics) resource_load_balancer.StickyCookiePoliciesValue {
	value, diagnostics := resource_load_balancer.NewStickyCookiePoliciesValue(
		resource_load_balancer.StickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_expiration_period": utils.FromIntPtrToTfInt64(elt.CookieExpirationPeriod),
			"policy_name":              types.StringPointerValue(elt.PolicyName),
		})
	diags.Append(diagnostics...)
	return value
}

func loadBalancerTags(ctx context.Context, tags types.Set) []api.ResourceTag {
	tfTags := make([]resource_load_balancer.TagsValue, 0, len(tags.Elements()))
	tags.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}
