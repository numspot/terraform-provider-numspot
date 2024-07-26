package virtualgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

/*
The Linking of Virtual Gateway with VPC is weird on Outscale side :
A Virtual Gateway can be linked to a single VPC, but vpcToVirtualGatewayLinks is an array of VPCs
This array contains all the VPC that has been linked to the Virtual Gateway (a given VPC can appear multiple time)
VPCs that have been unlinked have the state "detached", and the linked VPC (if any), have state "attached"

This function retrieve the first (single) vpcId that has state attached, if any
*/
func getVpcId(vpcToVirtualGatewayLinks *[]numspot.VpcToVirtualGatewayLink) *string {
	var vpcId *string

	if vpcToVirtualGatewayLinks != nil {
		vpcToVirtualGatewayLinksValue := *vpcToVirtualGatewayLinks

		for _, vpc := range vpcToVirtualGatewayLinksValue {
			if vpc.State != nil && vpc.VpcId != nil && *vpc.State == "attached" {
				vpcId = vpc.VpcId
				break
			}
		}
	}

	return vpcId
}

func VpcToVirtualGatewayLinksFromHttpToTf(ctx context.Context, http numspot.VpcToVirtualGatewayLink) (VpcToVirtualGatewayLinksValue, diag.Diagnostics) {
	return NewVpcToVirtualGatewayLinksValue(VpcToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"state":  types.StringPointerValue(http.State),
			"vpc_id": types.StringPointerValue(http.VpcId),
		},
	)
}

func VirtualGatewayFromHttpToTf(ctx context.Context, http *numspot.VirtualGateway) (*VirtualGatewayModel, diag.Diagnostics) {
	var (
		diags                              diag.Diagnostics
		tagsTf, vpcToVirtualGatewayLinksTf types.List
	)

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.VpcToVirtualGatewayLinks != nil {
		vpcToVirtualGatewayLinksTf, diags = utils.GenericListToTfListValue(ctx, VpcToVirtualGatewayLinksValue{}, VpcToVirtualGatewayLinksFromHttpToTf, *http.VpcToVirtualGatewayLinks)
		if diags.HasError() {
			return nil, diags
		}
	}

	vpcId := getVpcId(http.VpcToVirtualGatewayLinks)

	return &VirtualGatewayModel{
		ConnectionType:           types.StringPointerValue(http.ConnectionType),
		Id:                       types.StringPointerValue(http.Id),
		VpcToVirtualGatewayLinks: vpcToVirtualGatewayLinksTf,
		VpcId:                    types.StringPointerValue(vpcId),
		State:                    types.StringPointerValue(http.State),
		Tags:                     tagsTf,
	}, diags
}

func VirtualGatewayFromTfToCreateRequest(tf VirtualGatewayModel) numspot.CreateVirtualGatewayJSONRequestBody {
	return numspot.CreateVirtualGatewayJSONRequestBody{
		ConnectionType: tf.ConnectionType.ValueString(),
	}
}

func VirtualGatewaysFromTfToAPIReadParams(ctx context.Context, tf VirtualGatewaysDataSourceModel) numspot.ReadVirtualGatewaysParams {
	return numspot.ReadVirtualGatewaysParams{
		States:          utils.TfStringListToStringPtrList(ctx, tf.States),
		TagKeys:         utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:       utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:            utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:             utils.TfStringListToStringPtrList(ctx, tf.IDs),
		ConnectionTypes: utils.TfStringListToStringPtrList(ctx, tf.ConnectionTypes),
		LinkStates:      utils.TfStringListToStringPtrList(ctx, tf.LinkStates),
		LinkVpcIds:      utils.TfStringListToStringPtrList(ctx, tf.LinkVpcIds),
	}
}
