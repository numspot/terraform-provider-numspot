package natgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func publicIpFromApi(ctx context.Context, elt numspot.PublicIpLight) (PublicIpsValue, diag.Diagnostics) {
	return NewPublicIpsValue(
		PublicIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_ip":    types.StringPointerValue(elt.PublicIp),
			"public_ip_id": types.StringPointerValue(elt.PublicIpId),
		},
	)
}

func NatGatewayFromHttpToTf(ctx context.Context, http *numspot.NatGateway) (*NatGatewayModel, diag.Diagnostics) {
	var (
		tagsTf types.List
		diags  diag.Diagnostics
	)

	var publicIp []numspot.PublicIpLight
	if http.PublicIps != nil {
		publicIp = *http.PublicIps
	}
	// Public Ips
	publicIpsTf, diags := utils.GenericListToTfListValue(
		ctx,
		PublicIpsValue{},
		publicIpFromApi,
		publicIp,
	)
	if diags.HasError() {
		return nil, diags
	}

	// PublicIpId must be the id of the first public io
	var publicIpId *string
	if len(publicIp) > 0 {
		publicIpId = publicIp[0].PublicIpId
	} else {
		publicIpId = nil
	}

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &NatGatewayModel{
		Id:         types.StringPointerValue(http.Id),
		PublicIpId: types.StringPointerValue(publicIpId),
		PublicIps:  publicIpsTf,
		State:      types.StringPointerValue(http.State),
		SubnetId:   types.StringPointerValue(http.SubnetId),
		VpcId:      types.StringPointerValue(http.VpcId),
		Tags:       tagsTf,
	}, diags
}

func NatGatewayFromTfToCreateRequest(tf NatGatewayModel) numspot.CreateNatGatewayJSONRequestBody {
	return numspot.CreateNatGatewayJSONRequestBody{
		PublicIpId: tf.PublicIpId.ValueString(),
		SubnetId:   tf.SubnetId.ValueString(),
	}
}

func NatGatewaysFromTfToAPIReadParams(ctx context.Context, tf NatGatewaysDataSourceModel) numspot.ReadNatGatewayParams {
	return numspot.ReadNatGatewayParams{
		SubnetIds: utils.TfStringListToStringPtrList(ctx, tf.SubnetIds),
		VpcIds:    utils.TfStringListToStringPtrList(ctx, tf.VpcIds),
		States:    utils.TfStringListToStringPtrList(ctx, tf.States),
		TagKeys:   utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues: utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:      utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:       utils.TfStringListToStringPtrList(ctx, tf.IDs),
	}
}
