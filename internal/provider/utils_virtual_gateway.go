package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_virtual_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_virtual_gateway"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

/*
The Linking of Virtual Gateway with VPC is weird on Outscale side :
A Virtual Gateway can be linked to a single VPC, but vpcToVirtualGatewayLinks is an array of VPCs
This array contains all the VPC that has been linked to the Virtual Gateway (a given VPC can appear multiple time)
VPCs that have been unlinked have the state "detached", and the linked VPC (if any), have state "attached"

This function retrieve the first (single) vpcId that has state attached, if any
*/
func getVpcId(vpcToVirtualGatewayLinks *[]iaas.VpcToVirtualGatewayLink) *string {
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

func VirtualGatewayFromHttpToTf(ctx context.Context, http *iaas.VirtualGateway) (*resource_virtual_gateway.VirtualGatewayModel, diag.Diagnostics) {
	var (
		diags  diag.Diagnostics
		tagsTf types.List
	)

	vpcId := getVpcId(http.VpcToVirtualGatewayLinks)

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &resource_virtual_gateway.VirtualGatewayModel{
		ConnectionType: types.StringPointerValue(http.ConnectionType),
		Id:             types.StringPointerValue(http.Id),
		VpcId:          types.StringPointerValue(vpcId),
		State:          types.StringPointerValue(http.State),
		Tags:           tagsTf,
	}, diags
}

func VirtualGatewayFromTfToCreateRequest(tf resource_virtual_gateway.VirtualGatewayModel) iaas.CreateVirtualGatewayJSONRequestBody {
	return iaas.CreateVirtualGatewayJSONRequestBody{
		ConnectionType: tf.ConnectionType.ValueString(),
	}
}

func VirtualGatewaysFromTfToAPIReadParams(ctx context.Context, tf VirtualGatewaysDataSourceModel) iaas.ReadVirtualGatewaysParams {
	return iaas.ReadVirtualGatewaysParams{
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

func VirtualGatewaysFromHttpToTfDatasource(ctx context.Context, http *iaas.VirtualGateway) (*datasource_virtual_gateway.VirtualGatewayModel, diag.Diagnostics) {
	var (
		diags    diag.Diagnostics
		tagsList types.List
	)

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	vpcId := getVpcId(http.VpcToVirtualGatewayLinks)

	return &datasource_virtual_gateway.VirtualGatewayModel{
		Id:             types.StringPointerValue(http.Id),
		State:          types.StringPointerValue(http.State),
		ConnectionType: types.StringPointerValue(http.ConnectionType),
		VpcId:          types.StringPointerValue(vpcId),
		Tags:           tagsList,
	}, nil
}
