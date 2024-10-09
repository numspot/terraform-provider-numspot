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

func publicIpFromApi(ctx context.Context, elt numspot.PublicIpLight, diags *diag.Diagnostics) PublicIpsValue {
	value, diagnostics := NewPublicIpsValue(
		PublicIpsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"public_ip":    types.StringPointerValue(elt.PublicIp),
			"public_ip_id": types.StringPointerValue(elt.PublicIpId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func NatGatewayFromHttpToTf(ctx context.Context, http *numspot.NatGateway, diags *diag.Diagnostics) *NatGatewayModel {
	var tagsTf types.List

	var publicIp []numspot.PublicIpLight
	if http.PublicIps != nil {
		publicIp = *http.PublicIps
	}
	// Public Ips
	publicIpsTf := utils.GenericListToTfListValue(
		ctx,
		PublicIpsValue{},
		publicIpFromApi,
		publicIp,
		diags,
	)
	if diags.HasError() {
		return nil
	}

	// PublicIpId must be the id of the first public io
	var publicIpId *string
	if len(publicIp) > 0 {
		publicIpId = publicIp[0].PublicIpId
	} else {
		publicIpId = nil
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &NatGatewayModel{
		Id:         types.StringPointerValue(http.Id),
		PublicIps:  publicIpsTf,
		State:      types.StringPointerValue(http.State),
		SubnetId:   types.StringPointerValue(http.SubnetId),
		VpcId:      types.StringPointerValue(http.VpcId),
		Tags:       tagsTf,
		PublicIpId: types.StringPointerValue(publicIpId),
	}
}

func NatGatewayFromHttpToTfDatasource(ctx context.Context, http *numspot.NatGateway, diags *diag.Diagnostics) *NatGatewayModelDatasource {
	var tagsTf types.List

	var publicIp []numspot.PublicIpLight
	if http.PublicIps != nil {
		publicIp = *http.PublicIps
	}
	// Public Ips
	publicIpsTf := utils.GenericListToTfListValue(
		ctx,
		PublicIpsValue{},
		publicIpFromApi,
		publicIp,
		diags,
	)
	if diags.HasError() {
		return nil
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &NatGatewayModelDatasource{
		Id:        types.StringPointerValue(http.Id),
		PublicIps: publicIpsTf,
		State:     types.StringPointerValue(http.State),
		SubnetId:  types.StringPointerValue(http.SubnetId),
		VpcId:     types.StringPointerValue(http.VpcId),
		Tags:      tagsTf,
	}
}

func NatGatewayFromTfToCreateRequest(tf NatGatewayModel) numspot.CreateNatGatewayJSONRequestBody {
	return numspot.CreateNatGatewayJSONRequestBody{
		PublicIpId: tf.PublicIpId.ValueString(),
		SubnetId:   tf.SubnetId.ValueString(),
	}
}

func NatGatewaysFromTfToAPIReadParams(ctx context.Context, tf NatGatewaysDataSourceModel, diags *diag.Diagnostics) numspot.ReadNatGatewayParams {
	return numspot.ReadNatGatewayParams{
		SubnetIds: utils.TfStringListToStringPtrList(ctx, tf.SubnetIds, diags),
		VpcIds:    utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		States:    utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		TagKeys:   utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues: utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:      utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:       utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
	}
}
