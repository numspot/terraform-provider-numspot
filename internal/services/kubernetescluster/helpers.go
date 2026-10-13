package kubernetescluster

import (
	"time"

	"terraform-provider-numspot/internal/services/kubernetescluster/resource_kubernetes_cluster"
)

// createTimeout extracts the optional create timeout from the synthetic
// timeouts {} block and returns it as a *time.Duration for
// core.CreateKubernetesCluster. It returns nil when the value is unset/unknown
// or malformed, in which case the core falls back to its default retry timeout.
func createTimeout(plan *resource_kubernetes_cluster.KubernetesClusterModel) *time.Duration {
	if plan == nil || plan.Timeouts.IsNull() || plan.Timeouts.Create.IsNull() || plan.Timeouts.Create.IsUnknown() {
		return nil
	}
	d, err := time.ParseDuration(plan.Timeouts.Create.ValueString())
	if err != nil {
		return nil
	}
	return &d
}
