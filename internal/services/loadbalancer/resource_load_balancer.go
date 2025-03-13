package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewLoadBalancerResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	numSpotClient, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = numSpotClient
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), request, response)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_load_balancer"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = LoadBalancerResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan LoadBalancerModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
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

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state LoadBalancerModel

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

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan         LoadBalancerModel
		numSpotLoadBalancer *numspot.LoadBalancer
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

	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
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

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state LoadBalancerModel

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

func deserializeCreateNumSpotLoadBalancer(ctx context.Context, tf LoadBalancerModel, diags *diag.Diagnostics) numspot.CreateLoadBalancerJSONRequestBody {
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

func deserializeUpdateNumSpotLoadBalancer(ctx context.Context, tf LoadBalancerModel, diags *diag.Diagnostics) numspot.UpdateLoadBalancerJSONRequestBody {
	var (
		loadBalancerPort *int = nil
		policyNames           = make([]string, 0)
		hc                    = &numspot.HealthCheck{
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

	if !tf.SecurityGroups.IsUnknown() {
		sg := utils.TfStringListToStringList(ctx, tf.SecurityGroups, diags)
		securityGroups = &sg
	}

	listeners := utils.TfSetToGenericSet(func(elt ListenersValue) numspot.Listener {
		policyNamesListener := utils.TfStringListToStringList(ctx, elt.PolicyNames, diags)
		if policyNamesListener == nil {
			policyNamesListener = make([]string, 0)
		}
		return numspot.Listener{
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

	return numspot.UpdateLoadBalancerJSONRequestBody{
		HealthCheck:      hc,
		LoadBalancerPort: loadBalancerPort,
		PolicyNames:      &policyNames,
		PublicIp:         publicIp,
		SecuredCookies:   securedCookies,
		SecurityGroups:   securityGroups,
	}
}

func serializeNumSpotLoadBalancer(ctx context.Context, http *numspot.LoadBalancer, diags *diag.Diagnostics) LoadBalancerModel {
	var tagsTf types.List

	applicationStickyCookiePoliciesTypes := utils.GenericListToTfListValue(ctx, applicationStickyCookiePoliciesFromHTTP, *http.ApplicationStickyCookiePolicies, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	listeners := utils.GenericSetToTfSetValue(ctx, listenersFromHTTP, *http.Listeners, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	stickyCookiePolicies := utils.GenericListToTfListValue(ctx, stickyCookiePoliciesFromHTTP, *http.StickyCookiePolicies, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	backendIps := utils.FromStringListPointerToTfStringSet(ctx, http.BackendIps, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	backendVmIds := utils.FromStringListPointerToTfStringSet(ctx, http.BackendVmIds, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return LoadBalancerModel{}
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
		return LoadBalancerModel{}
	}

	securityGroups := utils.FromStringListPointerToTfStringList(ctx, http.SecurityGroups, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	sourceSecurityGroup := SourceSecurityGroupValue{
		SecurityGroupName: types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupName),
	}

	subnets := utils.FromStringListPointerToTfStringList(ctx, http.Subnets, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	return LoadBalancerModel{
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
		SourceSecurityGroup:             sourceSecurityGroup,
		StickyCookiePolicies:            stickyCookiePolicies,
		Subnets:                         subnets,
		Type:                            types.StringPointerValue(http.Type),
		Tags:                            tagsTf,
	}
}

func deserializeTagsToDelete(ctx context.Context, t types.List) []numspot.ResourceLoadBalancerTag {
	lbTags := make([]numspot.ResourceLoadBalancerTag, len(t.Elements()))
	stateTags := tags.TfTagsToApiTags(ctx, t)
	for idx, tag := range stateTags {
		key := &tag.Key
		lbTags[idx] = numspot.ResourceLoadBalancerTag{
			Key: key,
		}
	}
	return lbTags
}

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
