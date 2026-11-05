package natgateway

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
)

// firstPublicIpId returns the public IP ID of the first associated public IP,
// or null when none is present. NAT gateways echo the allocated public IP back
// inside PublicIps rather than as a top-level field, so the resource derives
// the singular public_ip_id from the first element.
func firstPublicIpId(ips *[]api.PublicIpLight) types.String {
	if ips == nil || len(*ips) == 0 {
		return types.StringNull()
	}
	return types.StringPointerValue((*ips)[0].PublicIpId)
}
