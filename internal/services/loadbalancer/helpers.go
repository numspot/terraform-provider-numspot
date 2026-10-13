package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/loadbalancer/resource_load_balancer"
	"terraform-provider-numspot/internal/utils"
)

// backendIps converts the backend_ips set into the plain []string the core
// link/unlink calls expect.
func backendIps(ctx context.Context, tf *resource_load_balancer.LoadBalancerModel, diags *diag.Diagnostics) []string {
	return utils.FromTfStringSetToStringList(ctx, tf.BackendIps, diags)
}

// backendVmIDs converts the backend_vm_ids set into the plain []string the core
// link/unlink calls expect.
func backendVmIDs(ctx context.Context, tf *resource_load_balancer.LoadBalancerModel, diags *diag.Diagnostics) []string {
	return utils.FromTfStringSetToStringList(ctx, tf.BackendVmIds, diags)
}

// azNamesToTfList serializes the read-only availability_zone_names list (a slice
// of the AvailabilityZoneName string alias) into a Terraform string list. The
// generator has no generic *[]api.AvailabilityZoneName -> types.List pair.
func azNamesToTfList(ctx context.Context, names *[]api.AvailabilityZoneName, diags *diag.Diagnostics) types.List {
	if names == nil {
		return types.ListNull(types.StringType)
	}
	strs := make([]string, len(*names))
	for i, n := range *names {
		strs[i] = string(n)
	}
	list, d := types.ListValueFrom(ctx, types.StringType, strs)
	diags.Append(d...)
	return list
}

// createListenersBody builds the create-body listeners, which use the divergent
// api.ListenerForCreation type (value ports, required protocol) rather than the
// read api.Listener.
func createListenersBody(ctx context.Context, listeners types.Set, diags *diag.Diagnostics) []api.ListenerForCreation {
	return utils.TfSetToGenericList(func(a resource_load_balancer.ListenersValue) api.ListenerForCreation {
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
	}, ctx, listeners, diags)
}

// deserializeUpdateLoadBalancer builds the update body shared by the
// attribute/security-group update paths and threaded into Create as an
// extra-arg. It carries the default health check plus the single-listener
// port/policy extraction the API requires.
func deserializeUpdateLoadBalancer(
	ctx context.Context,
	tf *resource_load_balancer.LoadBalancerModel,
	diags *diag.Diagnostics,
) *api.UpdateLoadBalancerJSONRequestBody {
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

	return &api.UpdateLoadBalancerJSONRequestBody{
		HealthCheck:      hc,
		LoadBalancerPort: loadBalancerPort,
		PolicyNames:      &policyNames,
		PublicIp:         publicIp,
		SecuredCookies:   securedCookies,
		SecurityGroups:   securityGroups,
	}
}
