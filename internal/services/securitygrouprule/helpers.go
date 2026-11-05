package securitygrouprule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
)

// securityGroupRuleID derives the synthetic resource id from a normalized
// (single IP range) SDK rule. It defers to core.SecurityGroupRuleID so the
// serialize-time id and the core's Read/Delete lookups stay in sync.
func securityGroupRuleID(http *api.SecurityGroupRule) types.String {
	if http == nil {
		return types.StringNull()
	}
	return types.StringValue(core.SecurityGroupRuleID(*http))
}

// firstIpRange maps the SDK's IpRanges slice onto the single ip_range attribute.
// Rules returned by core are always normalized to one IP range.
func firstIpRange(ipRanges *[]string) types.String {
	if ipRanges == nil || len(*ipRanges) == 0 {
		return types.StringNull()
	}
	return types.StringValue((*ipRanges)[0])
}
