package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

// loadBalancerResource is the Terraform resource implementation for numspot_load_balancer.
type loadBalancerResource struct {
	services.BaseResource
}

// NewLoadBalancerResource returns a new loadBalancerResource instance.
func NewLoadBalancerResource() resource.Resource {
	return &loadBalancerResource{}
}

// Metadata sets the resource type name.
func (r *loadBalancerResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_load_balancer"
}

// Schema defines the schema for the resource.
func (r *loadBalancerResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_load_balancer.LoadBalancerResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *loadBalancerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := loadbalancerTags(ctx, plan.Tags)
	numSpotLoadBalancer, err := core.CreateLoadBalancer(
		ctx,
		r.Provider,
		deserializeCreateLoadBalancerRequest(ctx, &plan, &response.Diagnostics),
		deserializeUpdateLoadBalancer(ctx, &plan, &response.Diagnostics),
		backendVmIDs(ctx, &plan, &response.Diagnostics),
		backendIps(ctx, &plan, &response.Diagnostics),
		tagsValue,
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create load_balancer", err.Error())
		return
	}

	state := serializeLoadBalancer(ctx, numSpotLoadBalancer, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *loadBalancerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotLoadBalancer, err := core.ReadLoadBalancer(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read load_balancer", err.Error())
		return
	}

	newState := serializeLoadBalancer(ctx, numSpotLoadBalancer, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *loadBalancerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                 error
		state, plan         resource_load_balancer.LoadBalancerModel
		numSpotLoadBalancer *api.LoadBalancer
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	loadbalancerID := state.Id.ValueString()
	planTags := loadbalancerTags(ctx, plan.Tags)
	stateTags := loadbalancerTags(ctx, state.Tags)

	if !plan.HealthCheck.Equal(state.HealthCheck) || !plan.PublicIp.Equal(state.PublicIp) {
		numSpotLoadBalancer, err = core.UpdateLoadBalancerAttributes(
			ctx,
			r.Provider,
			state.Id.ValueString(),
			deserializeUpdateLoadBalancer(ctx, &plan, &response.Diagnostics),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update load_balancer", err.Error())
			return
		}
	}

	if !plan.SecurityGroups.Equal(state.SecurityGroups) {
		numSpotLoadBalancer, err = core.UpdateLoadBalancerSecurityGroup(
			ctx,
			r.Provider,
			state.Id.ValueString(),
			deserializeUpdateLoadBalancer(ctx, &plan, &response.Diagnostics),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update load_balancer", err.Error())
			return
		}
	}

	if !plan.BackendVmIds.Equal(state.BackendVmIds) || !plan.BackendIps.Equal(state.BackendIps) {
		numSpotLoadBalancer, err = core.UpdateLoadBalancerBackend(
			ctx,
			r.Provider,
			state.Id.ValueString(),
			backendVmIDs(ctx, &state, &response.Diagnostics),
			backendVmIDs(ctx, &plan, &response.Diagnostics),
			backendIps(ctx, &state, &response.Diagnostics),
			backendIps(ctx, &plan, &response.Diagnostics),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update load_balancer", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numSpotLoadBalancer, err = core.UpdateLoadBalancerTags(ctx, r.Provider, loadbalancerID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update load_balancer tags", err.Error())
			return
		}
	}

	newState := serializeLoadBalancer(ctx, numSpotLoadBalancer, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *loadBalancerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteLoadBalancer(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete load_balancer", err.Error())
		return
	}
}

func serializeLoadBalancer(ctx context.Context, http *api.LoadBalancer, diags *diag.Diagnostics) resource_load_balancer.LoadBalancerModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil LoadBalancer response", "the API returned an empty LoadBalancer")
		return resource_load_balancer.LoadBalancerModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_load_balancer.LoadBalancerModel{
		ApplicationStickyCookiePolicies: utils.GenericListToTfListValue(
			ctx,
			serializeApplicationStickyCookiePolicie,
			*http.ApplicationStickyCookiePolicies,
			diags,
		),
		AvailabilityZoneNames: azNamesToTfList(ctx, http.AvailabilityZoneNames, diags),
		BackendIps:            utils.FromStringListPointerToTfStringSet(ctx, http.BackendIps, diags),
		BackendVmIds:          utils.FromStringListPointerToTfStringSet(ctx, http.BackendVmIds, diags),
		DnsName:               types.StringPointerValue(http.DnsName),
		HealthCheck:           serializeHealthCheck(ctx, http.HealthCheck, diags),
		Id:                    types.StringPointerValue(http.Name),
		Listeners:             utils.GenericSetToTfSetValue(ctx, serializeListener, *http.Listeners, diags),
		Name:                  types.StringPointerValue(http.Name),
		PublicIp:              types.StringPointerValue(http.PublicIp),
		SecuredCookies:        types.BoolPointerValue(http.SecuredCookies),
		SecurityGroups:        utils.FromStringListPointerToTfStringList(ctx, http.SecurityGroups, diags),
		StickyCookiePolicies:  utils.GenericListToTfListValue(ctx, serializeStickyCookiePolicie, *http.StickyCookiePolicies, diags),
		Subnets:               utils.FromStringListPointerToTfStringList(ctx, http.Subnets, diags),
		Type:                  types.StringPointerValue(http.Type),
		VpcId:                 types.StringPointerValue(http.VpcId),
		Tags:                  tagsTf,
	}
}

func deserializeCreateLoadBalancerRequest(
	ctx context.Context,
	tf *resource_load_balancer.LoadBalancerModel,
	diags *diag.Diagnostics,
) *api.CreateLoadBalancerJSONRequestBody {
	return &api.CreateLoadBalancerJSONRequestBody{
		Listeners:      createListenersBody(ctx, tf.Listeners, diags),
		Name:           tf.Name.ValueString(),
		PublicIp:       tf.PublicIp.ValueStringPointer(),
		SecurityGroups: utils.TfStringListToStringPtrList(ctx, tf.SecurityGroups, diags),
		Subnets:        utils.TfStringListToStringList(ctx, tf.Subnets, diags),
		Type:           tf.Type.ValueStringPointer(),
	}
}

func loadbalancerTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_load_balancer.TagsValue, 0, len(tagsSet.Elements()))
	tagsSet.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}

func serializeApplicationStickyCookiePolicie(
	ctx context.Context,
	item *api.ApplicationStickyCookiePolicy,
	diags *diag.Diagnostics,
) resource_load_balancer.ApplicationStickyCookiePoliciesValue {
	value, d := resource_load_balancer.NewApplicationStickyCookiePoliciesValue(
		resource_load_balancer.ApplicationStickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_name": types.StringPointerValue(item.CookieName),
			"policy_name": types.StringPointerValue(item.PolicyName),
		},
	)
	diags.Append(d...)
	return value
}

func serializeHealthCheck(ctx context.Context, item *api.HealthCheck, diags *diag.Diagnostics) resource_load_balancer.HealthCheckValue {
	if item == nil {
		return resource_load_balancer.NewHealthCheckValueNull()
	}
	value, d := resource_load_balancer.NewHealthCheckValue(resource_load_balancer.HealthCheckValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"check_interval":      utils.FromIntToTfInt64(item.CheckInterval),
		"healthy_threshold":   utils.FromIntToTfInt64(item.HealthyThreshold),
		"path":                types.StringPointerValue(item.Path),
		"port":                utils.FromIntToTfInt64(item.Port),
		"protocol":            types.StringValue(item.Protocol),
		"timeout":             utils.FromIntToTfInt64(item.Timeout),
		"unhealthy_threshold": utils.FromIntToTfInt64(item.UnhealthyThreshold),
	})
	diags.Append(d...)
	return value
}

func serializeListener(ctx context.Context, item *api.Listener, diags *diag.Diagnostics) resource_load_balancer.ListenersValue {
	value, d := resource_load_balancer.NewListenersValue(resource_load_balancer.ListenersValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"backend_port":           utils.FromIntPtrToTfInt64(item.BackendPort),
		"backend_protocol":       types.StringPointerValue(item.BackendProtocol),
		"load_balancer_port":     utils.FromIntPtrToTfInt64(item.LoadBalancerPort),
		"load_balancer_protocol": types.StringPointerValue(item.LoadBalancerProtocol),
		"policy_names":           utils.FromStringListPointerToTfStringList(ctx, item.PolicyNames, diags),
		"server_certificate_id":  types.StringPointerValue(item.ServerCertificateId),
	})
	diags.Append(d...)
	return value
}

func serializeStickyCookiePolicie(
	ctx context.Context,
	item *api.LoadBalancerStickyCookiePolicy,
	diags *diag.Diagnostics,
) resource_load_balancer.StickyCookiePoliciesValue {
	value, d := resource_load_balancer.NewStickyCookiePoliciesValue(
		resource_load_balancer.StickyCookiePoliciesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"cookie_expiration_period": utils.FromIntPtrToTfInt64(item.CookieExpirationPeriod),
			"policy_name":              types.StringPointerValue(item.PolicyName),
		},
	)
	diags.Append(d...)
	return value
}
