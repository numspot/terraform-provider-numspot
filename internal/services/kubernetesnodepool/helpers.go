package kubernetesnodepool

import (
	"time"

	"github.com/numspot/terraform-provider-numspot/internal/services/kubernetesnodepool/resource_kubernetes_nodepool"
)

// createTimeout extracts the optional create timeout from the synthetic
// timeouts {} block and returns it as a *time.Duration for
// core.CreateKubernetesNodePool. It returns nil when the value is unset/unknown
// or malformed, in which case the core falls back to its default retry timeout
// (1h — node pools provision their underlying machines for up to ~1h).
func createTimeout(plan *resource_kubernetes_nodepool.KubernetesNodepoolModel) *time.Duration {
	if plan == nil || plan.Timeouts.IsNull() || plan.Timeouts.Create.IsNull() || plan.Timeouts.Create.IsUnknown() {
		return nil
	}
	d, err := time.ParseDuration(plan.Timeouts.Create.ValueString())
	if err != nil {
		return nil
	}
	return &d
}
