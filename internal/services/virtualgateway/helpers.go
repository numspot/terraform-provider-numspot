package virtualgateway

import (
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
)

// getVpcId returns the ID of the VPC the virtual gateway is currently attached
// to, derived from the first "attached" link. The read model exposes the
// attachment only through VpcToVirtualGatewayLinks, not as a top-level field.
func getVpcId(vpcToVirtualGatewayLinks []api.VpcToVirtualGatewayLink) string {
	for _, vpc := range vpcToVirtualGatewayLinks {
		if vpc.VpcId != "" && vpc.State == "attached" {
			return vpc.VpcId
		}
	}
	return ""
}
